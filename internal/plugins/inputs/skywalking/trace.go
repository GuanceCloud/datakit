// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package skywalking

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	commonv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/common/v3"
	agentv3 "github.com/GuanceCloud/tracing-protos/skywalking-gen-go/language/agent/v3"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var traceOpts = []point.Option{}

func parseSegmentObjectV3(segment *agentv3.SegmentObject) itrace.DatakitTrace {
	var dktrace itrace.DatakitTrace
	for _, span := range segment.Spans {
		if span == nil {
			continue
		}

		spanKV := point.KVs{}
		spanKV = spanKV.Add(itrace.FieldTraceID, segment.TraceId, false, false).
			Add(itrace.FieldSpanid, fmt.Sprintf("%s%d", segment.TraceSegmentId, span.SpanId), false, false).
			AddTag(itrace.TagService, segment.Service).
			Add(itrace.FieldResource, span.OperationName, false, false).
			AddTag(itrace.TagOperation, span.OperationName).
			AddTag(itrace.TagSource, inputName).
			AddTag(itrace.TagSourceType, itrace.SpanSourceCustomer).
			Add(itrace.FieldStart, span.StartTime*int64(time.Microsecond), false, false).
			Add(itrace.FieldDuration, (span.EndTime-span.StartTime)*int64(time.Microsecond), false, false)

		if span.ParentSpanId < 0 {
			if len(span.Refs) > 0 {
				spanKV = spanKV.Add(itrace.FieldParentID,
					fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId, span.Refs[0].ParentSpanId), false, false)

				if span.Refs[0].RefType == agentv3.RefType_CrossProcess && strings.Contains(span.Refs[0].ParentService, "_rum_") {
					childSpanKV := point.KVs{}
					childSpanKV = childSpanKV.
						Add(itrace.FieldTraceID, segment.TraceId, false, false).
						Add(itrace.FieldParentID, "0", false, false).
						Add(itrace.FieldSpanid, fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId, span.Refs[0].ParentSpanId), false, false).
						AddTag(itrace.TagService, span.Refs[0].ParentService).
						Add(itrace.FieldResource, span.Refs[0].ParentService, false, false).
						AddTag(itrace.TagOperation, span.Refs[0].ParentService).
						AddTag(itrace.TagSource, inputName).
						AddTag(itrace.TagSpanType, itrace.SpanTypeEntry).
						AddTag(itrace.TagSourceType, itrace.SpanSourceWeb).
						Add(itrace.FieldStart, span.StartTime*int64(time.Microsecond)-int64(time.Microsecond), false, false).
						Add(itrace.FieldDuration, int64(time.Microsecond), false, false)

					childSpan := point.NewPointV2(inputName, childSpanKV, point.DefaultLoggingOptions()...)
					dktrace = append(dktrace, &itrace.DkSpan{Point: childSpan})

					if endpoint := span.Refs[0].GetNetworkAddressUsedAtPeer(); endpoint != "" {
						spanKV = spanKV.AddTag(itrace.TagEndpoint, endpoint)
					}
				}
			} else {
				spanKV = spanKV.Add(itrace.FieldParentID, "0", false, false)
			}
		} else {
			if len(span.Refs) > 0 {
				spanKV = spanKV.Add(itrace.FieldParentID, fmt.Sprintf("%s%d", span.Refs[0].ParentTraceSegmentId, span.Refs[0].ParentSpanId), false, false)
			} else {
				spanKV = spanKV.Add(itrace.FieldParentID, fmt.Sprintf("%s%d", segment.TraceSegmentId, span.ParentSpanId), false, false)
			}
		}

		if span.IsError {
			spanKV = spanKV.AddTag(itrace.TagSpanStatus, itrace.StatusErr)
		} else {
			spanKV = spanKV.AddTag(itrace.TagSpanStatus, itrace.StatusOk)
		}

		switch span.SpanType {
		case agentv3.SpanType_Entry:
			spanKV = spanKV.AddTag(itrace.TagSpanType, itrace.SpanTypeEntry)
		case agentv3.SpanType_Local:
			spanKV = spanKV.AddTag(itrace.TagSpanType, itrace.SpanTypeLocal)
		case agentv3.SpanType_Exit:
			spanKV = spanKV.AddTag(itrace.TagSpanType, itrace.SpanTypeExit)
		default:
			spanKV = spanKV.AddTag(itrace.TagSpanType, itrace.SpanTypeEntry)
		}

		for i := range plugins {
			if value, ok := getTagValue(span.Tags, plugins[i]); ok {
				spanKV = spanKV.MustAddTag(itrace.TagService, value).
					MustAddTag(itrace.TagSpanType, itrace.SpanTypeEntry).
					MustAddTag(itrace.TagSourceType, mapToSpanSourceType(span.SpanLayer))
				switch span.SpanLayer { // nolint: exhaustive
				case agentv3.SpanLayer_Database, agentv3.SpanLayer_Cache:
					if res, ok := getTagValue(span.Tags, "db.statement"); ok {
						spanKV = spanKV.Add(itrace.FieldResource, res, false, true)
					}
				case agentv3.SpanLayer_MQ:
				case agentv3.SpanLayer_Http:
				case agentv3.SpanLayer_RPCFramework:
				case agentv3.SpanLayer_FAAS:
				case agentv3.SpanLayer_Unknown:
				}
			}
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.Tags {
			sourceTags[tag.Key] = tag.Value
		}

		mTags, err := itrace.MergeInToCustomerTags(tags, sourceTags, ignoreTags)
		if err == nil {
			for k, v := range mTags {
				spanKV = spanKV.AddTag(k, v)
			}
		}

		if span.Peer != "" {
			spanKV = spanKV.AddTag(itrace.TagEndpoint, span.Peer)
		}
		if !delMessage {
			if buf, err := json.Marshal(span); err != nil {
				log.Warn(err.Error())
			} else {
				spanKV = spanKV.Add(itrace.FieldMessage, string(buf), false, false)
			}
		}

		pt := point.NewPointV2(inputName, spanKV, traceOpts...)
		dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
	}
	if len(dktrace) != 0 {
		dktrace[0].MustAdd(itrace.FieldPriority, itrace.PriorityAutoKeep)
	}

	return dktrace
}

func getTagValue(tags []*commonv3.KeyStringValuePair, key string) (value string, ok bool) {
	for i := range tags {
		if key == tags[i].Key {
			if len(tags[i].Value) == 0 {
				return "", false
			} else {
				return tags[i].Value, true
			}
		}
	}

	return "", false
}

func mapToSpanSourceType(layer agentv3.SpanLayer) string {
	switch layer {
	case agentv3.SpanLayer_Database:
		return itrace.SpanSourceDb
	case agentv3.SpanLayer_Cache:
		return itrace.SpanSourceCache
	case agentv3.SpanLayer_RPCFramework:
		return itrace.SpanSourceFramework
	case agentv3.SpanLayer_Http:
		return itrace.SpanSourceWeb
	case agentv3.SpanLayer_MQ:
		return itrace.SpanSourceMsgque
	case agentv3.SpanLayer_FAAS:
		return itrace.SpanSourceApp
	case agentv3.SpanLayer_Unknown:
		return itrace.SpanSourceCustomer
	default:
		return itrace.SpanSourceCustomer
	}
}
