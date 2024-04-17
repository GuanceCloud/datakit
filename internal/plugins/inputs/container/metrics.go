// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	collectCostVec *prometheus.SummaryVec
	collectPtsVec  *prometheus.CounterVec
	totalCostVec   *prometheus.SummaryVec
)

func setupMetrics() {
	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "container_collect_cost_seconds",
			Help:      "Container collect cost",

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

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "container_collect_pts_total",
			Help:      "Container collect point total",
		},
		[]string{
			"category",
		},
	)

	totalCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "container_total_collect_cost_seconds",
			Help:      "Total container collect cost",
		},
		[]string{
			"category",
		},
	)

	metrics.MustRegister(
		collectCostVec,
		collectPtsVec,
		totalCostVec,
	)
}
