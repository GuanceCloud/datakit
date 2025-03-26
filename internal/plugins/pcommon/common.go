// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pcommon share utils reused among various collectors
package pcommon

import (
	"fmt"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/disk"
)

var (
	once sync.Once
	l    = logger.DefaultSLogger("pcommon")
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

type DiskStats interface {
	Usage(path string) (*disk.UsageStat, error)
	Partitions() ([]disk.PartitionStat, error)
}

type FilesystemStats struct {
	Usage *disk.UsageStat
	Part  *disk.PartitionStat
}

func FilterUsage(ds DiskStats, hostRoot string) (arr []FilesystemStats, err error) {
	parts, err := ds.Partitions()
	if err != nil {
		return nil, fmt.Errorf("Partitions(): %w", err)
	}

	for i := range parts {
		p := &parts[i]

		du, err := ds.Usage(p.Mountpoint)
		if err != nil {
			l.Warnf("Usage on partition %+#v: %s, ignored", p, err)
			continue
		}

		p = TrimPartitionHostPath(hostRoot, &parts[i])

		// NOTE: prefer p.Path, du.Path may need to trim host-path prefix
		du.Path = p.Mountpoint

		l.Debugf("add partition %+#v, usage: %+#v", p, du)

		arr = append(arr, FilesystemStats{
			Usage: du,
			Part:  p,
		})
	}

	return arr, nil
}

type DiskStatsImpl struct{}

func (dk *DiskStatsImpl) Usage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

func (dk *DiskStatsImpl) Partitions() ([]disk.PartitionStat, error) {
	return disk.Partitions(true)
}

func SetLog() {
	once.Do(func() { l = logger.SLogger("pcommon") })
}
