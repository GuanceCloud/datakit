package io

import (
	"encoding/json"
	"testing"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/parser"
)

type dwMock struct{}

func (dw *dwMock) Pull() ([]byte, error) {
	filters := map[string][]string{
		"logging": {
			`{ source = "test1" and ( f1 in ["1", "2", "3"] )}`,
			`{ source = "test2" and ( f1 in ["1", "2", "3"] )}`,
		},

		"tracing": {
			`{ service = "test1" and ( f1 in ["1", "2", "3"] or t1 contain [ 'abc.*'])}`,
			`{ service = re("test2") and ( f1 in ["1", "2", "3"] or t1 contain [ 'def.*'])}`,
		},
	}

	return json.Marshal(&filterPull{Filters: filters, PullInterval: time.Second * 10})
}

func TestFilter(t *testing.T) {
	f := filter{
		conditions: map[string]parser.WhereConditions{},
		dw:         &dwMock{},
	}

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

			after := f.filter(tc.category, WrapPoint(pts))
			tu.Assert(t, len(after) == tc.expectPts, "expect %d pts, got %d", tc.expectPts, len(after))
		})
	}
}
