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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func JaegerTraceHandle(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("trace handle with path: %s", req.URL.Path)
	defer func() {
		resp.WriteHeader(http.StatusOK)
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	reqInfo, err := trace.ParseTraceInfo(req)
	if err != nil {
		log.Error(err.Error())

		return
	}

	if reqInfo.ContentType != "application/x-thrift" {
		log.Errorf("Jeager unsupported Content-Type: %s", reqInfo.ContentType)

		return
	}

	if err = parseJaegerThrift(reqInfo.Body); err != nil {
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

	group, err := batchToAdapters(batch)
	if err != nil {
		return err
	}

	if len(group) != 0 {
		trace.MkLineProto(group, inputName)
	} else {
		log.Warn("empty batch")
	}

	return nil
}

func batchToAdapters(batch *jaeger.Batch) ([]*trace.DatakitSpan, error) {
	project, version, env := getExpandInfo(batch)
	if project == "" {
		project = jaegerTags[trace.PROJECT]
	}
	if version == "" {
		version = jaegerTags[trace.VERSION]
	}
	if env == "" {
		env = jaegerTags[trace.ENV]
	}

	var (
		group              []*trace.DatakitSpan
		spanIDs, parentIDs = getSpanIDsAndParentIDs(batch.Spans)
	)
	for _, span := range batch.Spans {
		if span == nil {
			continue
		}

		tAdapter := &trace.DatakitSpan{
			TraceID:   trace.GetTraceStringID(span.TraceIdHigh, span.TraceIdLow),
			ParentID:  fmt.Sprintf("%d", span.ParentSpanId),
			SpanID:    fmt.Sprintf("%d", span.SpanId),
			Duration:  span.Duration * int64(time.Microsecond),
			Env:       env,
			Operation: span.OperationName,
			Project:   project,
			Service:   batch.Process.ServiceName,
			Source:    inputName,
			SpanType:  trace.FindIntIDSpanType(span.SpanId, span.ParentSpanId, spanIDs, parentIDs),
			Start:     span.StartTime * int64(time.Microsecond),
			Version:   version,
		}

		buf, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		tAdapter.Content = string(buf)

		tAdapter.Status = trace.STATUS_OK
		for _, tag := range span.Tags {
			if tag.Key == "error" {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}
		tAdapter.Tags = jaegerTags

		group = append(group, tAdapter)
	}

	return group, nil
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

func getExpandInfo(batch *jaeger.Batch) (project, ver, env string) {
	if batch.Process == nil {
		return
	}
	for _, tag := range batch.Process.Tags {
		if tag == nil {
			continue
		}

		if tag.Key == trace.PROJECT {
			project = fmt.Sprintf("%v", getValueString(tag))
		}

		if tag.Key == trace.VERSION {
			ver = fmt.Sprintf("%v", getValueString(tag))
		}

		if tag.Key == trace.ENV {
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
