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
	apiCounterVec,
	ptsCounterVec,
	bytesCounterVec,
	sinkCounterVec,
	sinkPtsVec *prometheus.CounterVec

	flushFailCacheVec,
	apiSumVec *prometheus.SummaryVec
)

// Metrics get all metrics aboud dataway.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		apiCounterVec,
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,
		sinkCounterVec,
		sinkPtsVec,
		flushFailCacheVec,
	}
}

func metricsReset() {
	apiCounterVec.Reset()
	ptsCounterVec.Reset()
	bytesCounterVec.Reset()
	apiSumVec.Reset()

	sinkCounterVec.Reset()
	flushFailCacheVec.Reset()
	sinkPtsVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		apiCounterVec,
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,

		flushFailCacheVec,
		sinkCounterVec,
		sinkPtsVec,
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

	apiCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_api_request_total",
			Help:      "dataway HTTP request processed, partitioned by status code and HTTP API(url path)",
		},
		[]string{"api", "status"},
	)

	ptsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_point_total",
			Help:      "dataway uploaded points, partitioned by category and send status(HTTP status)",
		},
		[]string{"category", "status"},
	)

	bytesCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_point_bytes_total",
			Help:      "dataway uploaded points bytes, partitioned by category and pint send status(HTTP status)",
		},
		[]string{"category", "status"},
	)

	apiSumVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_api_latency",
			Help:      "dataway HTTP request latency(ms) partitioned by HTTP API(method@url) and HTTP status",
		},
		[]string{"api", "status"},
	)

	sinkCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_sink_total",
			Help:      "dataway sink count, partitioned by category.",
		},
		[]string{"category"},
	)

	sinkPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_sink_point_total",
			Help:      "dataway sink points, partitioned by category and point send status(ok/failed/dropped)",
		},
		[]string{"category", "status"},
	)

	doRegister()
}
