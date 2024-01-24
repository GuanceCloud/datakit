// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ptsCounterVec,
	bytesCounterVec,
	writeDropPointsCounterVec,
	httpRetry *prometheus.CounterVec

	flushFailCacheVec,
	buildBodyCostVec,
	buildBodyBatchBytesVec,
	buildBodyBatchCountVec,
	groupedRequestVec,
	apiSumVec *prometheus.SummaryVec
)

func HTTPRetry() *prometheus.CounterVec {
	return httpRetry
}

func APISumVec() *prometheus.SummaryVec {
	return apiSumVec
}

// Metrics get all metrics aboud dataway.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		ptsCounterVec,
		bytesCounterVec,
		writeDropPointsCounterVec,
		apiSumVec,
		httpRetry,
		buildBodyCostVec,
		buildBodyBatchBytesVec,
		buildBodyBatchCountVec,
		groupedRequestVec,
		flushFailCacheVec,
	}
}

func metricsReset() {
	ptsCounterVec.Reset()
	bytesCounterVec.Reset()
	writeDropPointsCounterVec.Reset()
	apiSumVec.Reset()

	httpRetry.Reset()
	flushFailCacheVec.Reset()
	buildBodyCostVec.Reset()
	buildBodyBatchBytesVec.Reset()
	buildBodyBatchCountVec.Reset()
	groupedRequestVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		ptsCounterVec,
		bytesCounterVec,
		writeDropPointsCounterVec,
		apiSumVec,

		flushFailCacheVec,
		httpRetry,
		buildBodyCostVec,
		buildBodyBatchBytesVec,
		buildBodyBatchCountVec,
		groupedRequestVec,
	)
}

// nolint:gochecknoinits
func init() {
	flushFailCacheVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "flush_failcache_bytes",
			Help:      "IO flush fail-cache bytes(in gzip) summary",
		},
		[]string{"category"},
	)

	buildBodyCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_cost_seconds",
			Help:      "Build point HTTP body cost",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.75: 0.0075,
				0.95: 0.005,
			},
		},
		[]string{"category", "encoding"},
	)

	buildBodyBatchCountVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batches",
			Help:      "Batch HTTP body batches",
		},
		[]string{"category", "encoding"},
	)

	buildBodyBatchBytesVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batch_bytes",
			Help:      "Batch HTTP body size",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.75: 0.0075,
				0.95: 0.005,
			},
		},
		[]string{"category", "encoding", "gzip"},
	)

	ptsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_point_total",
			Help:      "Dataway uploaded points, partitioned by category and send status(HTTP status)",
		},
		[]string{"category", "status"},
	)

	bytesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_point_bytes_total",
			Help:      "Dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)",
		},
		[]string{"category", "enc", "status"},
	)

	writeDropPointsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_http_drop_point_total",
			Help:      "Dataway write drop points",
		},
		[]string{"category", "error"},
	)

	apiSumVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_api_latency_seconds",
			Help:      "Dataway HTTP request latency partitioned by HTTP API(method@url) and HTTP status",
		},
		[]string{"api", "status"},
	)

	httpRetry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_retry_total",
			Help:      "Dataway HTTP retried count",
		},
		[]string{"api", "status"},
	)

	groupedRequestVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "grouped_request",
			Help:      "Grouped requests under sinker",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
		},
	)

	doRegister()
}
