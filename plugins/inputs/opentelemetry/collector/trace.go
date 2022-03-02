// Package collector is trace and tags.
package collector

import (
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	DKtrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
)

func (s *SpansStorage) mkDKTrace(rss []*tracepb.ResourceSpans) []DKtrace.DatakitTrace {
	dkTraces := make([]DKtrace.DatakitTrace, 0)
	for _, spans := range rss {
		ls := spans.GetInstrumentationLibrarySpans()
		for _, librarySpans := range ls {
			dktrace := make([]*DKtrace.DatakitSpan, 0)
			for _, span := range librarySpans.Spans {
				dt := newEmptyTags(s.RegexpString, s.GlobalTags)
				dt.makeAllTags(span, spans.Resource.Attributes)
				dkSpan := &DKtrace.DatakitSpan{
					TraceID:            hex.EncodeToString(span.GetTraceId()),
					ParentID:           byteToString(span.GetParentSpanId()),
					SpanID:             byteToString(span.GetSpanId()),
					Service:            dt.getAttributeVal(otelResourceServiceKey),
					Resource:           librarySpans.InstrumentationLibrary.Name,
					Operation:          span.Name,
					Source:             inputName,
					SpanType:           span.Kind.String(),
					SourceType:         "",
					Env:                "",
					Project:            "",
					Version:            librarySpans.InstrumentationLibrary.Version,
					Tags:               dt.resource(),
					EndPoint:           "",
					HTTPMethod:         dt.getAttributeVal(otelResourceHTTPMethodKey),
					HTTPStatusCode:     dt.getAttributeVal(otelResourceHTTPStatusCodeKey),
					ContainerHost:      dt.getAttributeVal(otelResourceContainerNameKey),
					PID:                dt.getAttributeVal(otelResourceProcessPidKey),
					Start:              int64(span.StartTimeUnixNano),                        // 注意单位 nano
					Duration:           int64(span.EndTimeUnixNano - span.StartTimeUnixNano), // 单位 nano
					Status:             getDKSpanStatus(span.GetStatus()),                    // 使用 dk status
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

type dkTags struct {
	// 配置文件中的黑名单配置，通过正则过滤数据中的标签
	regexpString string

	// 配置文件中的全局标签
	globalTags map[string]string

	// 从span中获取的attribute 放到tags中
	tags map[string]string

	// 将`.`替换成`_`之后的map,防止二次遍历查找key,所以两者不可用一个map
	replaceTags map[string]string
}

func newEmptyTags(regexp string, globalTags map[string]string) *dkTags {
	return &dkTags{
		regexpString: regexp,
		globalTags:   globalTags,
		tags:         make(map[string]string),
		replaceTags:  make(map[string]string),
	}
}

func (dt *dkTags) makeAllTags(span *tracepb.Span, resourceAttr []*commonpb.KeyValue) {
	/*
		trace tags 处理逻辑:
			1 先将tags从resource中提取
			2 从span attributes中提取
			3 从span的剩余字段中提取
			4 统一的key处理
			5 过黑白名单
			6 添加global tags
		 如果要换成白名单机制，则顺序应该改变：步骤3在黑白名单之后
	*/
	dt.setAttributesToTags(resourceAttr).
		setAttributesToTags(span.Attributes).
		addOtherTags(span).
		checkAllTagsKey().
		checkCustomTags().
		addGlobalTags()
}

func (dt *dkTags) setAttributesToTags(attr []*commonpb.KeyValue) *dkTags {
	for _, kv := range attr {
		key := kv.Key
		switch t := kv.GetValue().Value.(type) {
		case *commonpb.AnyValue_StringValue:
			dt.tags[key] = kv.GetValue().GetStringValue()
		case *commonpb.AnyValue_BoolValue:
			dt.tags[key] = strconv.FormatBool(t.BoolValue)
		case *commonpb.AnyValue_IntValue:
			dt.tags[key] = strconv.FormatInt(t.IntValue, 10)
		case *commonpb.AnyValue_DoubleValue:
			// 保留两位小数
			dt.tags[key] = strconv.FormatFloat(t.DoubleValue, 'f', 2, 64)
		case *commonpb.AnyValue_ArrayValue:
			dt.tags[key] = t.ArrayValue.String()
		case *commonpb.AnyValue_KvlistValue:
			dt.setAttributesToTags(t.KvlistValue.Values)
		case *commonpb.AnyValue_BytesValue:
			dt.tags[key] = string(t.BytesValue)
		default:
			dt.tags[key] = kv.Value.GetStringValue()
		}
	}
	return dt
}

// checkCustomTags : 黑白名单机制
func (dt *dkTags) checkCustomTags() *dkTags {
	if dt.regexpString == "" {
		return dt
	}
	reg := regexp.MustCompile(dt.regexpString)
	for key := range dt.replaceTags {
		if reg.MatchString(key) {
			delete(dt.replaceTags, key)
		}
	}
	return dt
}

// addGlobalTags: 添加配置文件中的自定义tags
func (dt *dkTags) addGlobalTags() *dkTags {
	// set global tags
	for k, v := range dt.globalTags {
		dt.replaceTags[k] = v
	}
	return dt
}

// checkAllTagsKey 统一替换key 放进replaceTags中
func (dt *dkTags) checkAllTagsKey() *dkTags {
	for key, val := range dt.tags {
		dt.replaceTags[replace(key)] = val
	}
	return dt
}

func (dt *dkTags) addOtherTags(span *tracepb.Span) *dkTags {
	if span.DroppedAttributesCount != 0 {
		count := strconv.Itoa(int(span.DroppedAttributesCount))
		dt.tags[DroppedAttributesCount] = count
	}
	if span.DroppedEventsCount != 0 {
		count := strconv.Itoa(int(span.DroppedEventsCount))
		dt.tags[DroppedEventsCount] = count
	}
	if span.DroppedLinksCount != 0 {
		count := strconv.Itoa(int(span.DroppedLinksCount))
		dt.tags[DroppedLinksCount] = count
	}
	if len(span.Events) != 0 {
		count := strconv.Itoa(len(span.Events))
		dt.tags[Events] = count
	}
	if len(span.Links) != 0 {
		count := strconv.Itoa(len(span.Links))
		dt.tags[Links] = count
	}
	return dt
}

func (dt *dkTags) resource() map[string]string {
	return dt.replaceTags
}

func (dt *dkTags) getAttributeVal(keyName string) string {
	for k, v := range dt.tags {
		if k == keyName {
			return v
		}
	}
	if keyName == otelResourceServiceKey {
		return defaultServiceVal // set default to 'service.name'
	}
	return ""
}

func byteToString(bts []byte) string {
	hexCode := hex.EncodeToString(bts)
	if hexCode == "" {
		return "0"
	}
	return hexCode
}

// getDKSpanStatus 从otel的status转成dk的span_status
func getDKSpanStatus(statuspb *tracepb.Status) string {
	status := DKtrace.STATUS_INFO
	if statuspb == nil {
		return status
	}
	switch statuspb.Code {
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

// replace 行协议中的tag的key禁止有点 全部替换掉
func replace(key string) string {
	return strings.ReplaceAll(key, ".", "_")
}
