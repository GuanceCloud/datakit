// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package zipkin handle Zipkin APM traces.
package zipkin

import (
	"encoding/json"
	"fmt"
	"strconv"

	zpkmodel "github.com/openzipkin/zipkin-go/model"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

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
			ParentID:   span.ParentID.String(),
			SpanID:     span.ID.String(),
			Service:    service,
			Resource:   span.Name,
			Operation:  span.Name,
			Source:     inputName,
			SpanType:   itrace.FindSpanTypeInMultiServersStrSpanID(span.ID.String(), span.ParentID.String(), service, spanIDs, parentIDs),
			SourceType: itrace.SPAN_SOURCE_CUSTOMER,
			Tags:       tags,
			Start:      span.Timestamp.UnixNano(),
			Duration:   int64(span.Duration),
			Status:     itrace.STATUS_OK,
		}

		if isRootSpan(dkspan.ParentID) {
			dkspan.ParentID = "0"
		}

		if span.TraceID.High != 0 {
			dkspan.TraceID = fmt.Sprintf("%x%x", span.TraceID.High, span.TraceID.Low)
		} else {
			dkspan.TraceID = strconv.FormatUint(span.TraceID.Low, 16)
		}

		for tag := range span.Tags {
			if tag == itrace.STATUS_ERR {
				dkspan.Status = itrace.STATUS_ERR
				break
			}
		}

		var err error
		if dkspan.Tags, err = itrace.MergeInToCustomerTags(tags, span.Tags, ignoreTags, nil); err != nil {
			log.Debug(err.Error())
		}
		if span.RemoteEndpoint != nil {
			if endpoint := span.RemoteEndpoint.IPv4.String(); len(endpoint) != 0 {
				dkspan.Tags[itrace.TAG_ENDPOINT] = endpoint
			} else if endpoint = span.RemoteEndpoint.IPv6.String(); len(endpoint) != 0 {
				dkspan.Tags[itrace.TAG_ENDPOINT] = endpoint
			}
		}

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
		if span.ParentID != nil {
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
