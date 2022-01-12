package zipkin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	// nolint:staticcheck
	"github.com/golang/protobuf/proto"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	zipkin_proto3 "github.com/openzipkin/zipkin-go/proto/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func parseZipkinProtobuf3(octets []byte) (zss []*zipkinmodel.SpanModel, err error) {
	var listOfSpans zipkin_proto3.ListOfSpans
	if err := proto.Unmarshal(octets, &listOfSpans); err != nil {
		return nil, err
	}
	for _, zps := range listOfSpans.Spans {
		traceID, err := zipkinTraceIDFromHex(fmt.Sprintf("%x", zps.TraceId))
		if err != nil {
			return nil, fmt.Errorf("invalid TraceID: %w", err)
		}

		parentSpanID, _, err := zipkinSpanIDToModelSpanID(zps.ParentId)
		if err != nil {
			return nil, fmt.Errorf("invalid ParentID: %w", err)
		}
		spanIDPtr, spanIDBlank, err := zipkinSpanIDToModelSpanID(zps.Id) //nolint:stylecheck
		if err != nil {
			return nil, fmt.Errorf("invalid SpanID: %w", err)
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
			Duration:       time.Duration(zps.Duration) * time.Microsecond,
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

func zipkinSpanIDToModelSpanID(spanID []byte) (zid *zipkinmodel.ID, blank bool, err error) {
	if len(spanID) == 0 {
		return nil, true, nil
	}
	if len(spanID) != 8 {
		return nil, true, fmt.Errorf("has length %d yet wanted length 8", len(spanID))
	}

	// Converting [8]byte --> uint64
	u64 := binary.BigEndian.Uint64(spanID)
	zid_ := zipkinmodel.ID(u64)
	return &zid_, false, nil
}

func microsToTime(us uint64) time.Time {
	return time.Unix(0, int64(us*1e3)).UTC()
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

func protobufSpansToAdapters(zspans []*zipkinmodel.SpanModel,
	filters ...zipkinProtoBufV2SpansFilter) ([]*trace.TraceAdapter, error) {
	// run all filters
	for _, filter := range filters {
		if len(filter(zspans)) == 0 {
			return nil, nil
		}
	}

	var adapterGroup []*trace.TraceAdapter
	for _, span := range zspans {
		tAdapter := &trace.TraceAdapter{}
		tAdapter.Source = sourceZipkin

		tAdapter.Duration = int64(span.Duration)
		tAdapter.Start = span.Timestamp.UnixNano()
		sJSON, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		tAdapter.Content = string(sJSON)

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
		for tag := range span.Tags {
			if tag == statusErr {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}

		if span.RemoteEndpoint != nil {
			if len(span.RemoteEndpoint.IPv4) != 0 {
				tAdapter.EndPoint = span.RemoteEndpoint.IPv4.String()
			}

			if len(span.RemoteEndpoint.IPv6) != 0 {
				tAdapter.EndPoint = span.RemoteEndpoint.IPv6.String()
			}
		}

		if span.Kind == zipkinmodel.Undetermined {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		}
		tAdapter.Tags = zipkinTags

		adapterGroup = append(adapterGroup, tAdapter)
	}

	return adapterGroup, nil
}

func parseZipkinJSONV2(zspans []*zipkinmodel.SpanModel,
	filters ...zipkinJSONV2SpansFilter) ([]*trace.TraceAdapter, error) {
	// run all filters
	for _, filter := range filters {
		if len(filter(zspans)) == 0 {
			return nil, nil
		}
	}

	var adapterGroup []*trace.TraceAdapter
	for _, span := range zspans {
		tAdapter := &trace.TraceAdapter{}
		tAdapter.Source = sourceZipkin

		tAdapter.Duration = int64(span.Duration)
		tAdapter.Start = span.Timestamp.UnixNano()
		sJSON, err := json.Marshal(span)
		if err != nil {
			return nil, err
		}
		tAdapter.Content = string(sJSON)

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
		for tag := range span.Tags {
			if tag == statusErr {
				tAdapter.Status = trace.STATUS_ERR
				break
			}
		}

		if span.RemoteEndpoint != nil {
			if len(span.RemoteEndpoint.IPv4) != 0 {
				tAdapter.EndPoint = span.RemoteEndpoint.IPv4.String()
			}

			if len(span.RemoteEndpoint.IPv6) != 0 {
				tAdapter.EndPoint = span.RemoteEndpoint.IPv6.String()
			}
		}

		if span.Kind == zipkinmodel.Undetermined {
			tAdapter.SpanType = trace.SPAN_TYPE_LOCAL
		} else {
			tAdapter.SpanType = trace.SPAN_TYPE_ENTRY
		}

		tAdapter.Tags = zipkinTags

		adapterGroup = append(adapterGroup, tAdapter)
	}

	return adapterGroup, nil
}
