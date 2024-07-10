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
	tracingSamplerCount *prometheus.CounterVec
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

	tracingSamplerCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "sampler_total",
			Help:      "The sampler number of Trace processed by the trace module",
		},
		[]string{
			"input",
			"service",
		},
	)
}

func init() { //nolint:gochecknoinits
	metricsSetup()
	metrics.MustRegister(TracingProcessCount, tracingSamplerCount)
}
