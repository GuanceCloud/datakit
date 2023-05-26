// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"encoding/json"
	"expvar"
)

const (
	// EventTypeSnmpTraps is the event type for snmp traps.
	EventTypeSnmpTraps = "network-devices-snmp-traps"
)

var (
	trapsExpvars           = expvar.NewMap("snmp_traps")
	trapsPackets           = expvar.Int{}
	trapsPacketsAuthErrors = expvar.Int{}
)

func init() { // nolint:gochecknoinits
	trapsExpvars.Set("Packets", &trapsPackets)
	trapsExpvars.Set("PacketsAuthErrors", &trapsPacketsAuthErrors)
}

func getDroppedPackets() int64 {
	aggregatorMetrics, ok := expvar.Get("aggregator").(*expvar.Map)
	if !ok {
		return 0
	}

	epErrors, ok := aggregatorMetrics.Get("EventPlatformEventsErrors").(*expvar.Map)
	if !ok {
		return 0
	}

	droppedPackets, ok := epErrors.Get(EventTypeSnmpTraps).(*expvar.Int)
	if !ok {
		return 0
	}
	return droppedPackets.Value()
}

// GetStatus returns key-value data for use in status reporting of the traps server.
func GetStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	metricsJSON := []byte(expvar.Get("snmp_traps").String())
	metrics := make(map[string]interface{})
	if err := json.Unmarshal(metricsJSON, &metrics); err != nil {
		return nil, err
	}
	if dropped := getDroppedPackets(); dropped > 0 {
		metrics["PacketsDropped"] = dropped
	}
	status["metrics"] = metrics

	if errStart != nil {
		status["error"] = errStart.Error()
	}
	return status, nil
}
