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
