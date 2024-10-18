// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promv2

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var scrapeTotal *prometheus.SummaryVec

func setupMetrics() {
	scrapeTotal = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_promv2",
			Name:      "scrape_points",
			Help:      "The number of points scrape from endpoint",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"source",
			"remote",
		},
	)

	metrics.MustRegister(
		scrapeTotal,
	)
}
