package traceJaeger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/uber/jaeger-client-go/thrift"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

func JaegerTraceHandle(w http.ResponseWriter, r *http.Request) {
	log.Debugf("trace handle with path: %s", r.URL.Path)
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Stack crash: %v", r)
			log.Errorf("Stack info :%s", string(debug.Stack()))
		}
	}()

	if err := handleJaegerTrace(w, r); err != nil {
		io.FeedLastError(inputName, err.Error())
		log.Errorf("%v", err)
	}
}

func handleJaegerTrace(w http.ResponseWriter, r *http.Request) error {
	reqInfo, err := trace.ParseHttpReq(r)
	if err != nil {
		return err
	}

	if reqInfo.ContentType != "application/x-thrift" {
		return fmt.Errorf("Jeager unsupported Content-Type: %s", reqInfo.ContentType)
	}

	return parseJaegerThrift(reqInfo.Body)
}

func parseJaegerThrift(octets []byte) error {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(octets); err != nil {
		return err
	}
	transport := thrift.NewTBinaryProtocolConf(buffer, &thrift.TConfiguration{})
	batch := &j.Batch{}
	if err := batch.Read(context.TODO(), transport); err != nil {
		return err
	}

	groups, err := processBatch(batch)
	if err != nil {
		return err
	}

	trace.MkLineProto(groups, inputName)
	return nil
}

func processBatch(batch *j.Batch) ([]*trace.TraceAdapter, error) {
	adapterGroup := []*trace.TraceAdapter{}

	project, ver, env := getExpandInfo(batch)
	if project == "" {
		project = trace.GetFromPluginTag(JaegerTags, trace.PROJECT)
	}

	if ver == "" {
		ver = trace.GetFromPluginTag(JaegerTags, trace.VERSION)
	}

	if env == "" {
		env = trace.GetFromPluginTag(JaegerTags, trace.ENV)
	}

	for _, span := range batch.Spans {
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "jaeger"
		tAdpter.Project = project
		tAdpter.Version = ver
		tAdpter.Env = env

		tAdpter.Duration = span.Duration * 1000
		tAdpter.Start = span.StartTime * 1000
		sJson, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		tAdpter.Content = string(sJson)

		tAdpter.ServiceName = batch.Process.ServiceName
		tAdpter.OperationName = span.OperationName
		if span.ParentSpanId != 0 {
			tAdpter.ParentID = fmt.Sprintf("%d", span.ParentSpanId)
		}

		tAdpter.TraceID = fmt.Sprintf("%x%x", uint64(span.TraceIdHigh), uint64(span.TraceIdLow))
		tAdpter.SpanID = fmt.Sprintf("%d", span.SpanId)

		tAdpter.Status = trace.STATUS_OK
		for _, tag := range span.Tags {
			if tag.Key == "error" {
				tAdpter.Status = trace.STATUS_ERR
				break
			}
		}
		tAdpter.Tags = JaegerTags

		// run trace data sample
		if traceSampleConf != nil {
			if !trace.DefErrCheckHandler(trace.ErrMapper[tAdpter.Status]) && !trace.DefIgnoreTagsHandler(tAdpter.Tags, traceSampleConf.IgnoreTagsList) {
				if !trace.DefSampleHandler(uint64(trace.TraceStrIdToInt(tAdpter.TraceID)), traceSampleConf.Rate, traceSampleConf.Scope) {
					continue
				}
			}
		}

		adapterGroup = append(adapterGroup, tAdpter)
	}

	return adapterGroup, nil
}

func getExpandInfo(batch *j.Batch) (project, ver, env string) {
	if batch.Process == nil {
		return
	}
	for _, tag := range batch.Process.Tags {
		if tag == nil {
			continue
		}

		if tag.Key == trace.PROJECT {
			project = fmt.Sprintf("%v", getTagValue(tag))
		}

		if tag.Key == trace.VERSION {
			ver = fmt.Sprintf("%v", getTagValue(tag))
		}

		if tag.Key == trace.ENV {
			env = fmt.Sprintf("%v", getTagValue(tag))
		}
	}
	return
}

func getTagValue(tag *j.Tag) interface{} {
	switch tag.VType {
	case j.TagType_STRING:
		return *(tag.VStr)
	case j.TagType_DOUBLE:
		return *(tag.VDouble)
	case j.TagType_BOOL:
		return *(tag.VBool)
	case j.TagType_LONG:
		return *(tag.VLong)
	case j.TagType_BINARY:
		return tag.VBinary
	default:
		return nil
	}
}
