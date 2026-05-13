// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	logs "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/logs/v1"
)

const (
	DefaultLogSource        = "otel_logs"
	DefaultLogService       = "unSetServiceName"
	DefaultSchemaURLKey     = "schema_url"
	DefaultMessageField     = "message"
	DefaultTraceIDField     = "trace_id"
	DefaultSpanIDField      = "span_id"
	DefaultStatusTag        = "status"
	DefaultServiceTag       = "service"
	DefaultSourceTag        = "source"
	DefaultHostTag          = "host"
	DefaultFingerprintTag   = "dk_fingerprint"
	DefaultMaxMessageLength = 32 * 1024
)

type LogsParserOptions struct {
	CollectorSourceIP    string
	CollectorSourceIPTag string
	DKFingerprint        string
	DKFingerprintTag     string

	ResourceStringOptions StringMapOptions
	ScopeStringOptions    StringMapOptions
	RecordStringOptions   StringMapOptions

	SchemaURLKey  string
	ScopeNameKey  string
	MessageField  string
	TraceIDField  string
	SpanIDField   string
	StatusTag     string
	ServiceTag    string
	SourceTag     string
	HostTag       string
	MaxMessageLen int

	GlobalFields map[string]any
	GlobalTags   map[string]string

	IDConverter      func([]byte) string
	ServiceAndSource func([]*common.KeyValue) (service, source string)
	HostName         func([]*common.KeyValue) string
	SeverityMapper   func(logs.SeverityNumber, string) string
	PointOptions     func(ts time.Time) []point.Option
}

func defaultLogsParserOptions(opts LogsParserOptions) LogsParserOptions {
	if opts.CollectorSourceIPTag == "" {
		opts.CollectorSourceIPTag = DefaultCollectorSourceTag
	}
	if opts.DKFingerprintTag == "" {
		opts.DKFingerprintTag = DefaultFingerprintTag
	}
	if opts.SchemaURLKey == "" {
		opts.SchemaURLKey = DefaultSchemaURLKey
	}
	if opts.ScopeNameKey == "" {
		opts.ScopeNameKey = DefaultScopeNameKey
	}
	if opts.MessageField == "" {
		opts.MessageField = DefaultMessageField
	}
	if opts.TraceIDField == "" {
		opts.TraceIDField = DefaultTraceIDField
	}
	if opts.SpanIDField == "" {
		opts.SpanIDField = DefaultSpanIDField
	}
	if opts.StatusTag == "" {
		opts.StatusTag = DefaultStatusTag
	}
	if opts.ServiceTag == "" {
		opts.ServiceTag = DefaultServiceTag
	}
	if opts.SourceTag == "" {
		opts.SourceTag = DefaultSourceTag
	}
	if opts.HostTag == "" {
		opts.HostTag = DefaultHostTag
	}
	if opts.MaxMessageLen == 0 {
		opts.MaxMessageLen = DefaultMaxMessageLength
	}
	if opts.IDConverter == nil {
		opts.IDConverter = HexID
	}
	if opts.ServiceAndSource == nil {
		opts.ServiceAndSource = DefaultLogServiceAndSource
	}
	if opts.HostName == nil {
		opts.HostName = DefaultHostName
	}
	if opts.SeverityMapper == nil {
		opts.SeverityMapper = func(num logs.SeverityNumber, fallback string) string {
			return LogStatusFromSeverityNumber(int32(num), fallback)
		}
	}
	if opts.PointOptions == nil {
		opts.PointOptions = func(ts time.Time) []point.Option {
			opts := point.DefaultLoggingOptions()
			return append(opts, point.WithTime(ts))
		}
	}

	return opts
}

func ParseLogRequest(resourceLogs []*logs.ResourceLogs, opts LogsParserOptions) []*point.Point {
	opts = defaultLogsParserOptions(opts)

	pts := make([]*point.Point, 0)
	for _, resourceLog := range resourceLogs {
		resourceFields := AttributesToStringMap(resourceLog.GetResource().GetAttributes(), opts.ResourceStringOptions)
		if opts.SchemaURLKey != "" && resourceLog.GetSchemaUrl() != "" {
			resourceFields[opts.SchemaURLKey] = resourceLog.GetSchemaUrl()
		}

		service, source := opts.ServiceAndSource(resourceLog.GetResource().GetAttributes())
		host := opts.HostName(resourceLog.GetResource().GetAttributes())

		for _, scopeLog := range resourceLog.GetScopeLogs() {
			scopeFields := AttributesToStringMap(scopeLog.GetScope().GetAttributes(), opts.ScopeStringOptions)
			if scope := scopeLog.GetScope(); scope != nil && scope.GetName() != "" && opts.ScopeNameKey != "" {
				scopeFields[opts.ScopeNameKey] = scope.GetName()
			}

			for _, record := range scopeLog.GetLogRecords() {
				recordFields := AttributesToStringMap(record.GetAttributes(), opts.RecordStringOptions)
				messages := splitLogMessage(logBodyToString(record.GetBody()), opts.MaxMessageLen)

				for idx, msg := range messages {
					kvs := MergeStringMapsAsFields(resourceFields, scopeFields, recordFields)
					for key, value := range opts.GlobalFields {
						kvs = kvs.Add(key, value)
					}
					for key, value := range opts.GlobalTags {
						kvs = kvs.AddTag(key, value)
					}

					kvs = kvs.Add(opts.MessageField, msg).
						Add(opts.SpanIDField, opts.IDConverter(record.GetSpanId())).
						Add(opts.TraceIDField, opts.IDConverter(record.GetTraceId())).
						AddTag(opts.StatusTag, opts.SeverityMapper(record.GetSeverityNumber(), record.GetSeverityText())).
						AddTag(opts.ServiceTag, service).
						AddTag(opts.SourceTag, source)

					if opts.CollectorSourceIP != "" {
						kvs = kvs.AddTag(opts.CollectorSourceIPTag, opts.CollectorSourceIP)
					}
					if opts.DKFingerprint != "" {
						kvs = kvs.AddTag(opts.DKFingerprintTag, opts.DKFingerprint)
					}
					if host != "" {
						kvs = kvs.AddTag(opts.HostTag, host)
					}

					ts := time.Unix(0, int64(record.GetTimeUnixNano()))
					if record.GetTimeUnixNano() == 0 {
						ts = time.Unix(0, int64(record.GetObservedTimeUnixNano()))
					}
					ts = ts.Add(time.Millisecond * time.Duration(idx))

					pts = append(pts, point.NewPoint(source, kvs, opts.PointOptions(ts)...))
				}
			}
		}
	}

	return pts
}

func DefaultLogServiceAndSource(attrs []*common.KeyValue) (service string, source string) {
	for _, keyValue := range attrs {
		switch keyValue.GetKey() {
		case AttrServiceName:
			service = keyValue.GetValue().GetStringValue()
		case "log.source":
			source = keyValue.GetValue().GetStringValue()
		}
	}

	if source == "" {
		if service == "" {
			source = DefaultLogSource
		} else {
			source = service
		}
	}
	if service == "" {
		service = DefaultLogService
	}

	return service, source
}

func DefaultHostName(attrs []*common.KeyValue) string {
	for _, keyValue := range attrs {
		if keyValue.GetKey() == "host.name" {
			return keyValue.GetValue().GetStringValue()
		}
	}

	return ""
}

func logBodyToString(body *common.AnyValue) string {
	if body == nil {
		return ""
	}

	switch body.GetValue().(type) {
	case *common.AnyValue_StringValue:
		return body.GetStringValue()
	case *common.AnyValue_BytesValue:
		return string(body.GetBytesValue())
	case *common.AnyValue_DoubleValue:
		return fmt.Sprintf("%f", body.GetDoubleValue())
	case *common.AnyValue_IntValue:
		return strconv.FormatInt(body.GetIntValue(), 10)
	case *common.AnyValue_BoolValue:
		return strconv.FormatBool(body.GetBoolValue())
	case *common.AnyValue_ArrayValue:
		return body.GetArrayValue().String()
	case *common.AnyValue_KvlistValue:
		return body.GetKvlistValue().String()
	default:
		return ""
	}
}

func splitLogMessage(message string, maxLen int) []string {
	switch {
	case message == "":
		return []string{""}
	case maxLen <= 0:
		return []string{message}
	default:
		chunks := ChunkStringByRuneLength(message, maxLen)
		if len(chunks) == 0 {
			return []string{message}
		}
		return chunks
	}
}
