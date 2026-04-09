// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var otelMetricPoints = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace: "datakit",
		Subsystem: "input_opentelemetry",
		Name:      "metric_points",
		Help:      "Input opentelemetry collected metric points on single HTTP/gRPC request",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	},
	[]string{
		"type",
	},
)

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		otelMetricPoints,
	}
}

func doRegister() {
	metrics.MustRegister(
		otelMetricPoints,
	)
}

// nolint:gochecknoinits
func init() {
	doRegister()
}
