// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var collectDurationVec *prometheus.SummaryVec

func setupMetrics() {
	collectDurationVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "datakit",
			Subsystem:  "input_jvm",
			Name:       "collect_duration_seconds",
			Help:       "JVM input collect duration in seconds",
			Objectives: datakit.P8sStandardObjectives,
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
