// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	collectCostVec                   *prometheus.SummaryVec
	collectPtsVec                    *prometheus.CounterVec
	loggingDiscoveryCostVec          *prometheus.SummaryVec
	loggingDiscoveryScheduleDelayVec prometheus.Summary
)

func setupMetrics() {
	collectCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "container_collect_cost_seconds",
			Help:      "Total time (in seconds) spent collecting container metrics or objects",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"category",
		},
	)

	collectPtsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input",
			Name:      "container_collect_pts_total",
			Help:      "Total number of points collected from containers",
		},
		[]string{
			"category",
		},
	)

	loggingDiscoveryCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "input",
			Name:       "container_logging_discovery_cost_seconds",
			Help:       "Total time (in seconds) spent discovering container logs",
			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"trigger"},
	)

	loggingDiscoveryScheduleDelayVec = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "input",
			Name:       "container_logging_discovery_schedule_delay_seconds",
			Help:       "Delay (in seconds) between a scheduled container log discovery and its execution",
			Objectives: datakit.P8sStandardObjectives,
		},
	)

	metrics.MustRegister(
		collectCostVec,
		collectPtsVec,
		loggingDiscoveryCostVec,
		loggingDiscoveryScheduleDelayVec,
	)
}
