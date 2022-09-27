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

func TestLen(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "normal",
			pl: `abc = ["1", "2"]
			add_key(abc, len(abc))`,
			in:       `test`,
			expected: int64(2),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = []
			add_key(abc, len(abc))`,
			in:       `test`,
			expected: int64(0),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = {"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}
			add_key(abc, len(abc["a"]))`,
			in:       `test`,
			expected: int64(4),
			outkey:   "abc",
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

			tu.Equals(t, nil, err)

			t.Log(f)
			v := f[tc.outkey]
			// tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)

			t.Logf("[%d] PASS", idx)
		})
	}
}
