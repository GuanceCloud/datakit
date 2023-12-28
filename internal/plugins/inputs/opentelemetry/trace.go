// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"time"

	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/GuanceCloud/cliutils/point"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var traceOpts = []point.Option{}

func parseResourceSpans(resspans []*trace.ResourceSpans) itrace.DatakitTraces {
	var dktraces itrace.DatakitTraces
	spanIDs, parentIDs := getSpanIDsAndParentIDs(resspans)
	for _, spans := range resspans {
		log.Debugf("otel span: %s", spans.String())
		serviceName := "unknown_service"
		for _, v := range spans.Resource.Attributes {
			if v.Key == otelResourceServiceKey {
				serviceName = v.Value.GetStringValue()
			}
		}

		var dktrace itrace.DatakitTrace
		for _, scopeSpans := range spans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				var spanKV point.KVs
				spanKV = spanKV.Add(itrace.FieldTraceID, convert(span.GetTraceId()), false, false).
					Add(itrace.FieldParentID, convert(span.GetParentSpanId()), false, false).
					Add(itrace.FieldSpanid, convert(span.GetSpanId()), false, false).
					Add(itrace.FieldResource, span.Name, false, false).
					AddTag(itrace.TagOperation, span.Name).
					AddTag(itrace.TagSource, inputName).
					AddTag(itrace.TagService, serviceName).
					Add(itrace.FieldStart, int64(span.StartTimeUnixNano)/int64(time.Microsecond), false, false).
					Add(itrace.FieldDuration, int64(span.EndTimeUnixNano-span.StartTimeUnixNano)/int64(time.Microsecond), false, false).
					AddTag(itrace.TagSpanStatus, getDKSpanStatus(span.GetStatus())).
					AddTag(itrace.TagSpanType,
						itrace.FindSpanTypeStrSpanID(convert(span.GetSpanId()), convert(span.GetParentSpanId()), spanIDs, parentIDs))
				for k, v := range tags { // span.attribute 优先级大于全局tag。
					spanKV = spanKV.MustAddTag(k, v)
				}

				for i := range span.Events {
					if span.Events[i].Name == ExceptionEventName {
						for o, d := range otelErrKeyToDkErrKey {
							if attr, ok := getAttribute(o, span.Events[i].Attributes); ok {
								spanKV = spanKV.Add(d, attr.Value.GetStringValue(), false, false)
							}
						}
						break
					}
				}
				if kind, ok := spanKinds[int32(span.GetKind())]; ok {
					spanKV = spanKV.AddTag("span_kind", kind)
				}
				attrs := make([]*common.KeyValue, 0)
				spanKV, attrs = attributesToKVS(spanKV, attrs, spans.Resource.GetAttributes())
				spanKV, attrs = attributesToKVS(spanKV, attrs, scopeSpans.Scope.GetAttributes())
				spanKV, attrs = attributesToKVS(spanKV, attrs, span.GetAttributes())

				span.Attributes = attrs
				spanKV = spanKV.AddTag(itrace.TagSourceType, getSourceType(spanKV.Tags()))
				if !delMessage {
					if buf, err := protojson.Marshal(span); err != nil {
						log.Warn(err.Error())
					} else {
						spanKV = spanKV.Add(itrace.FieldMessage, string(buf), false, false)
					}
				}

				pt := point.NewPointV2(inputName, spanKV, traceOpts...)
				dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
			}
		}
		if len(dktrace) != 0 {
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
				spanIDs[convert(span.SpanId)] = true
				parentIDs[convert(span.ParentSpanId)] = true
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

func convert(id []byte) string {
	switch {
	case convertToDD:
		if len(id) >= 8 {
			bts := id[len(id)-8:]
			num := binary.BigEndian.Uint64(bts[:8])
			return strconv.FormatUint(num, 10)
		} else {
			log.Debugf("traceid or spanid is %s ,can not convert to [8]byte", string(id))
			return "0"
		}
	case convertToZhaoShang:
		if len(id) > 8 {
			return string(id)
		} else {
			return byteToString(id)
		}
	default:
		return byteToString(id)
	}
}

// getDKSpanStatus 从otel的status转成dk的span_status.
func getDKSpanStatus(statuspb *trace.Status) string {
	status := itrace.StatusInfo
	if statuspb == nil {
		return status
	}
	switch statuspb.Code {
	case trace.Status_STATUS_CODE_OK, trace.Status_STATUS_CODE_UNSET:
		status = itrace.StatusOk
	case trace.Status_STATUS_CODE_ERROR:
		status = itrace.StatusErr
	default:
	}

	return status
}

func getSourceType(tags point.KVs) string {
	for _, v := range tags {
		switch v.Key {
		case otelHTTPSchemeKey, otelHTTPMethodKey, otelRPCSystemKey:
			return itrace.SpanSourceWeb
		case otelDBSystemKey:
			return itrace.SpanSourceDb
		case otelMessagingSystemKey:
			return itrace.SpanSourceMsgque
		}
	}

	return itrace.SpanSourceCustomer
}
