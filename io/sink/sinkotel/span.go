// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkotel is for opentemetry
package sinkotel

import (
	"context"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	tracev1 "go.opentelemetry.io/otel/trace"
)

//nolint:stylecheck
const (
	// span status.
	STATUS_OK   = "ok"
	STATUS_INFO = "info"
	STATUS_WARN = "warning"
	STATUS_ERR  = "error"

	// line protocol tags.
	TAG_ENDPOINT    = "endpoint"
	TAG_OPERATION   = "operation"
	TAG_PROJECT     = "project"
	TAG_SERVICE     = "service"
	TAG_SOURCE_TYPE = "source_type"
	TAG_SPAN_STATUS = "status"
	TAG_SPAN_TYPE   = "span_type"
	TAG_VERSION     = "version"
	// line protocol fields.
	FIELD_DURATION           = "duration"
	FIELD_MSG                = "message"
	FIELD_PARENTID           = "parent_id"
	FIELD_PID                = "pid"
	FIELD_PRIORITY           = "priority"
	FIELD_RESOURCE           = "resource"
	FIELD_SAMPLE_RATE_GLOBAL = "sample_rate_global"
	FIELD_SPANID             = "span_id"
	FIELD_START              = "start"
	FIELD_TRACEID            = "trace_id"

	traceIDLen = 32
	spanIDLen  = 16

	defaultSpanID  = "0000000000000000"
	defaultTraceID = "00000000000000000000000000000000"
)

func pointToTrace(pts []sinkcommon.ISinkPoint) (roSpans []tracesdk.ReadOnlySpan) {
	if len(pts) == 0 {
		return nil
	}
	spans := tracetest.SpanStubs{}
	for _, point := range pts {
		pointJSON, err := point.ToJSON()
		if err != nil {
			continue
		}
		tags := pointJSON.Tags
		fields := pointJSON.Fields

		spanStub := setTagToSpan(tags)
		setFieldToSpan(fields, &spanStub)
		spans = append(spans, spanStub)
	}
	return spans.Snapshots()
}

func setFieldToSpan(fields map[string]interface{}, t *tracetest.SpanStub) {
	var duration int64
	var serviceName string
	var traceID tracev1.TraceID
	var spanID tracev1.SpanID
	var parentID tracev1.SpanID
	for key, i := range fields {
		switch key {
		case FIELD_DURATION: // duraion
			startTime, ok := i.(int64)
			if ok {
				duration = startTime * int64(time.Microsecond)
			}
		case FIELD_MSG:
		case FIELD_PARENTID:
			var err error
			id, ok := i.(string)
			traceStr := ""
			if ok {
				traceStr = id
			}
			if traceStr != "0" {
				parentID, err = tracev1.SpanIDFromHex(hexString8(traceStr))
				if err != nil {
					l.Errorf("trace id err =%v", err)
				}
			}
		case FIELD_PID:
		case FIELD_PRIORITY:
		case FIELD_RESOURCE:
			resourceF, ok := i.(string)
			if ok {
				serviceName = resourceF
			}
		case FIELD_SAMPLE_RATE_GLOBAL:
		case FIELD_SPANID:
			var err error
			id, ok := i.(string)
			var traceStr string
			if ok {
				traceStr = id
			}
			if id == "0" {
				spanID, err = tracev1.SpanIDFromHex(defaultSpanID)
				if err != nil {
					l.Errorf("trace id err =%v", err)
				}
			} else {
				spanID, err = tracev1.SpanIDFromHex(hexString8(traceStr))
				if err != nil {
					l.Errorf("trace id err =%v", err)
				}
			}
		case FIELD_START:
			startTime, ok := i.(int64)
			if ok {
				t.StartTime = time.Unix(0, startTime*int64(time.Microsecond))
			}
		case FIELD_TRACEID:
			var err error
			id, ok := i.(string)
			var traceStr string
			if ok && id != "0" {
				traceStr = id
				traceID, err = tracev1.TraceIDFromHex(hexString16(traceStr))
				if err != nil {
					l.Errorf("trace id err =%v", err)
				}
			}
		default:
			l.Debugf("other field key:%s", key)
		}
	}
	t.EndTime = time.Unix(0, t.StartTime.UnixNano()+duration)
	if t.Resource == nil {
		res, err := resource.New(context.Background(),
			resource.WithAttributes(
				// the service name used to display traces in backends
				semconv.ServiceNameKey.String(serviceName),
			),
			// resource.WithFromEnv(), // service name or service attributes
		)
		if err != nil {
			l.Errorf("resource err =%v", err)
		}
		t.Resource = res
	}

	spanCxt := tracev1.NewSpanContext(tracev1.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: 0,
		TraceState: tracev1.TraceState{},
		Remote:     false,
	})
	if parentID.IsValid() {
		parentCxt := tracev1.NewSpanContext(tracev1.SpanContextConfig{
			SpanID: parentID,
		})
		t.Parent = parentCxt
	}

	t.SpanContext = spanCxt
}

func hexString16(s string) string {
	if len(s) < traceIDLen {
		dst := defaultTraceID
		dst = dst[:traceIDLen-len(s)] + s
		return dst
	}
	return s
}

func hexString8(s string) string {
	if len(s) < spanIDLen {
		dst := defaultSpanID
		dst = dst[:spanIDLen-len(s)] + s
		return dst
	}
	return s
}

func setTagToSpan(tags map[string]string) (stub tracetest.SpanStub) {
	for key, val := range tags {
		switch key {
		case TAG_ENDPOINT:
			if val != "null" {
				stub.Attributes = append(stub.Attributes, attribute.String(key, val))
			}
		case TAG_OPERATION:
			stub.Name = val
		case TAG_PROJECT:
			stub.Attributes = append(stub.Attributes, attribute.String(key, val))
		case TAG_SERVICE:
			if val != "" {
				stub.Resource = resource.NewWithAttributes("", attribute.String("service.name", val))
			}
		case TAG_SOURCE_TYPE:
		case TAG_SPAN_STATUS:
			var code codes.Code
			switch val {
			case STATUS_OK, STATUS_INFO:
				code = 2
			case STATUS_ERR, STATUS_WARN:
				code = 1
			default:
				code = 0
			}
			stub.Status = tracesdk.Status{Code: code}
		case TAG_SPAN_TYPE:
		case TAG_VERSION:
		default:
			// 剩余的其他字段
			stub.Attributes = append(stub.Attributes, attribute.String(key, val))
		}
	}
	return stub
}
