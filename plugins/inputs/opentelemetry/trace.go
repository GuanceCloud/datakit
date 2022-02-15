package opentelemetry

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func mkDKTrace(rss []*tracepb.ResourceSpans) []DKtrace.DatakitTrace {
	dkTraces := make([]DKtrace.DatakitTrace, 0)
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibrarySpans()
		// opentelemetry/filter.go:15      resource = attributes:{key:"service.name" value:{string_value:"test-service"}}
		l.Infof("resource = %s", spans.Resource.String())

		service := getServiceName(spans.Resource.Attributes)
		for _, librarySpans := range ls {
			l.Infof("librarySpans.InstrumentationLibrary.Name = %s", librarySpans.InstrumentationLibrary.Name)
			l.Infof("librarySpans.InstrumentationLibrary.Version = %s", librarySpans.InstrumentationLibrary.Version)
			l.Infof("schemaurl = %s", librarySpans.SchemaUrl)
			dktrace := make([]*DKtrace.DatakitSpan, 0)
			for _, span := range librarySpans.Spans {
				tags := toDatakitTags(span.Attributes)
				tags = setTag(tags, spans.Resource.Attributes)
				dkSpan := &DKtrace.DatakitSpan{
					TraceID:            hex.EncodeToString(span.GetTraceId()),
					ParentID:           byteToInt64(span.GetParentSpanId()),
					SpanID:             byteToInt64(span.GetSpanId()),
					Service:            service,
					Resource:           librarySpans.InstrumentationLibrary.Name,
					Operation:          span.Name,
					Source:             inputName,
					SpanType:           span.Kind.String(),
					SourceType:         "",
					Env:                "",
					Project:            "",
					Version:            librarySpans.InstrumentationLibrary.Version,
					Tags:               tags,
					EndPoint:           "",
					HTTPMethod:         getHTTPMethod(span.Attributes),
					HTTPStatusCode:     getHTTPStatusCode(span.Attributes),
					ContainerHost:      getContainerHost(span.Attributes),
					PID:                getPID(span.Attributes),
					Start:              int64(span.StartTimeUnixNano),                        //  注意单位 nano
					Duration:           int64(span.EndTimeUnixNano - span.StartTimeUnixNano), // 单位 nano
					Status:             getDKSpanStatus(span.GetStatus().Code),               // 使用 dk status
					Content:            "",
					Priority:           0,
					SamplingRateGlobal: 0,
				}
				bts, err := json.Marshal(dkSpan)
				if err == nil {
					dkSpan.Content = string(bts)
				}
				l.Infof("dkspan = %+v", dkSpan)
				l.Infof("span = %+v", span)
				dktrace = append(dktrace, dkSpan)
			}
			dkTraces = append(dkTraces, dktrace)
		}
	}
	return dkTraces
}

// toDatakitTags : make attributes to tags
func toDatakitTags(attr []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(attr))
	for _, kv := range attr {
		// key := replace(kv.Key)
		m[kv.Key] = kv.GetValue().GetStringValue()
		/*
			switch kv.GetValue().Value.(type) {
			// For slice attributes, serialize as JSON list string.
			case *v1.AnyValue_StringValue:
				m[kv.Key] = kv.GetValue().GetStringValue()
			case *v1.AnyValue_BoolValue:
			case *v1.AnyValue_IntValue:
			case *v1.AnyValue_DoubleValue:
			case *v1.AnyValue_ArrayValue:
			case *v1.AnyValue_KvlistValue:
			case *v1.AnyValue_BytesValue:

			default:
				m[(string)(kv.Key)] = kv.Value.Emit()
			}
		*/
	}

	return m
}

func byteToInt64(bts []byte) string {
	hexCode := hex.EncodeToString(bts)
	if hexCode == "" {
		return "0"
	}
	return hexCode
}

// getDKSpanStatus 从otel的status转成dk的span_status
func getDKSpanStatus(code tracepb.Status_StatusCode) string {
	status := DKtrace.STATUS_INFO
	switch code {
	case tracepb.Status_STATUS_CODE_UNSET:
		status = DKtrace.STATUS_INFO
	case tracepb.Status_STATUS_CODE_OK:
		status = DKtrace.STATUS_OK
	case tracepb.Status_STATUS_CODE_ERROR:
		status = DKtrace.STATUS_ERR
	default:
	}
	return status
}

func getServiceName(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "service.name" {
			return kv.Value.GetStringValue()
		}
	}
	return "unknown.service"
}

func getHTTPMethod(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "http.method" { // see :vendor/go.opentelemetry.io/otel/semconv/v1.4.0/trace.go:742
			return kv.Value.GetStringValue()
		}
	}
	return ""
}

func getHTTPStatusCode(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "http.status_code" { //see :vendor/go.opentelemetry.io/otel/semconv/v1.4.0/trace.go:784
			return kv.Value.GetStringValue()
		}
	}
	return ""
}

func getContainerHost(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "container.name" {
			return kv.Value.GetStringValue()
		}
	}
	return ""
}

func getPID(attr []*commonpb.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "process.pid" { //see :vendor/go.opentelemetry.io/otel/semconv/v1.4.0/resource.go:686
			return kv.Value.GetStringValue()
		}
	}
	return ""
}

// replace 行协议中的tag的key禁止有点 全部替换掉
func replace(key string) string {
	return strings.ReplaceAll(key, ".", "_")
}
