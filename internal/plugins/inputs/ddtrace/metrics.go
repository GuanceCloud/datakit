// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	droppedTraces,
	truncatedTraceSpans *p8s.CounterVec
	traceSpans         *p8s.SummaryVec
	proxyTelemetryBody *p8s.SummaryVec
)

func metricsSetup() {
	truncatedTraceSpans = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_ddtrace",
			Name:      "truncated_spans_total",
			Help:      "Truncated trace spans",
		},
		[]string{"input"},
	)

	droppedTraces = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_ddtrace",
			Name:      "dropped_trace_total",
			Help:      "Dropped illegal traces",
		},
		[]string{"url"},
	)

	traceSpans = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_ddtrace",
			Name:      "trace_spans",
			Help:      "Trace spans(include truncated spans)",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"input"},
	)
	proxyTelemetryBody = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_ddtrace",
			Name:      "proxy_body",
			Help:      "Body length of DDTrace route:/telemetry/proxy/api/v2/apmtelemetry data",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"service"},
	)
}

// nolint: gochecknoinits
func init() {
	metricsSetup()

	metrics.MustRegister(
		truncatedTraceSpans,
		droppedTraces,
		traceSpans,
		proxyTelemetryBody,
	)
}
