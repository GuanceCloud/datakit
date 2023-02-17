// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/hex"
	"encoding/json"

	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	tracepb "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/opentelemetry/compiled/v1/trace"
)

func parseResourceSpans(resspans []*tracepb.ResourceSpans) itrace.DatakitTraces {
	var dktraces itrace.DatakitTraces
	spanIDs, parentIDs := getSpanIDsAndParentIDs(resspans)
	for _, spans := range resspans {
		var service, version, containerName, pid string
		if attr, ok := getAttribute(otelResourceServiceKey, spans.Resource.Attributes); ok {
			service = attr.Value.GetStringValue()
		}
		if attr, ok := getAttribute(otelResourceServiceVersionKey, spans.Resource.Attributes); ok {
			version = attr.Value.GetStringValue()
		}
		if attr, ok := getAttribute(otelResourceContainerNameKey, spans.Resource.Attributes); ok {
			containerName = attr.Value.GetStringValue()
		}
		if attr, ok := getAttribute(otelResourceProcessIDKey, spans.Resource.Attributes); ok {
			pid = attr.Value.GetStringValue()
		}

		restags, resfiedls := extractAtrribute(spans.Resource.Attributes)

		var dktrace itrace.DatakitTrace
		for _, scopeSpans := range spans.ScopeSpans {
			scopetags, scopefields := extractAtrribute(scopeSpans.Scope.Attributes)
			scopetags = itrace.MergeTags(restags, scopetags)
			scopefields = itrace.MergeFields(resfiedls, scopefields)

			for _, span := range scopeSpans.Spans {
				dkspan := &itrace.DatakitSpan{
					TraceID:   hex.EncodeToString(span.GetTraceId()),
					ParentID:  byteToString(span.GetParentSpanId()),
					SpanID:    byteToString(span.GetSpanId()),
					Service:   service,
					Resource:  span.Name,
					Operation: span.Name,
					Source:    inputName,
					Start:     int64(span.StartTimeUnixNano),
					Duration:  int64(span.EndTimeUnixNano - span.StartTimeUnixNano),
					// TODO: optimize status fetch
					Status: getDKSpanStatus(span.GetStatus()),
				}
				dkspan.SpanType = itrace.FindSpanTypeStrSpanID(dkspan.SpanID, dkspan.ParentID, spanIDs, parentIDs)

				// set all attributes into dk tags and fields respectively
				spantags, spanfields := extractAtrribute(span.Attributes)
				for i := range span.Events {
					eventtags, eventfields := extractAtrribute(span.Events[i].Attributes)
					spantags = itrace.MergeTags(spantags, eventtags)
					spanfields = itrace.MergeFields(spanfields, eventfields)
				}
				dkspan.Tags = spantags
				dkspan.Metrics = spanfields

				// TODO: get otel span source_type

				dkspan.Tags[itrace.TAG_VERSION] = version
				dkspan.Tags[itrace.TAG_CONTAINER_HOST] = containerName
				dkspan.Tags[itrace.TAG_PID] = pid

				if attr, ok := getAttribute(otelHTTPMethodKey, span.Attributes); ok {
					dkspan.Tags[itrace.TAG_HTTP_METHOD] = attr.Value.GetStringValue()
				}
				if attr, ok := getAttribute(otelHTTPStatusCodeKey, span.Attributes); ok {
					dkspan.Tags[itrace.TAG_HTTP_STATUS_CODE] = attr.Value.GetStringValue()
				}

				for i := range span.Events {
					if span.Events[i].Name == ExceptionEventName {
						for o, d := range otelErrKeyToDkErrKey {
							if attr, ok := getAttribute(o, span.Events[i].Attributes); ok {
								delete(dkspan.Tags, o)
								delete(dkspan.Metrics, o)
								dkspan.Metrics[d] = attr.Value.GetStringValue()
							}
						}
						break
					}
				}

				if buf, err := json.Marshal(span); err != nil {
					log.Warn(err.Error())
				} else {
					dkspan.Content = string(buf)
				}

				dktrace = append(dktrace, dkspan)
			}
		}
		if len(dktrace) != 0 {
			dktrace[0].Metrics[itrace.FIELD_PRIORITY] = itrace.PRIORITY_AUTO_KEEP
			dktraces = append(dktraces, dktrace)
		}
	}

	return dktraces
}

func getSpanIDsAndParentIDs(resspans []*tracepb.ResourceSpans) (map[string]bool, map[string]bool) {
	var (
		spanIDs   = make(map[string]bool)
		parentIDs = make(map[string]bool)
	)
	for _, spans := range resspans {
		for _, scopespans := range spans.ScopeSpans {
			for _, span := range scopespans.Spans {
				if span == nil {
					continue
				}
				spanIDs[byteToString(span.SpanId)] = true
				parentIDs[byteToString(span.ParentSpanId)] = true
			}
		}
	}

	return spanIDs, parentIDs
}

func byteToString(buf []byte) string {
	if len(buf) == 0 || string(buf) == "0" {
		return "0"
	}

	return hex.EncodeToString(buf)
}

// getDKSpanStatus 从otel的status转成dk的span_status.
func getDKSpanStatus(statuspb *tracepb.Status) string {
	status := itrace.STATUS_INFO
	if statuspb == nil {
		return status
	}
	switch statuspb.Code {
	case tracepb.Status_STATUS_CODE_OK, tracepb.Status_STATUS_CODE_UNSET:
		status = itrace.STATUS_OK
	case tracepb.Status_STATUS_CODE_ERROR:
		status = itrace.STATUS_ERR
	default:
	}

	return status
}

// type dkTags struct {
// 	// 配置文件中的黑名单配置，通过正则过滤数据中的标签
// 	regexpString string
// 	// 配置文件中的全局标签
// 	globalTags map[string]string
// 	// 从span中获取的attribute 放到tags中
// 	tags map[string]string
// 	// 将`.`替换成`_`之后的map,防止二次遍历查找key,所以两者不可用一个map
// 	replaceTags map[string]string
// }

// func newEmptyTags(regexp string, globalTags map[string]string) *dkTags {
// 	return &dkTags{
// 		regexpString: regexp,
// 		globalTags:   globalTags,
// 		tags:         make(map[string]string),
// 		replaceTags:  make(map[string]string),
// 	}
// }

// func (dt *dkTags) makeAllTags(span *tracepb.Span, resourceAttr []*commonpb.KeyValue) {
// 	/*
// 		trace tags 处理逻辑:
// 			1 先将tags从resource中提取
// 			2 从span attributes中提取
// 			3 从span的剩余字段中提取
// 			4 统一的key处理
// 			5 过黑白名单
// 			6 添加global tags
// 		 如果要换成白名单机制，则顺序应该改变：步骤3在黑白名单之后
// 	*/
// 	dt.setAttributesToTags(resourceAttr).
// 		setAttributesToTags(span.Attributes).
// 		addOtherTags(span).
// 		checkAllTagsKey().
// 		checkCustomTags().
// 		addGlobalTags()
// }

// func (dt *dkTags) setAttributesToTags(attr []*commonpb.KeyValue) *dkTags {
// 	for _, kv := range attr {
// 		key := kv.Key
// 		switch t := kv.GetValue().Value.(type) {
// 		case *commonpb.AnyValue_StringValue:
// 			dt.tags[key] = kv.GetValue().GetStringValue()
// 		case *commonpb.AnyValue_BoolValue:
// 			dt.tags[key] = strconv.FormatBool(t.BoolValue)
// 		case *commonpb.AnyValue_IntValue:
// 			dt.tags[key] = strconv.FormatInt(t.IntValue, 10)
// 		case *commonpb.AnyValue_DoubleValue:
// 			// 保留两位小数
// 			dt.tags[key] = strconv.FormatFloat(t.DoubleValue, 'f', 2, 64)
// 		case *commonpb.AnyValue_ArrayValue:
// 			dt.tags[key] = t.ArrayValue.String()
// 		case *commonpb.AnyValue_KvlistValue:
// 			dt.setAttributesToTags(t.KvlistValue.Values)
// 		case *commonpb.AnyValue_BytesValue:
// 			dt.tags[key] = string(t.BytesValue)
// 		default:
// 			dt.tags[key] = kv.Value.GetStringValue()
// 		}
// 	}
// 	return dt
// }

// // checkCustomTags : 黑白名单机制.
// func (dt *dkTags) checkCustomTags() *dkTags {
// 	if dt.regexpString == "" {
// 		return dt
// 	}
// 	reg := regexp.MustCompile(dt.regexpString)
// 	for key := range dt.replaceTags {
// 		if reg.MatchString(key) {
// 			delete(dt.replaceTags, key)
// 		}
// 	}
// 	return dt
// }

// // addGlobalTags: 添加配置文件中的自定义tags.
// func (dt *dkTags) addGlobalTags() *dkTags {
// 	for k, v := range dt.globalTags {
// 		dt.replaceTags[k] = v
// 	}
// 	return dt
// }

// // checkAllTagsKey 统一替换key 放进replaceTags中.
// func (dt *dkTags) checkAllTagsKey() *dkTags {
// 	for key, val := range dt.tags {
// 		dt.replaceTags[replace(key)] = val
// 	}
// 	return dt
// }

// func (dt *dkTags) addOtherTags(span *tracepb.Span) *dkTags {
// 	if span.DroppedAttributesCount != 0 {
// 		count := strconv.Itoa(int(span.DroppedAttributesCount))
// 		dt.tags[DroppedAttributesCount] = count
// 	}
// 	if span.DroppedEventsCount != 0 {
// 		count := strconv.Itoa(int(span.DroppedEventsCount))
// 		dt.tags[DroppedEventsCount] = count
// 	}
// 	if span.DroppedLinksCount != 0 {
// 		count := strconv.Itoa(int(span.DroppedLinksCount))
// 		dt.tags[DroppedLinksCount] = count
// 	}
// 	if len(span.Events) != 0 {
// 		count := strconv.Itoa(len(span.Events))
// 		dt.tags[Events] = count
// 	}
// 	if len(span.Links) != 0 {
// 		count := strconv.Itoa(len(span.Links))
// 		dt.tags[Links] = count
// 	}
// 	return dt
// }

// func (dt *dkTags) resource() map[string]string {
// 	return dt.replaceTags
// }

// func (dt *dkTags) getAttributeVal(keyName string) (string, bool) {
// 	for k, v := range dt.tags {
// 		if k == keyName {
// 			return v, true
// 		}
// 	}
// 	if keyName == otelResourceServiceKey {
// 		return otelUnknownServiceName, true
// 	}

// 	return "", false
// }

// func (dt *dkTags) getResourceType() string {
// 	for key := range dt.tags {
// 		switch key {
// 		case otelHTTPSchemeKey, otelHTTPMethodKey:
// 			return itrace.SPAN_SOURCE_WEB
// 		case otelDBSystemKey:
// 			return itrace.SPAN_SOURCE_DB
// 		}
// 	}

// 	return itrace.SPAN_SOURCE_CUSTOMER
// }
