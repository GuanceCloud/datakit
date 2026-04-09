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
	walQueueMemLenVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_wal_mem_len",
			Help:      "Dataway WAL's memory queue length",
		},
		[]string{"category"},
	)

	flushDroppedPackageVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "flush_drop_pkg_total",
			Help:      "WAL flush dropped packages count due to expiration",
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

	walPointCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "wal_point_total",
			Help:      "WAL queued points",
		},
		[]string{"category", "status"},
	)

	walPutRetriedVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "wal_put_retried",
			Help:      "WAL put retried on disk full",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"category"},
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
)

// Metrics get all metrics aboud dataway.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		walWorkerFlush,
		walPointCounterVec,
		walPutRetriedVec,
		writeDropPointsCounterVec,
		groupedRequestVec,
		flushFailCacheVec,
		walQueueMemLenVec,
		flushDroppedPackageVec,
	}
}

func metricsReset() {
	walWorkerFlush.Reset()
	walPointCounterVec.Reset()
	walPutRetriedVec.Reset()
	writeDropPointsCounterVec.Reset()

	flushFailCacheVec.Reset()
	walQueueMemLenVec.Reset()
	flushDroppedPackageVec.Reset()
	groupedRequestVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		walWorkerFlush,
		walPointCounterVec,
		walPutRetriedVec,
		writeDropPointsCounterVec,

		flushFailCacheVec,
		walQueueMemLenVec,
		flushDroppedPackageVec,
		groupedRequestVec,
	)
}

// nolint:gochecknoinits
func init() {
	doRegister()
}
