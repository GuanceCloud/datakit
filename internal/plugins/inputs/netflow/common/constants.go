// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package common

const (
	// InputName is the input's name.
	InputName = "netflow"

	// DefaultNamespace is default namespace.
	DefaultNamespace = "namespace"

	// DefaultStopTimeout is the default stop timeout in seconds.
	DefaultStopTimeout = 5

	// DefaultAggregatorFlushInterval is the default flush interval in seconds.
	DefaultAggregatorFlushInterval = 300 // 5min

	// DefaultAggregatorBufferSize is the default aggregator buffer size interval.
	DefaultAggregatorBufferSize = 10000

	// DefaultAggregatorPortRollupThreshold is the default aggregator port rollup threshold.
	DefaultAggregatorPortRollupThreshold = 10

	// DefaultAggregatorRollupTrackerRefreshInterval is the default aggregator rollup tracker refresh interval.
	DefaultAggregatorRollupTrackerRefreshInterval = 300 // 5min

	// DefaultBindHost is the default bind host used for flow listeners.
	DefaultBindHost = "0.0.0.0"

	// DefaultPrometheusListenerAddress is the default goflow prometheus listener address.
	DefaultPrometheusListenerAddress = "localhost:9090"
)

////////////////////////////////////////////////////////////////////////////////

type IfAdminStatus int

//nolint:stylecheck
const (
	AdminStatus_Up      IfAdminStatus = 1
	AdminStatus_Down    IfAdminStatus = 2
	AdminStatus_Testing IfAdminStatus = 3
)

//nolint:stylecheck
var adminStatus_StringMap map[IfAdminStatus]string = map[IfAdminStatus]string{
	AdminStatus_Up:      "up",
	AdminStatus_Down:    "down",
	AdminStatus_Testing: "testing",
}

func (i IfAdminStatus) AsString() string {
	status, ok := adminStatus_StringMap[i]
	if !ok {
		return "unknown"
	}
	return status
}

////////////////////////////////////////////////////////////////////////////////

type IfOperStatus int

//nolint:stylecheck
const (
	OperStatus_Up             IfOperStatus = 1
	OperStatus_Down           IfOperStatus = 2
	OperStatus_Testing        IfOperStatus = 3
	OperStatus_Unknown        IfOperStatus = 4
	OperStatus_Dormant        IfOperStatus = 5
	OperStatus_NotPresent     IfOperStatus = 6
	OperStatus_LowerLayerDown IfOperStatus = 7
)

//nolint:stylecheck
var operStatus_StringMap map[IfOperStatus]string = map[IfOperStatus]string{
	OperStatus_Up:             "up",
	OperStatus_Down:           "down",
	OperStatus_Testing:        "testing",
	OperStatus_Unknown:        "unknown",
	OperStatus_Dormant:        "dormant",
	OperStatus_NotPresent:     "not_present",
	OperStatus_LowerLayerDown: "lower_layer_down",
}

func (i IfOperStatus) AsString() string {
	status, ok := operStatus_StringMap[i]
	if !ok {
		return "unknown"
	}
	return status
}

////////////////////////////////////////////////////////////////////////////////
