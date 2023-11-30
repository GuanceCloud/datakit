// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc/metadata"
)

func processEvents(root *itrace.DkSpan, events []*ppv1.PSpanEvent, meta metadata.MD) itrace.DatakitTrace {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorf("panic err=%v", err)
		}
	}()

	if root == nil || len(events) == 0 {
		return nil
	}

	var (
		trace     itrace.DatakitTrace
		linkTrace itrace.DatakitTrace
		parentID  = root.GetFiledToString(itrace.FieldSpanid)
		startTime = root.GetFiledToInt64(itrace.FieldStart)
		traceID   = root.GetFiledToString(itrace.FieldTraceID)
		agentID   = ""
	)
	if vals := meta.Get("agentid"); len(vals) > 0 {
		agentID = vals[0]
	}

	agentCache.SetEvent(traceID, events, 0)
	eventsCache, _ := agentCache.GetEvent(traceID, startTime)
	rand.Seed(time.Now().UnixNano())
	for _, event := range eventsCache {
		// 先获取 parentID
		parentID = parentIDFromDepth(event, parentID, trace)

		spanKV := point.KVs{}
		spanID := strconv.FormatInt(rand.Int63(), 10) //nolint
		spanKV = spanKV.Add(itrace.FieldTraceID, traceID, false, false).
			Add(itrace.FieldSpanid, spanID, false, false).
			Add(itrace.FieldParentID, parentID, false, false).
			AddTag(itrace.TagService, getServiceType(event.ServiceType)).
			AddTag(itrace.TagSource, "pinpointV2").
			AddTag(itrace.TagSpanType, itrace.SpanTypeLocal).
			AddTag(itrace.TagSourceType, itrace.SpanSourceCustomer).
			Add(itrace.FieldStart, startTime+(int64(event.StartElapsed)*int64(time.Microsecond)), false, false).
			Add(itrace.FieldDuration, int64(event.EndElapsed)*int64(time.Microsecond), false, false).
			AddTag("sequence", strconv.Itoa(int(event.Sequence))).
			AddTag("depth", strconv.Itoa(int(event.Depth)))

		if event.ExceptionInfo != nil {
			spanKV = spanKV.Add(itrace.TagSpanStatus, itrace.StatusErr, true, true).
				Add(itrace.FieldErrMessage, event.ExceptionInfo.String(), false, false)
		} else {
			spanKV = spanKV.Add(itrace.TagSpanStatus, itrace.StatusOk, true, true)
		}

		res, opt, _ := agentCache.FindAPIInfo(agentID, event.ApiId)

		for _, anno := range event.Annotation { // todo 遇见http和sql 应当替换掉 resource
			if anno.Key == SQLAnnoKey {
				if val := anno.Value.GetIntStringStringValue(); val != nil {
					res, opt, _ = agentCache.FindSQLInfo(agentID, val.IntValue)
				}
			}
			spanKV = spanKV.AddTag(getAnnotationKey(anno.Key), anno.Value.String())
		}

		if res == "" {
			spanKV = spanKV.Add(itrace.FieldResource, strconv.Itoa(int(event.ApiId)), false, false)
		} else {
			spanKV = spanKV.Add(itrace.FieldResource, res, false, false)
		}
		spanKV = spanKV.AddTag(itrace.TagOperation, opt)

		info := agentCache.GetAgentInfo(agentID)
		if info != nil {
			infoTags := fromAgentTag(info)
			for k, v := range infoTags {
				spanKV = spanKV.AddTag(k, v)
			}
		}

		if event.NextEvent != nil {
			spanKV = spanKV.AddTag(itrace.TagSpanType, itrace.SpanTypeExit).
				AddTag("endpoint", event.NextEvent.GetMessageEvent().GetEndPoint()).
				AddTag("destination_id", event.NextEvent.GetMessageEvent().GetDestinationId())
			if msgEvnet := event.NextEvent.GetMessageEvent(); msgEvnet != nil {
				nextSpanID := msgEvnet.GetNextSpanId()
				if nextSpanID != 0 {
					// linkSpan：重要的一步，将关联的其他span追加到后面。
					log.Debugf("link Span ID=%s nextSpanID=%s", spanID, nextSpanID)
					linkTrace = append(linkTrace, linkSpan(spanID, nextSpanID)...)
				}
			}
		}

		if bts, err := json.Marshal(event); err == nil {
			spanKV = spanKV.Add(itrace.FieldMessage, string(bts), false, false)
		}
		pt := point.NewPointV2("pinpointV2", spanKV, point.DefaultLoggingOptions()...)
		trace = append(trace, &itrace.DkSpan{Point: pt})
	}
	// 最后将 linkTrace 放进去，防止上面 for 循环过程中按照 depth 取出错误的Event。
	trace = append(trace, linkTrace...)
	return trace
}

func parentIDFromDepth(event *ppv1.PSpanEvent, parentID string, trace itrace.DatakitTrace) string {
	switch event.Depth {
	case 1:
		return parentID
	case 0:
		for j := len(trace) - 1; j >= 0; j-- {
			if depth := trace[j].GetTag("depth"); depth == "0" {
				return trace[j].GetFiledToString(itrace.FieldSpanid)
			}
		}
	default:
		for j := len(trace) - 1; j >= 0; j-- {
			if depth := trace[j].GetTag("depth"); depth == strconv.Itoa(int(event.Depth)-1) {
				return trace[j].GetFiledToString(itrace.FieldSpanid)
			}
		}
	}
	return parentID
}
