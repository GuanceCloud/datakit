package traceJaeger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/uber/jaeger-client-go/thrift"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"

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
	adapterGroup := []*trace.TraceAdapter{}

	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(octets); err != nil {
		return err
	}
	transport := thrift.NewTBinaryProtocolTransport(buffer)
	batch := &j.Batch{}
	if err := batch.Read(transport); err != nil {
		return err
	}

	for _, s := range batch.Spans {
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "jaeger"

		tAdpter.Duration = s.Duration
		tAdpter.TimestampUs = s.StartTime
		sJson, err := json.Marshal(s)
		if err != nil {
			return err
		}
		tAdpter.Content = string(sJson)

		tAdpter.Class = "tracing"
		tAdpter.ServiceName = batch.Process.ServiceName
		tAdpter.OperationName = s.OperationName
		if s.ParentSpanId != 0 {
			tAdpter.ParentID = fmt.Sprintf("%d", s.ParentSpanId)
		}

		tAdpter.TraceID = fmt.Sprintf("%x%x", uint64(s.TraceIdHigh), uint64(s.TraceIdLow))
		tAdpter.SpanID = fmt.Sprintf("%d", s.SpanId)

		for _, tag := range s.Tags {
			if tag.Key == "error" {
				tAdpter.IsError = "true"
				break
			}
		}
		tAdpter.Tags = JaegerTags

		adapterGroup = append(adapterGroup, tAdpter)
	}

	trace.MkLineProto(adapterGroup, inputName)
	return nil
}
