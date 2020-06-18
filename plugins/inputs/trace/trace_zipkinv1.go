package trace

import (
	"encoding/json"
	"fmt"
	"unsafe"
	"time"

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
	BinaryAnnotations []*BinaryAnnotation `thrift:"binary_annotations,8" db:"binary_annotations" json:"binaryAnnotations"`
}

func getFirstTimestamp(zs *ZipkinSpanV1) int64 {
	for _, ano := range zs.Annotations {
		if ano.Timestamp != 0 {
			return ano.Timestamp
		}
	}

	return time.Now().UnixNano()/1000

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
		if tAdpter.timestampUs == 0 {
			tAdpter.timestampUs = getFirstTimestamp(zs)
		}

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

func zipkinConvThrift2Json(z *zipkincore.Span) *zipkincore.SpanJsonApater {
	zc := &zipkincore.SpanJsonApater{}
	zc.TraceID = uint64(z.TraceID)
	zc.Name    = z.Name
	zc.ID      = uint64(z.ID)
	if z.ParentID != nil {
		zc.ParentID= uint64(*z.ParentID)
	}

	for _, ano := range z.Annotations {
		jAno := zipkincore.AnnotationJsonApater{}
		jAno.Timestamp = uint64(ano.Timestamp)
		jAno.Value   = ano.Value
		if ano.Host != nil {
			ep := &zipkincore.EndpointJsonApater{}
			ep.ServiceName = ano.Host.ServiceName
			ep.Port = ano.Host.Port
			ep.Ipv6 = append(ep.Ipv6, ano.Host.Ipv6...)
			ptr := uintptr(unsafe.Pointer(&ano.Host.Ipv4))
			for i:= 0; i < 4; i++ {
				p := ptr + uintptr(i)
				ep.Ipv4 = append(ep.Ipv4, *((*byte)(unsafe.Pointer(p))))
			}
			jAno.Host = ep
		}
		zc.Annotations = append(zc.Annotations, jAno)
	}

	for _, bno := range z.BinaryAnnotations {
		jBno := zipkincore.BinaryAnnotationJsonApater{}
		jBno.Key = bno.Key
		jBno.Value = append(jBno.Value, bno.Value...)
		jBno.AnnotationType = bno.AnnotationType
		if bno.Host != nil {
			ep := &zipkincore.EndpointJsonApater{}
			ep.ServiceName = bno.Host.ServiceName
			ep.Port = bno.Host.Port
			ep.Ipv6 = append(ep.Ipv6, bno.Host.Ipv6...)
			ptr := uintptr(unsafe.Pointer(&bno.Host.Ipv4))
			for i:= 0; i < 4; i++ {
				p := ptr + uintptr(i)
				ep.Ipv4 = append(ep.Ipv4, *((*byte)(unsafe.Pointer(p))))
			}
			jBno.Host = ep
		}
		zc.BinaryAnnotations = append(zc.BinaryAnnotations, jBno)
	}
	zc.Debug   = z.Debug
	if z.Timestamp != nil {
		zc.Timestamp = uint64(*z.Timestamp)
	}

	if z.Duration != nil {
		zc.Duration  = uint64(*z.Duration)
	}

	if z.TraceIDHigh != nil {
		zc.TraceIDHigh = uint64(*z.TraceIDHigh)
	}

	return zc
}

func (z *ZipkinTracer) parseZipkinThriftV1(octets []byte) error{
	zspans, err := unmarshalZipkinThriftV1(octets)
	if err != nil {
		return err
	}

	for _, zs := range zspans {
		z := zipkinConvThrift2Json(zs)
		tAdpter := TraceAdapter{}
		tAdpter.source = "zipkin"

		tAdpter.duration = *zs.Duration
		tAdpter.timestampUs = *zs.Timestamp

		js, err := json.Marshal(z)
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
