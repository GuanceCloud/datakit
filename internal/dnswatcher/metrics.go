// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dnswatcher

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

var (
	watchRunCounter,
	dnsUpdateCounter *prometheus.CounterVec

	dnsDomainCounter prometheus.Counter

	watchLatency *prometheus.SummaryVec
)

//nolint:gochecknoinits
func init() {
	dnsDomainCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dns",
			Name:      "domain_total",
			Help:      "DNS watched domain counter",
		},
	)

	dnsUpdateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dns",
			Name:      "ip_updated_total",
			Help:      "Domain IP updated counter",
		},
		[]string{
			"domain",
		},
	)

	watchRunCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "dns",
			Name:      "watch_run_total",
			Help:      "Watch run counter",
		},
		[]string{
			"interval",
		},
	)

	watchLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "dns",
			Name:      "cost_seconds",
			Help:      "DNS IP lookup cost",

			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"domain",
			"status",
		},
	)

	metrics.MustRegister(Metrics()...)
}

func Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		dnsDomainCounter,
		dnsUpdateCounter,
		watchRunCounter,
		watchLatency,
	}
}

func MetricsReset() {
	dnsUpdateCounter.Reset()
	watchRunCounter.Reset()
	watchLatency.Reset()
}
