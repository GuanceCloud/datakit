// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	collectPtsVec    *prometheus.CounterVec
	collectCostVec   *prometheus.SummaryVec
	scraperNumberVec *prometheus.GaugeVec
	taskNumerVec     *prometheus.GaugeVec
)

func setupMetrics() {
	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "collect_pts_total",
			Help:      "The number of the points which have been sent",
		},
		[]string{
			"role",
			"name",
		},
	)

	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "collect_cost_seconds",
			Help:      "The collect cost in seconds",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"role",
			"name",
			"url",
		},
	)

	scraperNumberVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "scraper_number",
			Help:      "The number of the scraper",
		},
		[]string{
			"role",
			"name",
		},
	)

	taskNumerVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "task_number",
			Help:      "The number of the task",
		},
		[]string{
			"worker",
		},
	)

	metrics.MustRegister(
		collectPtsVec,
		collectCostVec,
		scraperNumberVec,
		taskNumerVec,
	)
}
