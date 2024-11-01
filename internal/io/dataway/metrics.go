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
	skippedPointVec,
	bodyCounterVec,
	ptsCounterVec,
	bytesCounterVec,
	writeDropPointsCounterVec,
	walPointCounterVec,
	httpRetry *prometheus.CounterVec

	walWorkerFlush,
	flushFailCacheVec,
	buildBodyCostVec,
	buildBodyBatchBytesVec,
	buildBodyBatchPointsVec,
	buildBodyBatchCountVec,
	groupedRequestVec,
	apiSumVec *prometheus.SummaryVec

	walQueueMemLenVec *prometheus.GaugeVec
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
		skippedPointVec,
		walWorkerFlush,
		bodyCounterVec,
		ptsCounterVec,
		walPointCounterVec,
		bytesCounterVec,
		writeDropPointsCounterVec,
		apiSumVec,
		httpRetry,
		buildBodyCostVec,
		buildBodyBatchBytesVec,
		buildBodyBatchPointsVec,
		buildBodyBatchCountVec,
		groupedRequestVec,
		flushFailCacheVec,
		walQueueMemLenVec,
	}
}

func metricsReset() {
	skippedPointVec.Reset()
	walWorkerFlush.Reset()
	bodyCounterVec.Reset()
	ptsCounterVec.Reset()
	walPointCounterVec.Reset()
	bytesCounterVec.Reset()
	writeDropPointsCounterVec.Reset()
	apiSumVec.Reset()

	httpRetry.Reset()
	flushFailCacheVec.Reset()
	walQueueMemLenVec.Reset()
	buildBodyCostVec.Reset()
	buildBodyBatchBytesVec.Reset()
	buildBodyBatchPointsVec.Reset()
	buildBodyBatchCountVec.Reset()
	groupedRequestVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		skippedPointVec,
		walWorkerFlush,
		bodyCounterVec,
		ptsCounterVec,
		walPointCounterVec,
		bytesCounterVec,
		writeDropPointsCounterVec,
		apiSumVec,

		flushFailCacheVec,
		walQueueMemLenVec,
		httpRetry,
		buildBodyCostVec,
		buildBodyBatchBytesVec,
		buildBodyBatchPointsVec,
		buildBodyBatchCountVec,
		groupedRequestVec,
	)
}

// nolint:gochecknoinits
func init() {
	skippedPointVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_skipped_point_total",
			Help:      "Skipped point count during encoding(protobuf) point",
		},
		[]string{"category"},
	)

	walQueueMemLenVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_wal_mem_len",
			Help:      "Dataway WAL's memory queue length",
		},
		[]string{"category"},
	)

	flushFailCacheVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "flush_failcache_bytes",
			Help:      "IO flush fail-cache bytes(in gzip) summary",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
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
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding", "stage"},
	)

	buildBodyBatchCountVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batches",
			Help:      "Batch HTTP body batches",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
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
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding", "type"},
	)

	buildBodyBatchPointsVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "build_body_batch_points",
			Help:      "Batch HTTP body points",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category", "encoding"},
	)

	walWorkerFlush = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_wal_flush",
			Help:      "Dataway WAL worker flushed bytes",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
			"gzip",
			"queue", // from walqueue disk or mem
		},
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

	bodyCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_body_total",
			Help:      "Dataway total body",
		},
		[]string{
			"from",
			"op",
			"type",
		},
	)

	walPointCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "wal_point_total",
			Help:      "WAL queued points",
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

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
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
