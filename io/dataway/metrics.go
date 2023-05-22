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
	sinkCounterVec,
	httpRetry,
	notSinkPtsVec,
	httpTCPConn,
	sinkPtsVec *prometheus.CounterVec

	httpConnReusedFromIdle prometheus.Counter
	httpConnIdleTime       prometheus.Summary

	flushFailCacheVec,
	apiSumVec *prometheus.SummaryVec

	// HTTP trace metrics.
	httpDNSCost,
	httpTLSHandshakeCost,
	httpConnectCost,
	httpGotFirstResponseByteCost prometheus.Summary
)

// Metrics get all metrics aboud dataway.
func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,
		sinkCounterVec,
		httpRetry,
		notSinkPtsVec,
		sinkPtsVec,
		flushFailCacheVec,

		httpDNSCost,
		httpTLSHandshakeCost,
		httpConnectCost,
		httpGotFirstResponseByteCost,

		httpTCPConn,
		httpConnReusedFromIdle,
		httpConnIdleTime,
	}
}

func metricsReset() {
	ptsCounterVec.Reset()
	bytesCounterVec.Reset()
	apiSumVec.Reset()
	httpTCPConn.Reset()

	httpRetry.Reset()
	sinkCounterVec.Reset()
	httpRetry.Reset()
	flushFailCacheVec.Reset()
	notSinkPtsVec.Reset()
	sinkPtsVec.Reset()
}

func doRegister() {
	metrics.MustRegister(
		ptsCounterVec,
		bytesCounterVec,
		apiSumVec,

		flushFailCacheVec,
		sinkCounterVec,
		httpRetry,
		notSinkPtsVec,
		sinkPtsVec,

		httpDNSCost,
		httpTLSHandshakeCost,
		httpConnectCost,
		httpGotFirstResponseByteCost,

		httpTCPConn,
		httpConnReusedFromIdle,
		httpConnIdleTime,
	)
}

// nolint:gochecknoinits
func init() {
	httpTCPConn = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_tcp_conn_total",
			Help:      "Dataway HTTP TCP connection count",
		},
		[]string{
			"remote",
			"type",
		},
	)

	httpConnReusedFromIdle = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "datakit",
		Subsystem: "io",
		Name:      "http_conn_reused_from_idle_total",
		Help:      "Dataway HTTP connection reused from idle count",
	})

	httpConnIdleTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "datakit",
		Subsystem: "io",
		Name:      "http_conn_idle_time_seconds",
		Help:      "Dataway HTTP connection idle time",
	})

	httpDNSCost = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_dns_cost_seconds",
			Help:      "Dataway HTTP DNS cost",
		})

	httpTLSHandshakeCost = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_tls_handshake_seconds",
			Help:      "Dataway TLS handshake cost",
		})

	httpConnectCost = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_connect_cost_seconds",
			Help:      "Dataway HTTP connect cost",
		})

	httpGotFirstResponseByteCost = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "http_got_first_resp_byte_cost_seconds",
			Help:      "Dataway got first response byte cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.75: 0.0075,
				0.95: 0.005,
			},

			MaxAge:     prometheus.DefMaxAge, // 10min
			AgeBuckets: prometheus.DefAgeBuckets,
		})

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

	sinkCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_sink_total",
			Help:      "Dataway Sinked count, partitioned by category.",
		},
		[]string{
			"category",
		},
	)

	notSinkPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_not_sink_point_total",
			Help:      "Dataway not-Sinked points(condition or category not match)",
		},
		[]string{
			"category",
		},
	)

	sinkPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "io",
			Name:      "dataway_sink_point_total",
			Help:      "Dataway Sinked points, partitioned by category and point send status(ok/failed/dropped)",
		},
		[]string{"category", "status"},
	)

	doRegister()
}
