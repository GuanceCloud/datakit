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

func TestNullIf(t *testing.T) {
	cases := []struct {
		name, pl, in string
		expected     interface{}
		fail         bool
		outkey       string
	}{
		{
			name: "normal",
			pl: `json(_, a.first, a_first)
			nullif(a_first, "1")`,
			in:       `{"a":{"first": 1,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: float64(1),
			outkey:   "a_first",
		},
		{
			name: "normal",
			pl: `json(_, a.first, a_first)
			nullif(a_first, 1)`,
			in:       `{"a":{"first": "1","second":2,"third":"aBC","forth":true},"age":47}`,
			expected: "1",
			outkey:   "a_first",
		},
		{
			name: "normal",
			pl: `json(_, a.first, a_first)
			nullif(a_first, "")`,
			in:       `{"a":{"first": "","second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a_first",
		},
		{
			name: "normal",
			pl: `json(_, a.first, a_first) 
			nullif(a_first, nil)`,
			in:       `{"a":{"first": null,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a_first",
		},
		{
			name: "normal",
			pl: `json(_, a.first, a_first) 
			nullif(a_first, true)`,
			in:       `{"a":{"first": true,"second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a_first",
		},
		{
			name: "normal",
			pl: `json(_, a.first, a_first) 
			nullif(a_first, 2.3)`,
			in:       `{"a":{"first": 2.3, "second":2,"third":"aBC","forth":true},"age":47}`,
			expected: nil,
			outkey:   "a_first",
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
			v := pt.Fields[tc.outkey]
			// tu.Equals(t, nil, err)
			tu.Equals(t, tc.expected, v)

			t.Logf("[%d] PASS", idx)
			ptinput.PutPoint(pt)
		})
	}
}
