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

func TestB64dec(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected interface{}
		fail     bool
	}{
		{
			name:     "normal",
			pl:       "json(_, `str`); b64dec(`str`)",
			in:       `{"str": "MTM4MzgxMzA1MTc="}`,
			outKey:   "str",
			expected: "13838130517",
			fail:     false,
		},
		{
			name:     "normal",
			pl:       "json(_, `str`); b64dec(`str`)",
			in:       `{"str": "aGVsbG8sIHdvcmxk"}`,
			outKey:   "str",
			expected: "hello, world",
			fail:     false,
		},
		{
			name:     "normal",
			pl:       "json(_, `str`); b64dec(`str`)",
			in:       `{"str": "5L2g5aW9"}`,
			outKey:   "str",
			expected: "你好",
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

			pt := ptinput.NewPlPoint(
				point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				t.Fatal(errR)
			}

			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, err)
			}
			v, _, _ := pt.Get(tc.outKey)
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}
