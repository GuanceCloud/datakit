// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpcli

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP client trace metrics.
	httpClientTCPConn,
	httpClientConnReusedFromIdle *prometheus.CounterVec
	httpClientConnIdleTime,
	httpClientDNSCost,
	httpClientTLSHandshakeCost,
	httpClientConnectCost,
	httpClientGotFirstResponseByteCost *prometheus.SummaryVec
)

const (
	subsystem = "httpcli"
)

// nolint:gochecknoinits
func init() {
	httpClientTCPConn = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "tcp_conn_total",
			Help:      "HTTP TCP connection count",
		},
		[]string{"from", "remote", "type"},
	)

	httpClientConnReusedFromIdle = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "conn_reused_from_idle_total",
			Help:      "HTTP connection reused from idle count",
		},
		[]string{"from"},
	)

	httpClientConnIdleTime = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "conn_idle_time_seconds",
			Help:      "HTTP connection idle time",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"from"},
	)

	httpClientDNSCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "dns_cost_seconds",
			Help:      "HTTP DNS cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"from"},
	)

	httpClientTLSHandshakeCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "tls_handshake_seconds",
			Help:      "HTTP TLS handshake cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"from"},
	)

	httpClientConnectCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "http_connect_cost_seconds",
			Help:      "HTTP connect cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"from"},
	)

	httpClientGotFirstResponseByteCost = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: subsystem,
			Name:      "got_first_resp_byte_cost_seconds",
			Help:      "Got first response byte cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},

			MaxAge:     prometheus.DefMaxAge, // 10min
			AgeBuckets: prometheus.DefAgeBuckets,
		},
		[]string{"from"},
	)

	metrics.MustRegister(
		httpClientTCPConn,
		httpClientConnReusedFromIdle,
		httpClientConnIdleTime,
		httpClientDNSCost,
		httpClientTLSHandshakeCost,
		httpClientConnectCost,
		httpClientGotFirstResponseByteCost,
	)
}
