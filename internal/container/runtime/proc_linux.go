// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package runtime

import (
	"fmt"
	"path/filepath"
	"strconv"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

const (
	cpuInfoPath     = "%s/proc/cpuinfo"
	memInfoPath     = "%s/proc/meminfo"
	networkStatPath = "%s/proc/%s/net/dev"
)

type cpuInfoProc struct {
	info *linuxproc.CPUInfo
}

func newCPUInfo(mountPoint string) (cpuInfo, error) {
	path := fmt.Sprintf(cpuInfoPath, mountPoint)
	info, err := linuxproc.ReadCPUInfo(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read info fail, err: %w", err)
	}

	return &cpuInfoProc{
		info: info,
	}, nil
}

func (c *cpuInfoProc) cores() int {
	return c.info.NumCore()
}

type memInfoProc struct {
	info *linuxproc.MemInfo
}

func newMemInfo(mountPoint string) (memInfo, error) {
	path := fmt.Sprintf(memInfoPath, mountPoint)
	info, err := linuxproc.ReadMemInfo(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read meminfo fail, err: %w", err)
	}

	return &memInfoProc{
		info: info,
	}, nil
}

func (m *memInfoProc) total() int64 {
	// While the file shows kilobytes (kB; 1 kB equals 1000 B), it
	// is actually kibibytes (KiB; 1 KiB equals 1024 B)
	// https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/6/html/deployment_guide/s2-proc-meminfo
	return int64(m.info.MemTotal << 10)
}

type networkStatProc struct {
	stats []linuxproc.NetworkStat
}

func newNetworkStat(mountPoint string, pid int) (networkStat, error) {
	path := fmt.Sprintf(networkStatPath, mountPoint, strconv.Itoa(pid))
	stats, err := linuxproc.ReadNetworkStat(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read network stat fail, err: %w", err)
	}

	return &networkStatProc{
		stats: stats,
	}, nil
}

func (n *networkStatProc) rxBytes(skipLoopback bool) int64 {
	var rx uint64
	for _, stat := range n.stats {
		if skipLoopback && stat.Iface == "lo" {
			continue
		}
		rx += stat.RxBytes
	}
	return int64(rx)
}

func (n *networkStatProc) txBytes(skipLoopback bool) int64 {
	var tx uint64
	for _, stat := range n.stats {
		if skipLoopback && stat.Iface == "lo" {
			continue
		}
		tx += stat.TxBytes
	}
	return int64(tx)
}
