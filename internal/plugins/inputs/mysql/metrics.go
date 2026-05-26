// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var sqlQueryCostSummary *prometheus.SummaryVec

func metricsSetup() {
	sqlQueryCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_mysql",
			Name:      "sql_query_cost_seconds",
			Help:      "Time cost to query SQL",

			Objectives: datakit.P8sStandardObjectives,
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
