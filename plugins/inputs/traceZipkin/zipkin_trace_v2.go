package traceZipkin

import (
	"fmt"
	"net"
	"time"
	"strconv"
	"encoding/json"
	"encoding/binary"

	"github.com/golang/protobuf/proto"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	zipkin_proto3 "github.com/openzipkin/zipkin-go/proto/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

func  parseZipkinJsonV2(octets []byte) error {
	log.Debugf("->|%s|<-", string(octets))
	spans := []*zipkinmodel.SpanModel{}
	if err := json.Unmarshal(octets, &spans); err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, zs := range spans {
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "zipkin"

		tAdpter.Duration    = int64(zs.Duration/time.Microsecond)
		tAdpter.TimestampUs = zs.Timestamp.UnixNano()/1000
		sJson, err := json.Marshal(zs)
		if err != nil {
			return err
		}
		tAdpter.Content = string(sJson)

		tAdpter.Class         = "tracing"
		if zs.LocalEndpoint != nil {
			tAdpter.ServiceName   = zs.LocalEndpoint.ServiceName
		}

		tAdpter.OperationName = zs.Name

		if zs.ParentID != nil {
			tAdpter.ParentID      = fmt.Sprintf("%x", uint64(*zs.ParentID))
		}

		if zs.TraceID.High != 0 {
			tAdpter.TraceID = fmt.Sprintf("%x%x", zs.TraceID.High, zs.TraceID.Low)
		} else {
			tAdpter.TraceID = fmt.Sprintf("%x", zs.TraceID.Low)
		}

		tAdpter.SpanID        = fmt.Sprintf("%x", uint64(zs.ID))

		for tag, _ := range zs.Tags {
			if tag == "error" {
				tAdpter.IsError = "true"
				break
			}
		}

		if zs.RemoteEndpoint != nil {
			if len(zs.RemoteEndpoint.IPv4) != 0 {
				tAdpter.EndPoint = fmt.Sprintf("%s", zs.RemoteEndpoint.IPv4)
			}

			if len(zs.RemoteEndpoint.IPv6) != 0 {
				tAdpter.EndPoint = fmt.Sprintf("%s", zs.RemoteEndpoint.IPv6)
			}
		}

		//tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
		//if zs.RemoteEndpoint == nil {
		//	if zs.LocalEndpoint == nil {
		//		tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		//	} else {
		//		if len(zs.LocalEndpoint.IPv4) == 0 && len(zs.LocalEndpoint.IPv6) == 0 {
		//			tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		//		}
		//	}
		//}
		if zs.Kind == zipkinmodel.Undetermined {
			tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
		}

		tAdpter.Tags = ZipkinTags

		adapterGroup = append(adapterGroup, tAdpter)
	}

	trace.MkLineProto(adapterGroup, inputName)
	return nil
}

func parseZipkinProtobufV2(octets []byte) error {
	zss, err := parseZipkinProtobuf3(octets)
	if err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, zs := range zss {
		tAdpter := &trace.TraceAdapter{}
		tAdpter.Source = "zipkin"

		tAdpter.Duration    = int64(zs.Duration/time.Microsecond)
		tAdpter.TimestampUs = zs.Timestamp.UnixNano()/1000
		sJson, err := json.Marshal(zs)
		if err != nil {
			return err
		}
		tAdpter.Content = string(sJson)

		tAdpter.Class         = "tracing"
		if zs.LocalEndpoint != nil {
			tAdpter.ServiceName   = zs.LocalEndpoint.ServiceName
		}
		tAdpter.OperationName = zs.Name

		if zs.ParentID != nil {
			tAdpter.ParentID      = fmt.Sprintf("%d", *zs.ParentID)
		}

		if zs.TraceID.High != 0 {
			tAdpter.TraceID = fmt.Sprintf("%d%d", zs.TraceID.High, zs.TraceID.Low)
		} else {
			tAdpter.TraceID = fmt.Sprintf("%d", zs.TraceID.Low)
		}

		tAdpter.SpanID        = fmt.Sprintf("%d", zs.ID)

		for tag, _ := range zs.Tags {
			if tag == "error" {
				tAdpter.IsError = "true"
				break
			}
		}

		if zs.RemoteEndpoint != nil {
			if len(zs.RemoteEndpoint.IPv4) != 0 {
				tAdpter.EndPoint = fmt.Sprintf("%s", zs.RemoteEndpoint.IPv4)
			}

			if len(zs.RemoteEndpoint.IPv6) != 0 {
				tAdpter.EndPoint = fmt.Sprintf("%s", zs.RemoteEndpoint.IPv6)
			}
		}

		//tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
		//if zs.RemoteEndpoint == nil {
		//	if zs.LocalEndpoint == nil {
		//		tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		//	} else {
		//		if len(zs.LocalEndpoint.IPv4) == 0 && len(zs.LocalEndpoint.IPv6) == 0 {
		//			tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		//		}
		//	}
		//}
		if zs.Kind == zipkinmodel.Undetermined {
			tAdpter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdpter.SpanType = trace.SPAN_TYPE_ENTRY
		}

		tAdpter.Tags = ZipkinTags
		adapterGroup = append(adapterGroup, tAdpter)

	}
	trace.MkLineProto(adapterGroup, inputName)
	return nil
}

func parseZipkinProtobuf3(octets []byte) (zss []*zipkinmodel.SpanModel, err error) {
	var listOfSpans zipkin_proto3.ListOfSpans
	if err := proto.Unmarshal(octets, &listOfSpans); err != nil {
		return nil, err
	}
	for _, zps := range listOfSpans.Spans {
		traceID, err := zipkinTraceIDFromHex(fmt.Sprintf("%x", zps.TraceId))
		if err != nil {
			return nil, fmt.Errorf("invalid TraceID: %v", err)
		}

		parentSpanID, _, err := zipkinSpanIDToModelSpanID(zps.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid ParentID: %v", err)
		}
		spanIDPtr, spanIDBlank, err := zipkinSpanIDToModelSpanID(zps.Id)
		if err != nil {
			return nil, fmt.Errorf("invalid SpanID: %v", err)
		}
		if spanIDBlank || spanIDPtr == nil {
			// This is a logical error
			return nil, fmt.Errorf("expected a non-nil SpanID")
		}

		zmsc := zipkinmodel.SpanContext{
			TraceID:  traceID,
			ID:       *spanIDPtr,
			ParentID: parentSpanID,
			Debug:    false,
		}
		zms := &zipkinmodel.SpanModel{
			SpanContext:    zmsc,
			Name:           zps.Name,
			Kind:           zipkinmodel.Kind(zps.Kind.String()),
			Timestamp:      microsToTime(zps.Timestamp),
			Tags:           zps.Tags,
			Duration:       microsToDuration(zps.Duration),
			LocalEndpoint:  protoEndpointToModelEndpoint(zps.LocalEndpoint),
			RemoteEndpoint: protoEndpointToModelEndpoint(zps.RemoteEndpoint),
			Shared:         zps.Shared,
			Annotations:    protoAnnotationsToModelAnnotations(zps.Annotations),
		}
		zss = append(zss, zms)
	}

	return zss, nil
}

func zipkinTraceIDFromHex(h string) (t zipkinmodel.TraceID, err error) {
	if len(h) > 16 {
		if t.High, err = strconv.ParseUint(h[0:len(h)-16], 16, 64); err != nil {
			return
		}
		t.Low, err = strconv.ParseUint(h[len(h)-16:], 16, 64)
		return
	}
	t.Low, err = strconv.ParseUint(h, 16, 64)
	return
}


func zipkinSpanIDToModelSpanID(spanId []byte) (zid *zipkinmodel.ID, blank bool, err error) {
	if len(spanId) == 0 {
		return nil, true, nil
	}
	if len(spanId) != 8 {
		return nil, true, fmt.Errorf("has length %d yet wanted length 8", len(spanId))
	}

	// Converting [8]byte --> uint64
	u64 := binary.BigEndian.Uint64(spanId)
	zid_ := zipkinmodel.ID(u64)
	return &zid_, false, nil
}
func microsToTime(us uint64) time.Time {
	return time.Unix(0, int64(us*1e3)).UTC()
}

func microsToDuration(us uint64) time.Duration {
	// us to ns; ns are the units of Duration
	return time.Duration(us * 1e3)
}

func protoEndpointToModelEndpoint(zpe *zipkin_proto3.Endpoint) *zipkinmodel.Endpoint {
	if zpe == nil {
		return nil
	}
	return &zipkinmodel.Endpoint{
		ServiceName: zpe.ServiceName,
		IPv4:        net.IP(zpe.Ipv4),
		IPv6:        net.IP(zpe.Ipv6),
		Port:        uint16(zpe.Port),
	}
}

func protoAnnotationsToModelAnnotations(zpa []*zipkin_proto3.Annotation) (zma []zipkinmodel.Annotation) {
	for _, za := range zpa {
		if za != nil {
			zma = append(zma, zipkinmodel.Annotation{
				Timestamp: microsToTime(za.Timestamp),
				Value:     za.Value,
			})
		}
	}

	if len(zma) == 0 {
		return nil
	}
	return zma
}

