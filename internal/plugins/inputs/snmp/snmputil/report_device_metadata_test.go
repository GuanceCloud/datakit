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

func Test_batchPayloads(t *testing.T) {
	collectTime := MockTimeNow()
	deviceID := "123"
	device := DeviceMetadata{ID: deviceID}

	var interfaces []InterfaceMetadata
	for i := 0; i < 350; i++ {
		interfaces = append(interfaces, InterfaceMetadata{DeviceID: deviceID, Index: int32(i)})
	}
	payloads := BatchPayloads("my-ns", "127.0.0.0/30", collectTime, 100, device, interfaces)

	assert.Equal(t, 4, len(payloads))

	assert.Equal(t, "my-ns", payloads[0].Namespace)
	assert.Equal(t, "127.0.0.0/30", payloads[0].Subnet)
	assert.Equal(t, int64(946684800), payloads[0].CollectTimestamp)
	assert.Equal(t, []DeviceMetadata{device}, payloads[0].Devices)
	assert.Equal(t, 99, len(payloads[0].Interfaces))
	assert.Equal(t, interfaces[0:99], payloads[0].Interfaces)

	assert.Equal(t, "127.0.0.0/30", payloads[1].Subnet)
	assert.Equal(t, int64(946684800), payloads[1].CollectTimestamp)
	assert.Equal(t, 0, len(payloads[1].Devices))
	assert.Equal(t, 100, len(payloads[1].Interfaces))
	assert.Equal(t, interfaces[99:199], payloads[1].Interfaces)

	assert.Equal(t, 0, len(payloads[2].Devices))
	assert.Equal(t, 100, len(payloads[2].Interfaces))
	assert.Equal(t, interfaces[199:299], payloads[2].Interfaces)

	assert.Equal(t, 0, len(payloads[3].Devices))
	assert.Equal(t, 51, len(payloads[3].Interfaces))
	assert.Equal(t, interfaces[299:350], payloads[3].Interfaces)
}
