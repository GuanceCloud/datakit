package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParseDuration(t *testing.T) {
	cases := []struct {
		name     string
		pl, in   string
		outKey   string
		expected interface{}
		fail     bool
	}{
		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1s"}`,
			outKey:   "str",
			expected: int64(time.Second),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1ms"}`,
			outKey:   "str",
			expected: int64(time.Millisecond),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1us"}`,
			outKey:   "str",
			expected: int64(time.Microsecond),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1µs"}`,
			outKey:   "str",
			expected: int64(time.Microsecond),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1m"}`,
			outKey:   "str",
			expected: int64(time.Minute),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1h"}`,
			outKey:   "str",
			expected: int64(time.Hour),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "-23h"}`,
			outKey:   "str",
			expected: -23 * int64(time.Hour),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "-23ns"}`,
			outKey:   "str",
			expected: int64(-23),
			fail:     false,
		},

		{
			name:     "normal",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "-2.3s"}`,
			outKey:   "str",
			expected: int64(time.Second*-2 - 300*time.Millisecond),
			fail:     false,
		},

		{
			name:     "invalid input string",
			pl:       `json(_, str) parse_duration(str)`,
			in:       `{"str": "1uuus"}`,
			outKey:   "str",
			expected: "1uuus",
			fail:     false,
		},

		{
			name: "too many args",
			pl:   `json(_, str) parse_duration(str, 1)`,
			in:   `{"str": "1uuus"}`,
			fail: true,
		},

		{
			name: "invalid input type",
			pl:   `json(_, str) parse_duration(str)`,
			in:   `{"str": 1}`,
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
