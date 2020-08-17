package trace

import (
	"encoding/json"
	"fmt"
	"github.com/uber/jaeger-client-go/thrift"
	j "github.com/uber/jaeger-client-go/thrift-gen/jaeger"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
		tAdpter.Source = "jaeger"

		tAdpter.Duration    = s.Duration
		tAdpter.TimestampUs = s.StartTime
		sJson, err := json.Marshal(s)
		if err != nil {
			return err
		}
		tAdpter.Content = string(sJson)

		tAdpter.Class         = "tracing"
		tAdpter.ServiceName   = batch.Process.ServiceName
		tAdpter.OperationName = s.OperationName
		if s.ParentSpanId != 0 {
			tAdpter.ParentID      = fmt.Sprintf("%d", s.ParentSpanId)
		}

		tAdpter.TraceID = fmt.Sprintf("%x%x", uint64(s.TraceIdHigh), uint64(s.TraceIdLow))
		tAdpter.SpanID  = fmt.Sprintf("%d", s.SpanId)

		for _, tag := range s.Tags {
			if tag.Key == "error" {
				tAdpter.IsError = "true"
				break
			}
		}
		tAdpter.Tags = JaegerTags
		pt, err := tAdpter.MkLineProto()
		if err != nil {
			log.Error(err)
			continue
		}
		if err := dkio.NamedFeed(pt, dkio.Logging, "tracing"); err != nil {
			log.Errorf("io feed err: %s", err)
		}
	}

	return nil
}
