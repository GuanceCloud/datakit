// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var collectDurationVec *prometheus.SummaryVec

func setupMetrics() {
	collectDurationVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_jvm",
			Name:      "collect_duration_seconds",
			Help:      "JVM input collect duration in seconds",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"url",
		},
	)

	metrics.MustRegister(
		collectDurationVec,
	)
}

//nolint:gochecknoinits
func init() {
	setupMetrics()
}
