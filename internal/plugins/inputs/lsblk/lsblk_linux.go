// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package lsblk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (ipt *Input) collectLsblkInfo() ([]BlockDeviceStat, error) {
	if ipt.partitionMap == nil {
		return nil, fmt.Errorf("no partition map")
	}
	mountedPartitions, err := ipt.getMountedPartitions(false)
	if err != nil {
		return nil, err
	}

	devices, err := ipt.getAllDevices()
	if err != nil {
		devices = mountedPartitions
	}

	setDeviceUUID(&devices)
	setDeviceLabel(&devices)
	setDeviceRQSize(&devices)

	// 处理mountpoint和parent都可能有多个值的情况
	devices = ipt.processDevices(devices)

	// 过滤exclude的device
	devices = ipt.filterDevices(devices)

	return devices, nil
}

func (ipt *Input) getAllDevices() ([]BlockDeviceStat, error) {
	path := hostSys("/dev/block")
	devices := make([]BlockDeviceStat, 0)

	entries, err := os.ReadDir(path)
	if err != nil {
		return devices, err
	}

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			deviceID := entry.Name() // 11:0
			linkPath, err := os.Readlink(filepath.Join(path, entry.Name()))
			if err != nil {
				continue
			}

			devicePath := filepath.Join(path, linkPath)
			var device BlockDeviceStat
			if ipt.partitionMap[deviceID] != nil {
				device = *ipt.partitionMap[deviceID]
			} else {
				device = BlockDeviceStat{
					MajMin:    deviceID,
					IsMounted: false,
				}
			}
			device.KName = filepath.Base(devicePath)

			if isDM(filepath.Base(devicePath)) {
				setDMDetailInfo(&device, devicePath)
			} else {
				setDetailInfo(&device, devicePath)
			}

			devices = append(devices, device)
		}
	}

	return devices, nil
}

func (ipt *Input) getMountedPartitions(all bool) ([]BlockDeviceStat, error) {
	filename := hostProc("self/mountinfo")
	lines, err := ReadLines(filename)
	if err != nil {
		return nil, err
	}

	fs, err := getFileSystems()
	if err != nil && !all {
		return nil, err
	}

	ret := make([]BlockDeviceStat, 0, len(lines))

	for _, line := range lines {
		var d BlockDeviceStat
		// a line of self/mountinfo has the following structure:
		// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
		// (1) (2) (3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

		// split the mountinfo line by the separator hyphen
		parts := strings.Split(line, " - ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", filename, line)
		}

		fields := strings.Fields(parts[0])
		blockDeviceID := fields[2]
		mountPoint := fields[4]

		fields = strings.Fields(parts[1])
		fstype := fields[0]
		device := fields[1]

		if ipt.partitionMap[blockDeviceID] != nil {
			ipt.partitionMap[blockDeviceID].MountPoints = append(ipt.partitionMap[blockDeviceID].MountPoints, unescapeFstab(mountPoint))
			continue
		}

		d = BlockDeviceStat{
			Name:       device,
			MajMin:     blockDeviceID,
			MountPoint: unescapeFstab(mountPoint),
			FSType:     fstype,
		}

		d.MountPoints = append(d.MountPoints, d.MountPoint)

		if !all {
			if d.Name == "none" || !StringsHas(fs, d.FSType) {
				continue
			}
		}

		if strings.HasPrefix(d.Name, "/dev/mapper/") {
			devpath, err := filepath.EvalSymlinks(filepath.Join("/dev", (strings.ReplaceAll(d.Name, "/dev", ""))))
			if err == nil {
				d.Name = devpath
			}
		}

		// /dev/root is not the real device name
		// so we get the real device name from its major/minor number
		if d.Name == "/dev/root" {
			devpath, err := os.Readlink("/sys/dev/block/" + blockDeviceID)
			if err != nil {
				return nil, err
			}
			d.Name = strings.Replace(d.Name, "root", filepath.Base(devpath), 1)
		}
		d.IsMounted = true

		if d.MajMin != "" {
			ipt.partitionMap[d.MajMin] = &d
		}

		ret = append(ret, d)
	}

	return ret, nil
}

func (ipt *Input) processDevices(devices []BlockDeviceStat) []BlockDeviceStat {
	var newDevices []BlockDeviceStat

	for _, device := range devices {
		tuList := cartesianProduct(device.MountPoints, device.Parents)
		for _, tu := range tuList {
			device.MountPoint = tu[0]
			device.Parent = tu[1]
			newDevices = append(newDevices, device)
		}
	}

	return newDevices
}

func (ipt *Input) filterDevices(device []BlockDeviceStat) []BlockDeviceStat {
	var newDevices []BlockDeviceStat

	excluded := func(x string, arr []string) bool {
		for _, fs := range arr {
			if strings.EqualFold(x, fs) {
				return true
			}
		}
		return false
	}

	for _, d := range device {
		if excluded(d.Name, ipt.ExcludeDevice) {
			continue
		}
		newDevices = append(newDevices, d)
	}
	return newDevices
}
