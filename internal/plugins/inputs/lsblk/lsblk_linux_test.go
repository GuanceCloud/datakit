// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package lsblk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectLsblkInfo(t *testing.T) {
	ipt := defaultInput()

	if _, err := ipt.collectLsblkInfo(); err != nil {
		t.Errorf("Failed to collect lsblk info: %v", err)
	}
	// fmt.Fprintln(os.Stdout, devices)
}

func TestSetFilesystemStatsFromStat(t *testing.T) {
	device := &BlockDeviceStat{}

	setFilesystemStatsFromStat(device, 100, 30, 20, 1024)

	assert.Equal(t, float64(100*1024), device.FSSize)
	assert.Equal(t, float64(20*1024), device.FSAvail)
	assert.Equal(t, float64(70*1024), device.FSUsed)
	assert.Equal(t, float64(70)/float64(70+20)*100.0, device.FSUsePercent)
	assert.NotEqual(t, float64(70)/float64(100)*100.0, device.FSUsePercent)
}
