package opentelemetry

import (
	"encoding/hex"
	"strings"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	common "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func mkDKTrace(rss []*tracepb.ResourceSpans) DKtrace.DatakitTraces {
	dkTraces := make([]DKtrace.DatakitTrace, 0)
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibrarySpans()
		//	spans.ProtoMessage()
		l.Infof("resource = %s", spans.Resource.String()) //opentelemetry/filter.go:15      resource = attributes:{key:"service.name" value:{string_value:"test-service"}}
		service := getServiceName(spans.Resource.Attributes)
		l.Infof("GetSchemaUrl = %s", spans.GetSchemaUrl()) // opentelemetry/filter.go:17      GetSchemaUrl =
		for _, librarySpans := range ls {
			l.Infof("librarySpans.InstrumentationLibrary.Name = %s", librarySpans.InstrumentationLibrary.Name)
			l.Infof("librarySpans.InstrumentationLibrary.Version = %s", librarySpans.InstrumentationLibrary.Version)
			l.Infof("schemaurl = %s", librarySpans.SchemaUrl)
			dktrace := make([]*DKtrace.DatakitSpan, 0)
			for _, span := range librarySpans.Spans {
				tags := toDatakitTags(span.Attributes)
				tags = setTag(tags, spans.Resource.Attributes)
				dkSpan := &DKtrace.DatakitSpan{
					TraceID:        hex.EncodeToString(span.GetTraceId()),
					ParentID:       hex.EncodeToString(span.GetParentSpanId()),
					SpanID:         hex.EncodeToString(span.GetSpanId()),
					Service:        service,
					Resource:       librarySpans.InstrumentationLibrary.Name,
					Operation:      span.Name,
					Source:         inputName,
					SpanType:       span.Kind.String(),
					SourceType:     "",
					Env:            "",
					Project:        "",
					Version:        librarySpans.InstrumentationLibrary.Version,
					Tags:           tags,
					EndPoint:       "",
					HTTPMethod:     "",
					HTTPStatusCode: "",
					ContainerHost:  "",
					PID:            "",
					Start:          int64(span.StartTimeUnixNano),                                //  注意单位 nano
					Duration:       int64(span.EndTimeUnixNano - span.StartTimeUnixNano),         // 单位 nano
					Status:         tracepb.Status_StatusCode_name[int32(span.GetStatus().Code)], // 使用 dk status
					Content:        "",
					SampleRate:     0,
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

func setTag(tags map[string]string, attr []*common.KeyValue) map[string]string {
	for _, kv := range attr {
		key := replace(kv.Key)
		tags[key] = kv.GetValue().String()
	}
	return tags
}

// toDatakitTags : make attributes to tags
func toDatakitTags(attr []*common.KeyValue) map[string]string {
	m := make(map[string]string, len(attr))
	for _, kv := range attr {
		key := replace(kv.Key)
		m[key] = kv.GetValue().String()
		/*switch kv.GetValue().Value.(type) {
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
		}*/
	}

	return m
}

func getServiceName(attr []*common.KeyValue) string {
	for _, kv := range attr {
		if kv.Key == "service.name" {
			if stringVal, ok := kv.Value.Value.(*common.AnyValue_StringValue); ok {
				return stringVal.StringValue
			}
		}
	}
	return ""
}

func replace(key string) string {
	return strings.ReplaceAll(key, ".", "")
}

// TODO :CalculatorFunc and FilterFunc
