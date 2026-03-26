// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package snmp

import (
	"github.com/GuanceCloud/cliutils/metrics"
	p8s "github.com/prometheus/client_golang/prometheus"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	discoveryCostVec     *p8s.SummaryVec
	collectCostVec       *p8s.SummaryVec
	deviceCollectCostVec *p8s.SummaryVec
	aliveDevicesVec      *p8s.GaugeVec
)

func metricsSetup() {
	discoveryCostVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_snmp",
			Name:      "discovery_cost",
			Help:      "Discovery cost(in second)",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"profile_type"},
	)

	collectCostVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_snmp",
			Name:      "collect_cost",
			Help:      "Every loop collect cost(in second)",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{},
	)

	deviceCollectCostVec = p8s.NewSummaryVec(
		p8s.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_snmp",
			Name:      "device_collect_cost",
			Help:      "Device collect cost(in second)",

			Objectives: datakit.P8sStandardObjectives,
		},
		[]string{"class"},
	)

	aliveDevicesVec = p8s.NewGaugeVec(
		p8s.GaugeOpts{
			Namespace: "datakit",
			Subsystem: "input_snmp",
			Name:      "alive_devices",
			Help:      "Alive devices",
		},
		[]string{"class"},
	)

	metrics.MustRegister(
		discoveryCostVec,
		collectCostVec,
		deviceCollectCostVec,
		aliveDevicesVec,
	)
}

//nolint:gochecknoinits
func init() {
	metricsSetup()
}
