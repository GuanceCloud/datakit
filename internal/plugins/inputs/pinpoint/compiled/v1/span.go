// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package v1 complement for PSpan conversion.
package v1

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

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

func (x *PSpan) ConvertToDKTrace(inputName string, meta metadata.MD, apiMetaTab *sync.Map) itrace.DatakitTrace {
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
		root.Tags[fmt.Sprintf("pspan_annotation:%d", anno.Key)] = anno.Value.String()
	}
	if bts, err := json.Marshal(x); err == nil {
		root.Content = string(bts)
	}

	trace := &itrace.DatakitTrace{root}
	if len(x.SpanEvent) != 0 {
		sort.Sort(PSpanEventList(x.SpanEvent))
		expandSpanEvents(root.Start, root, 0, x.SpanEvent, apiMetaTab, trace)
	}

	return *trace
}

func (x *PSpanChunk) ConvertToDKTrace(inputName string, meta metadata.MD, apiMetaTab *sync.Map) itrace.DatakitTrace {
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
		Duration:   int64(PSpanEventList(x.SpanEvent).Duration()) * time.Hour.Milliseconds(),
	}

	if len(x.SpanEvent) != 0 {
		root.Service = getServiceType(x.SpanEvent[0].ServiceType)
		root.Resource, root.Operation = PSpanEventList(x.SpanEvent).APIInfo(apiMetaTab)
	}
	if bts, err := json.Marshal(x); err == nil {
		root.Content = string(bts)
	}

	trace := &itrace.DatakitTrace{root}
	if len(x.SpanEvent) != 0 {
		sort.Sort(PSpanEventList(x.SpanEvent))
		expandSpanEvents(root.Start, root, 0, x.SpanEvent, apiMetaTab, trace)
	}

	return *trace
}

type PSpanEventList []*PSpanEvent

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

func (x PSpanEventList) APIInfo(apiMetaTab *sync.Map) (res, opt string) {
	res, opt, _ = getAPIInfo(x[0].ApiId, apiMetaTab)

	return res, opt
}

func expandSpanEvents(rootStart int64, parent *itrace.DatakitSpan, i int, events []*PSpanEvent, apiMetaTab *sync.Map, trace *itrace.DatakitTrace) {
	if parent == nil || trace == nil || i >= len(events) {
		return
	}

	e := events[i]
	dkspan := &itrace.DatakitSpan{
		TraceID:    parent.TraceID,
		ParentID:   parent.SpanID,
		SpanID:     strconv.FormatInt(rand.Int63(), 10), // nolint: gosec
		Service:    getServiceType(e.ServiceType),
		Source:     parent.Source,
		SpanType:   itrace.SPAN_TYPE_LOCAL,
		SourceType: itrace.SPAN_SOURCE_CUSTOMER,
		Tags:       make(map[string]string),
		Metrics:    make(map[string]interface{}),
		Start:      rootStart + int64(e.StartElapsed)*int64(time.Millisecond),
		Duration:   int64(e.EndElapsed) * int64(time.Millisecond),
		Status:     itrace.STATUS_OK,
	}

	if dkspan.Duration == 0 {
		dkspan.Duration = int64(500 * time.Microsecond)
	}
	if e.NextEvent != nil {
		dkspan.SpanType = itrace.SPAN_TYPE_EXIT
	}
	if res, opt, ok := getAPIInfo(e.ApiId, apiMetaTab); ok {
		dkspan.Resource = res
		dkspan.Operation = opt
	}
	if e.ExceptionInfo != nil {
		dkspan.Status = itrace.STATUS_ERR
		dkspan.Metrics[itrace.FIELD_ERR_MESSAGE] = e.ExceptionInfo.String()
	}
	for _, anno := range e.Annotation {
		dkspan.Tags[fmt.Sprintf("pspan_event_annotation:%d", anno.Key)] = anno.Value.String()
	}
	if bts, err := json.Marshal(e); err == nil {
		dkspan.Content = string(bts)
	}

	*trace = append(*trace, dkspan)

	expandSpanEvents(rootStart, dkspan, i+1, events, apiMetaTab, trace)
}

func getTraceID(transid *PTransactionId, meta metadata.MD) string {
	if tid := grpcMeta(meta).Get("x-b3-traceid"); len(tid) != 0 {
		return tid
	}
	if transid != nil {
		return fmt.Sprintf("%s^%d^%d", transid.AgentId, transid.AgentStartTime, transid.Sequence)
	} else {
		return fmt.Sprintf("unknow-pinpoint-agent^%d^1", time.Now().UnixMilli())
	}
}

func NewMetaData(id int32, meta interface{}) *RequestMetaData {
	return &RequestMetaData{ID: id, Meta: meta}
}

type RequestMetaData struct {
	ID   int32
	Meta interface{}
}

func (md *RequestMetaData) GetSqlMetaData() (*PSqlMetaData, bool) { // nolint: stylecheck
	v, ok := md.Meta.(*PSqlMetaData)

	return v, ok
}

func (md *RequestMetaData) GetApiMetaData() (*PApiMetaData, bool) { // nolint: stylecheck
	v, ok := md.Meta.(*PApiMetaData)

	return v, ok
}

func (md *RequestMetaData) GetStringMetaData() (*PStringMetaData, bool) { // nolint: stylecheck
	v, ok := md.Meta.(*PStringMetaData)

	return v, ok
}

func getAPIInfo(apiID int32, apiMetaTab *sync.Map) (res, opt string, ok bool) {
	var meta interface{}
	if meta, ok = apiMetaTab.Load(apiID); ok {
		var md *RequestMetaData
		if md, ok = meta.(*RequestMetaData); ok {
			var amd *PApiMetaData
			if amd, ok = md.GetApiMetaData(); ok {
				opt = fmt.Sprintf("id:%d line:%d %s:%s", apiID, amd.Line, amd.Location, amd.ApiInfo)
				res = amd.ApiInfo
				ok = true

				return
			}
		}
	}

	return
}
