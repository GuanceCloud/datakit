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
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestCollect(t *testing.T) {
	i := defaultInput()

	for x := 0; x < 1; x++ {
		if err := i.collect(); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second * 1)
	}
	if len(i.collectCache) == 0 {
		t.Log("nfs data is empty!")
		return
	}
	tmap := map[string]bool{}
	for _, pt := range i.collectCache {
		tmap[pt.Time().String()] = true
	}
	if len(tmap) != 1 {
		t.Error("Need to clear collectCache.")
	}
}

func TestEnableNFSd(t *testing.T) {
	_, err := os.Stat("/proc/net/rpc/nfsd")
	if os.IsNotExist(err) {
		t.Skip("Skipping test: no /proc/net/rpc/nfsd file or directory because the NFS is not installed.")
	}
	i := defaultInput()

	// Test with NFSd enabled.
	var pts []*point.Point
	i.NFSd = true
	if err = i.collect(); err != nil {
		t.Error(err)
	}
	for _, pt := range i.collectCache {
		if pt.Name() == "nfsd" {
			pts = append(pts, pt)
		}
	}
	assert.Greater(t, len(pts), 0, "nfsd metric should not be empty with NFSd enabled.")

	// Test with NFSd disabled.
	pts = []*point.Point{}
	i.NFSd = false
	if err = i.collect(); err != nil {
		t.Error(err)
	}
	for _, pt := range i.collectCache {
		if pt.Name() == "nfsd" {
			pts = append(pts, pt)
		}
	}
	assert.Len(t, pts, 0, "nfsd metric should be empty with NFSd disabled.")
}

func TestEnabledMountstatsRW(t *testing.T) {
	i := defaultInput()
	i.MountstatsMetric.Rw = true

	if err := i.collect(); err != nil {
		t.Error(err)
	}
}
