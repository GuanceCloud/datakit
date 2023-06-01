// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	inputInstanceVec *p8s.GaugeVec
	inputsPanicVec   *p8s.CounterVec
)

func metricsSetup() {
	inputInstanceVec = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "inputs",
			Name:      "instance",
			Help:      "Input instance count",
		},
		[]string{
			"input",
		},
	)

	inputsPanicVec = p8s.NewCounterVec(
		p8s.CounterOpts{
			Namespace: "datakit",
			Subsystem: "inputs",
			Name:      "crash_total",
			Help:      "Input crash count",
		},
		[]string{
			"input",
		},
	)

	metrics.MustRegister(inputInstanceVec, inputsPanicVec)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
