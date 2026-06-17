// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"io/ioutil"
	"net"
	"os"
	T "testing"

	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type trackingDiskStatsMock struct {
	partitions []disk.PartitionStat
	usagePath  []string
}

type singleDiskStatsMock struct {
	usage *disk.UsageStat
}

func (m *singleDiskStatsMock) Usage(_, _ string) (*disk.UsageStat, error) {
	return m.usage, nil
}

func (m *singleDiskStatsMock) Partitions() ([]disk.PartitionStat, error) {
	return []disk.PartitionStat{
		{Device: "/dev/dm-0", Mountpoint: "/data", Fstype: "ext4"},
	}, nil
}

func (m *trackingDiskStatsMock) Usage(path, _ string) (*disk.UsageStat, error) {
	m.usagePath = append(m.usagePath, path)
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

func (m *trackingDiskStatsMock) Partitions() ([]disk.PartitionStat, error) {
	return m.partitions, nil
}

func TestGetNetInfo(t *T.T) {
	ifs, err := interfaces()
	if err != nil {
		l.Errorf("fail to get interfaces, %s", err)
	}
	var infos []*netInfo

	// netVIfaces := map[string]bool{}
	netVIfaces, _ := NetIgnoreIfaces()

	for _, it := range ifs {
		if _, ok := netVIfaces[it.Name]; ok {
			continue
		}
		i := &netInfo{
			Index:        it.Index,
			MTU:          it.MTU,
			Name:         it.Name,
			HardwareAddr: it.HardwareAddr,
			Flags:        it.Flags,
		}
		for _, ad := range it.Addrs {
			ip, _, _ := net.ParseCIDR(ad.Addr)
			if ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				i.IP4 = ad.Addr
				i.IP4All = append(i.IP4All, ad.Addr)
			} else if ip.To16() != nil {
				i.IP6 = ad.Addr
				i.IP6All = append(i.IP6All, ad.Addr)
			}
		}
		infos = append(infos, i)
	}
	if len(infos) == 0 {
		t.Skip("no non-ignored network interfaces found")
	}
}

func createTempFile(t *T.T, content []byte) string {
	t.Helper()

	tempFile, err := ioutil.TempFile("", "testfile*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	return tempFile.Name()
}

func TestGetConfigFile(t *T.T) {
	testCases := []struct {
		name    string
		content []byte
		isValid bool
	}{
		{"ValidTextFile", []byte("This is a text file."), true},
		{"LargeFile", make([]byte, 5*1024), false}, // 大于 4KB 的文件
		{"NonTextFile", []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F}, false}, // 非文本文件
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			filePath := createTempFile(t, tc.content)
			defer os.Remove(filePath)

			ipt := &Input{
				ConfigPath: []string{filePath},
			}

			files := ipt.getConfigFile()

			require.Equal(t, tc.isValid, len(files) == 1)

			if tc.isValid {
				if content, ok := files[filePath]; !ok || content != string(tc.content) {
					t.Errorf("failed to read content from %s", filePath)
				}
			}
		})
	}
}

func TestGetDiskInfoSkipsIgnoredBeforeUsage(t *T.T) {
	stats := &trackingDiskStatsMock{
		partitions: []disk.PartitionStat{
			{Device: "/dev/sda1", Mountpoint: "/data", Fstype: "xfs"},
			{Device: "systemd-1", Mountpoint: "/proc/sys/fs/binfmt_misc", Fstype: "autofs"},
			{Device: "binfmt_misc", Mountpoint: "/proc/sys/fs/binfmt_misc", Fstype: "binfmt_misc"},
			{Device: "/dev/sdb1", Mountpoint: "/usr/local/datakit/123", Fstype: "xfs"},
		},
	}

	ipt := defaultInput()
	ipt.diskStats = stats
	ipt.setup()

	disks, _, err := ipt.getDiskInfo()
	require.NoError(t, err)
	require.Len(t, disks, 1)
	assert.Equal(t, "/data", disks[0].MountPoint)
	assert.Equal(t, []string{"/data"}, stats.usagePath)
}

func TestGetDiskInfoUsesAvailableSpaceForUsedPercent(t *T.T) {
	usage := &disk.UsageStat{
		Total:             100,
		Free:              20,
		Used:              70,
		UsedPercent:       100.0 * 70.0 / (70.0 + 20.0),
		InodesTotal:       100,
		InodesFree:        50,
		InodesUsed:        50,
		InodesUsedPercent: 50,
	}

	ipt := defaultInput()
	ipt.diskStats = &singleDiskStatsMock{usage: usage}
	ipt.setup()

	disks, usedPercent, err := ipt.getDiskInfo()
	require.NoError(t, err)
	require.Len(t, disks, 1)

	assert.Equal(t, usage.UsedPercent, disks[0].UsedPercent)
	assert.InDelta(t, usage.UsedPercent, usedPercent, 0.000001)
	assert.NotEqual(t, float64(usage.Used)/float64(usage.Total)*100.0, usedPercent)
}
