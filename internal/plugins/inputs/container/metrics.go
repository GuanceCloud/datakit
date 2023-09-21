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
)

func setupMetrics() {
	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "container",
			Name:      "collect_cost_seconds",
			Help:      "Container collect cost",
		},
		[]string{
			"category",
		},
	)

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "container",
			Name:      "collect_pts_total",
			Help:      "Container collect point total",
		},
		[]string{
			"category",
		},
	)

	metrics.MustRegister(
		collectCostVec,
		collectPtsVec,
	)
}
