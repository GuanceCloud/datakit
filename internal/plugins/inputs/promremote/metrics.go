// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promremote

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	collectPointsTotalVec *p8s.SummaryVec
	httpTimeDiffVec       *p8s.SummaryVec
	noTimePointsVec       *p8s.CounterVec
)

func metricsSetup() {
	collectPointsTotalVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_promremote",
			Name:      "collect_points",
			Help:      "Total number of promremote collection points",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},
		},

		[]string{"source"},
	)

	httpTimeDiffVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_promremote",
			Name:      "time_diff_in_second",
			Help:      "Time diff with local time",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.90: 0.01,
				0.99: 0.001,
			},

			MaxAge:     p8s.DefMaxAge, // 10min
			AgeBuckets: p8s.DefAgeBuckets,
		},
		[]string{"source"},
	)

	noTimePointsVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_promremote",
			Name:      "no_time_points_total",
			Help:      "Total number of promremote collection no time points",
		},
		[]string{"source"},
	)

	metrics.MustRegister(
		collectPointsTotalVec,
		httpTimeDiffVec,
		noTimePointsVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
