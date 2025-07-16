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
	T "testing"
	"time"

	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestCollect(t *T.T) {
	i := defaultInput()
	var ts time.Time

	for x := 0; x < 1; x++ {
		tn := time.Now()
		ts = inputs.AlignTime(tn, ts, i.Interval)
		if err := i.collect(ts.UnixNano()); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second * 1)
	}
	if len(i.collectCache) < 1 {
		t.Error("Failed to collect, no data returned")
	}
	tmap := map[string]bool{}
	for _, pt := range i.collectCache {
		tmap[pt.Time().String()] = true
	}
	if len(tmap) != 1 {
		t.Error("Need to clear collectCache.")
	}
}

type diskStatsMock struct {
	mountinfo string
}

func (m *diskStatsMock) Usage(path, _ string) (*disk.UsageStat, error) {
	return &disk.UsageStat{
		Total:             100,
		Free:              10,
		Used:              90,
		UsedPercent:       .9,
		InodesTotal:       1 << 32,
		InodesUsed:        1 << 20,
		InodesFree:        (1 << 32) - (1 << 20),
		InodesUsedPercent: float64(1<<20) / float64(1<<32),
	}, nil
}

func (m *diskStatsMock) Partitions() ([]disk.PartitionStat, error) {
	var arr []disk.PartitionStat
	for _, fstype := range []string{
		// linux common fstype
		"ext4", "btrfs", "xfs", "zfs", "f2fs", "overlay", "quashfs", "vfat", "exfat", "ntfs", "tmpfs", "proc", "sysfs",
		// macos
		"hfs", "apfs", "msdos",
	} {
		arr = append(arr, disk.PartitionStat{
			Device:     fmt.Sprintf("/dev/%s/device", fstype),
			Mountpoint: fmt.Sprintf("/some/%s/mountpoint", fstype),
			Fstype:     fstype,
		})

		arr = append(arr, disk.PartitionStat{
			Device:     fmt.Sprintf("/dev/%s/device", fstype),
			Mountpoint: "/usr/local/datakit/123",
			Fstype:     fstype,
		})

		arr = append(arr, disk.PartitionStat{
			Device:     fmt.Sprintf("/dev/%s/device", fstype),
			Mountpoint: "/run/containerd/1",
			Fstype:     fstype,
		})
	}

	return arr, nil
}

func TestFilterUsage(t *T.T) {
	t.Run(`ignore-fstype`, func(t *T.T) {
		ipt := defaultInput()
		ipt.diskStats = &diskStatsMock{}
		ipt.setup()

		arr, err := ipt.filterUsage()
		require.NoError(t, err)
		for _, x := range arr {
			t.Logf("usage: %+#v, part: %+#v", x.Usage, x.Part)

			assert.NotEqual(t, "overlay", x.Part.Fstype)
			assert.NotEqual(t, "tmpfs", x.Part.Fstype)
			assert.NotEqual(t, "proc", x.Part.Fstype)
			assert.NotEqual(t, "sysfs", x.Part.Fstype)
		}
	})

	t.Run(`ignore-mountpoint`, func(t *T.T) {
		ipt := defaultInput()
		ipt.diskStats = &diskStatsMock{}
		ipt.setup()

		arr, err := ipt.filterUsage()
		require.NoError(t, err)
		for _, x := range arr {
			t.Logf("usage: %+#v, part: %+#v", x.Usage, x.Part)

			assert.NotContains(t, x.Part.Mountpoint, "/usr/local/datakit")
			assert.NotContains(t, x.Part.Mountpoint, "/run/containerd")
		}
	})
}

func TestLinuxFilterUsage(t *T.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipped on ", runtime.GOOS)
	}

	t.Run(`all-ignored`, func(t *T.T) {
		ipt := defaultInput()

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "1")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/1/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.all-ignored")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		_, err = ipt.filterUsage()
		require.NoError(t, err)
	})

	t.Run("case2", func(t *T.T) {
		ipt := defaultInput()

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "1")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/1/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.2")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		arr, err := ipt.filterUsage()
		require.NoError(t, err)

		for _, fs := range arr {
			t.Logf("usage: %s, part: %s", fs.Usage, fs.Part)
		}
	})
}
