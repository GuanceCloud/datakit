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
			Name:      "last_update",
			Help:      "filter last update time(in unix timestamp second)",
		})

	filterPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "point_total",

			Help: "Filter points of filters",
		},
		[]string{
			"category",
			"filters",
			"source",
		},
	)

	filterDroppedPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "point_dropped_total",

			Help: "Dropped points of filters",
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
			Name:      "pull_latency",

			Help: "Filter pull(remote) latency(ms)",
		},
		[]string{"status"},
	)

	filterLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "filter",
			Name:      "latency",

			Help: "Filter latency(us) of these filters",
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

			Help: "Filters(remote) updated count",
		},
	)

	metrics.MustRegister(
		filterDroppedPtsVec,
		filterPtsVec,
		lastUpdate,
		filterPullLatencyVec,
		filterLatencyVec,
		filtersUpdateCount,
	)
}
