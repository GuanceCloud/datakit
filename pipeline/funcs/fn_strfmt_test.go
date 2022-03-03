package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
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
json(_, a.second)
json(_, a.third)
cast(a.second, "int")
json(_, a.forth)
strfmt(bb, "%d %s %v", a.second, a.third, a.forth)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2 abc true",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first)
cast(a.first, "float")
strfmt(bb, "%.4f", a.first)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.3000",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first)
cast(a.first, "float")
strfmt(bb, "%.4f%d", a.first, 3)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.30003",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.first)
cast(a.first, "float")
strfmt(bb, "%.4f%.1f", a.first, 3.5)
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "2.30003.5",
			fail:     false,
		},

		{
			name: "normal",
			pl: `
json(_, a.forth)
strfmt(bb, "%v%s", a.forth, "tone")
`,
			in:       `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			outKey:   "bb",
			expected: "true'tone'",
			fail:     false,
		},

		{
			name: "not enough arg",
			pl: `
		json(_, a.first)
		cast(a.first, "float")
		strfmt(bb)
		`,
			in:   `{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}`,
			fail: true,
		},

		{
			name: "incorrect arg type",
			pl: `
		json(_, a.first)
		cast(a.first, "float")
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

			err = runner.Run(tc.in)
			if err != nil {
				if tc.fail {
					t.Logf("[%d]expect error: %s", idx, err)
				} else {
					t.Error(err)
				}
			} else {
				t.Log(runner.Result())
				v, _ := runner.GetContent(tc.outKey)
				tu.Equals(t, tc.expected, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
