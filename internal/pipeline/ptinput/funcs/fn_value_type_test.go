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

func TestVauleType(t *testing.T) {
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
			d = load_json(_) 
			add_key("val_type", value_type(d))
	`,
			outkey: "val_type",
			expect: "map",
			fail:   false,
		},
		{
			name: "list",
			in:   `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: `
			d = load_json(_)

			if value_type(d) == "map" && "a" in d && 
				value_type(d["a"]) == "map" && "first" in d["a"] {
				add_key("val_type", value_type(d["a"]["first"]))
			}
	`,
			outkey: "val_type",
			expect: "list",
			fail:   false,
		},
		{
			name: "map_2",
			in:   `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: `
			d = load_json(_)

			if value_type(d) == "map" && "a" in d  {
				add_key("val_type", value_type(d["a"]))
			}
	`,
			outkey: "val_type",
			expect: "map",
			fail:   false,
		},
		{
			name: "list-not-in",
			in:   `{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: `
			d = load_json(_)

			if "a" in d && "first" in d["a"] {
				add_key("val_type", value_type(d["a"]["fist"])) # "fist" not in d["a"]
			}
	`,
			outkey: "val_type",
			expect: "",
			fail:   false,
		},
		{
			name: "int->float",
			in:   `{"a":{"first": [2.2, 1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			d = load_json(_) 
			add_key("val_type", value_type(d["a"]["first"][1]))
	`,
			outkey: "val_type",
			expect: "float", // not int
			fail:   false,
		},
		{
			name: "float",
			in:   `{"a":{"first": [2.2, 1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			d = load_json(_) 
			add_key("val_type", value_type(d["a"]["first"][0]))
	`,
			outkey: "val_type",
			expect: "float",
			fail:   false,
		},
		{
			name: "int",
			in:   ``,
			pl: ` 
			d = {"a" : 1}
			add_key("val_type", value_type(d["a"]))
	`,
			outkey: "val_type",
			expect: "int",
			fail:   false,
		},
		{
			name: "bool",
			in:   `{"a":{"first": [true, 1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			d = load_json(_) 
			add_key("val_type", value_type(d["a"]["first"][0]))
	`,
			outkey: "val_type",
			expect: "bool",
			fail:   false,
		},
		{
			name: "str",
			in:   `{"a":{"first": ["true", 1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}`,
			pl: ` 
			d = load_json(_) 
			add_key("val_type", value_type(d["a"]["first"][0]))
	`,
			outkey: "val_type",
			expect: "str",
			fail:   false,
		},
		{
			name: "empty_nil",
			in:   ``,
			pl: `  
			add_key("val_type", value_type(nil))
	`,
			outkey: "val_type",
			expect: "",
			fail:   false,
		},
		{
			name: "empty_nil",
			in:   ``,
			pl: `  
			add_key("val_type", value_type(x))
	`,
			outkey: "val_type",
			expect: "",
			fail:   false,
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
				t.Fatal(errR)
			} else {
				v, _, _ := pt.Get(tc.outkey)
				tu.Equals(t, tc.expect, v)
				t.Logf("[%d] PASS", idx)
			}
		})
	}
}
