// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package lsblk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
)

func TestCollect(t *testing.T) {
	i := defaultInput()
	for x := 0; x < 2; x++ {
		i.ptsTime = ntp.Now()
		if err := i.collect(); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second * 1) // with sleep to update ptsTime
	}

	assert.NotEmpty(t, i.collectCache)

	tmap := map[int64]bool{}
	for _, pt := range i.collectCache { // test if all point's time the same.
		tmap[pt.Time().UnixNano()] = true
	}

	for key, value := range tmap {
		t.Log(key, value)
	}

	// if point's time the same, the map should only 1 elem.
	assert.Lenf(t, tmap, 1, "Need to clear collectCache.")
}
