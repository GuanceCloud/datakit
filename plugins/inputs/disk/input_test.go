// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"math"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestCollect(t *testing.T) {
	i := newDefaultInput()
	for x := 0; x < 1; x++ {
		if err := i.Collect(); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second * 1)
	}
	if len(i.collectCache) < 1 {
		t.Error("Failed to collect, no data returned")
	}
	tmap := map[string]bool{}
	for _, v := range i.collectCache {
		m, ok := v.(*diskMeasurement)
		if !ok {
			t.Error("v expect to be *diskMeasurement")
			continue
		}

		tmap[m.ts.String()] = true
	}
	if len(tmap) != 1 {
		t.Error("Need to clear collectCache.")
	}
}

func TestWrapUint64(t *testing.T) {
	tu.Assert(t, wrapUint64(math.MaxInt64+1) == -1, "")
	tu.Assert(t, wrapUint64(math.MaxInt64-1) == math.MaxInt64-1, "")
	tu.Assert(t, wrapUint64(1023) == 1023, "")
}
