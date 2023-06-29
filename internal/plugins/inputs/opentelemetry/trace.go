// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/hex"
	"encoding/json"

	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func parseResourceSpans(resspans []*trace.ResourceSpans) itrace.DatakitTraces {
	var dktraces itrace.DatakitTraces
	spanIDs, parentIDs := getSpanIDsAndParentIDs(resspans)
	for _, spans := range resspans {
		resattrs := extractAtrributes(spans.Resource.Attributes)

		var dktrace itrace.DatakitTrace
		for _, scopeSpans := range spans.ScopeSpans {
			scpattrs := extractAtrributes(scopeSpans.Scope.Attributes)

			for _, span := range scopeSpans.Spans {
				spattrs := extractAtrributes(span.Attributes)

				dkspan := &itrace.DatakitSpan{
					TraceID:   hex.EncodeToString(span.GetTraceId()),
					ParentID:  byteToString(span.GetParentSpanId()),
					SpanID:    byteToString(span.GetSpanId()),
					Resource:  span.Name,
					Operation: span.Name,
					Source:    inputName,
					Tags:      make(map[string]string),
					Metrics:   make(map[string]interface{}),
					Start:     int64(span.StartTimeUnixNano),
					Duration:  int64(span.EndTimeUnixNano - span.StartTimeUnixNano),
					Status:    getDKSpanStatus(span.GetStatus()),
				}
				dkspan.SpanType = itrace.FindSpanTypeStrSpanID(dkspan.SpanID, dkspan.ParentID, spanIDs, parentIDs)

				attrs := newAttributes(resattrs).merge(scpattrs...).merge(spattrs...)
				if kv, i := attrs.find(otelResourceServiceKey); i != -1 {
					dkspan.Service = kv.Value.GetStringValue()
				}
				if kv, i := attrs.find(otelResourceServiceVersionKey); i != -1 {
					dkspan.Tags[itrace.TAG_VERSION] = kv.Value.GetStringValue()
				}
				if kv, i := attrs.find(otelResourceProcessIDKey); i != -1 {
					dkspan.Tags[itrace.TAG_PID] = kv.Value.GetStringValue()
				}
				if kv, i := attrs.find(otelResourceContainerNameKey); i != -1 {
					dkspan.Tags[itrace.TAG_CONTAINER_HOST] = kv.Value.GetStringValue()
				}
				if kv, i := attrs.find(otelHTTPMethodKey); i != -1 {
					dkspan.Tags[itrace.TAG_HTTP_METHOD] = kv.Value.GetStringValue()
					attrs.remove(otelHTTPMethodKey)
				}
				if kv, i := attrs.find(otelHTTPStatusCodeKey); i != -1 {
					dkspan.Tags[itrace.TAG_HTTP_STATUS_CODE] = kv.Value.GetStringValue()
					attrs.remove(otelHTTPStatusCodeKey)
				}

				for i := range span.Events {
					if span.Events[i].Name == ExceptionEventName {
						for o, d := range otelErrKeyToDkErrKey {
							if attr, ok := getAttribute(o, span.Events[i].Attributes); ok {
								dkspan.Metrics[d] = attr.Value.GetStringValue()
							}
						}
						break
					}
				}

				attrtags, attrfields := attrs.splite()
				dkspan.Tags = itrace.MergeTags(tags, dkspan.Tags, attrtags)
				dkspan.Metrics = itrace.MergeFields(dkspan.Metrics, attrfields)

				dkspan.SourceType = getSourceType(dkspan.Tags)

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

func getSpanIDsAndParentIDs(resspans []*trace.ResourceSpans) (map[string]bool, map[string]bool) {
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
func getDKSpanStatus(statuspb *trace.Status) string {
	status := itrace.STATUS_INFO
	if statuspb == nil {
		return status
	}
	switch statuspb.Code {
	case trace.Status_STATUS_CODE_OK, trace.Status_STATUS_CODE_UNSET:
		status = itrace.STATUS_OK
	case trace.Status_STATUS_CODE_ERROR:
		status = itrace.STATUS_ERR
	default:
	}

	return status
}

func getSourceType(tags map[string]string) string {
	for key := range tags {
		switch key {
		case otelHTTPSchemeKey, otelHTTPMethodKey, otelRPCSystemKey:
			return itrace.SPAN_SOURCE_WEB
		case otelDBSystemKey:
			return itrace.SPAN_SOURCE_DB
		case otelMessagingSystemKey:
			return itrace.SPAN_SOURCE_MSGQUE
		}
	}

	return itrace.SPAN_SOURCE_CUSTOMER
}
