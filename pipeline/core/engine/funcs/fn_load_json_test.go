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

func TestLoadJson(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "normal",
			pl: `abc = load_json(_)
			add_key(abc, abc["a"]["first"])`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: float64(2.3),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = load_json(_)
			add_key(abc, abc["a"]["first"][-1])
			add_key(len_abc, len(load_json(abc["a"]["ff"])))`,
			in:       `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			expected: int64(2),
			outkey:   "len_abc",
		},
		{
			name: "normal",
			pl: `abc = load_json(_)
			add_key(abc, abc[-1])`,
			in:       `[2.2, 1.1]`,
			expected: float64(1.1),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = load_json(_)
			add_key(abc, len(abc))`,
			in:       `[]`,
			expected: int64(0),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = load_json("1")
			add_key(abc)`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: float64(1),
			outkey:   "abc",
		},
		{
			name: "normal",
			pl: `abc = load_json("true")
			add_key(abc)`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: true,
			outkey:   "abc",
		},

		{
			name: "normal",
			pl: `abc = load_json("null")
			add_key(abc)`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
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
			_, _, f, _, _, err := engine.RunScript(runner, "test", nil, map[string]interface{}{
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
