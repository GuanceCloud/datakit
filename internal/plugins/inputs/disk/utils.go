// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	//nolint
	"github.com/shirou/gopsutil/disk"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
			if strings.EqualFold(x, fs) {
				return true
			}
		}
		return false
	}

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat

	for i := range parts {
		p := parts[i]

		if strings.HasPrefix(p.Device, "/rootfs/dev/") {
			p.Device = p.Device[7:]
		}

		l.Debugf("disk fstype: %s, device: %s, mountpoint: %s", p.Fstype, p.Device, p.Mountpoint)

		// nolint
		if !strings.HasPrefix(p.Device, "/dev/") && runtime.GOOS != datakit.OSWindows && !excluded(p.Device, dk.ipt.ExtraDevice) {
			continue // ignore the partition
		}

		if excluded(p.Device, dk.ipt.ExcludeDevice) {
			continue
		}

		mergerFlag := false
		// merger device
		for _, cont := range partitions {
			if cont.Device == p.Device {
				mergerFlag = true
				break
			}
		}

		if mergerFlag {
			continue
		}

		du, err := dk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

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

func unique(strSlice []string) []string {
	keys := make(map[string]interface{})
	var list []string
	for _, entry := range strSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = nil
			list = append(list, entry)
		}
	}
	return list
}

func hostSys(combineWith ...string) string {
	value := os.Getenv("HOST_SYS")
	if value == "" {
		value = "/sys"
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

func findDiskFromDM(dmDevice string) ([]string, error) {
	if !strings.HasPrefix(dmDevice, "dm-") {
		return nil, fmt.Errorf("invalid dm partition path: %s", dmDevice)
	}
	sysBlockPath := hostSys("block")
	slavesPath := filepath.Join(sysBlockPath, dmDevice, "slaves")

	slaveList, err := os.ReadDir(slavesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", slavesPath, err)
	}

	var physicalDisks []string
	for _, slave := range slaveList {
		diskNameList, err := findDiskFromBlock(slave.Name())
		if err != nil {
			return nil, err
		}
		physicalDisks = append(physicalDisks, diskNameList...)
	}
	return physicalDisks, nil
}

func findDiskFromBlock(partitionName string) ([]string, error) {
	sysBlockPath := hostSys("block")
	blockDeviceList, err := os.ReadDir(sysBlockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", sysBlockPath, err)
	}

	// sort block devices by name in descending order.
	sort.Slice(blockDeviceList, func(i, j int) bool { return blockDeviceList[i].Name() > blockDeviceList[j].Name() })
	for _, blockDevice := range blockDeviceList {
		blockDeviceName := blockDevice.Name()
		if strings.HasPrefix(partitionName, blockDeviceName) {
			return []string{"/dev/" + blockDeviceName}, nil
		}
	}
	return nil, fmt.Errorf("no disk found matching partition %s", partitionName)
}

func GetMapperPath(dmID string) (string, error) {
	dmID = strings.TrimPrefix(dmID, "/dev/")
	sysBlockPath := hostSys("block")
	dmPath := filepath.Join(sysBlockPath, dmID, "dm", "name")

	data, err := os.ReadFile(dmPath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to read %s", dmPath)
	}

	mapperPath := fmt.Sprintf("/dev/mapper/%s", strings.TrimSpace(string(data)))
	_, err = os.Stat(mapperPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("mapper path %s does not exist", mapperPath)
	}
	return mapperPath, nil
}
