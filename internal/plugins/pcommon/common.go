// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pcommon share utils reused among various collectors
package pcommon

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
	Usage(path, hostPath string) (*disk.UsageStat, error)
	Partitions() ([]disk.PartitionStat, error)
}

type FilesystemStats struct {
	Usage *disk.UsageStat
	Part  *disk.PartitionStat
}

func FilterUsage(ds DiskStats, hostPath string) (arr []FilesystemStats, err error) {
	parts, err := ds.Partitions()
	if err != nil {
		return nil, fmt.Errorf("Partitions(): %w", err)
	}

	for i := range parts {
		p := &parts[i]

		du, err := ds.Usage(p.Mountpoint, hostPath)
		if err != nil {
			l.Warnf("Usage on partition %+#v: %s, ignored", p, err)
			continue
		}

		p = TrimPartitionHostPath(hostPath, &parts[i])

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

type NSEnterDiskstatsImpl struct{}

func (*NSEnterDiskstatsImpl) Usage(path, _ string) (*disk.UsageStat, error) {
	// NOTE: nsenter do not need hostPath.
	return nsenterDiskStat(path)
}

func parseStatF(output string) (*disk.UsageStat, error) {
	var (
		stats = &disk.UsageStat{}
		lines = strings.Split(output, "\n")
		blockSize,
		freeBlocks,
		totalInodes,
		freeInodes,
		totalBlocks uint64
	)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		arr := strings.Split(line, " ")

		for i, elem := range arr {
			val, err := strconv.ParseUint(elem, 10, 64)
			if err != nil {
				l.Warnf("skil elem(at #%d): %s: %s, ignored", i, elem, err.Error())
				continue
			}

			switch i {
			case 0: // block-size
				blockSize = val
			case 1: // blokcs
				totalBlocks = val
			case 2: // blocks-free
				freeBlocks = val
			case 3: // inodes-total
				totalInodes = val
			case 4: // inodes-free
				freeInodes = val
			}
		}
	}

	if blockSize == 0 {
		return nil, fmt.Errorf("failed to parse block size from stat -f output: %s",
			output)
	}

	stats.Total = totalBlocks * blockSize
	stats.Used = (totalBlocks - freeBlocks) * blockSize
	stats.Free = freeBlocks * blockSize
	stats.UsedPercent = float64(stats.Used) / float64(stats.Used+stats.Free) * 100

	stats.InodesTotal = totalInodes
	stats.InodesFree = freeInodes
	stats.InodesUsed = stats.InodesTotal - stats.InodesFree
	if stats.InodesTotal != 0 && stats.InodesUsed != 0 { // 0.0/0.0 may cause NaN, this will lead to point encode erorr:
		stats.InodesUsedPercent = 100.0 * float64(stats.InodesUsed) / float64(stats.InodesTotal)
	}
	return stats, nil
}

func nsenterDiskStat(path string) (*disk.UsageStat, error) {
	if runtime.GOOS != "linux" || !datakit.Docker {
		return nil, fmt.Errorf("nsenter not implemented under %q", runtime.GOOS)
	}

	cmd := exec.Command("nsenter", "--target", "1", "--mount", "--", "stat",
		"-c",
		// block-size/blocks/block-free/inodes-total/inodes-free
		"%s %b %f %c %d",
		"-f", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run nsenter command for '%s': %w, stderr: %s", path, err, stderr.String())
	}

	return parseStatF(stdout.String())
}

func (*NSEnterDiskstatsImpl) Partitions() ([]disk.PartitionStat, error) {
	return disk.Partitions(true) // still use gopsutil
}

type DiskStatsImpl struct{}

func (*DiskStatsImpl) Usage(path, hostPath string) (*disk.UsageStat, error) {
	if hostPath != "" && !strings.HasPrefix(path, hostPath) {
		// add `/rootfs' to access mountpoint within /proc/1/mountpoint.
		path = filepath.Join(hostPath, path)
	}

	return disk.Usage(path)
}

func (*DiskStatsImpl) Partitions() ([]disk.PartitionStat, error) {
	return disk.Partitions(true)
}

func SetLog() {
	once.Do(func() { l = logger.SLogger("pcommon") })
}
