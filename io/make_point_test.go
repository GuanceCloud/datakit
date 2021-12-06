package io

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestMakePoint(t *testing.T) {
	cases := []struct {
		tname string
		fail  bool

		name   string
		tags   map[string]string
		fields map[string]interface{}
		extags map[string]string
		t      time.Time
		expect string
		maxTags,
		maxFields int
	}{

		{
			tname: "with point in tag/field key",

			name:   "abc",
			tags:   map[string]string{"t1": "abc", "t.2": "xyz"},
			fields: map[string]interface{}{"f1": 123, "f.2": "def"},
			fail:   true,
		},

		{
			tname: "both exceed max field/tag count",

			name:      "abc",
			tags:      map[string]string{"t1": "abc", "t2": "xyz"},
			fields:    map[string]interface{}{"f1": 123, "f2": "def"},
			fail:      true,
			maxFields: 1,
			maxTags:   1,
		},

		{
			tname: "exceed max field count",

			name:      "abc",
			tags:      map[string]string{"t1": "abc", "t2": "xyz"},
			fields:    map[string]interface{}{"f1": 123, "f2": "def"},
			fail:      true,
			maxFields: 1,
		},

		{
			tname: "exceed max tag count",

			name:    "abc",
			tags:    map[string]string{"t1": "abc", "t2": "xyz"},
			fields:  map[string]interface{}{"f1": 123},
			fail:    true,
			maxTags: 1,
		},

		{
			tname: "with disabled tag key",

			name:   "abc",
			tags:   map[string]string{"source": "s1"},
			fields: map[string]interface{}{"f1": 123},
			fail:   true,
		},

		{
			tname: "with disabled tag key",

			name:   "abc",
			tags:   map[string]string{"source": "s1"},
			fields: map[string]interface{}{"f1": 123},
			fail:   true,
		},
		{
			tname: "with disabled field key",

			name:   "abc",
			tags:   map[string]string{},
			fields: map[string]interface{}{"class": 123},
			fail:   true,
		},
		{
			tname: "normal",

			name:   "abc",
			tags:   map[string]string{},
			fields: map[string]interface{}{"f1": 123},
			t:      time.Unix(0, 123),
			expect: "abc f1=123i 123",
		},
	}

	for _, tc := range cases {
		t.Run(tc.tname, func(t *testing.T) {
			if tc.maxTags > 0 {
				__maxTags = tc.maxTags
			}

			if tc.maxFields > 0 {
				__maxFields = tc.maxFields
			}

			p, err := MakePoint(tc.name, tc.tags, tc.fields, tc.t)
			if tc.fail {
				tu.Assert(t, err != nil, "expect error, got nothing")
				t.Logf("[expect error] %s", err)
			} else {
				tu.Assert(t, err == nil, "expect ok, got error: %s", err)
				tu.Equals(t, tc.expect, p.String())
			}
		})
		__maxTags = MAX_TAGS
		__maxFields = MAX_FIELDS
	}
}
