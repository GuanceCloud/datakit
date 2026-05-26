// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/prometheus/client_golang/prometheus"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

/*
指标
trace_hits
trace_hits_by_http_status
trace_latency
trace_errors
trace_errors_by_http_status
trace_apdex
*/

var (
	reg                     = prometheus.NewRegistry()
	traceHits               *prometheus.CounterVec
	TraceHitsByHTTPStatus   *prometheus.CounterVec
	traceLatency            *prometheus.HistogramVec
	traceErrors             *prometheus.CounterVec
	traceErrorsByHTTPStatus *prometheus.CounterVec
	traceApdex              *prometheus.HistogramVec

	labels []string
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
			if val := span.GetTag(label); val != "" {
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
	if len(sourceLabels) != len(values) {
		log.Errorf("spanMetrics: len(sourceLabels) != len(values)")
		return
	}

	resource := span.Get(itrace.FieldResource).(string)
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
		).Observe(float64(duration))
	}
	if isError {
		traceErrors.WithLabelValues(values...).Inc()
	}
	traceLatency.WithLabelValues(values...).Observe(float64(duration))

	traceHits.WithLabelValues(values...).Inc()
}

func traceMetric(trace DDTrace, sourceLabels []string, values []string) {
	if len(sourceLabels) == 0 {
		return
	}

	for _, span := range trace {
		isHTTP := false
		isError := false
		values = values[:0]
		// other tag from meta
		for _, label := range sourceLabels {
			switch label {
			case itrace.TagOperation:
				values = append(values, span.GetName())
			case itrace.TagService:
				values = append(values, span.GetService())
			case itrace.FieldResource:
				values = append(values, strings.ReplaceAll(span.Resource, "\n", " "))
			case "source":
				values = append(values, inputName)
			case itrace.TagSpanStatus:
				if span.Error != 0 {
					isError = true
					values = append(values, itrace.StatusErr)
				} else {
					values = append(values, itrace.StatusOk)
				}
			default:
				// 判断在不在meta中  如果不在，给一个空值
				meta := span.GetMeta()
				val := ""
				for key, repKey := range ddTags {
					if repKey == label {
						if x, ok := meta[key]; ok {
							val = x
						}
					}
					if repKey == itrace.TagHttpStatusCode {
						isHTTP = true
					}
				}
				values = append(values, val)
			}
		}
		// 确保sourceLabels和values长度一致
		if len(sourceLabels) != len(values) {
			log.Errorf("traceMetric: len(sourceLabels):%d != len(values):%d", len(sourceLabels), len(values))
			return
		}

		if isHTTP {
			TraceHitsByHTTPStatus.WithLabelValues(values...).Inc()
			if isError {
				traceErrorsByHTTPStatus.WithLabelValues(values...).Inc()
			}
			traceApdex.WithLabelValues(
				span.Service, span.Meta["env"], span.Meta["version"], span.Resource, inputName,
			).Observe(float64(span.Duration))
		}

		if isError {
			traceErrors.WithLabelValues(values...).Inc()
		}
		traceLatency.WithLabelValues(values...).Observe(float64(span.Duration))

		traceHits.WithLabelValues(values...).Inc()
	}
}
