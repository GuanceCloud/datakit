// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var sqlQueryCostSummary *prometheus.SummaryVec

func metricsSetup() {
	sqlQueryCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_postgresql",
			Name:      "sql_query_cost_seconds",
			Help:      "Time cost to query sql",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"metric_name", "sql_name"},
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister([]prometheus.Collector{
		sqlQueryCostSummary,
	}...)
}
