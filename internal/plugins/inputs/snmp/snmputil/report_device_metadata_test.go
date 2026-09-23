// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNetworkInterfacesMetadataWithTypeAndIsPhysical(t *testing.T) {
	store := NewMetadataStore()
	store.AddColumnValue("interface.name", "1", ResultValue{Value: "eth0"})
	store.AddColumnValue("interface.type", "1", ResultValue{Value: float64(6)})
	store.AddColumnValue("interface.name", "2", ResultValue{Value: "tun0"})
	store.AddColumnValue("interface.type", "2", ResultValue{Value: float64(24)})
	store.AddColumnValue("interface.name", "3", ResultValue{Value: "unknown0"})

	interfaces := BuildNetworkInterfacesMetadata("device-id", store)

	trueVal := true
	falseVal := false
	assert.Equal(t, []InterfaceMetadata{
		{
			DeviceID:   "device-id",
			Index:      1,
			Name:       "eth0",
			Type:       6,
			IsPhysical: &trueVal,
		},
		{
			DeviceID:   "device-id",
			Index:      2,
			Name:       "tun0",
			Type:       24,
			IsPhysical: &falseVal,
		},
		{
			DeviceID: "device-id",
			Index:    3,
			Name:     "unknown0",
		},
	}, interfaces)
}
