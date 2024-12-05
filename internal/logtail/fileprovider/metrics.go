// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	scanCostVec  *prometheus.SummaryVec
	scanTotalVec *prometheus.SummaryVec
)

func setupMetrics() {
	scanCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_tailer_scanner",
			Name:      "cost_seconds",
			Help:      "Scanning costs seconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},
		},
		[]string{"pattern"},
	)

	scanTotalVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_tailer_scanner",
			Name:      "files",
			Help:      "Total number of scanned files",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},
		},
		[]string{"pattern"},
	)

	metrics.MustRegister(
		scanCostVec,
		scanTotalVec,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
