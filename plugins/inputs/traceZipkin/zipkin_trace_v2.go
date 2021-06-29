package traceZipkin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	zipkin_proto3 "github.com/openzipkin/zipkin-go/proto/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

func parseZipkinJsonV2(octets []byte) error {
	log.Debugf("->|%s|<-", string(octets))
	spans := []*zipkinmodel.SpanModel{}
	if err := json.Unmarshal(octets, &spans); err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, span := range spans {
		tAdapter := &trace.TraceAdapter{}
		tAdapter.Source = "zipkin"

		tAdapter.Duration = int64(span.Duration)
		tAdapter.Start = span.Timestamp.UnixNano()
		sJson, err := json.Marshal(span)
		if err != nil {
			return err
		}
		tAdapter.Content = string(sJson)

		if span.LocalEndpoint != nil {
			tAdapter.ServiceName = span.LocalEndpoint.ServiceName
		}

		tAdapter.OperationName = span.Name

		if span.ParentID != nil {
			tAdapter.ParentID = fmt.Sprintf("%x", uint64(*span.ParentID))
		}

		if span.TraceID.High != 0 {
			tAdapter.TraceID = fmt.Sprintf("%x%x", span.TraceID.High, span.TraceID.Low)
		} else {
			tAdapter.TraceID = fmt.Sprintf("%x", span.TraceID.Low)
		}

		tAdapter.SpanID = fmt.Sprintf("%x", uint64(span.ID))

		tAdapter.Status = trace.STATUS_OK
		for tag, _ := range span.Tags {
			if tag == "error" {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}

		if span.RemoteEndpoint != nil {
			if len(span.RemoteEndpoint.IPv4) != 0 {
				tAdapter.EndPoint = fmt.Sprintf("%s", span.RemoteEndpoint.IPv4)
			}

			if len(span.RemoteEndpoint.IPv6) != 0 {
				tAdapter.EndPoint = fmt.Sprintf("%s", span.RemoteEndpoint.IPv6)
			}
		}

		//tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		//if span.RemoteEndpoint == nil {
		//	if span.LocalEndpoint == nil {
		//		tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		//	} else {
		//		if len(span.LocalEndpoint.IPv4) == 0 && len(span.LocalEndpoint.IPv6) == 0 {
		//			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		//		}
		//	}
		//}
		if span.Kind == zipkinmodel.Undetermined {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		}

		tAdapter.Tags = ZipkinTags

		// run trace data sample
		if traceSampleConf.SampleFilter(tAdapter.Status, tAdapter.Tags, tAdapter.TraceID) {
			adapterGroup = append(adapterGroup, tAdapter)
		}
	}
	trace.MkLineProto(adapterGroup, inputName)

	return nil
}

func parseZipkinProtobufV2(octets []byte) error {
	spans, err := parseZipkinProtobuf3(octets)
	if err != nil {
		return err
	}

	adapterGroup := []*trace.TraceAdapter{}
	for _, span := range spans {
		tAdapter := &trace.TraceAdapter{}
		tAdapter.Source = "zipkin"

		tAdapter.Duration = int64(span.Duration)
		tAdapter.Start = span.Timestamp.UnixNano()
		sJson, err := json.Marshal(span)
		if err != nil {
			return err
		}
		tAdapter.Content = string(sJson)

		if span.LocalEndpoint != nil {
			tAdapter.ServiceName = span.LocalEndpoint.ServiceName
		}
		tAdapter.OperationName = span.Name

		if span.ParentID != nil {
			tAdapter.ParentID = fmt.Sprintf("%d", *span.ParentID)
		}

		if span.TraceID.High != 0 {
			tAdapter.TraceID = fmt.Sprintf("%d%d", span.TraceID.High, span.TraceID.Low)
		} else {
			tAdapter.TraceID = fmt.Sprintf("%d", span.TraceID.Low)
		}

		tAdapter.SpanID = fmt.Sprintf("%d", span.ID)

		tAdapter.Status = trace.STATUS_OK
		for tag, _ := range span.Tags {
			if tag == "error" {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}

		if span.RemoteEndpoint != nil {
			if len(span.RemoteEndpoint.IPv4) != 0 {
				tAdapter.EndPoint = fmt.Sprintf("%s", span.RemoteEndpoint.IPv4)
			}

			if len(span.RemoteEndpoint.IPv6) != 0 {
				tAdapter.EndPoint = fmt.Sprintf("%s", span.RemoteEndpoint.IPv6)
			}
		}

		//tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		//if span.RemoteEndpoint == nil {
		//	if span.LocalEndpoint == nil {
		//		tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		//	} else {
		//		if len(span.LocalEndpoint.IPv4) == 0 && len(span.LocalEndpoint.IPv6) == 0 {
		//			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		//		}
		//	}
		//}
		if span.Kind == zipkinmodel.Undetermined {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		}

		tAdapter.Tags = ZipkinTags

		// run trace data sample
		if traceSampleConf.SampleFilter(tAdapter.Status, tAdapter.Tags, tAdapter.TraceID) {
			adapterGroup = append(adapterGroup, tAdapter)
		}
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
