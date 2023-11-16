// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var setUlimitVec *prometheus.GaugeVec

func metricsSetup() {
	setUlimitVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "config",
			Name:      "datakit_ulimit",
			Help:      "Datakit ulimit",
		},
		[]string{"status"},
	)
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		setUlimitVec,
	}
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
	metrics.MustRegister(Metrics()...)
}
