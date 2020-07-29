package trace

import (
	"encoding/json"
	"fmt"
	"github.com/uber/jaeger-client-go/thrift"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
)

func (z *JaegerTracer) Decode(octets []byte) error {
	switch z.ContentType {
	case "application/x-thrift":
		return z.parseJaegerThrift(octets)
	default:
		return fmt.Errorf("Jaeger unsupported Content-Type: %s", z.ContentType)
	}
}

func (z *JaegerTracer) parseJaegerThrift(octets []byte) error {
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
		tAdpter := TraceAdapter{}
		tAdpter.source = "jaeger"

		tAdpter.duration    = s.Duration
		tAdpter.timestampUs = s.StartTime
		sJson, err := json.Marshal(s)
		if err != nil {
			return err
		}
		tAdpter.content = string(sJson)

		tAdpter.class         = "tracing"
		tAdpter.serviceName   = batch.Process.ServiceName
		tAdpter.operationName = s.OperationName
		if s.ParentSpanId != 0 {
			tAdpter.parentID      = fmt.Sprintf("%d", s.ParentSpanId)
		}

		tAdpter.traceID = fmt.Sprintf("%x%x", uint64(s.TraceIdHigh), uint64(s.TraceIdLow))
		tAdpter.spanID  = fmt.Sprintf("%d", s.SpanId)

		for _, tag := range s.Tags {
			if tag.Key == "error" {
				tAdpter.isError = "true"
				break
			}
		}

		tAdpter.mkLineProto()
	}

	return nil
}
