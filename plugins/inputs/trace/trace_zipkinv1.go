package trace

import (
	"encoding/json"
	"fmt"

	"github.com/apache/thrift/lib/go/thrift"

	zipkincore "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace/zipkinV1_core"
)


type Endpoint struct {
	ServiceName string `json:"serviceName"`
	Ipv4        string `json:"ipv4"`
	Ipv6        string `json:"ipv6,omitempty"`
	Port        int16  `json:"port"`
}

type Annotation struct {
	Timestamp int64     `json:"timestamp"`
	Value     string    `json:"value"`
	Host      *Endpoint `json:"endpoint,omitempty"`
}

type BinaryAnnotation struct {
	Key            string         `json:"key"`
	Value          string         `json:"value"`
	Host           *Endpoint      `json:"endpoint,omitempty"`
}

type ZipkinSpanV1 struct {
	TraceID   string `thrift:"traceId,1" db:"traceId" json:"traceId"`
	Name      string `thrift:"name,3" db:"name" json:"name"`
	ParentID  string `thrift:"parentId,5" db:"parentId" json:"parentId,omitempty"`
	ID        string `thrift:"id,4" db:"id" json:"id"`
	Timestamp int64  `thrift:"timestamp,10" db:"timestamp" json:"timestamp,omitempty"`
	Duration  int64  `thrift:"duration,11" db:"duration" json:"duration,omitempty"`
	Debug     bool   `thrift:"debug,9" db:"debug" json:"debug,omitempty"`

	Annotations []*Annotation `thrift:"annotations,6" db:"annotations" json:"annotations"`
	BinaryAnnotations []*BinaryAnnotation `thrift:"binary_annotations,8" db:"binary_annotations" json:"binary_annotations"`
}

func (z *ZipkinTracer) parseZipkinJsonV1(octets []byte) error {
	spans := []*ZipkinSpanV1{}
	if err := json.Unmarshal(octets, &spans); err != nil {
		return err
	}

	for _, zs := range spans {
		tAdpter := TraceAdapter{}
		tAdpter.source = "zipkin"

		tAdpter.duration = zs.Duration
		tAdpter.timestampUs = zs.Timestamp

		js, err := json.Marshal(zs)
		if err != nil {
			return err
		}
		tAdpter.content = string(js)

		tAdpter.class = "tracing"
		tAdpter.operationName = zs.Name
		tAdpter.parentID = zs.ParentID
		tAdpter.traceID = zs.TraceID
		tAdpter.spanID = zs.ID

		for _, ano := range zs.Annotations {
			if ano.Host != nil && ano.Host.ServiceName != "" {
				tAdpter.serviceName = ano.Host.ServiceName
				break
			}
		}

		if tAdpter.serviceName == "" {
			for _, bno := range zs.BinaryAnnotations {
				if bno.Host != nil && bno.Host.ServiceName != "" {
					tAdpter.serviceName = bno.Host.ServiceName
					break
				}
			}
		}

		for _, bno := range zs.BinaryAnnotations {
			if bno != nil && bno.Key == "error" {
				tAdpter.isError = "true"
				break
			}
		}

		tAdpter.mkLineProto()
	}
	return nil
}

func (z *ZipkinTracer) parseZipkinThriftV1(octets []byte) error{
	zspans, err := unmarshalZipkinThriftV1(octets)
	if err != nil {
		return err
	}

	for _, zs := range zspans {
		tAdpter := TraceAdapter{}
		tAdpter.source = "zipkin"

		tAdpter.duration = *zs.Duration
		tAdpter.timestampUs = *zs.Timestamp

		js, err := json.Marshal(zs)
		if err != nil {
			return err
		}
		tAdpter.content = string(js)

		tAdpter.class = "tracing"
		tAdpter.operationName = zs.Name
		if zs.ParentID != nil {
			tAdpter.parentID = fmt.Sprintf("%d", uint64(*zs.ParentID))
		}

		tAdpter.traceID = fmt.Sprintf("%d", uint64(zs.TraceID))
		tAdpter.spanID = fmt.Sprintf("%d", uint64(zs.ID))

		for _, ano := range zs.Annotations {
			if ano.Host != nil && ano.Host.ServiceName != "" {
				tAdpter.serviceName = ano.Host.ServiceName
				break
			}
		}

		if tAdpter.serviceName == "" {
			for _, bno := range zs.BinaryAnnotations {
				if bno.Host != nil && bno.Host.ServiceName != "" {
					tAdpter.serviceName = bno.Host.ServiceName
					break
				}
			}
		}

		for _, bno := range zs.BinaryAnnotations {
			if bno != nil && bno.Key == "error" {
				tAdpter.isError = "true"
				break
			}
		}

		tAdpter.mkLineProto()
	}
	return nil
}

func unmarshalZipkinThriftV1(octets []byte) ([]*zipkincore.Span, error) {
	buffer := thrift.NewTMemoryBuffer()
	if _, err := buffer.Write(octets); err != nil {
		return nil, err
	}

	transport := thrift.NewTBinaryProtocolTransport(buffer)
	_, size, err := transport.ReadListBegin()
	if err != nil {
		return nil, err
	}

	spans := make([]*zipkincore.Span, size)
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			return nil, err
		}
		spans[i] = zs
	}

	if err = transport.ReadListEnd(); err != nil {
		return nil, err
	}

	return spans, nil
}
