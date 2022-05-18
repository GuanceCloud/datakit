// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"strings"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestUppercase(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected string
		fail     bool
	}{
		{
			name: "normal",
			pl: `
json(_, a.third)
uppercase(a.third)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "a.third",
			expected: "ABC",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, age)
uppercase(age)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "age",
			expected: "47",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.forth)
uppercase(a.forth)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":"1a2B3c/d"},"age":47}`,
			outKey:   "a.forth",
			expected: strings.ToUpper("1a2B3C/d"),
			fail:     false,
		},

		{
			name: "too many args",
			pl: `
		json(_, a.forth)
		uppercase(a.forth, "someArg")
		`,
			in:   `{"a":{"first":2.3,"second":2,"third":"abc","forth":"1a2B3c/d"},"age":47}`,
			fail: true,
		},

		{
			name: "invalid arg type",
			pl: `
		json(_, a.forth)
		uppercase("hello")
		`,
			in:   `{"a":{"first":2.3,"second":2,"third":"abc","forth":"1a2B3c/d"},"age":47}`,
			fail: true,
		},
	}

	for idx, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.pl)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Errorf("[%d] failed: %s", idx, err)
				}
				return
			}
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Error(err)
				}
			} else {
				t.Log(ret)
				v := ret.Fields[tc.outKey]
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
