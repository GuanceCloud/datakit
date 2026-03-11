// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	sqlQueryCostSummary  *prometheus.SummaryVec
	dbmSQLQueryDuration  *prometheus.SummaryVec
	dbmObfuscateDuration *prometheus.SummaryVec
)

func metricsSetup() {
	sqlQueryCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_sqlserver",
			Name:      "sql_query_cost_seconds",
			Help:      "Time cost to query SQL",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"metric_name", "sql_name"},
	)

	dbmSQLQueryDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_sqlserver",
			Name:      "dbm_sql_query_duration_seconds",
			Help:      "Time cost to query database for DBM metrics",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"dbm_type", "sql_type"},
	)

	dbmObfuscateDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_sqlserver",
			Name:      "dbm_obfuscate_duration_seconds",
			Help:      "Time cost to obfuscate SQL text",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"dbm_type", "sql_type"}, // dbm_type: activity, statement; sql_type: statement, procedure
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister([]prometheus.Collector{
		sqlQueryCostSummary,
		dbmSQLQueryDuration,
		dbmObfuscateDuration,
	}...)
}
