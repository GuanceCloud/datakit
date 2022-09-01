// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

type PSDiskStats interface {
	Usage(path string) (*disk.UsageStat, error)
	FilterUsage() ([]*disk.UsageStat, []*disk.PartitionStat, error)
	OSGetenv(key string) string
	Partitions(all bool) ([]disk.PartitionStat, error)
}

type PSDisk struct {
	ipt *Input
}

func (dk *PSDisk) Usage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

func (dk *PSDisk) OSGetenv(key string) string {
	return os.Getenv(key)
}

func (dk *PSDisk) Partitions(all bool) ([]disk.PartitionStat, error) {
	return disk.Partitions(all)
}

func (dk *PSDisk) FilterUsage() ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	parts, err := dk.Partitions(!dk.ipt.OnlyPhysicalDevice)
	if err != nil {
		return nil, nil, err
	}

	excluded := func(x string, arr []string) bool {
		for _, fs := range arr {
			if x == fs {
				return true
			}
		}

		return false
	}

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat
	hostMountPrefix := dk.OSGetenv("HOST_MOUNT_PREFIX")

	for i := range parts {
		p := parts[i]
		if len(dk.ipt.Mountpoints) != 0 {
			if !excluded(p.Mountpoint, dk.ipt.Mountpoints) {
				continue
			}
		} else if excluded(p.Mountpoint, dk.ipt.IgnoreMountPoints) {
			continue
		}

		// If the mount point is a member of the exclude set, don't gather info on it.
		if len(dk.ipt.Fs) != 0 {
			if !excluded(p.Fstype, dk.ipt.Fs) {
				continue
			}
		} else if excluded(p.Fstype, dk.ipt.IgnoreFS) {
			continue
		}

		du, err := dk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		du.Path = filepath.Join("/", strings.TrimPrefix(p.Mountpoint, hostMountPrefix))
		du.Fstype = p.Fstype
		usage = append(usage, du)
		partitions = append(partitions, &p)
	}

	return usage, partitions, nil
}

type MountOptions []string

func (opts MountOptions) Mode() string {
	switch {
	case opts.exists("rw"):
		return "rw"
	case opts.exists("ro"):
		return "ro"
	default:
		return "unknown"
	}
}

func (opts MountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func parseOptions(opts string) MountOptions {
	return strings.Split(opts, ",")
}
