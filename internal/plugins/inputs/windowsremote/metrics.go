// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	wmiConnCount        *prometheus.CounterVec
	reachableIPCountVec *prometheus.GaugeVec
)

func setupMetrics() {
	reachableIPCountVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "windowsremote",
			Name:      "reachable_ip",
			Help:      "Total number of IPs reachable through tcp/upd port verification.",
		},
		[]string{},
	)
	wmiConnCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "datakit",
			Subsystem: "windowsremote",
			Name:      "wmi_conn_total",
			Help:      "The WMI connects count",
		}, []string{"status"})

	metrics.MustRegister(
		reachableIPCountVec,
		wmiConnCount,
	)
}
