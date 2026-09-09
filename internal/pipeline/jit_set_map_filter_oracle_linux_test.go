// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

// Supplementary combinations of the upstream pt_kvs_del implementation; these
// are not counted as migrated original upstream table cases.
func TestJITNestedPointDeleteOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "return_nil", source: `pt_kvs_set("victim",7); add_key(result,pt_kvs_del("victim") == nil)`, input: "test", key: "result", want: true},
		{name: "missing", source: `add_key(result,pt_kvs_del("absent") == nil)`, input: "test", key: "result", want: true},
		{name: "tag", source: `pt_kvs_set("victim",7,as_tag=true); add_key(result,pt_kvs_del("victim") == nil)`, input: "test", key: "victim", want: nil},
		{name: "short_circuit", source: `pt_kvs_set("victim",7); result=false && pt_kvs_del("victim") == nil; add_key(result,result)`, input: "test", key: "victim", want: int64(7)},
		{name: "later_error", source: `pt_kvs_set("victim",7); result=(pt_kvs_del("victim") == nil) && 1/divisor == 0`, input: "test", key: "victim", want: nil, fails: true},
		{name: "bad_name", source: `pt_kvs_set("victim",7); name=1; add_key(result,pt_kvs_del(name) == nil)`, input: "test", key: "victim", want: int64(7), fails: true},
		{name: "raw_name", source: `pt_kvs_set(message,7); add_key(result,pt_kvs_del(message) == nil)`, input: "\xff", key: "\xff", want: nil},
	})
}

func TestJITNestedSetMapOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "return_value", source: `add_key(count,pt_kvs_set_map({"a":7,"b":8},include_keys=["a"]))`, input: "test", key: "count", want: int64(1)},
		{name: "nested_write", source: `pt_kvs_set("count",pt_kvs_set_map({"a":7},include_keys=["a"]))`, input: "test", key: "count", want: int64(1)},
		{name: "condition", source: `if pt_kvs_set_map({"a":7},include_keys=["a"]) == 1 { add_key(selected,true) }`, input: "test", key: "selected", want: true},
		{name: "short_circuit", source: `result=false && pt_kvs_set_map({"a":7},include_keys=["a"]) == 1; add_key(result,result)`, input: "test", key: "a", want: nil},
		{name: "no_filter", source: `add_key(count,pt_kvs_set_map({"a":7}))`, input: "test", key: "count", want: int64(0)},
		{name: "dynamic_error", source: `f=[1]; add_key(count,pt_kvs_set_map({"a":7},include_keys=f))`, input: "test", key: "a", want: nil, fails: true},
		{name: "later_expression_error", source: `result=pt_kvs_set_map({"a":7},include_keys=["a"]) + 1/divisor`, input: "test", key: "a", want: int64(7), fails: true},
		{name: "raw_payload", source: `add_key(count,pt_kvs_set_map({"raw":message},include_keys=["raw"]))`, input: "\xff", key: "raw", want: "\xff"},
		{name: "named_order", source: `add_key(count,pt_kvs_set_map({},raw=pt_kvs_set("order","raw"),as_tag=pt_kvs_set("order","tag")))`, input: "test", key: "order", want: "raw"},
	})
}

func TestJITPointReadArgumentEffectsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "get_bad_name", source: `name=1; v=pt_kvs_get(name,raw=pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "get_bad_raw", source: `raw=1; v=pt_kvs_get("message",raw=raw); add_key(value,v)`, input: "test", fails: true},
		{name: "keys_bad_tags", source: `flag=1; v=pt_kvs_keys(tags=flag,fields=pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "keys_named_order", source: `v=pt_kvs_keys(fields=pt_kvs_set("order","fields"),tags=pt_kvs_set("order","tags"))`, input: "test", key: "order", want: "fields"},
		{name: "keys_defaults", source: `v=pt_kvs_keys(); found=false; for key in v { if key == "message" { found=true } }; add_key(found,found)`, input: "test", key: "found", want: true},
		{name: "delete_bad_name", source: `name=1; pt_kvs_del(name)`, input: "test", fails: true},
	})
}

func TestJITSetMapArgumentEffectsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "bad_values", source: `v=1; pt_kvs_set_map(v,as_tag=pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "bad_filter_type", source: `f=1; pt_kvs_set_map({},include_keys=f,as_tag=pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "bad_filter_element", source: `f=[1]; pt_kvs_set_map({},include_keys=f,as_tag=pt_kvs_set("later",true))`, input: "test", key: "later", want: true, fails: true},
		{name: "named_order", source: `f=1; pt_kvs_set_map({},as_tag=pt_kvs_set("later",true),include_keys=f)`, input: "test", key: "later", want: nil, fails: true},
		{name: "valid_order", source: `pt_kvs_set_map({},raw=pt_kvs_set("order","raw"),as_tag=pt_kvs_set("order","tag"))`, input: "test", key: "order", want: "raw"},
		{name: "bad_flag", source: `flag=1; pt_kvs_set_map({},as_tag=flag,raw=pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "set_named_order", source: `pt_kvs_set("target",7,raw=pt_kvs_set("order","raw"),as_tag=pt_kvs_set("order","tag"))`, input: "test", key: "order", want: "raw"},
		{name: "set_bad_name", source: `name=1; pt_kvs_set(name,pt_kvs_set("later",true))`, input: "test", key: "later", want: nil, fails: true},
		{name: "nested_write_value", source: `pt_kvs_set("outer",pt_kvs_set("inner",7))`, input: "test", key: "outer", want: true},
	})
}

func TestJITSetMapDynamicFilterOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, parameter := range []string{"include_keys", "key_patterns"} {
		for _, filter := range []struct {
			name, expression string
			fails            bool
		}{
			{"nil", "nil", false}, {"empty", "[]", false}, {"valid", `["a"]`, false},
			{"string", `"a"`, true}, {"integer", "1", true}, {"map", `{"a":1}`, true},
			{"bad_element", `["a",1]`, true}, {"nil_element", `[nil]`, true},
		} {
			for _, values := range []struct{ name, expression string }{{"empty", "{}"}, {"full", `{"a":7}`}} {
				cases = append(cases, jsonOracleCase{name: parameter + "/" + filter.name + "/" + values.name,
					source: "filters=" + filter.expression + "; n=pt_kvs_set_map(" + values.expression + ", " + parameter + "=filters); add_key(count,n)", input: "input", fails: filter.fails})
			}
		}
	}
	cases = append(cases,
		jsonOracleCase{name: "raw_include", source: `filters=[message,"a"]; n=pt_kvs_set_map({"a":7,"�":8},include_keys=filters); add_key(count,n)`, input: "\xff", key: "count", want: int64(1)},
		jsonOracleCase{name: "raw_pattern", source: `filters=[message]; n=pt_kvs_set_map({"a":7,"�":8},key_patterns=filters); add_key(count,n)`, input: "\xff", key: "�", want: int64(8)},
		jsonOracleCase{name: "truncated_pattern", source: `filters=[message]; n=pt_kvs_set_map({"�":7,"��":8},key_patterns=filters); add_key(count,n)`, input: "\xe2\x82", key: "��", want: int64(8)},
		jsonOracleCase{name: "failure_after_write", source: `filters=["a"]; pt_kvs_set_map({"a":7},include_keys=filters); filters[0]=1; pt_kvs_set_map({"b":8},key_patterns=filters)`, input: "input", key: "a", want: int64(7), fails: true},
	)
	runJSONOracle(t, cases)
}
