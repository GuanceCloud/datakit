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
	"time"

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
		afterGatherRun.Run(inputName, itrace.DatakitTraces{dktrace}, false)
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
			ParentID:   strconv.FormatUint(uint64(span.ParentSpanId), 16),
			SpanID:     strconv.FormatUint(uint64(span.SpanId), 16),
			Service:    batch.Process.ServiceName,
			Resource:   span.OperationName,
			Operation:  span.OperationName,
			Source:     inputName,
			SourceType: itrace.SPAN_SOURCE_CUSTOMER,
			SpanType:   itrace.FindSpanTypeIntSpanID(uint64(span.SpanId), uint64(span.ParentSpanId), spanIDs, parentIDs),
			Start:      span.StartTime * int64(time.Microsecond),
			Duration:   span.Duration * int64(time.Microsecond),
		}

		if span.TraceIdHigh != 0 {
			dkspan.TraceID = fmt.Sprintf("%x%x", uint64(span.TraceIdHigh), uint64(span.TraceIdLow))
		} else {
			dkspan.TraceID = strconv.FormatUint(uint64(span.TraceIdLow), 16)
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
		if project != "" {
			dkspan.Tags[itrace.PROJECT] = project
		}
		if version != "" {
			dkspan.Tags[itrace.TAG_VERSION] = version
		}
		if env != "" {
			dkspan.Tags[itrace.TAG_ENV] = env
		}

		dkJSpan := &DkJaegerSpan{
			TraceIdLow:   uint64(span.TraceIdLow),
			TraceIdHigh:  uint64(span.TraceIdHigh),
			SpanId:       uint64(span.SpanId),
			ParentSpanId: uint64(span.ParentSpanId),
			Span:         span,
		}
		if buf, err := json.Marshal(dkJSpan); err != nil {
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
