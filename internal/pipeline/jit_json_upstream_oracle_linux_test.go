// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// TestJSON fixtures copied from pipeline-go v1.4.3 ptinput/funcs/fn_json_test.go.
// MIT License; Copyright 2021-present Guance, Inc.
// Original inputs, scripts, flags and expectations are retained verbatim.
// The adapter fixes the Point clock and adds complete Go/native Point comparison.
import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITJSONUpstreamOracle(t *testing.T) {
	testCase := []struct {
		name, in, script, key string
		expected              any
		fail                  bool
	}{
		{
			in: `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script: `json(_, name)
			json(name, first)`,
			expected: "Tom",
			key:      "first",
		},
		{
			in: `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script: `json(_, friends)
			json(friends, .[1].first, f_first)`,
			expected: "Roger",
			key:      "f_first",
		},
		{
			in: `[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, .[0].nets[-1])`,
			expected: "tw",
			key:      "[0].nets[-1]",
		},
		{
			in: `[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, .[1].age)`,
			expected: float64(68),
			key:      "[1].age",
		},
		{
			name:     "trim_space auto",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item)`,
			key:      "item",
			expected: "not_space",
		},
		{
			name:     "trim_space disable",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item, false)`,
			key:      "item",
			expected: " not_space ",
		},
		{
			name:     "trim_space enable",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item, true)`,
			key:      "item",
			expected: "not_space",
		},
		{
			name:     "path_with_dot_in_key",
			in:       `{"a.b": 123}`,
			script:   "json(_, `a.b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_numeric_key",
			in:       `{"0": 123}`,
			script:   "json(_, `0`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "duplicate_key_uses_last_value",
			in:       `{"a": 1, "a": 2}`,
			script:   "json(_, a, out)",
			key:      "out",
			expected: float64(2),
		},
		{
			name:     "nested_duplicate_key_uses_last_value",
			in:       `{"root": {"a": 1, "a": 2}}`,
			script:   "json(_, root.a, out)",
			key:      "out",
			expected: float64(2),
		},
		{
			name:     "path_with_wildcard_in_key",
			in:       `{"a*b": 123}`,
			script:   "json(_, `a*b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_question_in_key",
			in:       `{"a?b": 123}`,
			script:   "json(_, `a?b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_array_count_in_key",
			in:       `{"a#b": 123}`,
			script:   "json(_, `a#b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_pipe_in_key",
			in:       `{"a|b": 123}`,
			script:   "json(_, `a|b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_backslash_in_key",
			in:       `{"a\\b": 123}`,
			script:   "json(_, `a\\b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_modifier_name_in_key",
			in:       `{"@this": 123}`,
			script:   "json(_, `@this`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_static_prefix_in_key",
			in:       `{"!foo": 123}`,
			script:   "json(_, `!foo`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_multipath_prefix_in_key",
			in:       `{"[key": 123}`,
			script:   "json(_, `[key`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "path_with_object_multipath_prefix_in_key",
			in:       `{"{key": 123}`,
			script:   "json(_, `{key`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_multipath_prefix_in_key",
			in:       `{"root": {"[key": 123}}`,
			script:   "json(_, root.`[key`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_object_multipath_prefix_in_key",
			in:       `{"root": {"{key": 123}}`,
			script:   "json(_, root.`{key`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_dot_in_key",
			in:       `{"root": {"a.b": 123}}`,
			script:   "json(_, root.`a.b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_wildcard_in_key",
			in:       `{"root": {"a*b": 123}}`,
			script:   "json(_, root.`a*b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_question_in_key",
			in:       `{"root": {"a?b": 123}}`,
			script:   "json(_, root.`a?b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_array_count_in_key",
			in:       `{"root": {"a#b": 123}}`,
			script:   "json(_, root.`a#b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_pipe_in_key",
			in:       `{"root": {"a|b": 123}}`,
			script:   "json(_, root.`a|b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_backslash_in_key",
			in:       `{"root": {"a\\b": 123}}`,
			script:   "json(_, root.`a\\b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_modifier_name_in_key",
			in:       `{"root": {"@this": 123}}`,
			script:   "json(_, root.`@this`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "nested_path_with_static_prefix_in_key",
			in:       `{"root": {"!foo": 123}}`,
			script:   "json(_, root.`!foo`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "special_key_map_value",
			in:       `{"a|b": {"inner": 123}}`,
			script:   "json(_, `a|b`, out)",
			key:      "out",
			expected: `{"inner":123}`,
		},
		{
			name:     "special_key_list_value",
			in:       `{"a|b": [1, 2, 3]}`,
			script:   "json(_, `a|b`, out)",
			key:      "out",
			expected: `[1,2,3]`,
		},
		{
			name:     "special_key_then_index",
			in:       `{"a|b": [1, 2, 3]}`,
			script:   "json(_, `a|b`[1], out)",
			key:      "out",
			expected: float64(2),
		},
		{
			name:     "index_then_special_key",
			in:       `[{"a|b": 123}]`,
			script:   "json(_, .[0].`a|b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name:     "negative_index_then_special_key",
			in:       `[{"a|b": 123}]`,
			script:   "json(_, .[-1].`a|b`, out)",
			key:      "out",
			expected: float64(123),
		},
		{
			name: "source_replaced_between_json_calls",
			in:   `{"a":{"first":1}}`,
			script: `json(_, a.first, first)
			add_key("message", "{\"a\":{\"second\":2}}")
			json(_, a.second, second)`,
			key:      "second",
			expected: float64(2),
		},
		{
			name:     "map_delete_after",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item, true, true)`,
			key:      "message",
			expected: "{}",
			fail:     false,
		},
		{
			name:     "map_delete_after1",
			in:       `{"item": " not_space ", "item2":{"item3": [123]}}`,
			script:   `json(_, item2.item3, item, delete_after_extract = true)`,
			key:      "message",
			expected: `{"item":" not_space ","item2":{}}`,
		},
		{
			name:     "list_delete_after1",
			in:       `{"item": " not_space ", "item2": [[1,2,3,4,5],[6]]}`,
			script:   `json(_, .[0].item2[0][2].a[0], item, true, true)`,
			key:      "item",
			expected: "1",
			fail:     true,
		},
		{
			name:     "list_delete_after2",
			in:       `{"item": " not_space ", "item2": [[1,2,3,4,5],[6]]}`,
			script:   `json(_, .[0], item, true, true)`,
			key:      "item",
			expected: "1",
			fail:     true,
		},
		{
			name:     "list_delete_after3",
			in:       `{"item": " not_space ", "item2": [[1,2,3,4,5],[6]]}`,
			script:   `json(_, a[0][1], item, true, true)`,
			key:      "item",
			expected: "1",
			fail:     true,
		},
	}

	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	var cases []jsonOracleCase
	for idx, tc := range testCase {
		name := fmt.Sprintf("upstream/%02d/%s", idx, tc.name)
		if tc.fail {
			t.Run(name, func(t *testing.T) {
				if _, err := NewPlScriptSimple(point.Logging, "json-upstream.p", tc.script); err == nil {
					t.Fatal("Go must reject upstream compile-failure case")
				}
				if check := runner.Check(tc.script); check.Route == pljit.RouteJITNative {
					t.Fatalf("JIT accepted Go-invalid source: %+v", check)
				}
			})
			continue
		}
		t.Run(name+"/upstream_assertion", func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, "json-upstream.p", tc.script)
			if err != nil {
				t.Fatal(err)
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": tc.in}, time.Unix(1700000000, 123))
			if err := script.Run(pt, nil, nil); err != nil {
				t.Fatal(err)
			}
			got, _, err := pt.Get(tc.key)
			if err != nil || !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("original upstream assertion: got=%#v error=%v want=%#v", got, err, tc.expected)
			}
		})
		// Upstream asserts PlInputPt.Get, which stringifies composite values.
		// The production oracle compares the actual typed Point separately;
		// do not compare an upstream string to Point.Get's map/list value.
		cases = append(cases, jsonOracleCase{name, tc.script, tc.in, "", nil, false})
	}
	runJSONOracle(t, cases)
}

// The two additional non-benchmark tests in the same upstream source assert
// missing output, not an execution error, for object-key/index type mismatch.
func TestJITJSONUpstreamPathTypeOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"TestJSONIndexDoesNotMatchObjectNumericKey", "json(_, .[0], out)", "{\"0\":123}", "out", nil, false},
		{"TestJSONNumericKeyDoesNotMatchArrayIndex", "json(_, `0`, out)", "[123]", "out", nil, false},
	})
}

func TestJITJSONLiteralPathBoundaryOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct{ name, input, path string }{
		{"dot", `{"a.b":123}`, "`a.b`"},
		{"bracket", `{"[key":123}`, "`[key`"},
		{"root_marker", `{"$":123}`, "`$`"},
		{"space", `{" a ":123}`, "` a `"},
		{"nested_dot", `{"root":{"a.b":123}}`, "root.`a.b`"},
		{"nested_bracket", `{"root":{"[key":123}}`, "root.`[key`"},
	} {
		cases = append(cases,
			jsonOracleCase{tc.name + "/explicit", "json(_, " + tc.path + ", out)", tc.input, "out", float64(123), false},
			jsonOracleCase{tc.name + "/default", "json(_, " + tc.path + ")", tc.input, "", nil, false},
			jsonOracleCase{tc.name + "/delete", "json(_, " + tc.path + ", out, delete_after_extract=true)", tc.input, "out", float64(123), false},
		)
	}
	runJSONOracle(t, cases)
}

func TestJITJSONCompositeArrayOutputOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"objects", `json(_, items, out)`, `{"items":[{"a":1},{"b":2}]}`, "out", `[{"a":1},{"b":2}]`, false},
		{"nested_arrays", `json(_, items, out)`, `{"items":[[1,2],[3]]}`, "out", `[[1,2],[3]]`, false},
		{"mixed_containers", `json(_, items, out)`, `{"items":[1,{"b":2},[3]]}`, "out", `[1,{"b":2},[3]]`, false},
		{"raw_nested_string", `json(_, items, out)`, "{\"items\":[{\"a\":\"\xff\"}]}", "out", `[{"a":"\ufffd"}]`, false},
		{"object_array_read", `json(_, items, out); add_key(kind, value_type(out))`, `{"items":[{"a":1}]}`, "kind", "str", false},
		{"nested_map", `json(_, object, out)`, `{"object":{"b":2,"a":[1,true,null]}}`, "out", `{"a":[1,true,null],"b":2}`, false},
		{"flat_map_read", `json(_, object, out); add_key(kind, value_type(out))`, `{"object":{"a":1}}`, "kind", "str", false},
		{"flat_list_read", `json(_, items, out); add_key(kind, value_type(out))`, `{"items":[1,2]}`, "kind", "str", false},
	})
}

func TestJITJSONContainerConsumerOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, input := range []struct{ name, value string }{
		{"map", `{"a":1}`}, {"list", `[1,2]`}, {"nested", `[{"a":1}]`},
	} {
		for _, consumer := range []struct{ name, script string }{
			{"copy", `add_key(copy, out)`},
			{"len", `add_key(size, len(out))`},
			{"decode", `decoded=load_json(out); add_key(kind, value_type(decoded))`},
			{"raw_kind", `raw=pt_kvs_get("out",true); add_key(kind,value_type(raw))`},
			{"raw_copy", `raw=pt_kvs_get("out",true); pt_kvs_set("copy",raw,false,true)`},
			{"normal_get", `add_key(copy,pt_kvs_get("out",false))`},
			{"overwrite_snapshot", `saved=out; add_key(out,"replaced"); add_key(copy,saved)`},
			{"cache", `cache_set("json-value",out); add_key(copy,cache_get("json-value"))`},
		} {
			cases = append(cases, jsonOracleCase{
				name:   input.name + "/" + consumer.name,
				source: `json(_, value, out); ` + consumer.script,
				input:  `{"value":` + input.value + `}`,
			})
		}
	}
	runJSONOracle(t, cases)
}

func TestJITJSONRawReadSnapshotOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"map", `json(_,value,out); raw=pt_kvs_get("out",true); raw["a"]=9; json(out,a,result)`, `{"value":{"a":1}}`, "result", float64(1), false},
		{"list", `json(_,value,out); raw=pt_kvs_get("out",true); raw[0]=9; json(out,.[0],result)`, `{"value":[1,2]}`, "result", float64(1), false},
	})
}
