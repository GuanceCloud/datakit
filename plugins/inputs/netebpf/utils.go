package netebpf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/host"
)

const (
	// major.minor.patch  (linux kernel version).
	majorK           = uint64(4)
	minorK           = uint64(0)
	patchK           = uint64(0)
	minKernelVersion = majorK<<(16*3) | minorK<<(16*2) | patchK<<16 // 0x00004, 0x0009, 0x0000, 0x0000
	// minKernelVersionB16 = 0x0004000000000000.
)

// If the value of the parameter "version" is an empty string,
// it will be automatically obtained.
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
	var kVersion uint64 = 0 // major(off +0), minor(off +16), patch(off +32), 0(off +48)
	if len(kVersionStrArr) == 3 {
		for i, vStr := range kVersionStrArr {
			if v, err := strconv.Atoi(vStr); err != nil {
				err = fmt.Errorf("linux kernel version parsing failed: %s", kVersionStr)
				return false, err
			} else {
				kVersion |= uint64(v) << (16 * (3 - i))
			}
		}
	}

	if kVersion < minKernelVersion {
		return false, fmt.Errorf("the current kernel version is lower than the minimum requirement: %d.%d.%d < %d.%d.%d",
			(kVersion>>48)&0xFFFF, (kVersion>>32)&0xFFFF, (kVersion>>16)&0xFFFF, majorK, minorK, patchK)
	}

	return true, nil
}

func checkIsCentos76Ubuntu1604(platform string, version string) bool {
	sArr := strings.Split(version, ".")
	if len(sArr) < 2 {
		return false
	}
	major, err := strconv.Atoi(sArr[0])
	if err != nil {
		return false
	}
	minor, err := strconv.Atoi(sArr[1])
	if err != nil {
		return false
	}

	if platform == "centos" {
		if (major == 7 && minor >= 6) || major > 7 {
			return true
		}
	}
	if platform == "ubuntu" {
		if (major == 16 && minor >= 4) || major > 16 {
			return true
		}
	}
	return false
}
