// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	logs "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/logs/v1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	kb = 1024
	// logMaxLen 默认500kb.
	logMaxLen = 500 * kb
)

func (ipt *Input) parseLogRequest(resourceLogss []*logs.ResourceLogs) []*point.Point {
	pts := make([]*point.Point, 0)
	for _, resourceLogs := range resourceLogss {
		resourceTags := attributesToTag(resourceLogs.GetResource().GetAttributes()) // resource Attr
		resourceTags["schema_url"] = resourceLogs.GetSchemaUrl()
		service, source := getServiceAndSource(resourceLogs.GetResource().GetAttributes())
		host := getHostName(resourceLogs.GetResource().GetAttributes())

		for _, scope := range resourceLogs.GetScopeLogs() {
			scopeTags := attributesToTag(scope.GetScope().GetAttributes()) // scope Attr
			scopeTags["scope_name"] = scope.GetScope().GetName()
			for _, record := range scope.GetLogRecords() {
				ptTags := attributesToTag(record.GetAttributes())

				body := record.GetBody()
				message := ""
				switch body.GetValue().(type) {
				case *common.AnyValue_StringValue:
					message = body.GetStringValue()
				case *common.AnyValue_BytesValue:
					message = string(body.GetBytesValue())
				case *common.AnyValue_DoubleValue:
					message = fmt.Sprintf("%f", body.GetDoubleValue())
				case *common.AnyValue_IntValue:
					message = strconv.FormatInt(body.GetIntValue(), 10)
				case *common.AnyValue_BoolValue:
					message = strconv.FormatBool(body.GetBoolValue())
				case *common.AnyValue_ArrayValue:
					message = body.GetArrayValue().String()
				case *common.AnyValue_KvlistValue:
					message = body.GetKvlistValue().String()
				}
				messages := splitByByteLength(message, logMaxLen)
				for i, msg := range messages {
					kvs := mergeTagsToField(resourceTags, scopeTags, ptTags)
					for k, v := range ipt.Tags { // span.attribute 优先级大于全局tag。
						kvs = kvs.AddV2(k, v, false)
					}
					kvs = kvs.Add("message", msg, false, false).
						AddV2(itrace.FieldSpanid, ipt.convertBinID(record.GetSpanId()), false).
						AddV2(itrace.FieldTraceID, ipt.convertBinID(record.GetTraceId()), false).
						AddTag("status", getStatus(record.GetSeverityNumber(), record.GetSeverityText())).
						AddTag("service", service).
						AddTag(itrace.TagSource, source).
						AddTag(itrace.TagDKFingerprintKey, datakit.DatakitHostName+"_"+datakit.Version)

					if host != "" {
						kvs = kvs.AddTag("host", host)
					}
					ts := time.Unix(0, int64(record.GetTimeUnixNano()))
					if record.GetTimeUnixNano() == 0 {
						ts = time.Unix(0, int64(record.GetObservedTimeUnixNano()))
					}
					opts := point.DefaultLoggingOptions()
					opts = append(opts, point.WithTime(ts.Add(time.Millisecond*time.Duration(i))))
					pts = append(pts, point.NewPointV2(source, kvs, opts...))
				}
			}
		}
	}
	return pts
}

func splitByByteLength(s string, length int) []string {
	runes := []rune(s)
	var chunks []string
	for i := 0; i < len(runes); i += length {
		end := i + length
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}

func getStatus(severityNum logs.SeverityNumber, level string) string {
	switch severityNum {
	case logs.SeverityNumber_SEVERITY_NUMBER_TRACE,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE2,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE3,
		logs.SeverityNumber_SEVERITY_NUMBER_TRACE4:
		return "trace"
	case logs.SeverityNumber_SEVERITY_NUMBER_DEBUG,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG2,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG3,
		logs.SeverityNumber_SEVERITY_NUMBER_DEBUG4:
		return "debug"
	case logs.SeverityNumber_SEVERITY_NUMBER_INFO,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO2,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO3,
		logs.SeverityNumber_SEVERITY_NUMBER_INFO4:
		return "info"
	case logs.SeverityNumber_SEVERITY_NUMBER_WARN,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN2,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN3,
		logs.SeverityNumber_SEVERITY_NUMBER_WARN4:
		return "warn"
	case logs.SeverityNumber_SEVERITY_NUMBER_ERROR,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR2,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR3,
		logs.SeverityNumber_SEVERITY_NUMBER_ERROR4:
		return "error"
	case logs.SeverityNumber_SEVERITY_NUMBER_FATAL,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL2,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL3,
		logs.SeverityNumber_SEVERITY_NUMBER_FATAL4:
		return "fatal"
	case logs.SeverityNumber_SEVERITY_NUMBER_UNSPECIFIED:
		return "unknown"
	}

	return level
}

func getServiceAndSource(attr []*common.KeyValue) (service string, source string) {
	for _, keyValue := range attr {
		if keyValue.GetKey() == otelResourceServiceKey {
			service = keyValue.GetValue().GetStringValue()
		}
		if keyValue.GetKey() == "log.source" {
			source = keyValue.GetValue().GetStringValue()
		}
	}

	if source == "" {
		if service == "" {
			source = "otel_logs"
		} else {
			source = service
		}
	}
	if service == "" {
		service = "unSetServiceName"
	}
	return
}

func getHostName(attr []*common.KeyValue) (hostName string) {
	for _, keyValue := range attr {
		if keyValue.GetKey() == "host.name" {
			hostName = keyValue.GetValue().GetStringValue()
		}
	}

	return
}
