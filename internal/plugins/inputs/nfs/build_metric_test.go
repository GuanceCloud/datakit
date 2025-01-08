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
)

func TestNFSMountStats(t *testing.T) {
	ipt := defaultInput()

	if _, err := ipt.buildMountStats(); err != nil {
		t.Skipf("collect nfs mountstats failed: %v", err)
	}
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
