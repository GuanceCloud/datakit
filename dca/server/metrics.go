// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package server is DCA's HTTP server
package server

import (
	"github.com/GuanceCloud/cliutils/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

var connectionTotalGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "connection_number",
	Help: "The total number of connections.",
})

//nolint:gochecknoinits
func init() {
	metrics.MustRegister([]prometheus.Collector{connectionTotalGauge}...)
}
