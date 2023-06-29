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

func TestVaildJson(t *testing.T) {
	cases := []struct {
		name, pl, in string
		outkey       string
		expect       interface{}
		fail         bool
	}{
		{
			name: "map",
			in:   `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			if vaild_json(_) {
				d = load_json(_)
				add_key("abc", d["a"]["first"][0])
			}
			`,
			outkey: "abc",
			expect: 2.2,
		},
		{
			name: "map",
			in:   `{"a"??:{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			if vaild_json(_) {
			} else {
				d = load_json(_)
				add_key("abc", d["a"]["first"][0])
			}
			`,
			outkey: "abc",
			expect: 2.2,
			fail:   true,
		},
		{
			name:   "map",
			in:     ``,
			pl:     "add_key(`in`, vaild_json(_))",
			outkey: "in",
			expect: false,
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

			t.Log(pt.Fields())
			if errR != nil {
				if tc.fail {
					return
				}
				t.Fatal(errR)
			} else {
				v, _, _ := pt.Get(tc.outkey)
				tu.Equals(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
