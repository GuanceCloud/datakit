package jaeger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

func JaegerTraceHandle(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("%s: listen on path: %s", inputName, req.URL.Path)

	defer func() {
		resp.WriteHeader(http.StatusOK)
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	treqinfo, err := itrace.ParseTraceInfo(req)
	if err != nil {
		log.Error(err.Error())

		return
	}

	if treqinfo.ContentType != "application/x-thrift" {
		log.Errorf("Jeager unsupported Content-Type: %s", treqinfo.ContentType)

		return
	}

	if err = parseJaegerThrift(treqinfo.Body); err != nil {
		log.Error(err.Error())
	}
}

func parseJaegerThrift(octets []byte) error {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(octets); err != nil {
		return err
	}
	transport := thrift.NewTBinaryProtocolConf(buffer, &thrift.TConfiguration{})
	batch := &jaeger.Batch{}
	if err := batch.Read(context.TODO(), transport); err != nil {
		return err
	}

	if dktrace := batchToDkTrace(batch); len(dktrace) == 0 {
		log.Warn("empty datakit trace")
	} else {
		afterGather.Run(inputName, dktrace, false)
	}

	return nil
}

func batchToDkTrace(batch *jaeger.Batch) itrace.DatakitTrace {
	project, version, env := getExpandInfo(batch)
	if project == "" {
		project = tags[itrace.PROJECT]
	}
	if version == "" {
		version = tags[itrace.VERSION]
	}
	if env == "" {
		env = tags[itrace.ENV]
	}

	var (
		dktrace            itrace.DatakitTrace
		spanIDs, parentIDs = getSpanIDsAndParentIDs(batch.Spans)
	)
	for _, span := range batch.Spans {
		if span == nil {
			continue
		}

		dkspan := &itrace.DatakitSpan{
			TraceID:   itrace.GetTraceStringID(span.TraceIdHigh, span.TraceIdLow),
			ParentID:  fmt.Sprintf("%d", span.ParentSpanId),
			SpanID:    fmt.Sprintf("%d", span.SpanId),
			Env:       env,
			Operation: span.OperationName,
			Project:   project,
			Service:   batch.Process.ServiceName,
			Source:    inputName,
			SpanType:  itrace.FindSpanTypeInt(span.SpanId, span.ParentSpanId, spanIDs, parentIDs),
			Start:     span.StartTime * int64(time.Microsecond),
			Duration:  span.Duration * int64(time.Microsecond),
			Version:   version,
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

		if defSampler != nil {
			dkspan.Priority = defSampler.Priority
			dkspan.SamplingRateGlobal = defSampler.SamplingRateGlobal
		}

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}

	return dktrace
}

func getSpanIDsAndParentIDs(trace []*jaeger.Span) (map[int64]bool, map[int64]bool) {
	var (
		spanIDs   = make(map[int64]bool)
		parentIDs = make(map[int64]bool)
	)
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
