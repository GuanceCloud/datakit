// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	reg                     = prometheus.NewRegistry()
	traceHits               *prometheus.CounterVec
	TraceHitsByHTTPStatus   *prometheus.CounterVec
	traceLatency            *prometheus.HistogramVec
	traceErrors             *prometheus.CounterVec
	traceErrorsByHTTPStatus *prometheus.CounterVec
	traceApdex              *prometheus.HistogramVec
)

func addCollect(collect prometheus.Collector) {
	if err := reg.Register(collect); err != nil {
		log.Errorf("register prometheus collector error:%v", err)
	}
}

func initP8SMetrics(labels []string) {
	traceHits = itrace.NewTraceHits(labels)
	addCollect(traceHits)

	TraceHitsByHTTPStatus = itrace.NewTraceHitsByHTTPStatus(labels)
	addCollect(TraceHitsByHTTPStatus)

	traceLatency = itrace.NewTraceLatency(labels)
	addCollect(traceLatency)

	traceErrors = itrace.NewTraceErrors(labels)
	addCollect(traceErrors)

	traceErrorsByHTTPStatus = itrace.NewTraceErrorsByHTTPStatus(labels)
	addCollect(traceErrorsByHTTPStatus)

	traceApdex = itrace.NewTraceApdex()
	addCollect(traceApdex)
}

func reset() {
	traceHits.Reset()
	TraceHitsByHTTPStatus.Reset()
	traceLatency.Reset()
	traceErrors.Reset()
	traceErrorsByHTTPStatus.Reset()
	traceApdex.Reset()
}

func spanMetrics(span *point.Point, sourceLabels []string, values []string) {
	if len(sourceLabels) == 0 {
		return
	}
	isHTTP := false
	isError := false

	duration := span.Get(itrace.FieldDuration).(int64)
	for _, label := range sourceLabels {
		if label == itrace.TagHttpStatusCode {
			code := span.GetTag(itrace.TagHttpStatusCode)
			if code != "" {
				span.AddTag(itrace.TagHttpStatusClass, itrace.GetClass(fmt.Sprintf("%v", code)))
				isHTTP = true
			}
		}
		if label == itrace.TagSpanStatus {
			val := span.GetTag(label)
			if val == itrace.StatusErr {
				isError = true
			}
		}
		val := span.Get(label)
		if val != nil {
			values = append(values, fmt.Sprintf("%v", val))
		} else {
			values = append(values, "")
		}
	}

	resource := fmt.Sprintf("%v", span.Get(itrace.FieldResource))
	if isHTTP {
		TraceHitsByHTTPStatus.WithLabelValues(values...).Inc()
		if isError {
			traceErrorsByHTTPStatus.WithLabelValues(values...).Inc()
		}
		traceApdex.WithLabelValues(
			span.GetTag(itrace.TagService),
			span.GetTag(itrace.TagEnv),
			span.GetTag(itrace.TagVersion),
			resource,
			span.GetTag(itrace.TagSource),
			span.GetTag(itrace.TagRemoteIP),
		).Observe(float64(duration))
	}
	if isError {
		traceErrors.WithLabelValues(values...).Inc()
	}
	traceLatency.WithLabelValues(values...).Observe(float64(duration))

	traceHits.WithLabelValues(values...).Inc()
}
