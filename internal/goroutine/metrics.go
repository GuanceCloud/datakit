// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package goroutine

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	goroutineGroups     p8s.Gauge
	goroutineCostVec    *p8s.SummaryVec
	goroutineStoppedVec *p8s.CounterVec
	goroutineCounterVec *p8s.GaugeVec
)

func metricsSetup() {
	goroutineCounterVec = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "goroutine",
			Name:      "alive",
			Help:      "Alive Goroutines",
		},
		[]string{
			"name",
		},
	)

	goroutineStoppedVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "goroutine",
			Name:      "stopped_total",
			Help:      "Stopped Goroutines",
		},
		[]string{
			"name",
		},
	)

	goroutineGroups = p8s.NewGauge(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "goroutine",
			Name:      "groups",
			Help:      "Goroutine group count",
		},
	)

	goroutineCostVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "goroutine",
			Name:      "cost",
			Help:      "Goroutine running time(in nanosecond)",
		},
		[]string{
			"name",
		},
	)

	metrics.MustRegister(
		goroutineGroups,
		goroutineCostVec,
		goroutineCounterVec,
		goroutineStoppedVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
