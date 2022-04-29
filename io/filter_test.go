// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/json"
	"testing"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type dwMock struct {
	pullCount int
}

func (dw *dwMock) Pull() ([]byte, error) {
	filters := map[string][]string{
		"logging": {
			`{ source = "test1" and ( f1 in ["1", "2", "3"] )}`,
			`{ source = "test2" and ( f1 in ["1", "2", "3"] )}`,
		},

		"tracing": {
			`{ service = "test1" and ( f1 in ["1", "2", "3"] or t1 match [ 'abc.*'])}`,
			`{ service = re("test2") and ( f1 in ["1", "2", "3"] or t1 match [ 'def.*'])}`,
		},
	}

	dw.pullCount++

	return json.Marshal(&filterPull{Filters: filters, PullInterval: time.Duration(dw.pullCount) * time.Millisecond})
}

func TestFilter(t *testing.T) {
	f := newFilter(&dwMock{})

	cases := []struct {
		name      string
		pts       string
		category  string
		expectPts int
	}{
		{
			pts: `test1 f1="1",f2=2i,f3=3 124
test1 f1="2",f2=2i,f3=3 124
test1 f1="3",f2=2i,f3=3 125
			`,
			category:  datakit.Logging,
			expectPts: 0,
		},

		{
			pts: `test1,service=test1 f1="1",f2=2i,f3=3 123
test1,service=test1 f1="1",f2=2i,f3=3 124
test1,service=test1 f1="1",f2=2i,f3=3 125`,
			category:  datakit.Tracing,
			expectPts: 0,
		},
	}

	f.pull()
	for k, v := range f.conditions {
		t.Logf("%s: %s", k, v)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts, err := lp.ParsePoints([]byte(tc.pts), nil)
			if err != nil {
				t.Error(err)
				return
			}

			after, _ := f.doFilter(tc.category, WrapPoint(pts))
			tu.Assert(t, len(after) == tc.expectPts, "expect %d pts, got %d", tc.expectPts, len(after))
		})
	}
}

func TestPull(t *testing.T) {
	f := newFilter(&dwMock{})

	round := 3
	for i := 0; i < round; i++ {
		f.pull()
		// test if reset tick ok
		<-f.tick.C
	}

	tu.Assert(t, f.pullInterval == time.Millisecond*time.Duration(round), "expect %ds, got %s", round, f.pullInterval)
}
