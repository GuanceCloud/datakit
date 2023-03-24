// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ptinput"
)

func TestKV(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     []interface{}
		fail         bool
		outkey       []string
	}{
		{
			name:     "normal",
			pl:       `kv_split(_)`,
			in:       `a=1 b=2 c=3`,
			expected: []any{nil, nil, nil},
			outkey:   []string{"a", "b", "c"},
		},
		{
			name: "normal1",
			pl: `kv_split(_,
				include_keys=["a", "b","c"])`,
			in:       `a=1, b=2 c=3`,
			expected: []any{"1,", "2", "3"},
			outkey:   []string{"a", "b", "c"},
		},
		{
			name: "trim_value",
			pl: `kv_split(_, trim_value=",", 
			include_keys=["a", "b","c"])`,
			in:       `a=1, b=2 c=3`,
			expected: []any{"1", "2", "3"},
			outkey:   []string{"a", "b", "c"},
		},
		{
			name: "trim_key_skip_empty_str_key",
			pl: `kv_split(_, trim_value=",", trim_key="a",
			include_keys=["a", "b","c"])`,
			in:       `a=1, b=2 c=3`,
			expected: []any{nil, nil, "2", "3"},
			outkey:   []string{"", "a", "b", "c"},
		},
		{
			name: "prefix",
			pl: `kv_split(_, prefix="prefix_",trim_value=",", trim_key="a",
			include_keys=["a", "b","c"])`,
			in:       `a=1, b=2 c=3`,
			expected: []any{nil, nil, "2", "3"},
			outkey:   []string{"", "prefix_", "prefix_b", "prefix_c"},
		},
		{
			name:     "prefix",
			pl:       `kv_split(_, include_keys= ["b", ],prefix="prefix_",trim_value=",", trim_key="a")`,
			in:       `a=1, b=2 c=3`,
			expected: []any{nil, nil, "2", nil},
			outkey:   []string{"", "prefix_", "prefix_b", "prefix_c"},
		},
		{
			name: "value_split_pattern",
			pl: `kv_split(_, include_keys= ["b", ], field_split_pattern="\\+", value_split_pattern="::",
						prefix="prefix_",trim_value=",", trim_key="a")`,
			in:       `a::1,+b::2+c::3`,
			expected: []any{nil, nil, "2", nil},
			outkey:   []string{"", "prefix_", "prefix_b", "prefix_c"},
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

			for i := range tc.outkey {
				// tu.Equals(t, nil, err)
				t.Log(tc.outkey[i])
				t.Log(pt.Fields)
				tu.Equals(t, tc.expected[i], pt.Fields[tc.outkey[i]])
				t.Logf("[%d] PASS", idx)
			}
			ptinput.PutPoint(pt)
		})
	}
}
