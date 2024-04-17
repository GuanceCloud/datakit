// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package proxy

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	proxyReqVec        *prometheus.CounterVec
	proxyConnectVec    *prometheus.CounterVec
	proxyReqLatencyVec *prometheus.SummaryVec
)

func metricsSetup() {
	proxyConnectVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_proxy",
			Name:      "connect_total",
			Help:      "Proxy connect(method CONNECT)",
		},
		[]string{
			"client_ip",
		},
	)

	proxyReqVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_proxy",
			Name:      "api_total",
			Help:      "Proxy API total",
		},
		[]string{
			"api",
			"method",
		},
	)

	proxyReqLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_proxy",
			Name:      "api_latency_seconds",
			Help:      "Proxy API latency",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"api",
			"method",
			"status",
		},
	)
}

func allMetrics() []prometheus.Collector {
	return []prometheus.Collector{proxyReqVec, proxyReqLatencyVec, proxyConnectVec}
}

func resetMetrics() {
	proxyReqLatencyVec.Reset()
	proxyReqVec.Reset()
	proxyConnectVec.Reset()
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister(allMetrics()...)
}
