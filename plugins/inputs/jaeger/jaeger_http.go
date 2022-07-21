// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
)

func handleJaegerTrace(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### received tracing data from path: %s", req.URL.Path)

	var (
		readbodycost = time.Now()
		buf          = thrift.NewTMemoryBuffer()
	)
	if _, err := buf.ReadFrom(req.Body); err != nil {
		log.Error(err.Error())
		resp.WriteHeader(http.StatusBadRequest)

		return
	}

	log.Debugf("### path: %s, Content-Type: %s, body-size: %d, read-body-cost: %dms",
		req.URL.Path, req.Header.Get("Content-Type"), buf.Len(), time.Since(readbodycost)/time.Millisecond)

	if wpool == nil {
		if err := parseJaegerTrace(buf); err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
	} else {
		job, err := workerpool.NewJob(workerpool.WithInput(buf),
			workerpool.WithProcess(parseJaegerTraceAdapter),
			workerpool.WithProcessCallback(func(input, output interface{}, cost time.Duration, isTimeout bool) {
				log.Debugf("### job status: input: %v, output: %v, cost: %dms, timeout: %v", input, output, cost/time.Millisecond, isTimeout)
			}),
			workerpool.WithTimeout(jobTimeout),
		)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}

		if err = wpool.MoreJob(job); err != nil {
			log.Error(err)
			resp.WriteHeader(http.StatusTooManyRequests)

			return
		}
	}

	resp.WriteHeader(http.StatusOK)
}

func parseJaegerTraceAdapter(input interface{}) (output interface{}) {
	buf, ok := input.(*thrift.TMemoryBuffer)
	if !ok {
		return errors.New("type assertion failed")
	}

	return parseJaegerTrace(buf)
}

func parseJaegerTrace(buf *thrift.TMemoryBuffer) error {
	var (
		transport = thrift.NewTBinaryProtocolConf(buf, &thrift.TConfiguration{})
		batch     = &jaeger.Batch{}
	)
	if err := batch.Read(context.TODO(), transport); err != nil {
		return err
	}

	if dktrace := batchToDkTrace(batch); len(dktrace) != 0 && afterGatherRun != nil {
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
	}

	return nil
}

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

		dkspan := &itrace.DatakitSpan{
			ParentID:   fmt.Sprintf("%x", uint64(span.ParentSpanId)),
			SpanID:     fmt.Sprintf("%x", uint64(span.SpanId)),
			Service:    batch.Process.ServiceName,
			Resource:   span.OperationName,
			Operation:  span.OperationName,
			Source:     inputName,
			SourceType: itrace.SPAN_SOURCE_CUSTOMER,
			SpanType:   itrace.FindSpanTypeIntSpanID(span.SpanId, span.ParentSpanId, spanIDs, parentIDs),
			Env:        env,
			Project:    project,
			Start:      span.StartTime * int64(time.Microsecond),
			Duration:   span.Duration * int64(time.Microsecond),
			Version:    version,
		}

		if span.TraceIdHigh != 0 {
			dkspan.TraceID = fmt.Sprintf("%x%x", uint64(span.TraceIdHigh), uint64(span.TraceIdLow))
		} else {
			dkspan.TraceID = fmt.Sprintf("%x", uint64(span.TraceIdLow))
		}

		dkspan.Status = itrace.STATUS_OK
		for _, tag := range span.Tags {
			if tag.Key == "error" {
				dkspan.Status = itrace.STATUS_ERR
				break
			}
		}

		sourceTags := make(map[string]string)
		for _, tag := range span.Tags {
			sourceTags[tag.Key] = tag.String()
		}
		dkspan.Tags = itrace.MergeInToCustomerTags(customerKeys, tags, sourceTags)

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}
	if len(dktrace) != 0 {
		dktrace[0].Metrics = make(map[string]interface{})
		dktrace[0].Metrics[itrace.FIELD_PRIORITY] = itrace.PRIORITY_AUTO_KEEP
	}

	return dktrace
}

func gatherSpansInfo(trace []*jaeger.Span) (parentIDs map[int64]bool, spanIDs map[int64]bool) {
	parentIDs = make(map[int64]bool)
	spanIDs = make(map[int64]bool)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[span.SpanId] = true
		if span.ParentSpanId != 0 {
			parentIDs[span.ParentSpanId] = true
		}
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
		case itrace.PROJECT:
			project = fmt.Sprintf("%v", getValueString(tag))
		case itrace.VERSION:
			version = fmt.Sprintf("%v", getValueString(tag))
		case itrace.ENV:
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
