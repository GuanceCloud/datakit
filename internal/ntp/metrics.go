// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package ntp

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	ntpSyncCount = p8s.NewCounter(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "ntp",
			Name:      "sync_total",
			Help:      "Total count synced with remote NTP server",
		},
	)

	ntpSyncSummary = p8s.NewSummary(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "ntp",
			Name:      "time_diff",
			Help:      "Time difference(seconds) between remote NTP server",
			Objectives: map[float64]float64{
				0.5:  .05,
				0.9:  .01,
				0.99: .001,
			},
		},
	)
)

// nolint:gochecknoinits
func init() {
	metrics.MustRegister(
		ntpSyncCount,
		ntpSyncSummary,
	)
}
