// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	collectPointsTotalVec *p8s.SummaryVec
	httpGetBytesVec       *p8s.SummaryVec
	httpLatencyVec        *p8s.SummaryVec
)

func metricsSetup() {
	collectPointsTotalVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "prom",
			Name:      "collect_points",
			Help:      "Total number of prom collection points",
		},
		[]string{
			"source",
		},
	)

	httpGetBytesVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "prom",
			Name:      "http_get_bytes",
			Help:      "HTTP get bytes",
		}, []string{
			"source",
		},
	)

	httpLatencyVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "prom",
			Name:      "http_latency_in_second",
			Help:      "HTTP latency(in second)",
		}, []string{
			"source",
		},
	)

	metrics.MustRegister(
		collectPointsTotalVec,
		httpGetBytesVec,
		httpLatencyVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
