// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	kvUpdateCount      prometheus.Counter
	kvPullLatencyVec   *prometheus.SummaryVec
	kvLastUpdate       prometheus.Gauge
	kvInputReloadCount prometheus.CounterVec
	kvInputLastReload  prometheus.GaugeVec

	setUlimitVec *prometheus.GaugeVec
)

func metricsSetup() {
	setUlimitVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "config",
			Name:      "datakit_ulimit",
			Help:      "Datakit ulimit",
		},
		[]string{"status"},
	)

	kvLastUpdate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "kv",
			Name:      "last_update_timestamp_seconds",
			Help:      "KV last update time",
		})

	kvPullLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "kv",
			Name:      "pull_latency_seconds",
			Help:      "KV pull latency",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"status"},
	)

	kvUpdateCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kv",
			Name:      "update_total",
			Help:      "KV updated count",
		},
	)

	kvInputLastReload = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "kv",
			Name:      "input_last_reload_timestamp_seconds",
			Help:      "Input last reload timestamp",
		},
		[]string{"source"},
	)

	kvInputReloadCount = *prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "kv",
			Name:      "input_reload_total",
			Help:      "Input reload count",
		},
		[]string{"source"},
	)
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		setUlimitVec,
		kvUpdateCount,
		kvPullLatencyVec,
		kvLastUpdate,
		kvInputLastReload,
		kvInputReloadCount,
	}
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister(Metrics()...)
}
