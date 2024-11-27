// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"fmt"
	T "testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestTickPass(t *T.T) {
	cases := []struct {
		tickms int
	}{
		{1},
		{2},
		{3},
		{4},
		{5},
		{6},
		{7},
		{8},
		{9},
		{10},

		{11},
		{12},
		{13},
		{14},
		{15},
		{16},
		{17},
		{18},
		{19},
		{110},

		{111},
		{112},
		{113},
		{114},
		{115},
		{116},
		{117},
		{118},
		{119},
		{1110},
	}

	// nolint: durationcheck
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%dms", tc.tickms), func(t *T.T) {
			t.Skip()
			n := time.Duration(tc.tickms)
			sleep := n * time.Millisecond / 10

			tick := time.NewTicker(n * time.Millisecond)
			lastts := time.Now().UnixMilli()
			notRound := 0

			for i := 0; i < 100; i++ {
				for tt := range tick.C {
					now := tt.UnixMilli()
					round := (now - lastts) % int64(n)
					if round != 0 {
						notRound++
					}

					nextts := inputs.AlignTimeMillSec(tt, lastts, int64(n))
					lastts = nextts

					// working time
					time.Sleep(sleep)
				}
			}

			t.Logf("not round: %d", notRound)
		})
	}
}
