// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	resp.WriteHeader(http.StatusOK)
}

func handleJaegerTrace(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### receiving trace data from path: %s", req.URL.Path)

	pbuf := bufpool.GetBuffer()
	defer bufpool.PutBuffer(pbuf)

	_, err := io.Copy(pbuf, req.Body)
	if err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	param := &itrace.TraceParameters{Body: pbuf}
	if err = parseJaegerTrace(param); err != nil {
		log.Errorf("### parse jaeger trace from HTTP failed: %s", err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	resp.WriteHeader(http.StatusOK)
}

func parseJaegerTrace(param *itrace.TraceParameters) error {
	tmbuf := thrift.NewTMemoryBuffer()
	_, err := tmbuf.ReadFrom(param.Body)
	if err != nil {
		return err
	}

	var (
		transport = thrift.NewTBinaryProtocolConf(tmbuf, &thrift.TConfiguration{})
		batch     = &jaeger.Batch{}
	)
	if err = batch.Read(context.TODO(), transport); err != nil {
		return err
	}

	if dktrace := batchToDkTrace(batch); len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace})
	}

	return nil
}

type DkJaegerSpan struct {
	TraceIdLow   uint64 `json:"traceIdLow"`   //nolint: stylecheck
	TraceIdHigh  uint64 `json:"traceIdHigh"`  //nolint: stylecheck
	SpanId       uint64 `json:"spanId"`       //nolint: stylecheck
	ParentSpanId uint64 `json:"parentSpanId"` //nolint: stylecheck
	*jaeger.Span
}

var traceOpts = []point.Option{}

func batchToDkTrace(batch *jaeger.Batch) itrace.DatakitTrace {
	var (
		project, version, env = getExpandInfo(batch)
		dktrace               itrace.DatakitTrace
		spanIDs, parentIDs    = gatherSpansInfo(batch.Spans)
	)
	for _, span := range batch.Spans {
		if span == nil {
			continue
		}

		spanKV := point.KVs{}
		spanKV = spanKV.Add(itrace.FieldParentID, strconv.FormatUint(uint64(span.ParentSpanId), 16)).
			Add(itrace.FieldSpanid, strconv.FormatUint(uint64(span.SpanId), 16)).
			AddTag(itrace.TagService, batch.Process.ServiceName).
			Add(itrace.FieldResource, span.OperationName).
			AddTag(itrace.TagOperation, span.OperationName).
			AddTag(itrace.TagSource, inputName).
			AddTag(itrace.TagSourceType, itrace.SpanSourceCustomer).
			AddTag(itrace.TagSpanType, itrace.FindSpanTypeIntSpanID(uint64(span.SpanId), uint64(span.ParentSpanId), spanIDs, parentIDs)).
			Add(itrace.FieldStart, span.StartTime).
			Add(itrace.FieldDuration, span.Duration)

		if span.TraceIdHigh != 0 {
			spanKV = spanKV.Add(itrace.FieldTraceID, toTraceIDString(uint64(span.TraceIdHigh), uint64(span.TraceIdLow)))
		} else {
			spanKV = spanKV.Add(itrace.FieldTraceID, strconv.FormatUint(uint64(span.TraceIdLow), 16))
		}

		spanKV = spanKV.AddTag(itrace.TagSpanStatus, itrace.StatusOk)
		for _, tag := range span.Tags {
			if tag.Key == "error" {
				spanKV = spanKV.SetTag(itrace.TagSpanStatus, itrace.StatusErr)
				break
			}
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.Tags {
			t := tag.GetVType()
			switch t {
			case jaeger.TagType_STRING:
				sourceTags[tag.Key] = tag.GetVStr()
			case jaeger.TagType_DOUBLE:
				sourceTags[tag.Key] = strconv.FormatFloat(tag.GetVDouble(), 'f', -1, 64)
			case jaeger.TagType_BOOL:
				sourceTags[tag.Key] = strconv.FormatBool(tag.GetVBool())
			case jaeger.TagType_LONG:
				sourceTags[tag.Key] = strconv.FormatInt(tag.GetVLong(), 10)
			case jaeger.TagType_BINARY:
				sourceTags[tag.Key] = string(tag.GetVBinary())
			}
		}

		if mTags, err := itrace.MergeInToCustomerTags(tags, sourceTags, ignoreTags); err == nil {
			for k, v := range mTags {
				spanKV = spanKV.AddTag(k, v)
			}
		}
		if project != "" {
			spanKV = spanKV.AddTag(itrace.Project, project)
		}
		if version != "" {
			spanKV = spanKV.AddTag(itrace.TagVersion, version)
		}
		if env != "" {
			spanKV = spanKV.AddTag(itrace.TagEnv, env)
		}

		dkJSpan := &DkJaegerSpan{
			TraceIdLow:   uint64(span.TraceIdLow),
			TraceIdHigh:  uint64(span.TraceIdHigh),
			SpanId:       uint64(span.SpanId),
			ParentSpanId: uint64(span.ParentSpanId),
			Span:         span,
		}
		if !delMessage {
			if buf, err := json.Marshal(dkJSpan); err != nil {
				log.Warn(err.Error())
			} else {
				spanKV = spanKV.Add(itrace.FieldMessage, string(buf))
			}
		}

		t := time.UnixMicro(span.StartTime)
		pt := point.NewPoint(inputName, spanKV, append(traceOpts, point.WithTime(t))...)
		dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
	}

	return dktrace
}

func toTraceIDString(traceIDHigh, traceIDLow uint64) string {
	highStr := strconv.FormatUint(traceIDHigh, 16)
	highStr = strings.Repeat("0", 16-len(highStr)) + highStr

	lowStr := strconv.FormatUint(traceIDLow, 16)
	lowStr = strings.Repeat("0", 16-len(lowStr)) + lowStr

	return highStr + lowStr
}

func gatherSpansInfo(trace []*jaeger.Span) (parentIDs map[uint64]bool, spanIDs map[uint64]bool) {
	parentIDs = make(map[uint64]bool)
	spanIDs = make(map[uint64]bool)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[uint64(span.SpanId)] = true
		parentIDs[uint64(span.ParentSpanId)] = true
	}

	return spanIDs, parentIDs
}

func getExpandInfo(batch *jaeger.Batch) (project, version, env string) {
	if batch.Process == nil {
		return
	}

	for _, tag := range batch.Process.Tags {
		if tag == nil {
			continue
		}

		switch tag.Key {
		case itrace.Project:
			project = fmt.Sprintf("%v", getValueString(tag))
		case itrace.Version:
			version = fmt.Sprintf("%v", getValueString(tag))
		case itrace.Env:
			env = fmt.Sprintf("%v", getValueString(tag))
		}
	}

	return
}

func getValueString(tag *jaeger.Tag) interface{} {
	switch tag.VType {
	case jaeger.TagType_STRING:
		return *(tag.VStr)
	case jaeger.TagType_DOUBLE:
		return *(tag.VDouble)
	case jaeger.TagType_BOOL:
		return *(tag.VBool)
	case jaeger.TagType_LONG:
		return *(tag.VLong)
	case jaeger.TagType_BINARY:
		return tag.VBinary
	default:
		return nil
	}
}
