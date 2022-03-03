package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestGroupIn(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name:     "normal",
			pl:       `json(_, status) group_in(status, [true], "ok", "newkey")`,
			in:       `{"status": true,"age":"47"}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_in(status, [true], "ok", "newkey")`,
			in:       `{"status": true,"age":"47"}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_in(status, [true], "ok", "newkey")`,
			in:       `{"status": "aa","age":"47"}`,
			expected: "aa",
			outkey:   "status",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_in(status, ["aa"], "ok", "newkey")`,
			in:       `{"status": "aa","age":"47"}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name:     "normal",
			in:       `{"status": "test","age":"47"}`,
			pl:       `json(_, status) group_in(status, [200, "test"], 119)`,
			expected: "test",
			outkey:   "status",
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

			v, err := runner.GetContent(tc.outkey)
			tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)

			t.Logf("[%d] PASS", idx)
		})
	}
}
