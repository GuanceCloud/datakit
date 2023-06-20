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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
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
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1s"}`,
			outKey:   "str",
			expected: int64(time.Second),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1ms"}`,
			outKey:   "str",
			expected: int64(time.Millisecond),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1us"}`,
			outKey:   "str",
			expected: int64(time.Microsecond),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1µs"}`,
			outKey:   "str",
			expected: int64(time.Microsecond),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1m"}`,
			outKey:   "str",
			expected: int64(time.Minute),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1h"}`,
			outKey:   "str",
			expected: int64(time.Hour),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "-23h"}`,
			outKey:   "str",
			expected: -23 * int64(time.Hour),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "-23ns"}`,
			outKey:   "str",
			expected: int64(-23),
			fail:     false,
		},

		{
			name: "normal",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "-2.3s"}`,
			outKey:   "str",
			expected: int64(time.Second*-2 - 300*time.Millisecond),
			fail:     false,
		},

		{
			name: "invalid input string",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:       `{"str": "1uuus"}`,
			outKey:   "str",
			expected: "1uuus",
			fail:     false,
		},

		{
			name: "too many args",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
			in:   `{"str": "1uuus"}`,
			fail: true,
		},

		{
			name: "invalid input type",
			pl: "json(_, `str`)\n" +
				"parse_duration(`str`)",
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

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR)
			}

			if v, _, err := pt.Get(tc.outKey); err != nil {
				if !tc.fail {
					t.Errorf("[%d]expect error: %s", idx, err)
				}
			} else {
				tu.Equals(t, tc.expected, v)
			}
		})
	}
}
