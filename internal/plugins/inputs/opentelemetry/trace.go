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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
)

func (ipt *Input) parseResourceSpans(resspans []*trace.ResourceSpans) itrace.DatakitTraces {
	var (
		dktraces           itrace.DatakitTraces
		spanIDs, parentIDs = ipt.getSpanIDsAndParentIDs(resspans)
	)

	// otel trace are 3 level:
	//   level-1: resource spans
	//   level-2: scope spans
	//   level-3: spans within a single scope
	for _, spans := range resspans {
		var (
			serviceName          = "unknown_service"
			runtimeIDInitialized = false

			dktrace   itrace.DatakitTrace
			runtimeID string
		)

		for idx, v := range spans.Resource.Attributes {
			if v == nil {
				continue
			}

			switch v.Key {
			case otelResourceServiceKey:
				serviceName = v.Value.GetStringValue()

				if ipt.CleanMessage {
					spans.Resource.Attributes[idx] = nil
				}

			case itrace.FieldRuntimeID:
				runtimeID = v.Value.GetStringValue()

				if ipt.CleanMessage {
					spans.Resource.Attributes[idx] = nil
				}

			default: // pass
			}
		}

		for _, scopeSpans := range spans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				var (
					spanKVs     point.KVs
					mergedAttrs = make([]*common.KeyValue, 0)
					spanID      = ipt.convertBinID(span.GetSpanId())
					parenID     = ipt.convertBinID(span.GetParentSpanId())
					traceID     = ipt.convertBinID(span.GetTraceId())
					spanAttrs   = span.GetAttributes()
					spanTime    = int64(span.StartTimeUnixNano)
				)

				// extra add converted trace-id to attrs for global text search.
				spanAttrs = append(spanAttrs, &common.KeyValue{
					Key:   itrace.FieldTraceID,
					Value: &common.AnyValue{Value: &common.AnyValue_StringValue{StringValue: traceID}},
				})

				// commom tags & fields.
				spanKVs = spanKVs.Add(itrace.FieldTraceID, traceID).
					Add(itrace.FieldParentID, parenID).
					Add(itrace.FieldSpanid, spanID).
					Add(itrace.FieldResource, span.Name).
					Add(itrace.FieldStart, int64(span.StartTimeUnixNano)/int64(time.Microsecond)).
					Add(itrace.FieldDuration, int64(span.EndTimeUnixNano-span.StartTimeUnixNano)/int64(time.Microsecond)).
					AddTag(itrace.TagSpanStatus, getDKSpanStatus(span.GetStatus())).
					AddTag(itrace.TagSpanType, itrace.FindSpanTypeStrSpanID(spanID, parenID, spanIDs, parentIDs)).
					AddTag(itrace.TagDKFingerprintKey, datakit.DKHost+"_"+datakit.Version).
					AddTag(itrace.TagOperation, span.Name).
					AddTag(itrace.TagSource, inputName).
					AddTag(itrace.TagService, serviceName)

				// service_name from xx.system.
				if ipt.SplitServiceName {
					baseService := ipt.getServiceNameBySystem(spanAttrs)
					if baseService != "" { // 只有存在中间件的时候才不为空。
						spanKVs = spanKVs.SetTag(itrace.TagService, baseService).AddTag(itrace.TagBaseService, serviceName)
					}
				}

				for k, v := range ipt.Tags { // span.attribute 优先级大于全局tag。
					spanKVs = spanKVs.SetTag(k, v)
				}

				if runtimeID == "" && !runtimeIDInitialized {
					if v, idx := getAttr(itrace.FieldRuntimeID, spanAttrs); v != nil {
						runtimeID = v.Value.GetStringValue()
						if ipt.CleanMessage {
							spanAttrs[idx] = nil
						}
					}
					runtimeIDInitialized = true
				}

				if runtimeID != "" {
					spanKVs = spanKVs.Set(ext.RuntimeID, runtimeID). // NOTE: ext.RuntimeID deprecated
												Set(itrace.FieldRuntimeID, runtimeID)
				}

				// extract exception event and related fields
				for i := range span.Events {
					if span.Events[i] == nil {
						continue
					}

					if span.Events[i].Name == ExceptionEventName {
						evtAttrs := span.Events[i].Attributes
						for key, alias := range otelExceptionAliasMap {
							if v, idx := getAttr(key, evtAttrs); v != nil {
								spanKVs = spanKVs.Set(alias, v.Value.GetStringValue())
								if ipt.CleanMessage {
									evtAttrs[idx] = nil
								}
							}
						}

						if ipt.CleanMessage {
							span.Events[i] = nil
						}
						break
					}
				}

				if kind, ok := spanKinds[int32(span.GetKind())]; ok {
					spanKVs = spanKVs.AddTag(itrace.TagSpanKind, kind)
				}

				// Extract selected attrs as tags, other attrs are merged into current span's attrs.
				//
				// NOTE:
				//   - all resource attrs(expect selected) are copied into all spans of current trace
				//   - all scoped attrs(except selected) are copied into all spans of current scope spans
				for _, x := range [][]*common.KeyValue{
					spans.Resource.GetAttributes(),
					scopeSpans.Scope.GetAttributes(),
					spanAttrs,
				} {
					newkv, newattr := ipt.selectAttrs(x)
					spanKVs = append(spanKVs, newkv...)
					mergedAttrs = append(mergedAttrs, newattr...)
				}

				if len(span.GetEvents()) > 0 {
					for _, event := range span.GetEvents() {
						newkv, newattr := ipt.selectAttrs(event.GetAttributes())
						spanKVs = append(spanKVs, newkv...)
						mergedAttrs = append(mergedAttrs, newattr...)
					}
				}

				if dbHost := getDBHost(spanAttrs); dbHost != "" {
					spanKVs = spanKVs.AddTag("db_host", dbHost)
				}

				span.Attributes = mergedAttrs

				spanKVs = spanKVs.AddTag(itrace.TagSourceType, getSourceType(spanKVs.Tags()))
				if !ipt.DelMessage {
					if ipt.CleanMessage {
						span = ipt.cleanSpan(span)
					}

					if buf, err := ipt.jmarshaler.Marshal(span); err != nil {
						log.Warn(err.Error())
					} else {
						spanKVs = spanKVs.Add(itrace.FieldMessage, string(buf))
					}
				}

				pt := point.NewPoint(inputName, spanKVs,
					append(ipt.ptsOpts, point.WithTimestamp(spanTime))...)
				dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
			}
		}

		if len(dktrace) != 0 {
			dktraces = append(dktraces, dktrace)
		}
	}

	return dktraces
}

// cleanSpan try to remove fields of the span and make the marshaled json smaller.
func (ipt *Input) cleanSpan(span *trace.Span) *trace.Span {
	span.TraceId = nil
	span.SpanId = nil
	span.ParentSpanId = nil
	span.Name = ""
	span.Kind = trace.Span_SPAN_KIND_UNSPECIFIED
	span.StartTimeUnixNano = 0
	span.EndTimeUnixNano = 0
	span.Status = nil

	return span
}

func (ipt *Input) getSpanIDsAndParentIDs(resspans []*trace.ResourceSpans) (map[string]bool, map[string]bool) {
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
				spanIDs[ipt.convertBinID(span.SpanId)] = true
				parentIDs[ipt.convertBinID(span.ParentSpanId)] = true
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

func (ipt *Input) convertBinID(id []byte) string {
	switch {
	case ipt.CompatibleDDTrace:
		if len(id) >= 8 {
			bts := id[len(id)-8:]
			num := binary.BigEndian.Uint64(bts[:8])
			return strconv.FormatUint(num, 10)
		} else {
			log.Debugf("traceid or spanid is %s ,can not convert to [8]byte", string(id))
			return "0"
		}

	case ipt.CompatibleZhaoShang:
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
