package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestNullIf(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, "1")`,
			in:       `{"a":{"first": 1,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: float64(1),
			outkey:   "a.first",
		},
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, 1)`,
			in:       `{"a":{"first": "1","second":2,"third":"aBC","forth":true},"age":47}`,
			expected: "1",
			outkey:   "a.first",
		},
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, "")`,
			in:       `{"a":{"first": "","second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a.first",
		},
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, nil)`,
			in:       `{"a":{"first": null,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a.first",
		},
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, true)`,
			in:       `{"a":{"first": true,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a.first",
		},
		{
			name:     "normal",
			pl:       `json(_, a.first) nullif(a.first, 2.3)`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a.first",
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
			tu.Equals(t, nil, err)
			t.Log(runner.Result())

			v, _ := runner.GetContent(tc.outkey)
			// tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)

			t.Logf("[%d] PASS", idx)
		})
	}
}
