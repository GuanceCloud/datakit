// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdatrace

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

type PointSink struct {
	inputName string
	feeder    dkio.Feeder
	baseTags  map[string]string
}

type pointMessage struct {
	Meta    map[string]string  `json:"meta"`
	Metrics map[string]float64 `json:"metrics"`
}

func NewPointSink(inputName string, feeder dkio.Feeder, baseTags map[string]string) *PointSink {
	cloned := map[string]string{}
	for k, v := range baseTags {
		cloned[k] = v
	}
	return &PointSink{
		inputName: inputName,
		feeder:    feeder,
		baseTags:  cloned,
	}
}

func (s *PointSink) Consume(_ context.Context, spans []Span) error {
	pts := make([]*point.Point, 0, len(spans))
	for _, span := range spans {
		pts = append(pts, s.toPoint(span))
	}
	return s.feeder.Feed(point.Tracing, pts,
		dkio.WithSyncSend(true),
		dkio.WithElection(false),
		dkio.WithSource(s.inputName))
}

func (s *PointSink) toPoint(span Span) *point.Point {
	kvs := make(point.KVs, 0, 16+len(s.baseTags)+len(span.Meta))
	kvs = kvs.Add(itrace.FieldTraceID, strconv.FormatUint(span.TraceID, 10)).
		Add(itrace.FieldParentID, strconv.FormatUint(span.ParentID, 10)).
		Add(itrace.FieldSpanid, strconv.FormatUint(span.SpanID, 10)).
		Add(itrace.FieldResource, span.Resource).
		Add(itrace.FieldStart, span.Start/int64(time.Microsecond)).
		Add(itrace.FieldDuration, span.Duration/int64(time.Microsecond)).
		AddTag(itrace.TagService, span.Service).
		AddTag(itrace.TagOperation, span.Name).
		AddTag(itrace.TagSource, s.inputName).
		AddTag(itrace.TagSpanType, spanType(span)).
		AddTag(itrace.TagSourceType, sourceType(span)).
		AddTag(itrace.TagSpanStatus, spanStatus(span))

	for k, v := range s.baseTags {
		kvs = kvs.AddTag(k, v)
	}
	for k, v := range span.Meta {
		if v != "" {
			kvs = kvs.AddTag(k, v)
		}
	}
	if len(span.Meta) > 0 || len(span.Metrics) > 0 {
		if msg, err := json.Marshal(pointMessage{Meta: span.Meta, Metrics: span.Metrics}); err == nil {
			kvs = kvs.Add(itrace.FieldMessage, string(msg))
		}
	}

	opts := append(point.CommonLoggingOptions(), point.WithTime(time.Unix(0, span.Start)))
	return point.NewPoint(s.inputName, kvs, opts...)
}

func spanStatus(span Span) string {
	if span.Error != 0 {
		return itrace.StatusErr
	}
	return itrace.StatusOk
}

func spanType(span Span) string {
	switch span.Name {
	case "aws.lambda":
		if span.ParentID == 0 {
			return itrace.SpanTypeEntry
		}
		return itrace.SpanTypeLocal
	case "aws.lambda.cold_start", "aws.lambda.snapstart_restore":
		return itrace.SpanTypeEntry
	default:
		return itrace.SpanTypeEntry
	}
}

func sourceType(span Span) string {
	switch span.Meta["trigger"] {
	case "api-gateway-rest", "api-gateway-http", "lambda-function-url", "alb":
		return itrace.SpanSourceWeb
	case "sqs", "sns", "eventbridge", "kinesis", "dynamodb", "msk", "step-functions", "s3":
		return itrace.SpanSourceMsgque
	default:
		return itrace.SpanSourceCustomer
	}
}
