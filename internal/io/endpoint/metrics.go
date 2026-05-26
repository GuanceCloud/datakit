// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package endpoint

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	bytesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "endpoint_point_bytes_total",
			Help:      "Uploaded points bytes, partitioned by category and pint send status(HTTP status)",
		},
		[]string{
			"category",
			"enc",
			"owner",
			"status",
		},
	)

	ptsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "endpoint_point_total",
			Help:      "Uploaded points, partitioned by category and send status(HTTP status)",
		},
		[]string{
			"category",
			"owner",
			"status",
		},
	)

	apiSumVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "endpoint_api_latency_seconds",
			Help:      "HTTP request latency partitioned by HTTP API(method@url) and HTTP status",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"api",
			"owner",
			"status",
		},
	)

	httpRetry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_retry_total",
			Help:      "HTTP retried count",
		},
		[]string{
			"api",
			"owner",
			"status",
		},
	)
)

func HTTPRetry() *prometheus.CounterVec {
	return httpRetry
}

func APISumVec() *prometheus.SummaryVec {
	return apiSumVec
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,
		httpRetry,
	}
}

func doRegister() {
	metrics.MustRegister(Metrics()...)
}

func MetricsReset() {
	ptsCounterVec.Reset()
	bytesCounterVec.Reset()
	apiSumVec.Reset()
	httpRetry.Reset()
}

// nolint:gochecknoinits
func init() {
	doRegister()
}
