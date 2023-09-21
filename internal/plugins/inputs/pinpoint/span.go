// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pinpoint handle Pinpoint APM traces.
package pinpoint

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	ppv1 "github.com/GuanceCloud/tracing-protos/pinpoint-gen-go/v1"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"google.golang.org/grpc/metadata"
)

type grpcMeta metadata.MD

func (md grpcMeta) Get(key string) string {
	for k, v := range md {
		if k == key && len(v) != 0 {
			return v[0]
		}
	}

	return ""
}

func ConvertPSpanToDKTrace(inputName string, x *ppv1.PSpan, meta metadata.MD) itrace.DatakitTrace {
	root := &itrace.DatakitSpan{
		TraceID:    getTraceID(x.TransactionId, meta),
		SpanID:     strconv.FormatUint(uint64(x.SpanId), 10),
		Service:    grpcMeta(meta).Get("applicationname"),
		Source:     inputName,
		SpanType:   itrace.SPAN_TYPE_ENTRY,
		SourceType: getServiceType(x.ServiceType),
		Tags:       make(map[string]string),
		Metrics:    map[string]interface{}{itrace.FIELD_PRIORITY: itrace.PRIORITY_AUTO_KEEP},
		Start:      x.StartTime * int64(time.Millisecond),
		Duration:   int64(x.Elapsed) * int64(time.Millisecond),
		Status:     itrace.STATUS_OK,
	}

	if x.ParentSpanId == -1 {
		root.ParentID = "0"
	} else {
		root.ParentID = strconv.FormatInt(x.ParentSpanId, 10)
	}
	if x.AcceptEvent != nil {
		root.Resource = x.AcceptEvent.Rpc
		root.Operation = x.AcceptEvent.EndPoint
	}
	if x.Err != 0 {
		root.Status = itrace.STATUS_ERR
		root.Metrics[itrace.FIELD_ERR_MESSAGE] = x.ExceptionInfo.String()
	}
	for _, anno := range x.Annotation {
		root.Tags[getAnnotationKey(anno.Key)] = anno.Value.String()
	}
	if bts, err := json.Marshal(x); err == nil {
		root.Content = string(bts)
	}

	trace := itrace.DatakitTrace{root}
	if len(x.SpanEvent) != 0 {
		sort.Sort(PSpanEventList(x.SpanEvent))
		etrace := expandSpanEventsToDKTrace(root, x.SpanEvent)
		trace = append(trace, etrace...)
	}

	return trace
}

func ConvertPSpanChunkToDKTrace(inputName string, x *ppv1.PSpanChunk, meta metadata.MD) itrace.DatakitTrace {
	root := &itrace.DatakitSpan{
		TraceID:    getTraceID(x.TransactionId, meta),
		ParentID:   "0",
		SpanID:     strconv.FormatUint(uint64(x.SpanId), 10),
		Service:    grpcMeta(meta).Get("applicationname"),
		Source:     inputName,
		SpanType:   itrace.SPAN_TYPE_ENTRY,
		SourceType: getServiceType(x.ApplicationServiceType),
		Metrics:    map[string]interface{}{itrace.FIELD_PRIORITY: itrace.PRIORITY_AUTO_KEEP},
		Start:      x.KeyTime * int64(time.Millisecond),
		Duration:   int64(PSpanEventList(x.SpanEvent).Duration()) * int64(time.Millisecond),
	}

	if len(x.SpanEvent) != 0 {
		root.Service = getServiceType(x.SpanEvent[0].ServiceType)
		root.Resource, root.Operation = PSpanEventList(x.SpanEvent).APIInfo()
	}
	if bts, err := json.Marshal(x); err == nil {
		root.Content = string(bts)
	}

	trace := itrace.DatakitTrace{root}
	if len(x.SpanEvent) != 0 {
		sort.Sort(PSpanEventList(x.SpanEvent))
		etrace := expandSpanEventsToDKTrace(root, x.SpanEvent)
		trace = append(trace, etrace...)
	}

	return trace
}

type PSpanEventList []*ppv1.PSpanEvent

func (x PSpanEventList) Len() int { return len(x) }

func (x PSpanEventList) Less(i, j int) bool {
	return x[i].Depth < x[j].Depth && x[i].StartElapsed < x[j].StartElapsed
}

func (x PSpanEventList) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x PSpanEventList) Duration() int32 {
	var last int32
	for i := range x {
		if x[i].EndElapsed > last {
			last = x[i].EndElapsed
		}
	}

	return last
}

func (x PSpanEventList) APIInfo() (res, opt string) {
	res, opt, _ = findAPIInfo(x[0].ApiId)

	return res, opt
}

func expandSpanEventsToDKTrace(root *itrace.DatakitSpan, events PSpanEventList) itrace.DatakitTrace {
	if root == nil || len(events) == 0 {
		return nil
	}

	var (
		trace        itrace.DatakitTrace
		parentSpanID = root.SpanID
	)
	for _, e := range events {
		dkspan := &itrace.DatakitSpan{
			TraceID:    root.TraceID,
			ParentID:   parentSpanID,
			SpanID:     strconv.FormatInt(rand.Int63(), 10), // nolint: gosec
			Service:    getServiceType(e.ServiceType),
			Source:     root.Source,
			SpanType:   itrace.SPAN_TYPE_LOCAL,
			SourceType: itrace.SPAN_SOURCE_CUSTOMER,
			Tags:       make(map[string]string),
			Metrics:    make(map[string]interface{}),
			Start:      root.Start + int64(e.StartElapsed)*int64(time.Millisecond),
			Duration:   int64(e.EndElapsed) * int64(time.Millisecond),
			Status:     itrace.STATUS_OK,
		}
		parentSpanID = dkspan.SpanID

		if dkspan.Duration == 0 {
			dkspan.Duration = int64(500 * time.Microsecond)
		}
		if e.NextEvent != nil {
			dkspan.SpanType = itrace.SPAN_TYPE_EXIT
		}
		if res, opt, ok := findAPIInfo(e.ApiId); ok {
			dkspan.Resource = res
			dkspan.Operation = opt
		}
		if e.ExceptionInfo != nil {
			dkspan.Status = itrace.STATUS_ERR
			dkspan.Metrics[itrace.FIELD_ERR_MESSAGE] = e.ExceptionInfo.String()
		}
		for _, anno := range e.Annotation {
			if anno.Key == SQLAnnoKey {
				if val := anno.Value.GetIntStringStringValue(); val != nil {
					res, opt, ok := findAPIInfo(val.IntValue)
					if ok {
						dkspan.Resource = strings.ReplaceAll(res, "\n", " ")
						dkspan.Operation = opt
					}
				}
			}
			dkspan.Tags[getAnnotationKey(anno.Key)] = anno.Value.String()
		}
		if bts, err := json.Marshal(e); err == nil {
			dkspan.Content = string(bts)
		}

		trace = append(trace, dkspan)
	}

	return trace
}

func getTraceID(transid *ppv1.PTransactionId, meta metadata.MD) string {
	if transid != nil {
		return fmt.Sprintf("%s^%d^%d", transid.AgentId, transid.AgentStartTime, transid.Sequence)
	} else {
		return fmt.Sprintf("unknow-pinpoint-agent^%d^1", time.Now().UnixMilli())
	}
}
