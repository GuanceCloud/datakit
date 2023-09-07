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
	httpRetry,
	groupedRequestVec *prometheus.CounterVec

	flushFailCacheVec,
	apiSumVec *prometheus.SummaryVec
)

// Metrics get all metrics aboud dataway.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,
		httpRetry,
		groupedRequestVec,
		flushFailCacheVec,
	}
}

func metricsReset() {
	ptsCounterVec.Reset()
	bytesCounterVec.Reset()
	apiSumVec.Reset()

	httpRetry.Reset()
	httpRetry.Reset()
	flushFailCacheVec.Reset()
	groupedRequestVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,

		flushFailCacheVec,
		httpRetry,
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

	groupedRequestVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "grouped_request_total",
			Help:      "Grouped requests under sinker",
		},
		[]string{
			"category",
		},
	)

	doRegister()
}
