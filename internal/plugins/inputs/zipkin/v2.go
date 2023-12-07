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

	"github.com/GuanceCloud/cliutils/point"
	zpkmodel "github.com/openzipkin/zipkin-go/model"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var traceOpts = []point.Option{}

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
		spanKV := point.KVs{}
		spanKV = spanKV.
			Add(itrace.FieldParentID, span.ParentID.String(), false, false).
			Add(itrace.FieldSpanid, span.ID.String(), false, false).
			AddTag(itrace.TagService, service).
			Add(itrace.FieldResource, span.Name, false, false).
			AddTag(itrace.TagOperation, span.Name).
			AddTag(itrace.TagSpanType, itrace.FindSpanTypeInMultiServersStrSpanID(span.ID.String(), span.ParentID.String(), service, spanIDs, parentIDs)).
			AddTag(itrace.TagSource, inputName).
			AddTag(itrace.TagSourceType, itrace.SpanSourceCustomer).
			Add(itrace.FieldStart, span.Timestamp.UnixMicro(), false, false).
			Add(itrace.FieldDuration, int64(span.Duration)/1000, false, true).
			AddTag(itrace.TagSpanStatus, itrace.StatusOk)

		if isRootSpan(span.ParentID.String()) {
			spanKV = spanKV.Add(itrace.FieldParentID, "0", false, true)
		}

		if span.TraceID.High != 0 {
			spanKV = spanKV.Add(itrace.FieldTraceID, fmt.Sprintf("%x%x", span.TraceID.High, span.TraceID.Low), false, true)
		} else {
			spanKV = spanKV.Add(itrace.FieldTraceID, strconv.FormatUint(span.TraceID.Low, 16), false, true)
		}

		for tag := range span.Tags {
			if tag == itrace.StatusErr {
				spanKV = spanKV.MustAddTag(itrace.TagSpanStatus, itrace.StatusErr)
				break
			}
		}

		if mTags, err := itrace.MergeInToCustomerTags(tags, span.Tags, ignoreTags); err == nil {
			for k, v := range mTags {
				spanKV = spanKV.AddTag(k, v)
			}
		}
		if !delMessage {
			if buf, err := json.Marshal(span); err != nil {
				log.Warn(err.Error())
			} else {
				spanKV = spanKV.Add(itrace.FieldMessage, string(buf), false, false)
			}
		}

		pt := point.NewPointV2(inputName, spanKV, traceOpts...)
		dktrace = append(dktrace, &itrace.DkSpan{Point: pt})
	}
	if len(dktrace) != 0 {
		dktrace[0].MustAdd(itrace.FieldPriority, itrace.PriorityAutoKeep)
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
