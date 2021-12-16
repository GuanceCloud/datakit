package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestGroup(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [200, 400], false, newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: false,
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [200, 400], 10, newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: int64(10),
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [200, 400], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [200, 299], "ok")`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "status",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [200, 299], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: "ok",
			outkey:   "newkey",
		},
		{
			name:     "normal",
			pl:       `json(_, status) group_between(status, [300, 400], "ok", newkey)`,
			in:       `{"status": 200,"age":47}`,
			expected: nil,
			outkey:   "newkey",
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
