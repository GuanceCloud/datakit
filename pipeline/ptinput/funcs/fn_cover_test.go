// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestDz(t *testing.T) {
	cases := []struct {
		name     string
		outKey   string
		pl, in   string
		expected interface{}
		fail     bool
	}{
		{
			name:     "normal",
			pl:       "json(_, `str`); cover(`str`, [8, 13])",
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1383813****",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       "json(_, `str`) ;cover(`str`, [8, 11])",
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1383813****",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       "json(_, `str`); cover(`str`, [2, 4])",
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "1***8130517",
			fail:     false,
		},

		{
			name:     "normal",
			pl:       "json(_, `str`) ;cover(`str`, [1, 1])",
			in:       `{"str": "13838130517"}`,
			outKey:   "str",
			expected: "*3838130517",
			fail:     false,
		},

		{
			name:     "odd range",
			pl:       "json(_, `str`); cover(`str`, [1, 100])",
			in:       `{"str": "刘少波"}`,
			outKey:   "str",
			expected: "＊＊＊",
			fail:     false,
		},

		{
			name: "invalid range",
			pl:   "json(_, `str`); cover(`str`, [3, 2])",
			in:   `{"str": "刘少波"}`,
			fail: true,
		},

		{
			name: "not enough args",
			pl:   "json(_, `str`) ;cover(`str`)",
			in:   `{"str": "刘少波"}`,
			fail: true,
		},

		{
			name: "invalid range",
			pl:   "json(_, `str`); cover(`str`, [\"刘\", \"波\"])",
			in:   `{"str": "刘少波"}`,
			fail: true,
		},

		{
			name: "normal",
			pl:   "json(_, `str`) ;cover(`str`, [1, 2])",
			in:   `{"str": 123456}`,
			fail: true,
		},

		{
			name: "normal",
			pl:   "json(_, `str`) ;cover(`str`, [-1, -2])",
			in:   `{"str": 123456}`,
			fail: true,
		},

		{
			name: "normal",
			pl:   "json(_, `str`); cover(`str`, [-1, -2])",
			in:   `{"str": 123456}`,
			fail: true,
		},

		{
			name:     "normal",
			pl:       "json(_, `str`) ; cast(`str`, \"int\"); cover(`str`, [-2, 10000])",
			in:       `{"str": 123456}`,
			outKey:   "str",
			expected: int64(123456),
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

			pt := ptinput.GetPoint()
			ptinput.InitPt(pt, "test", nil, map[string]any{"message": tc.in}, time.Now())
			errR := runScript(runner, pt)

			if errR != nil {
				ptinput.PutPoint(pt)
				t.Fatal(errR)
			}

			if tc.fail {
				t.Logf("[%d]expect error: %s", idx, err)
			}
			v := pt.Fields[tc.outKey]
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}
