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
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, time.Now())
			tu.Equals(t, nil, err)
			tu.Equals(t, nil, ret.Error)

			t.Log(ret)
			v, ok := ret.Fields[tc.outkey]
			tu.Equals(t, true, ok)
			tu.Equals(t, tc.expected, v)
			t.Logf("[%d] PASS", idx)
		})
	}
}
