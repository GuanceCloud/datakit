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
	collectPtsCounter  *prometheus.CounterVec
	scrapeTargetNumber *prometheus.GaugeVec
)

func setupMetrics() {
	collectPtsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "resource_collect_pts_total",
			Help:      "The number of the points which have been sent",
		},
		[]string{
			"role",
			"name",
		},
	)

	scrapeTargetNumber = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_kubernetesprometheus",
			Name:      "resource_target_number",
			Help:      "The number of the target",
		},
		[]string{
			"role",
			"name",
		},
	)

	metrics.MustRegister(
		collectPtsCounter,
		scrapeTargetNumber,
	)
}
