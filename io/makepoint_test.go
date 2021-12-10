package io

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestMakePoint(t *testing.T) {
	cases := []struct {
		tname, name, mtype, expect string
		t                          map[string]string
		globalTags                 map[string]string
		f                          map[string]interface{}
		ts                         time.Time
		fail                       bool
		withoutGlobalTags          bool
	}{
		{
			tname:  "base",
			name:   "abc",
			ts:     time.Unix(0, 123),
			t:      map[string]string{"t1": "tval1"},
			f:      map[string]interface{}{"f1": 12},
			expect: "abc,t1=tval1 f1=12i 123",
		},

		{
			tname:  "metric with point in field key",
			name:   "abc",
			mtype:  datakit.Metric,
			ts:     time.Unix(0, 123),
			t:      map[string]string{"t1": "tval1"},
			f:      map[string]interface{}{"f.1": 12},
			expect: "abc,t1=tval1 f.1=12i 123",
		},

		{
			tname:  "metric with point in tag key",
			name:   "abc",
			mtype:  datakit.Metric,
			ts:     time.Unix(0, 123),
			t:      map[string]string{"t.1": "tval1"},
			f:      map[string]interface{}{"f1": 12},
			expect: "abc,t.1=tval1 f1=12i 123",
		},

		{
			tname: "with point in t/f key on non-metric type",
			name:  "abc",
			ts:    time.Unix(0, 123),
			t:     map[string]string{"t1": "tval1"},
			f:     map[string]interface{}{"f.1": 12},
			fail:  true,
		},

		{
			tname:      "with global tags added",
			name:       "abc",
			ts:         time.Unix(0, 123),
			t:          map[string]string{"t1": "tval1"},
			globalTags: map[string]string{"gt1": "gtval1"},
			f:          map[string]interface{}{"f1": 12},
			expect:     "abc,gt1=gtval1,t1=tval1 f1=12i 123",
		},

		{
			tname:             "without global tags",
			name:              "abc",
			ts:                time.Unix(0, 123),
			t:                 map[string]string{"t1": "tval1"},
			globalTags:        map[string]string{"gt1": "gtval1"},
			f:                 map[string]interface{}{"f1": 12},
			expect:            "abc,t1=tval1 f1=12i 123",
			withoutGlobalTags: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var pt *Point
			var err error
			extraTags = nil

			if tc.globalTags != nil {
				extraTags = tc.globalTags
			}

			switch tc.mtype {
			case datakit.Metric:
				pt, err = MakeTypedPoint(tc.name, tc.mtype, tc.t, tc.f, tc.ts)
			default:
				if tc.withoutGlobalTags {
					pt, err = MakePointWithoutGlobalTags(tc.name, tc.t, tc.f, tc.ts)
				} else {
					pt, err = MakePoint(tc.name, tc.t, tc.f, tc.ts)
				}
			}

			if tc.fail {
				tu.NotOk(t, err, "")
				t.Logf("[expected] %s", err)
				return
			}

			tu.Ok(t, err)
			x := pt.String()
			tu.Equals(t, tc.expect, x)
		})
	}
}
