// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package remotejob

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var jobRunVec *prometheus.SummaryVec

func setupMetrics() {
	jobRunVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "remote_job",
			Name:      "jvm_dump",
			Help:      "JVM dump job execution time statistics",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{
			"name",
			"status",
		},
	)
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	metrics.MustRegister(jobRunVec)
}
