package io

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestNewPoint(t *testing.T) {
	cases := []struct {
		tname, name, mtype, expect string
		t                          map[string]string
		globalTags                 map[string]string
		f                          map[string]interface{}
		ts                         time.Time
		fail                       bool
		withoutGlobalTags          bool
		maxTags, maxFields         int
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
			mtype: datakit.Object,
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

		{
			tname: "with point in tag-field key",
			name:  "abc",
			mtype: datakit.Logging,
			t:     map[string]string{"t1": "abc", "t.2": "xyz"},
			f:     map[string]interface{}{"f1": 123, "f.2": "def"},
			fail:  true,
		},

		{
			tname:     "both exceed max field/tag count",
			name:      "abc",
			t:         map[string]string{"t1": "abc", "t2": "xyz"},
			f:         map[string]interface{}{"f1": 123, "f2": "def"},
			fail:      true,
			maxFields: 1,
			maxTags:   1,
		},

		{
			tname: "exceed max field count",

			name:      "abc",
			t:         map[string]string{"t1": "abc", "t2": "xyz"},
			f:         map[string]interface{}{"f1": 123, "f2": "def"},
			fail:      true,
			maxFields: 1,
		},

		{
			tname:   "exceed max tag count",
			name:    "abc",
			t:       map[string]string{"t1": "abc", "t2": "xyz"},
			f:       map[string]interface{}{"f1": 123},
			fail:    true,
			maxTags: 1,
		},

		{
			tname: "with disabled tag key source",
			name:  "abc",
			mtype: datakit.Logging,
			t:     map[string]string{"source": "s1"},
			f:     map[string]interface{}{"f1": 123},
			fail:  true,
		},

		{
			tname: "with disabled field key",
			name:  "abc",
			mtype: datakit.Object,
			t:     map[string]string{},
			f:     map[string]interface{}{"class": 123},
			fail:  true,
		},

		{
			tname: "with disabled tag key log_type",
			name:  "abc",
			mtype: datakit.Logging,
			t:     map[string]string{"log_type": "s1"},
			f:     map[string]interface{}{"f1": 123},
			fail:  true,
		},

		{
			tname: "with disabled field key log_type",
			name:  "abc",
			mtype: datakit.Logging,
			t:     map[string]string{},
			f:     map[string]interface{}{"log_type": 123},
			fail:  true,
		},

		{
			tname: "normal",

			name:   "abc",
			t:      map[string]string{},
			f:      map[string]interface{}{"f1": 123},
			ts:     time.Unix(0, 123),
			expect: "abc f1=123i 123",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			var pt *Point
			var err error
			extraTags = nil

			MaxTags = 256
			MaxFields = 1024

			if tc.maxTags > 0 {
				MaxTags = tc.maxTags
			}

			if tc.maxFields > 0 {
				MaxFields = tc.maxFields
			}

			if tc.globalTags != nil {
				extraTags = tc.globalTags
			}

			switch tc.mtype {
			case datakit.Metric, datakit.Logging, datakit.Object:
				pt, err = NewPoint(tc.name, tc.t, tc.f,
					&PointOption{
						Category: tc.mtype,
						Time:     tc.ts,
					})
			default:
				if tc.withoutGlobalTags {
					pt, err = NewPoint(tc.name, tc.t, tc.f,
						&PointOption{
							Time:              tc.ts,
							DisableGlobalTags: true,
							Category:          datakit.Metric,
						})
				} else {
					pt, err = NewPoint(tc.name, tc.t, tc.f, &PointOption{
						Time:     tc.ts,
						Category: datakit.Metric,
					})
				}
			}

			if tc.fail {
				// tu.NotOk(t, err, "")
				t.Logf("[expected] %s", err)
				return
			}

			tu.Ok(t, err)
			x := pt.String()
			tu.Equals(t, tc.expect, x)
		})
	}
}

var benchCases = []struct {
	name     string
	t        map[string]string
	f        map[string]interface{}
	category string
}{
	{
		name: "3-tag-10-field-logging",
		t: map[string]string{
			"t1": "val1",
			"t2": "val2",
			"t3": "val3",
		},
		f: map[string]interface{}{
			"f1":  123,
			"f2":  "abc",
			"f3":  45.6,
			"f4":  123,
			"f5":  "abc",
			"f6":  45.6,
			"f7":  123,
			"f8":  "abc",
			"f9":  45.6,
			"f10": false,
		},
		category: datakit.Logging,
	},

	{
		name: "3-tag-10-field-metric",
		t: map[string]string{
			"t1": "val1",
			"t2": "val2",
			"t3": "val3",
		},
		f: map[string]interface{}{
			"f1":  123,
			"f2":  "abc",
			"f3":  45.6,
			"f4":  123,
			"f5":  "abc",
			"f6":  45.6,
			"f7":  123,
			"f8":  "abc",
			"f9":  45.6,
			"f10": false,
		},
		category: datakit.Metric,
	},

	{
		name: "3-tag-10-field-object",
		t: map[string]string{
			"t1": "val1",
			"t2": "val2",
			"t3": "val3",
		},
		f: map[string]interface{}{
			"f1":  123,
			"f2":  "abc",
			"f3":  45.6,
			"f4":  123,
			"f5":  "abc",
			"f6":  45.6,
			"f7":  123,
			"f8":  "abc",
			"f9":  45.6,
			"f10": false,
		},
		category: datakit.Object,
	},

	{
		name: "3-tag-10-long-field-object",
		t: map[string]string{
			"t1": "val1",
			"t2": "val2",
			"t3": "val3",
		},
		f: map[string]interface{}{
			"f1122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f2122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f3122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f4122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f5122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f6122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f7122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f8122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f9122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f10": false,
		},
		category: datakit.Object,
	},

	{
		name: "3-long-tag-10-long-field-object",
		t: map[string]string{
			"t1122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "val1",
			"t2122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "val2",
			"t3122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "val3",
		},
		f: map[string]interface{}{
			"f1122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f2122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f3122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f4122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f5122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f6122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f7122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 123,
			"f8122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": "abc",
			"f9122342143214321412342314321423423143214321432143214321432143214321432j14h32jkl14h32jkl": 45.6,
			"f10": false,
		},
		category: datakit.Object,
	},
}

func BenchmarkMakePoint(b *testing.B) {
	for _, tc := range benchCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := MakePoint(tc.name, tc.t, tc.f)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}

func BenchmarkNewPoint(b *testing.B) {
	for _, tc := range benchCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := NewPoint(tc.name, tc.t, tc.f, &PointOption{
					Category: tc.category,
				})
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}
