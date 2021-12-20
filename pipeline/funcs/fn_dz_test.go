package funcs

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestDz(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected string
		fail     bool
	}{
		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [8, 13])`,
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1383813****",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [8, 11])`,
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1383813****",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [2, 4])`,
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1***8130517",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [1, 1])`,
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "*3838130517",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [0, 3])`,
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "***38130517",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) cover(str, [2, 2])`,
			in:       `{"str": "刘少波"}`,
			outKey:   "str",
			expected: "刘＊波",
			fail:     false,
		},

		{
			name:     "odd range",
			pl:       `json(_, str) cover(str, [1, 100])`,
			in:       `{"str": "刘少波"}`,
			outKey:   "str",
			expected: "＊＊＊",
			fail:     false,
		},

		{
			name:     "odd range",
			pl:       `json(_, str) cover(str, [-1, 3])`,
			in:       `{"str": "刘少波"}`,
			outKey:   "str",
			expected: "＊＊＊",
			fail:     false,
		},

		{
			name: "invalid range",
			pl:   `json(_, str) cover(str, [3, 2])`,
			in:   `{"str": "刘少波"}`,
			fail: true,
		},

		{
			name: "not enough args",
			pl:   `json(_, str) cover(str)`,
			in:   `{"str": "刘少波"}`,
			fail: true,
		},

		{
			name: "invalid range",
			pl:   `json(_, str) cover(str, ["刘", "波"])`,
			in:   `{"str": "刘少波"}`,
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
