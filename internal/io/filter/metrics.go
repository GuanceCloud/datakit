// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	filterDroppedPtsVec,
	filterPtsVec *prometheus.CounterVec

	filtersUpdateCount prometheus.Counter

	filterPullLatencyVec,
	filterLatencyVec *prometheus.SummaryVec

	lastUpdate prometheus.Gauge

	filterParseErrorVec *prometheus.GaugeVec
)

const (
	sourceLocal  = "local"
	sourceRemote = "remote"
)

func setupMetrics() {
	lastUpdate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "last_update_timestamp_seconds",
			Help:      "Filter last update time",
		})

	filterPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "point_total",
			Help:      "Filter points of filters",
		},
		[]string{
			"category",
			"filters",
			"source",
		},
	)

	filterParseErrorVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "parse_error",
			Help:      "Filter parse error",
		},
		[]string{
			"error",
			"filters",
		},
	)

	filterDroppedPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "point_dropped_total",
			Help:      "Dropped points of filters",
		},
		[]string{
			"category",
			"filters",
			"source",
		},
	)

	filterPullLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "pull_latency_seconds",
			Help:      "Filter pull(remote) latency",
		},
		[]string{"status"},
	)

	filterLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "latency_seconds",
			Help:      "Filter latency of these filters",
		},
		[]string{
			"category",
			"filters",
			"source",
		},
	)

	filtersUpdateCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "update_total",
			Help:      "Filters(remote) updated count",
		},
	)

	metrics.MustRegister(
		filterDroppedPtsVec,
		filterPtsVec,
		filterParseErrorVec,
		lastUpdate,
		filterPullLatencyVec,
		filterLatencyVec,
		filtersUpdateCount,
	)
}
