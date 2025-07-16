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

func Test_parseStatF(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		output := `1048576 10737418240 10736238651 1000000000 997250393`

		stat, err := parseStatF(output)
		assert.NoError(t, err)

		t.Logf("stat: %+#v", stat)

		assert.Equal(t, uint64(1048576*10737418240), stat.Total)
		assert.Equal(t, uint64(1048576*10736238651), stat.Free)
		assert.Equal(t, uint64(1048576*(10737418240-10736238651)), stat.Used)
		assert.Equal(t, uint64(1000000000), stat.InodesTotal)
		assert.Equal(t, uint64(997250393), stat.InodesFree)
		assert.Equal(t, uint64(1000000000-997250393), stat.InodesUsed)
		assert.Equal(t, float64(1000000000-997250393)/float64(1000000000)*100.0, stat.InodesUsedPercent)
	})
}
