package opentelemetry

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func mkDKTrace(rss []*tracepb.ResourceSpans) []DKtrace.DatakitTrace {
	dkTraces := make([]DKtrace.DatakitTrace, 0)
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibrarySpans()
		l.Infof("resource = %s", spans.Resource.String())

		tags := make(map[string]string)
		setResourceAttributes(tags, spans.Resource.GetAttributes())
		service := getServiceName(spans.Resource.Attributes)
		for _, librarySpans := range ls {
			dktrace := make([]*DKtrace.DatakitSpan, 0)
			for _, span := range librarySpans.Spans {
				//	tags := toDatakitTags(span.Attributes)
				tags := toDataKitTagsV2(span, spans.Resource.Attributes)

				//	tags = setTag(tags, spans.Resource.Attributes) // 将用户自定义和global tag添加到tags中
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
					Start:              int64(span.StartTimeUnixNano),                        // 注意单位 nano
					Duration:           int64(span.EndTimeUnixNano - span.StartTimeUnixNano), // 单位 nano
					Status:             getDKSpanStatus(span.GetStatus().Code),               // 使用 dk status
					Content:            "",
					Priority:           0,
					SamplingRateGlobal: 0,
				}
				bts, err := json.Marshal(span)
				if err == nil {
					dkSpan.Content = string(bts)
				}

				l.Infof("dkspan = %+v", dkSpan)
				l.Infof("")
				l.Infof("span = %s", span.String())
				dktrace = append(dktrace, dkSpan)
			}
			dkTraces = append(dkTraces, dktrace)
		}
	}
	return dkTraces
}

func toDataKitTagsV2(span *tracepb.Span, resourceAttr []*commonpb.KeyValue) map[string]string {
	/*
		tags :
			1 先将tags从resource中提取
			2 从span attributes中提取
			3 从span的剩余字段中提取
			4 统一的key处理
			5 过黑白名单
			6 添加global tags
	*/
	tags := make(map[string]string)
	setResourceAttributes(tags, resourceAttr)
	setSpanAttributes(tags, span.Attributes)
	addOtherTags(tags, span)
	tags = checkAllTagsKey(tags)
	checkCustomTags(tags)
	addGlobalTags(tags)
	return tags
}

func setSpanAttributes(tags map[string]string, attr []*commonpb.KeyValue) {
	for _, kv := range attr {
		//	if _, ok := customTags[kv.Key]; ok { // 白名单机制
		//	key := replace(kv.Key) // 统一将`.`换成 `_`
		key := kv.Key
		switch t := kv.GetValue().Value.(type) {
		case *commonpb.AnyValue_StringValue:
			tags[key] = kv.GetValue().GetStringValue()
		case *commonpb.AnyValue_BoolValue:
			tags[key] = strconv.FormatBool(t.BoolValue)
		case *commonpb.AnyValue_IntValue:
			tags[key] = strconv.FormatInt(t.IntValue, 10)
		case *commonpb.AnyValue_DoubleValue:
			tags[key] = strconv.FormatFloat(t.DoubleValue, 'f', 2, 64)
		case *commonpb.AnyValue_ArrayValue:
			tags[key] = t.ArrayValue.String()
		case *commonpb.AnyValue_KvlistValue:
			setSpanAttributes(tags, t.KvlistValue.Values)
			for s, s2 := range tags {
				tags[s] = s2
			}
		case *commonpb.AnyValue_BytesValue:
			tags[key] = string(t.BytesValue)
		default:
			tags[key] = kv.Value.GetStringValue()
		}
		//	}
	}
}

func setResourceAttributes(tags map[string]string, attr []*commonpb.KeyValue) {
	// todo
}

// todo 黑白名单机制
func checkCustomTags(tags map[string]string) {

}

// setGlobalTags: 添加配置文件中的自定义tags
func addGlobalTags(tags map[string]string) {
	// set global tags
	for k, v := range globalTags {
		tags[k] = v
	}
}

// 统一做替换
func checkAllTagsKey(tags map[string]string) map[string]string {
	newTags := make(map[string]string)
	for key, val := range tags {
		newTags[replace(key)] = val
	}
	return newTags
}

func addOtherTags(tags map[string]string, span *tracepb.Span) {
	if span.DroppedAttributesCount != 0 {
		count := strconv.Itoa(int(span.DroppedAttributesCount))
		tags[DroppedAttributesCount] = count
	}
	if span.DroppedEventsCount != 0 {
		count := strconv.Itoa(int(span.DroppedEventsCount))
		tags[DroppedEventsCount] = count
	}
	if span.DroppedLinksCount != 0 {
		count := strconv.Itoa(int(span.DroppedLinksCount))
		tags[DroppedLinksCount] = count
	}
	if len(span.Events) != 0 {
		count := strconv.Itoa(len(span.Events))
		tags[Events] = count
	}
	if len(span.Links) != 0 {
		count := strconv.Itoa(len(span.Links))
		tags[Links] = count
	}
}

// toDatakitTags : make attributes to tags
func toDatakitTags(attr []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(attr))
	for _, kv := range attr {
		if _, ok := customTags[kv.Key]; ok { // 白名单机制
			key := replace(kv.Key) // 统一将`.`换成 `_`
			switch t := kv.GetValue().Value.(type) {
			case *commonpb.AnyValue_StringValue:
				m[key] = kv.GetValue().GetStringValue()
			case *commonpb.AnyValue_BoolValue:
				m[key] = strconv.FormatBool(t.BoolValue)
			case *commonpb.AnyValue_IntValue:
				m[key] = strconv.FormatInt(t.IntValue, 10)
			case *commonpb.AnyValue_DoubleValue:
				m[key] = strconv.FormatFloat(t.DoubleValue, 'f', 2, 64)
			case *commonpb.AnyValue_ArrayValue:
				m[key] = t.ArrayValue.String()
			case *commonpb.AnyValue_KvlistValue:
				tags := toDatakitTags(t.KvlistValue.Values)
				for s, s2 := range tags {
					m[s] = s2
				}
			case *commonpb.AnyValue_BytesValue:
				m[key] = string(t.BytesValue)
			default:
				m[key] = kv.Value.GetStringValue()
			}
		}
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
