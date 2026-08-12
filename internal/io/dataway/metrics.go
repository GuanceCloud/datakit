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
	datawayDialTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dataway",
			Name:      "dial_total",
			Help:      "Dataway TCP dial attempts partitioned by IP family and result",
		},
		[]string{"family", "result"},
	)

	datawayIPv4FallbackTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dataway",
			Name:      "ipv4_fallback_total",
			Help:      "Dataway IPv4 fallback attempts started after an IPv6 attempt",
		},
		[]string{},
	)

	datawayDialSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "datakit",
			Subsystem: "dataway",
			Name:      "dial_seconds",
			Help:      "Dataway TCP dial latency partitioned by IP family",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"family"},
	)

	datawayActiveConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dataway",
			Name:      "active_connections",
			Help:      "Current active Dataway TCP connections partitioned by IP family",
		},
		[]string{"family"},
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
			Help:      "IO flush fail-cache bytes summary",

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

	walWorkerFlushByCompression = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_wal_flush_bytes",
			Help:      "Dataway WAL worker flushed bytes by compression",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"category",
			"compression",
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
		datawayDialTotal,
		datawayIPv4FallbackTotal,
		datawayDialSeconds,
		datawayActiveConnections,
		walWorkerFlush,
		walWorkerFlushByCompression,
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
	datawayDialTotal.Reset()
	datawayIPv4FallbackTotal.Reset()
	datawayDialSeconds.Reset()
	datawayActiveConnections.Reset()
	initIPFamilyMetrics()

	walWorkerFlush.Reset()
	walWorkerFlushByCompression.Reset()
	walPointCounterVec.Reset()
	walPutRetriedVec.Reset()
	writeDropPointsCounterVec.Reset()

	flushFailCacheVec.Reset()
	walQueueMemLenVec.Reset()
	flushDroppedPackageVec.Reset()
	groupedRequestVec.Reset()
}

func doRegister() {
	initIPFamilyMetrics()
	metrics.MustRegister(
		datawayDialTotal,
		datawayIPv4FallbackTotal,
		datawayDialSeconds,
		datawayActiveConnections,
		walWorkerFlush,
		walWorkerFlushByCompression,
		walPointCounterVec,
		walPutRetriedVec,
		writeDropPointsCounterVec,

		flushFailCacheVec,
		walQueueMemLenVec,
		flushDroppedPackageVec,
		groupedRequestVec,
	)
}

func initIPFamilyMetrics() {
	for _, family := range []string{"ipv4", "ipv6"} {
		for _, result := range []string{"success", "failed", "canceled"} {
			datawayDialTotal.WithLabelValues(family, result).Add(0)
		}
		datawayDialSeconds.WithLabelValues(family)
		datawayActiveConnections.WithLabelValues(family).Set(0)
	}
	datawayIPv4FallbackTotal.WithLabelValues().Add(0)
}

// nolint:gochecknoinits
func init() {
	doRegister()
}
