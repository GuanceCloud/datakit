// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package nfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestNFSMountStats(t *testing.T) {
	ipt := defaultInput()

	if _, err := ipt.buildMountStats(); err != nil {
		t.Skipf("collect nfs mountstats failed: %v", err)
	}
}

func TestFilesystemUsageFromStat(t *testing.T) {
	total, avail, used, usedPercent := filesystemUsageFromStat(unix.Statfs_t{
		Bsize:  1024,
		Blocks: 100,
		Bfree:  30,
		Bavail: 20,
	})

	assert.Equal(t, uint64(100*1024), total)
	assert.Equal(t, uint64(20*1024), avail)
	assert.Equal(t, uint64(70*1024), used)
	assert.Equal(t, float64(70)/float64(70+20)*100.0, usedPercent)
	assert.NotEqual(t, float64(70)/float64(100)*100.0, usedPercent)
}

func TestNFSBase(t *testing.T) {
	_, err := os.Stat("/proc/net/rpc/nfs")
	if os.IsNotExist(err) {
		t.Skip("Skipping test: no /proc/net/rpc/nfs file or directory because the nfs is not installed.")
	}
	ipt := defaultInput()

	nfsPts, err := ipt.buildBaseMetric()
	if err != nil {
		t.Errorf("Failed to collect nfs: %v", err)
	}
	assert.Greater(t, len(nfsPts), 0, "nfsd metric should not be empty with NFSd enabled.")
}

func TestNFSd(t *testing.T) {
	_, err := os.Stat("/proc/net/rpc/nfsd")
	if os.IsNotExist(err) {
		t.Skip("Skipping test: no /proc/net/rpc/nfsd file or directory because the nfs is not installed.")
	}
	ipt := defaultInput()

	nfsdPts, err := ipt.buildNFSdMetric()
	if err != nil {
		t.Errorf("Failed to collect nfsd: %v", err)
	}
	assert.Greater(t, len(nfsdPts), 0, "nfsd metric should not be empty with NFSd enabled.")
}
