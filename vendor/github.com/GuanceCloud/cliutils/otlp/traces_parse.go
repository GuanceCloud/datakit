// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultTracePointName      = "tracing"
	DefaultUnknownService      = "unknown_service"
	DefaultTraceSource         = "opentelemetry"
	DefaultParentSpanID        = "0"
	DefaultTraceBaseServiceTag = "base_service"
	DefaultTraceStatusTag      = "status"
	DefaultTraceSpanTypeTag    = "span_type"
	DefaultTraceSpanKindTag    = "span_kind"
	DefaultTraceSourceTypeTag  = "source_type"
	DefaultTraceFingerprintTag = "dk_fingerprint"
	DefaultTraceServiceTag     = "service"
	DefaultTraceSourceTag      = "source"
	DefaultTraceTraceIDField   = "trace_id"
	DefaultTraceParentIDField  = "parent_id"
	DefaultTraceSpanIDField    = "span_id"
	DefaultTraceResourceField  = "resource"
	DefaultTraceStartField     = "start"
	DefaultTraceDurationField  = "duration"
	DefaultTraceMessageField   = "message"
	DefaultTraceRuntimeIDField = "runtime_id"
	DefaultTraceDBHostTag      = "db_host"
	TraceSourceTypeWeb         = "web"
	TraceSourceTypeDB          = "db"
	TraceSourceTypeMessage     = "message_queue"
	TraceSourceTypeCustomer    = "custom"
)

type TraceBatch []*point.Point

type TracesParserOptions struct {
	PointName string
	Source    string

	CollectorSourceIP    string
	CollectorSourceIPTag string
	DKFingerprint        string
	DKFingerprintTag     string

	UnknownService string

	ServiceTag     string
	BaseServiceTag string
	SpanStatusTag  string
	SpanTypeTag    string
	SpanKindTag    string
	SourceTag      string
	SourceTypeTag  string
	DBHostTag      string

	TraceIDField   string
	ParentIDField  string
	SpanIDField    string
	ResourceField  string
	StartField     string
	DurationField  string
	MessageField   string
	RuntimeIDField string

	GlobalTags map[string]string

	IDConverter    func([]byte) string
	SelectAttrs    func([]*common.KeyValue) (point.KVs, []*common.KeyValue)
	ScopeKVs       func(*trace.ScopeSpans) point.KVs
	SpanType       func(spanID, parentID string, spanIDs, parentIDs map[string]bool) string
	SpanStatus     func(*trace.Status) string
	SpanKind       func(int32) string
	SourceType     func(point.KVs) string
	BaseService    func(map[string]string) string
	DecorateKVs    func(point.KVs, *trace.ResourceSpans, *trace.ScopeSpans, *trace.Span) point.KVs
	MessageEncoder func(*trace.Span) (string, error)
	CleanSpan      func(*trace.Span) *trace.Span
}

func defaultTracesParserOptions(opts TracesParserOptions) TracesParserOptions {
	if opts.PointName == "" {
		opts.PointName = DefaultTracePointName
	}
	if opts.Source == "" {
		opts.Source = DefaultTraceSource
	}
	if opts.CollectorSourceIPTag == "" {
		opts.CollectorSourceIPTag = DefaultCollectorSourceTag
	}
	if opts.DKFingerprintTag == "" {
		opts.DKFingerprintTag = DefaultTraceFingerprintTag
	}
	if opts.UnknownService == "" {
		opts.UnknownService = DefaultUnknownService
	}
	if opts.ServiceTag == "" {
		opts.ServiceTag = DefaultTraceServiceTag
	}
	if opts.BaseServiceTag == "" {
		opts.BaseServiceTag = DefaultTraceBaseServiceTag
	}
	if opts.SpanStatusTag == "" {
		opts.SpanStatusTag = DefaultTraceStatusTag
	}
	if opts.SpanTypeTag == "" {
		opts.SpanTypeTag = DefaultTraceSpanTypeTag
	}
	if opts.SpanKindTag == "" {
		opts.SpanKindTag = DefaultTraceSpanKindTag
	}
	if opts.SourceTag == "" {
		opts.SourceTag = DefaultTraceSourceTag
	}
	if opts.SourceTypeTag == "" {
		opts.SourceTypeTag = DefaultTraceSourceTypeTag
	}
	if opts.DBHostTag == "" {
		opts.DBHostTag = DefaultTraceDBHostTag
	}
	if opts.TraceIDField == "" {
		opts.TraceIDField = DefaultTraceTraceIDField
	}
	if opts.ParentIDField == "" {
		opts.ParentIDField = DefaultTraceParentIDField
	}
	if opts.SpanIDField == "" {
		opts.SpanIDField = DefaultTraceSpanIDField
	}
	if opts.ResourceField == "" {
		opts.ResourceField = DefaultTraceResourceField
	}
	if opts.StartField == "" {
		opts.StartField = DefaultTraceStartField
	}
	if opts.DurationField == "" {
		opts.DurationField = DefaultTraceDurationField
	}
	if opts.MessageField == "" {
		opts.MessageField = DefaultTraceMessageField
	}
	if opts.RuntimeIDField == "" {
		opts.RuntimeIDField = DefaultTraceRuntimeIDField
	}
	if opts.IDConverter == nil {
		opts.IDConverter = HexID
	}
	if opts.SelectAttrs == nil {
		opts.SelectAttrs = SelectPublicTraceAttributes
	}
	if opts.SpanStatus == nil {
		opts.SpanStatus = func(status *trace.Status) string {
			if status == nil {
				return SpanStatusInfo
			}
			return TraceStatusFromCode(int32(status.GetCode()))
		}
	}
	if opts.SpanKind == nil {
		opts.SpanKind = SpanKindName
	}
	if opts.SourceType == nil {
		opts.SourceType = TraceSourceTypeFromTags
	}
	if opts.BaseService == nil {
		opts.BaseService = SystemNameForService
	}
	if opts.MessageEncoder == nil {
		opts.MessageEncoder = marshalTraceSpan
	}

	return opts
}

func ParseResourceSpans(resourceSpans []*trace.ResourceSpans, opts TracesParserOptions) []TraceBatch {
	opts = defaultTracesParserOptions(opts)
	spanIDs, parentIDs := collectSpanIDs(resourceSpans, opts.IDConverter)

	batches := make([]TraceBatch, 0, len(resourceSpans))
	for _, resourceSpan := range resourceSpans {
		serviceName, runtimeID := serviceAndRuntimeFromResource(resourceSpan.GetResource().GetAttributes(), opts.UnknownService)
		resourceAttrMap := AttributesToStringMap(resourceSpan.GetResource().GetAttributes(), StringMapOptions{})

		var batch TraceBatch
		for _, scopeSpans := range resourceSpan.GetScopeSpans() {
			for _, span := range scopeSpans.GetSpans() {
				spanID := opts.IDConverter(span.GetSpanId())
				parentID := opts.IDConverter(span.GetParentSpanId())
				traceID := opts.IDConverter(span.GetTraceId())
				if parentID == "" {
					parentID = DefaultParentSpanID
				}

				spanAttrs := append([]*common.KeyValue{}, span.GetAttributes()...)
				spanAttrs = append(spanAttrs, &common.KeyValue{
					Key: opts.TraceIDField,
					Value: &common.AnyValue{
						Value: &common.AnyValue_StringValue{StringValue: traceID},
					},
				})

				mergedAttrs := make([]*common.KeyValue, 0)
				kvs := point.KVs{}
				kvs = kvs.Add(opts.TraceIDField, traceID).
					Add(opts.ParentIDField, parentID).
					Add(opts.SpanIDField, spanID).
					Add(opts.ResourceField, span.GetName()).
					Add(opts.StartField, int64(span.GetStartTimeUnixNano())/int64(time.Microsecond)).
					Add(opts.DurationField, int64(span.GetEndTimeUnixNano()-span.GetStartTimeUnixNano())/int64(time.Microsecond)).
					AddTag(opts.SpanStatusTag, opts.SpanStatus(span.GetStatus())).
					AddTag(opts.SourceTag, opts.Source).
					AddTag(opts.ServiceTag, serviceName)

				if opts.CollectorSourceIP != "" {
					kvs = kvs.AddTag(opts.CollectorSourceIPTag, opts.CollectorSourceIP)
				}
				if opts.DKFingerprint != "" {
					kvs = kvs.AddTag(opts.DKFingerprintTag, opts.DKFingerprint)
				}
				if opts.SpanType != nil {
					kvs = kvs.AddTag(opts.SpanTypeTag, opts.SpanType(spanID, parentID, spanIDs, parentIDs))
				}

				baseService := opts.BaseService(AttributesToStringMap(spanAttrs, StringMapOptions{}))
				if baseService != "" {
					kvs = kvs.SetTag(opts.ServiceTag, baseService).AddTag(opts.BaseServiceTag, serviceName)
				}
				if runtimeID == "" {
					runtimeID = runtimeIDFromAttributes(spanAttrs)
				}
				if runtimeID != "" {
					kvs = kvs.Add(opts.RuntimeIDField, runtimeID)
				}

				if kind := opts.SpanKind(int32(span.GetKind())); kind != "" {
					kvs = kvs.AddTag(opts.SpanKindTag, kind)
				}
				if opts.ScopeKVs != nil {
					kvs = append(kvs, opts.ScopeKVs(scopeSpans)...)
				}

				for _, attrs := range [][]*common.KeyValue{
					resourceSpan.GetResource().GetAttributes(),
					scopeSpans.GetScope().GetAttributes(),
					spanAttrs,
				} {
					newKVs, newAttrs := opts.SelectAttrs(attrs)
					kvs = append(kvs, newKVs...)
					mergedAttrs = append(mergedAttrs, newAttrs...)
				}

				for _, event := range span.GetEvents() {
					if event == nil {
						continue
					}
					newKVs, newAttrs := opts.SelectAttrs(event.GetAttributes())
					kvs = append(kvs, newKVs...)
					mergedAttrs = append(mergedAttrs, newAttrs...)
					if event.GetName() == EventException {
						kvs = append(kvs, exceptionEventKVs(event.GetAttributes())...)
					}
				}

				if dbHost := DBHostFromAttributes(resourceAttrMap, AttributesToStringMap(span.GetAttributes(), StringMapOptions{})); dbHost != "" {
					kvs = kvs.AddTag(opts.DBHostTag, dbHost)
				}

				kvs = kvs.AddTag(opts.SourceTypeTag, opts.SourceType(kvs))
				for key, value := range opts.GlobalTags {
					kvs = kvs.SetTag(key, value)
				}

				span.Attributes = mergedAttrs
				msgSpan := span
				if opts.CleanSpan != nil {
					msgSpan = opts.CleanSpan(protoCloneTraceSpan(span))
				}
				if msg, err := opts.MessageEncoder(msgSpan); err == nil && msg != "" {
					kvs = kvs.Add(opts.MessageField, msg)
				}
				if opts.DecorateKVs != nil {
					kvs = opts.DecorateKVs(kvs, resourceSpan, scopeSpans, span)
				}

				ts := int64(span.GetStartTimeUnixNano())
				batch = append(batch, point.NewPoint(opts.PointName, kvs, append(point.CommonLoggingOptions(), point.WithTimestamp(ts))...))
			}
		}

		if len(batch) > 0 {
			batches = append(batches, batch)
		}
	}

	return batches
}

func SelectPublicTraceAttributes(attrs []*common.KeyValue) (point.KVs, []*common.KeyValue) {
	kvs := point.KVs{}
	merged := make([]*common.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if attr == nil {
			continue
		}
		alias, ok := PublicAttributeAlias(attr.GetKey())
		if !ok {
			merged = append(merged, attr)
			continue
		}
		value, ok := AnyValueToInterface(attr.GetValue())
		if !ok {
			merged = append(merged, attr)
			continue
		}
		if alias == "http_status_code" {
			kvs = kvs.AddTag(alias, toString(value))
		} else {
			kvs = kvs.Add(alias, value)
		}
	}

	return kvs, merged
}

func TraceSourceTypeFromTags(tags point.KVs) string {
	for _, tag := range tags {
		switch tag.Key {
		case "http_scheme", "http_method", "rpc_system":
			return TraceSourceTypeWeb
		case "db_system":
			return TraceSourceTypeDB
		case "messaging_system":
			return TraceSourceTypeMessage
		}
	}

	return TraceSourceTypeCustomer
}

func DBHostFromAttributes(resourceAttrs, spanAttrs map[string]string) string {
	if resourceAttrs["db.system"] == "" && spanAttrs["db.system"] == "" {
		return ""
	}
	if host := spanAttrs["server.address"]; host != "" {
		return host
	}
	if host := spanAttrs["net.peer.name"]; host != "" {
		return host
	}
	if host := resourceAttrs["server.address"]; host != "" {
		return host
	}
	return resourceAttrs["net.peer.name"]
}

func collectSpanIDs(resourceSpans []*trace.ResourceSpans, idConverter func([]byte) string) (map[string]bool, map[string]bool) {
	spanIDs := make(map[string]bool)
	parentIDs := make(map[string]bool)

	for _, resourceSpan := range resourceSpans {
		for _, scopeSpans := range resourceSpan.GetScopeSpans() {
			for _, span := range scopeSpans.GetSpans() {
				if span == nil {
					continue
				}
				spanIDs[idConverter(span.GetSpanId())] = true
				parentIDs[idConverter(span.GetParentSpanId())] = true
			}
		}
	}

	return spanIDs, parentIDs
}

func exceptionEventKVs(attrs []*common.KeyValue) point.KVs {
	kvs := point.KVs{}
	for _, attr := range attrs {
		switch attr.GetKey() {
		case AttrExceptionType:
			kvs = kvs.Add("error_type", attr.GetValue().GetStringValue())
		case AttrExceptionMessage:
			kvs = kvs.Add("error_message", attr.GetValue().GetStringValue())
		case AttrExceptionStacktrace:
			kvs = kvs.Add("error_stack", attr.GetValue().GetStringValue())
		}
	}

	return kvs
}

func serviceAndRuntimeFromResource(attrs []*common.KeyValue, fallback string) (service string, runtimeID string) {
	service = fallback
	for _, attr := range attrs {
		switch attr.GetKey() {
		case AttrServiceName:
			if name := attr.GetValue().GetStringValue(); name != "" {
				service = name
			}
		case DefaultTraceRuntimeIDField:
			runtimeID = attr.GetValue().GetStringValue()
		}
	}

	return service, runtimeID
}

func runtimeIDFromAttributes(attrs []*common.KeyValue) string {
	for _, attr := range attrs {
		if attr.GetKey() == DefaultTraceRuntimeIDField {
			return attr.GetValue().GetStringValue()
		}
	}

	return ""
}

func marshalTraceSpan(span *trace.Span) (string, error) {
	if span == nil {
		return "", nil
	}
	if bts, err := protojson.Marshal(span); err == nil {
		return string(bts), nil
	}

	bts, err := json.Marshal(span)
	if err != nil {
		return "", err
	}
	return string(bts), nil
}

func protoCloneTraceSpan(span *trace.Span) *trace.Span {
	if span == nil {
		return nil
	}

	return proto.Clone(span).(*trace.Span)
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprint(v)
	}
}
