// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc/metadata"
)

type emptyTransaction struct {
	sync.Mutex
	now int64
	seq int64
}

func (uaid *emptyTransaction) NextID() string {
	uaid.Lock()
	defer uaid.Unlock()

	ts := time.Now().UnixMilli()
	if ts == uaid.now {
		uaid.seq = (uaid.seq + 1) & math.MaxInt64
		if uaid.seq == 0 {
			for ts == uaid.now {
				ts = time.Now().UnixMilli()
			}
		}
	}
	uaid.now = ts

	return fmt.Sprintf("empty-transid^%d^%d", uaid.now, uaid.seq)
}

type grpcMeta metadata.MD

func (md grpcMeta) Get(key string) string {
	for k, v := range md {
		if k == key && len(v) != 0 {
			return v[0]
		}
	}

	return ""
}

func ConvertPSpanToDKTrace(x *ppv1.PSpan, meta metadata.MD) {
	log.Debugf("ConvertPSpanToDKTrace x=%+v", x)

	root := creatRootSpan(x, meta)
	// log.Debugf("root span=%s", root.LineProto())
	trace := itrace.DatakitTrace{root}
	if len(x.SpanEvent) != 0 {
		etrace := processEvents(root, x.SpanEvent, meta)
		trace = append(trace, etrace...)
	}

	if spanSender != nil {
		spanSender.Append(trace...)
	}
}

func ConvertPSpanChunkToDKTrace(x *ppv1.PSpanChunk, meta metadata.MD) {
	log.Debugf("ConvertPSpanChunkToDKTrace x=%+v", x)
	traceID := getTraceID(x.TransactionId, meta)
	if len(x.GetSpanEvent()) > 2 {
		startTime := x.SpanEvent[0].StartElapsed
		for i := 1; i < len(x.SpanEvent)-1; i++ {
			startTime += x.SpanEvent[i].StartElapsed
			x.SpanEvent[i].StartElapsed = startTime
		}
	}

	// agentCache.SetEvent(traceID, x.SpanEvent, 0)
	agentCache.SetSpanChunk(traceID, x, 0)
}

var emptyTrans = emptyTransaction{}

func getTraceID(transid *ppv1.PTransactionId, meta metadata.MD) string {
	if transid != nil {
		if transid.AgentId == "" {
			if agid := grpcMeta(meta).Get("agentid"); agid != "" {
				transid.AgentId = agid
			} else if agid = grpcMeta(meta).Get("applicationname"); agid != "" {
				transid.AgentId = agid
			} else {
				transid.AgentId = "unknown-pp-agent"
			}
		}

		return fmt.Sprintf("%s^%d^%d", transid.AgentId, transid.AgentStartTime, transid.Sequence)
	}

	log.Debugf("### empty transaction id %s", transid.String())

	return emptyTrans.NextID()
}

func creatRootSpan(pspan *ppv1.PSpan, meta metadata.MD) *itrace.DkSpan {
	spanKV := point.KVs{}
	spanKV = spanKV.Add(itrace.FieldTraceID, getTraceID(pspan.TransactionId, meta), false, false).
		Add(itrace.FieldSpanid, strconv.FormatInt(pspan.SpanId, 10), false, false).
		AddTag(itrace.TagService, grpcMeta(meta).Get("applicationname")).
		AddTag(itrace.TagSource, "pinpointV2").
		AddTag(itrace.TagSpanType, itrace.SpanTypeEntry).
		AddTag(itrace.TagSourceType, getServiceType(pspan.ServiceType)).
		Add(itrace.FieldStart, pspan.StartTime*int64(time.Microsecond), false, false).
		Add(itrace.FieldDuration, int64(pspan.Elapsed)*int64(time.Microsecond), false, false)

	if pspan.ParentSpanId == -1 {
		spanKV = spanKV.Add(itrace.FieldParentID, "0", false, false)
	} else {
		spanKV = spanKV.Add(itrace.FieldParentID, strconv.FormatInt(pspan.ParentSpanId, 10), false, false)
	}
	if pspan.AcceptEvent != nil {
		spanKV = spanKV.Add(itrace.FieldResource, pspan.AcceptEvent.Rpc, false, true).
			Add(itrace.TagOperation, pspan.AcceptEvent.EndPoint, true, true)
	}

	if pspan.Err != 0 {
		spanKV = spanKV.Add(itrace.TagSpanStatus, itrace.StatusErr, true, true).
			Add(itrace.FieldErrMessage, pspan.ExceptionInfo.String(), false, false)
	} else {
		spanKV = spanKV.Add(itrace.TagSpanStatus, itrace.StatusOk, true, true)
	}
	for k, v := range tags {
		spanKV = spanKV.AddTag(k, v)
	}
	for _, anno := range pspan.Annotation {
		spanKV = spanKV.AddTag(getAnnotationKey(anno.Key), anno.Value.String())
	}

	if bts, err := json.Marshal(pspan); err == nil {
		spanKV = spanKV.Add(itrace.FieldMessage, string(bts), false, false)
	}
	if vals := meta.Get("agentid"); len(vals) > 0 {
		info := agentCache.GetAgentInfo(vals[0])
		if info != nil {
			maps := fromAgentTag(info)
			for k, v := range maps {
				// root.Tags[k] = v
				spanKV = spanKV.AddTag(k, v)
			}
		}
	}

	return &itrace.DkSpan{Point: point.NewPointV2("pinpointV2", spanKV, traceOpts...)}
}

func fromAgentTag(agentInfo *ppv1.PAgentInfo) map[string]string {
	infoTags := make(map[string]string)
	if v := agentInfo.GetHostname(); v != "" {
		infoTags["hostname"] = v
	}
	if v := agentInfo.GetIp(); v != "" {
		infoTags["ip"] = v
	}
	if v := agentInfo.GetPid(); v != 0 {
		infoTags["pid"] = strconv.Itoa(int(v))
	}
	if v := agentInfo.GetPorts(); v != "" {
		infoTags["ports"] = v
	}

	infoTags["container"] = strconv.FormatBool(agentInfo.GetContainer())

	if v := agentInfo.GetAgentVersion(); v != "" {
		infoTags["agentVersion"] = v
	}

	return infoTags
}

func linkSpan(parentID string, nextSpanID int64) []*itrace.DkSpan {
	dktrace := make([]*itrace.DkSpan, 0)
	spanItem, ok := agentCache.GetSpan(nextSpanID)
	if !ok {
		log.Debugf("can not get spanID=%d from cache", nextSpanID)
		return dktrace
	}

	root := creatRootSpan(spanItem.Span, spanItem.Meta)
	// 关键一步，重置 parentID， 才能让两个 span 关联起来。
	// root.ParentID = parentID
	root.MustAdd(itrace.FieldParentID, parentID)
	// log.Debugf("link span root=%s", root.LineProto())
	dktrace = append(dktrace, root)
	if len(spanItem.Span.SpanEvent) > 0 {
		etrace := processEvents(root, spanItem.Span.SpanEvent, spanItem.Meta)

		dktrace = append(dktrace, etrace...)
	}

	return dktrace
}
