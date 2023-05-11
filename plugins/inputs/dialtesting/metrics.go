// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/GuanceCloud/cliutils/metrics"
)

var (
	taskGauge               *prometheus.GaugeVec
	taskPullCostSummary     *prometheus.SummaryVec
	taskSynchronizedCounter *prometheus.CounterVec
	taskCheckCostSummary    *prometheus.SummaryVec
	taskRunCostSummary      *prometheus.SummaryVec
	taskInvalidCounter      *prometheus.CounterVec
)

func metricsSetup() {
	taskGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_number",
			Help:      "the number of tasks",
		},
		[]string{"region", "protocol"},
	)

	taskPullCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "pull_cost",
			Help:      "time cost to pull tasks(in nanosecond)",
		},
		[]string{"region", "is_first"},
	)

	taskSynchronizedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_synchronized_total",
			Help:      "task synchronized number",
		},
		[]string{"region", "protocol"},
	)

	taskInvalidCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_invalid_total",
			Help:      "invalid task number",
		},
		[]string{"region", "protocol", "fail_reason"},
	)

	taskCheckCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_check_cost",
			Help:      "task check time(in nanosecond)",
		},
		[]string{"region", "protocol", "status"},
	)

	taskRunCostSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dialtesting",
			Name:      "task_run_cost",
			Help:      "task run time(in nanosecond)",
		},
		[]string{"region", "protocol"},
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister([]prometheus.Collector{
		taskGauge,
		taskSynchronizedCounter,
		taskPullCostSummary,
		taskCheckCostSummary,
		taskRunCostSummary,
		taskInvalidCounter,
	}...)
}
