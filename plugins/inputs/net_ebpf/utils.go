package net_ebpf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/host"
)

const (
	// major.minor.patch  (linux kernel version).
	majorK = 4
	minorK = 9
	patchK = 0
)

// version == "", auto to get.
func checkLinuxKernelVesion(version string) (bool, error) {
	var err error
	kVersionStr := version
	if version == "" {
		kVersionStr, err = host.KernelVersion()
		if err != nil {
			return false, err
		}
	}

	kVersionStrArr := strings.Split(strings.Split(kVersionStr, "-")[0], ".")
	var kVersion []int
	for _, vStr := range kVersionStrArr {
		if v, err := strconv.Atoi(vStr); err != nil {
			err = fmt.Errorf("linux kernel version parsing failed: %s", kVersionStr)
			return false, err
		} else {
			kVersion = append(kVersion, v)
		}
	}
	if len(kVersion) != 3 ||
		(kVersion[0] < majorK) ||
		(kVersion[0] == majorK && kVersion[1] < minorK) ||
		(kVersion[0] == majorK && kVersion[1] == minorK && kVersion[2] < patchK) {
		err = fmt.Errorf("the current kernel version is lower than the minimum requirement: %d.%d.%d < %d.%d.%d",
			kVersion[0], kVersion[1], kVersion[2], majorK, minorK, patchK)
		return false, err
	}
	return true, nil
}
