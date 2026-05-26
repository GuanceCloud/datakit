// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	collectPointsTotalVec *p8s.SummaryVec
	httpGetBytesVec       *p8s.SummaryVec
)

func metricsSetup() {
	collectPointsTotalVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_statsd",
			Name:      "collect_points",
			Help:      "Total number of statsd collection points",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{},
	)

	httpGetBytesVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_statsd",
			Name:      "accept_bytes",
			Help:      "Accept bytes from network",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{},
	)

	metrics.MustRegister(
		collectPointsTotalVec,
		httpGetBytesVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
