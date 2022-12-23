// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestB64enc(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected interface{}
		fail     bool
	}{
		{
			name:     "normal",
			pl:       "json(_, `str`); b64enc(`str`)",
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "MTM4MzgxMzA1MTc=",
			fail:     false,
		},
		{
			name:     "normal",
			pl:       "json(_, `str`); b64enc(`str`)",
			in:       `{"str": "hello, world"}`,
			outKey:   "str",
			expected: "aGVsbG8sIHdvcmxk",
			fail:     false,
		},
		{
			name:     "normal",
			pl:       "json(_, `str`); b64enc(`str`)",
			in:       `{"str": "你好"}`,
			outKey:   "str",
			expected: "5L2g5aW9",
			fail:     false,
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

			_, _, f, _, _, err := runScript(runner, "test", nil, map[string]interface{}{
				"message": tc.in,
			}, time.Now())
			if err != nil {
				t.Fatal(err)
			}

			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, err)
			}
			t.Log(f)
			v := f[tc.outKey]
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}
