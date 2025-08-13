// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package trace prom metrics
package trace

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TracingProcessCount *prometheus.CounterVec
	tracingDropVec      *prometheus.SummaryVec
	grpcPayloadSizeVec  *prometheus.SummaryVec
)

func metricsSetup() {
	TracingProcessCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "tracing_total",
			Help:      "The total links number of Trace processed by the trace module",
		},
		[]string{
			"input",
			"service",
		},
	)

	tracingDropVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "drop_number",
			Help:      "The drop number of Trace processed by the trace filter",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"input",
			"service",
			"reason",
		},
	)

	grpcPayloadSizeVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "grpc",
			Name:      "trace_payload_bytes",
			Help:      "The payload size of gRPC request send to DataKit",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"method",
		},
	)
}

func init() { //nolint:gochecknoinits
	metricsSetup()
	metrics.MustRegister(Metrics()...)
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		TracingProcessCount, tracingDropVec, grpcPayloadSizeVec,
	}
}
