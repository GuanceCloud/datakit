package traceZipkin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/apache/thrift/lib/go/thrift"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
	zipkincore "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/traceZipkin/zipkinV1_core"
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
	Key   string    `json:"key"`
	Value string    `json:"value"`
	Host  *Endpoint `json:"endpoint,omitempty"`
}

type ZipkinSpanV1 struct {
	TraceID   string `thrift:"traceId,1" db:"traceId" json:"traceId"`
	Name      string `thrift:"name,3" db:"name" json:"name"`
	ParentID  string `thrift:"parentId,5" db:"parentId" json:"parentId,omitempty"`
	ID        string `thrift:"id,4" db:"id" json:"id"`
	Timestamp int64  `thrift:"timestamp,10" db:"timestamp" json:"timestamp,omitempty"`
	Duration  int64  `thrift:"duration,11" db:"duration" json:"duration,omitempty"`
	Debug     bool   `thrift:"debug,9" db:"debug" json:"debug,omitempty"`

	Annotations       []*Annotation       `thrift:"annotations,6" db:"annotations" json:"annotations"`
	BinaryAnnotations []*BinaryAnnotation `thrift:"binary_annotations,8" db:"binary_annotations" json:"binaryAnnotations"`
}

func getFirstTimestamp(zs *ZipkinSpanV1) int64 {
	var ts int64
	var isFound bool
	ts = 0x7FFFFFFFFFFFFFFF
	for _, ano := range zs.Annotations {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp < ts {
			isFound = true
			ts = ano.Timestamp
		}
	}

	if isFound {
		return ts
	}
	return time.Now().UnixNano() / 1000
}
func parseZipkinJsonV1(octets []byte) error {
	log.Debugf("->|%v|<-", string(octets))
	
	spans := []*ZipkinSpanV1{}
	if err := json.Unmarshal(octets, &spans); err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, zs := range spans {
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "zipkin"

		tAdpter.Duration = zs.Duration
		tAdpter.TimestampUs = zs.Timestamp
		if tAdpter.TimestampUs == 0 {
			tAdpter.TimestampUs = getFirstTimestamp(zs)
		}

		js, err := json.Marshal(zs)
		if err != nil {
			return err
		}
		tAdpter.Content = string(js)

		tAdpter.Class = "tracing"
		tAdpter.OperationName = zs.Name
		tAdpter.ParentID = zs.ParentID
		tAdpter.TraceID = zs.TraceID
		tAdpter.SpanID = zs.ID

		for _, ano := range zs.Annotations {
			if ano.Host != nil && ano.Host.ServiceName != "" {
				tAdpter.ServiceName = ano.Host.ServiceName
				break
			}
		}

		if tAdpter.ServiceName == "" {
			for _, bno := range zs.BinaryAnnotations {
				if bno.Host != nil && bno.Host.ServiceName != "" {
					tAdpter.ServiceName = bno.Host.ServiceName
					break
				}
			}
		}

		for _, bno := range zs.BinaryAnnotations {
			if bno != nil && bno.Key == "error" {
				tAdpter.IsError = "true"
				break
			}
		}

		if tAdpter.Duration == 0  {
			tAdpter.Duration = getDurationByAno(zs.Annotations)
		}
		tAdpter.Tags = ZipkinTags

		adapterGroup = append(adapterGroup, tAdpter)
	}

	trace.MkLineProto(adapterGroup, inputName)
	return nil
}

func getDurationByAno(anos []*Annotation) int64 {
	if len(anos) < 2 {
		return 0
	}

	var startTs, stopTs int64
	startTs = 0x7FFFFFFFFFFFFFFF
	for _, ano := range anos {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp > stopTs {
			stopTs = ano.Timestamp
		}

		if ano.Timestamp < startTs {
			startTs = ano.Timestamp
		}
	}
	if stopTs > startTs {
		return stopTs - startTs
	}
	return 0
}

func int2ip(i uint32) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, i)
	return bs
}

func zipkinConvThrift2Json(z *zipkincore.Span) *zipkincore.SpanJsonApater {
	zc := &zipkincore.SpanJsonApater{}
	zc.TraceID = uint64(z.TraceID)
	zc.Name = z.Name
	zc.ID = uint64(z.ID)
	if z.ParentID != nil {
		zc.ParentID = uint64(*z.ParentID)
	}

	for _, ano := range z.Annotations {
		jAno := zipkincore.AnnotationJsonApater{}
		jAno.Timestamp = uint64(ano.Timestamp)
		jAno.Value = ano.Value
		if ano.Host != nil {
			ep := &zipkincore.EndpointJsonApater{}
			ep.ServiceName = ano.Host.ServiceName
			ep.Port = ano.Host.Port
			ep.Ipv6 = append(ep.Ipv6, ano.Host.Ipv6...)

			ipbytes := int2ip(uint32(ano.Host.Ipv4))
			ep.Ipv4 = net.IP(ipbytes)
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

			ipbytes := int2ip(uint32(bno.Host.Ipv4))
			ep.Ipv4 = net.IP(ipbytes)

			jBno.Host = ep
		}
		zc.BinaryAnnotations = append(zc.BinaryAnnotations, jBno)
	}
	zc.Debug = z.Debug
	if z.Timestamp != nil {
		zc.Timestamp = uint64(*z.Timestamp)
	}

	if z.Duration != nil {
		zc.Duration = uint64(*z.Duration)
	}

	if z.TraceIDHigh != nil {
		zc.TraceIDHigh = uint64(*z.TraceIDHigh)
	}

	return zc
}

func parseZipkinThriftV1(octets []byte) error {
	zspans, err := unmarshalZipkinThriftV1(octets)
	if err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, zs := range zspans {
		z := zipkinConvThrift2Json(zs)
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "zipkin"

		if zs.Duration != nil {
			tAdpter.Duration = *zs.Duration
		}

		if zs.Timestamp != nil {
			tAdpter.TimestampUs = *zs.Timestamp
		} else {
			//tAdpter.TimestampUs = time.Now().UnixNano() / 1000
			tAdpter.TimestampUs = getStartTimestamp(zs)
		}

		js, err := json.Marshal(z)
		if err != nil {
			return err
		}
		tAdpter.Content = string(js)

		tAdpter.Class = "tracing"
		tAdpter.OperationName = zs.Name
		if zs.ParentID != nil {
			tAdpter.ParentID = fmt.Sprintf("%d", uint64(*zs.ParentID))
		}

		tAdpter.TraceID = fmt.Sprintf("%d", uint64(zs.TraceID))
		tAdpter.SpanID = fmt.Sprintf("%d", uint64(zs.ID))

		for _, ano := range zs.Annotations {
			if ano.Host != nil && ano.Host.ServiceName != "" {
				tAdpter.ServiceName = ano.Host.ServiceName
				break
			}
		}

		if tAdpter.ServiceName == "" {
			for _, bno := range zs.BinaryAnnotations {
				if bno.Host != nil && bno.Host.ServiceName != "" {
					tAdpter.ServiceName = bno.Host.ServiceName
					break
				}
			}
		}

		for _, bno := range zs.BinaryAnnotations {
			if bno != nil && bno.Key == "error" {
				tAdpter.IsError = "true"
				break
			}
		}
		if tAdpter.Duration == 0 {
			tAdpter.Duration = getDurationThriftAno(zs.Annotations)
		}


		tAdpter.Tags = ZipkinTags
		adapterGroup = append(adapterGroup, tAdpter)
	}

	trace.MkLineProto(adapterGroup, inputName)
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

	spans := make([]*zipkincore.Span, 0)
	for i := 0; i < size; i++ {
		zs := &zipkincore.Span{}
		if err = zs.Read(transport); err != nil {
			return nil, err
		}
		spans = append(spans, zs)
	}

	if err = transport.ReadListEnd(); err != nil {
		return nil, err
	}

	return spans, nil
}

func getStartTimestamp(zs *zipkincore.Span) int64 {
	var ts int64
	var isFound bool
	ts = 0x7FFFFFFFFFFFFFFF
	for _, ano := range zs.Annotations {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp < ts {
			isFound = true
			ts = ano.Timestamp
		}
	}

	if isFound {
		return ts
	}
	return time.Now().UnixNano() / 1000
}

func getDurationThriftAno(anos []*zipkincore.Annotation) int64 {
	if len(anos) < 2 {
		return 0
	}

	var startTs, stopTs int64
	startTs = 0x7FFFFFFFFFFFFFFF
	for _, ano := range anos {
		if ano.Timestamp == 0 {
			continue
		}
		if ano.Timestamp > stopTs {
			stopTs = ano.Timestamp
		}

		if ano.Timestamp < startTs {
			startTs = ano.Timestamp
		}
	}
	if stopTs > startTs {
		return stopTs - startTs
	}
	return 0
}
