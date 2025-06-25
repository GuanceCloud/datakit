// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/GuanceCloud/cliutils/point"
	common "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/common/v1"
	trace "github.com/GuanceCloud/tracing-protos/opentelemetry-gen-go/trace/v1"
	jsoniter "github.com/json-iterator/go"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
)

type protojsonMarshaler struct{}

func (j *protojsonMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return protojson.Marshal(x)
}

type gojsonMarshaler struct{}

func (j *gojsonMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return json.Marshal(x)
}

func initJSONIter() jsoniter.API {
	customExtension := jsoniter.DecoderExtension{}

	// 注册所有OTel关键类型的序列化优化
	json := jsoniter.Config{
		EscapeHTML:                    false,
		TagKey:                        "json",
		OnlyTaggedField:               false,
		SortMapKeys:                   false,
		IndentionStep:                 0,
		ValidateJsonRawMessage:        true,
		ObjectFieldMustBeSimpleString: true,
	}.Froze()

	json.RegisterExtension(customExtension)
	return json
}

var otelJSON = initJSONIter()

type jsoniterMarshaler struct{}

func (m *jsoniterMarshaler) Marshal(x proto.Message) ([]byte, error) {
	return otelJSON.Marshal(x)
}

func (ipt *Input) parseResourceSpans(resspans []*trace.ResourceSpans) itrace.DatakitTraces {
	var (
		dktraces           itrace.DatakitTraces
		spanIDs, parentIDs = ipt.getSpanIDsAndParentIDs(resspans)
	)

	for _, spans := range resspans {
		var (
			serviceName          = "unknown_service"
			runtimeID            = ""
			runtimeIDInitialized = false
			dktrace              itrace.DatakitTrace
		)

		for _, v := range spans.Resource.Attributes {
			if v.Key == otelResourceServiceKey {
				serviceName = v.Value.GetStringValue()
			} else if v.Key == itrace.FieldRuntimeID {
				runtimeID = v.Value.GetStringValue()
			}
		}

		for _, scopeSpans := range spans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				var (
					spanKV  point.KVs
					attrs   = make([]*common.KeyValue, 0)
					spanID  = ipt.convertBinID(span.GetSpanId())
					parenID = ipt.convertBinID(span.GetParentSpanId())
					traceID = ipt.convertBinID(span.GetTraceId())
				)

				spanKV = spanKV.Add(itrace.FieldTraceID, traceID, false, false).
					Add(itrace.FieldParentID, parenID, false, false).
					Add(itrace.FieldSpanid, spanID, false, false).
					Add(itrace.FieldResource, span.Name, false, false).
					AddTag(itrace.TagOperation, span.Name).
					AddTag(itrace.TagSource, inputName).
					AddTag(itrace.TagService, serviceName).
					Add(itrace.FieldStart, int64(span.StartTimeUnixNano)/int64(time.Microsecond), false, false).
					Add(itrace.FieldDuration, int64(span.EndTimeUnixNano-span.StartTimeUnixNano)/int64(time.Microsecond), false, false).
					AddTag(itrace.TagSpanStatus, getDKSpanStatus(span.GetStatus())).
					AddTag(itrace.TagSpanType,
						itrace.FindSpanTypeStrSpanID(spanID, parenID, spanIDs, parentIDs)).
					AddTag(itrace.TagDKFingerprintKey, datakit.DatakitHostName+"_"+datakit.Version)

				// service_name from xx.system.
				if ipt.SpiltServiceName {
					spanKV = spanKV.MustAddTag(itrace.TagService, getServiceNameBySystem(span.GetAttributes(), serviceName)).
						AddTag(itrace.TagBaseService, serviceName)
				}

				for k, v := range ipt.Tags { // span.attribute 优先级大于全局tag。
					spanKV = spanKV.MustAddTag(k, v)
				}

				if runtimeID == "" && !runtimeIDInitialized {
					if attrRuntimeID, ok := getAttr(itrace.FieldRuntimeID, span.Attributes); ok {
						runtimeID = attrRuntimeID.Value.GetStringValue()
					}
					runtimeIDInitialized = true
				}
				if runtimeID != "" {
					spanKV = spanKV.AddV2(ext.RuntimeID, runtimeID, true).AddV2(itrace.FieldRuntimeID, runtimeID, true)
				}

				for i := range span.Events {
					if span.Events[i].Name == ExceptionEventName {
						for o, d := range otelErrKeyToDkErrKey {
							if attr, ok := getAttr(o, span.Events[i].Attributes); ok {
								spanKV = spanKV.Add(d, attr.Value.GetStringValue(), false, false)
							}
						}
						break
					}
				}

				if kind, ok := spanKinds[int32(span.GetKind())]; ok {
					spanKV = spanKV.AddTag("span_kind", kind)
				}

				for _, x := range [][]*common.KeyValue{
					spans.Resource.GetAttributes(),
					scopeSpans.Scope.GetAttributes(),
					span.GetAttributes(),
				} {
					newkv, newattr := ipt.attributesToKVS(x)
					log.Debugf("newkv: %d, newattr: %d", len(newkv), len(newattr))
					spanKV = append(spanKV, newkv...)
					attrs = append(attrs, newattr...)
				}

				if len(span.GetEvents()) > 0 {
					for _, event := range span.GetEvents() {
						newkv, newattr := ipt.attributesToKVS(event.GetAttributes())
						log.Debugf("newkv: %d, newattr: %d", len(newkv), len(newattr))
						spanKV = append(spanKV, newkv...)
						attrs = append(attrs, newattr...)
					}
				}

				if dbHost := getDBHost(span.GetAttributes()); dbHost != "" {
					spanKV = spanKV.AddTag("db_host", dbHost)
				}

				span.Attributes = attrs
				spanKV = spanKV.AddTag(itrace.TagSourceType, getSourceType(spanKV.Tags()))
				if !ipt.DelMessage {
					if buf, err := ipt.jmarshaler.Marshal(span); err != nil {
						log.Warn(err.Error())
					} else {
						spanKV = spanKV.Add(itrace.FieldMessage, string(buf), false, false)
					}
				}
				t := time.Unix(int64(span.StartTimeUnixNano)/1e9, int64(span.StartTimeUnixNano)%1e9)
				pt := point.NewPointV2(inputName, spanKV, append(ipt.ptsOpts, point.WithTime(t))...)
				dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
			}
		}

		if len(dktrace) != 0 {
			dktraces = append(dktraces, dktrace)
		}
	}

	return dktraces
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
