package jaeger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/uber/jaeger-client-go/thrift"
	"github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
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

	reqInfo, err := trace.ParseHTTPReq(req)
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
		log.Debug("empty batch")
	}

	return nil
}

func batchToAdapters(batch *jaeger.Batch) ([]*trace.TraceAdapter, error) {
	project, ver, env := getExpandInfo(batch)
	if project == "" {
		project = jaegerTags[trace.PROJECT]
	}

	if ver == "" {
		ver = jaegerTags[trace.VERSION]
	}

	if env == "" {
		env = jaegerTags[trace.ENV]
	}

	var adapterGroup []*trace.TraceAdapter
	for _, span := range batch.Spans {
		tAdapter := &trace.TraceAdapter{}
		tAdapter.Source = "jaeger"
		tAdapter.Project = project
		tAdapter.Version = ver
		tAdapter.Env = env

		tAdapter.Duration = span.Duration * 1000
		tAdapter.Start = span.StartTime * 1000
		sJSON, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		tAdapter.Content = string(sJSON)

		tAdapter.ServiceName = batch.Process.ServiceName
		tAdapter.OperationName = span.OperationName
		if span.ParentSpanId != 0 {
			tAdapter.ParentID = fmt.Sprintf("%d", span.ParentSpanId)
		}

		// tAdapter.TraceID = fmt.Sprintf("%x%x", uint64(span.TraceIdHigh), uint64(span.TraceIdLow))
		tAdapter.TraceID = trace.GetStringTraceID(span.TraceIdHigh, span.TraceIdLow)
		tAdapter.SpanID = fmt.Sprintf("%d", span.SpanId)

		tAdapter.Status = trace.STATUS_OK
		for _, tag := range span.Tags {
			if tag.Key == "error" {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}
		tAdapter.Tags = jaegerTags

		adapterGroup = append(adapterGroup, tAdapter)
	}

	return adapterGroup, nil
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
