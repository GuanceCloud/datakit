// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pcommon

import (
	T "testing"

	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
)

func TestTrimPartitionHostPath(t *T.T) {
	hostpath := "/rootfs"
	p := &disk.PartitionStat{
		Device:     hostpath + "/dev/sda",
		Mountpoint: hostpath + "/var/lib/containerd/container_logs",
	}

	p = TrimPartitionHostPath(hostpath, p)

	assert.Equal(t, "/dev/sda", p.Device)
	assert.Equal(t, "/var/lib/containerd/container_logs", p.Mountpoint)

	// pure `/rootfs`
	p = &disk.PartitionStat{
		Device:     hostpath,
		Mountpoint: hostpath,
	}
	p = TrimPartitionHostPath(hostpath, p)

	assert.Equal(t, p.Device, "/")
	assert.Equal(t, p.Mountpoint, "/")
}
