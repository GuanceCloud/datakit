// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
	zpkmodel "github.com/openzipkin/zipkin-go/model"
	zpkprotov2 "github.com/openzipkin/zipkin-go/proto/v2"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func parseZipkinProtobuf3(body []byte) (zss []*zpkmodel.SpanModel, err error) {
	var listOfSpans zpkprotov2.ListOfSpans
	if err := proto.Unmarshal(body, &listOfSpans); err != nil {
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

		zmsc := zpkmodel.SpanContext{
			TraceID:  traceID,
			ID:       *spanIDPtr,
			ParentID: parentSpanID,
			Debug:    false,
		}
		zms := &zpkmodel.SpanModel{
			SpanContext:    zmsc,
			Name:           zps.Name,
			Kind:           zpkmodel.Kind(zps.Kind.String()),
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

func zipkinTraceIDFromHex(h string) (t zpkmodel.TraceID, err error) {
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

func zipkinSpanIDToModelSpanID(spanID []byte) (zid *zpkmodel.ID, blank bool, err error) {
	if len(spanID) == 0 {
		return nil, true, nil
	}
	if len(spanID) != 8 {
		return nil, true, fmt.Errorf("has length %d yet wanted length 8", len(spanID))
	}

	// Converting [8]byte --> uint64
	u64 := binary.BigEndian.Uint64(spanID)
	zid_ := zpkmodel.ID(u64)

	return &zid_, false, nil
}

func microsToTime(us uint64) time.Time {
	return time.Unix(0, int64(us*1e3)).UTC()
}

func protoEndpointToModelEndpoint(zpe *zpkprotov2.Endpoint) *zpkmodel.Endpoint {
	if zpe == nil {
		return nil
	}

	return &zpkmodel.Endpoint{
		ServiceName: zpe.ServiceName,
		IPv4:        net.IP(zpe.Ipv4),
		IPv6:        net.IP(zpe.Ipv6),
		Port:        uint16(zpe.Port),
	}
}

func protoAnnotationsToModelAnnotations(zpa []*zpkprotov2.Annotation) (zma []zpkmodel.Annotation) {
	for _, za := range zpa {
		if za != nil {
			zma = append(zma, zpkmodel.Annotation{
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

func spanModeleV2ToDkTrace(zpktrace []*zpkmodel.SpanModel) itrace.DatakitTrace {
	var (
		dktrace            itrace.DatakitTrace
		parentIDs, spanIDs = gatherSpanModelsInfo(zpktrace)
	)
	for _, span := range zpktrace {
		if span == nil {
			continue
		}

		if span.ParentID == nil {
			span.ParentID = new(zpkmodel.ID)
		}
		service := getServiceFromSpanModel(span)
		dkspan := &itrace.DatakitSpan{
			ParentID:  span.ParentID.String(),
			SpanID:    span.ID.String(),
			Service:   service,
			Resource:  span.Name,
			Operation: span.Name,
			Source:    inputName,
			SpanType:  itrace.FindSpanTypeInMultiServersStrSpanID(span.ID.String(), span.ParentID.String(), service, spanIDs, parentIDs),
			Status:    itrace.STATUS_OK,
			Start:     span.Timestamp.UnixNano(),
			Duration:  int64(span.Duration),
			Tags:      tags,
		}

		if isRootSpan(dkspan.ParentID) {
			dkspan.ParentID = "0"
		}

		if span.TraceID.High != 0 {
			dkspan.TraceID = fmt.Sprintf("%x%x", span.TraceID.High, span.TraceID.Low)
		} else {
			dkspan.TraceID = fmt.Sprintf("%x", span.TraceID.Low)
		}

		for tag := range span.Tags {
			if tag == itrace.STATUS_ERR {
				dkspan.Status = itrace.STATUS_ERR
				break
			}
		}

		if span.RemoteEndpoint != nil {
			if len(span.RemoteEndpoint.IPv4) != 0 {
				dkspan.EndPoint = span.RemoteEndpoint.IPv4.String()
			}
			if len(span.RemoteEndpoint.IPv6) != 0 {
				dkspan.EndPoint = span.RemoteEndpoint.IPv6.String()
			}
		}

		dkspan.Tags = itrace.MergeInToCustomerTags(customerKeys, tags, span.Tags)

		if buf, err := json.Marshal(span); err != nil {
			log.Warn(err.Error())
		} else {
			dkspan.Content = string(buf)
		}

		dktrace = append(dktrace, dkspan)
	}
	if len(dktrace) != 0 {
		dktrace[0].Metrics = make(map[string]interface{})
		dktrace[0].Metrics[itrace.FIELD_PRIORITY] = itrace.PRIORITY_AUTO_KEEP
	}

	return dktrace
}

func gatherSpanModelsInfo(trace []*zpkmodel.SpanModel) (parentIDs map[string]bool, spanIDs map[string]string) {
	parentIDs = make(map[string]bool)
	spanIDs = make(map[string]string)
	for _, span := range trace {
		if span == nil {
			continue
		}
		spanIDs[span.ID.String()] = getServiceFromSpanModel(span)
		if span.ParentID != nil && span.ParentID.String() != "0" {
			parentIDs[span.ParentID.String()] = true
		}
	}

	return
}

func getServiceFromSpanModel(span *zpkmodel.SpanModel) string {
	if span.LocalEndpoint != nil {
		return span.LocalEndpoint.ServiceName
	} else {
		return "zipkin_v2_unknow_service"
	}
}
