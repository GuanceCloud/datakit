// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
)

func TestStrfmt(t *testing.T) {
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
json(_, a.second, a_second)
json(_, a.third, a_third)
cast(a_second, "int")
json(_, a.forth, a_forth)
strfmt(bb, "%d %s %v", a_second, a_third, a_forth)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2 abc true",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first, a_first)
cast(a_first, "float")
strfmt(bb, "%.4f", a_first)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.3000",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first, a_first)
cast(a_first, "float")
strfmt(bb, "%.4f%d", a_first, 3)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.30003",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first, a_first)
cast(a_first, "float")
strfmt(bb, "%.4f%.1f", a_first, 3.5)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.30003.5",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.forth, a_forth)
strfmt(bb, "%v%s", a_forth, "tone")
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abcd","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "truetone",
			fail:     false,
		},

		{
			name: "not enough arg",
			pl: `
		json(_, a.first, a_first)
		cast(a_first, "float")
		strfmt(bb)
		`,
			in:   `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			fail: true,
		},

		{
			name: "incorrect arg type",
			pl: `
		json(_, a.first, a_first)
		cast(a_first, "float")
		strfmt(bb, 1)
		`,
			in:   `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
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

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR.Error())
			}

			assert.Equal(t, "test", pt.GetPtName())

			if v, _, err := pt.Get(tc.outKey); err != nil {
				if !tc.fail {
					t.Errorf("[%d]expect error: %s", idx, err)
				}
			} else {
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
