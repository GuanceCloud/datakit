// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pcommon share utils reused among various collectors
package pcommon

import (
	"strings"

	"github.com/shirou/gopsutil/disk"
)

// TrimPartitionHostPath remove host-path prefix in p's device name and mountpoint.
func TrimPartitionHostPath(hostpath string, p *disk.PartitionStat) *disk.PartitionStat {
	if hostpath == "" {
		return p
	}

	// we need to trim prefix of host-root path: /rootfs/dev/vdb -> /dev/vdb
	if strings.HasPrefix(p.Device, hostpath) {
		if p.Device == hostpath {
			// NOTE: should not been here, all device should like /rootfs/...
			p.Device = "/"
		} else {
			p.Device = p.Device[len(hostpath):]
		}
	}

	// we need to trim prefix of host-root mountpoint: /rootfs/etc/host -> /etc/host
	if strings.HasPrefix(p.Mountpoint, hostpath) {
		if p.Mountpoint == hostpath {
			p.Mountpoint = "/"
		} else {
			p.Mountpoint = p.Mountpoint[len(hostpath):]
		}
	}

	return p
}
