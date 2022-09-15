// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/engine"
)

func TestGrok(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "normal",
			pl: `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")`,
			in:       "12:13:14.123",
			expected: "14.123",
			outkey:   "second",
		},
		{
			name: "normal",
			pl: `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")`,
			in:       "12:13:14",
			expected: "13",
			outkey:   "minute",
		},
		{
			name: "normal",
			pl: `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")`,
			in:       "12:13:14",
			expected: "12",
			outkey:   "hour",
		},
		{
			name: "normal",
			pl: `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour}:%{_minute:minute}(?::%{_second:second})([^0-9]?)")
grok(_, "%{time}")`,
			in:       "12:13:14",
			expected: "14",
			outkey:   "second",
		},
		{
			name: "normal",
			pl: `
add_pattern("time", "%{NUMBER:time:float}")
grok(_, '''%{time}
%{WORD:word:string}
	%{WORD:code:int}
%{WORD:w1}''')`,
			in: `1.1
s
	123cvf
aa222`,
			expected: int64(0),
			outkey:   "code",
		},
		{
			name: "normal",
			pl: `
add_pattern("time", "%{NUMBER:time:float}")
grok(_, '''%{time}
%{WORD:word:string}
	%{WORD:code:int}
%{WORD:w1}''')`,
			in: `1.1
s
	123
aa222`,
			expected: int64(123),
			outkey:   "code",
		},
		{
			name: "normal",
			pl: `
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour:string}:%{_minute:minute:int}(?::%{_second:second:float})([^0-9]?)")
grok(_, "%{WORD:date} %{time}")`,
			in:       "2021/1/11 2:13:14.123",
			expected: float64(14.123),
			outkey:   "second",
		},
		{
			name: "trim_space",
			in:   " not_space ",
			pl: `add_pattern("d", "[\\s\\S]*")
			grok(_, "%{d:item}")`,
			expected: "not_space",
			outkey:   "item",
		},
		{
			name: "trim_space, enable",
			in:   " not_space ",
			pl: `add_pattern("d", "[\\s\\S]*")
			grok(_, "%{d:item}", true)`,
			expected: "not_space",
			outkey:   "item",
		},
		{
			name: "trim_space, disable",
			in:   " not_space ",
			pl: `add_pattern("d", "[\\s\\S]*")
			grok(_, "%{d:item}", false)`,
			expected: " not_space ",
			outkey:   "item",
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

			_, _, f, _, _, err := engine.RunScript(runner, "test", nil, map[string]interface{}{
				"message": tc.in,
			}, time.Now())

			tu.Equals(t, err, nil)

			t.Log(f)

			v, ok := f[tc.outkey]
			tu.Equals(t, true, ok)
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}
