// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"fmt"
	"os"
	"os/user"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/disk"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	FSTypeUnknown = "unknown"
)

func GetFileSystemType(path string) (string, error) {
	ptr := 0
	if runtime.GOOS == datakit.OSWindows {
		info, err := disk.Partitions(true)
		if err != nil {
			return FSTypeUnknown, fmt.Errorf("error get windows disk information: %w", err)
		}
		for i := 0; i < len(info); i++ {
			if strings.Contains(path, info[i].Device) {
				ptr = i
			}
		}
		return info[ptr].Fstype, nil
	} else {
		infos, err := disk.Usage(path)
		if err != nil {
			return FSTypeUnknown, fmt.Errorf("error get %s disk information: %w", runtime.GOOS, err)
		}
		return infos.Fstype, nil
	}
}

func GetFileOwnership(path string, host string) (string, error) {
	uid, err := Getuid(path, host)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error get uid: %w", err)
	}
	u, err := user.LookupId(uid)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error look for uid: %w", err)
	}
	return u.Username, nil
}

func Getuid(path string, host string) (string, error) {
	var uid string
	info, err := os.Stat(path)
	if host == "linux" || host == "darwin" {
		a := reflect.ValueOf(info.Sys()).Elem()
		uid = strconv.Itoa(int(a.FieldByName("Uid").Uint()))
	}
	return uid, err
}

func Getdirmode(path string) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return FSTypeUnknown, fmt.Errorf("error get os stat information:%w", err)
	}
	mode := fileInfo.Mode()
	return mode.String(), nil
}

func getFileSystemInfo(path string, fileSize int64, usedInode int64, kvs *point.KVs) error {
	usage, err := disk.Usage(path)
	if err != nil {
		return fmt.Errorf("error get disk usages: %w", err)
	}

	var usedPercent float64
	if usage.Used+usage.Free > 0 {
		usedPercent = float64(fileSize) /
			(float64(usage.Used) + float64(usage.Free)) * 100
	}
	*kvs = kvs.Add("total", usage.Total, false, true)
	*kvs = kvs.Add("free", usage.Free, false, true)
	*kvs = kvs.Add("used_percent", usedPercent, false, true)

	if runtime.GOOS != datakit.OSWindows {
		var inodesUsedPercent float64
		if usage.Used+usage.Free > 0 {
			inodesUsedPercent = float64(usedInode) /
				(float64(usage.InodesTotal)) * 100
		}

		*kvs = kvs.Add("inodes_total", usage.InodesTotal, false, true)
		*kvs = kvs.Add("inodes_free", usage.InodesFree, false, true)
		*kvs = kvs.Add("inodes_used", usedInode, false, true)
		*kvs = kvs.Add("inodes_used_percent", inodesUsedPercent, false, true) // float64

		partitions, err := disk.Partitions(true)
		if err != nil {
			return fmt.Errorf("error get disk partitions: %w", err)
		}

		mountpoint := ""
		for _, partition := range partitions {
			if strings.HasPrefix(path, partition.Mountpoint) && len(partition.Mountpoint) > len(mountpoint) {
				mountpoint = partition.Mountpoint
			}
		}
		if mountpoint != "" {
			*kvs = kvs.Add("mount_point", mountpoint, true, true)
		}
	}

	return nil
}
