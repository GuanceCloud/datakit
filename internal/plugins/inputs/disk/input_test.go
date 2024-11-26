// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestCollect(t *testing.T) {
	i := defaultInput()
	intervalMillSec := i.Interval.Milliseconds()
	var lastAlignTime int64

	for x := 0; x < 1; x++ {
		tn := time.Now()
		lastAlignTime = inputs.AlignTimeMillSec(tn, lastAlignTime, intervalMillSec)
		if err := i.collect(lastAlignTime * 1e6); err != nil {
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
