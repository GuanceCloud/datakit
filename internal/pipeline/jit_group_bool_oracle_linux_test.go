// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"github.com/GuanceCloud/cliutils/point"
	"testing"
)

func TestJITGroupInNestedOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "match", source: `add_key(returned,group_in(message,["hit"],true,result) == nil)`, input: "hit", key: "result", want: true},
		{name: "miss", source: `add_key(returned,group_in(message,["hit"],true,result) == nil)`, input: "miss", key: "result", want: nil},
		{name: "variable", source: `choices=["hit"]; add_key(returned,group_in(message,choices,"yes",result) == nil)`, input: "hit", key: "result", want: "yes"},
		{name: "short_circuit", source: `r=false && group_in(message,["hit"],true,result) == nil; add_key(returned,r)`, input: "hit", key: "result", want: nil},
		{name: "overwrite", source: `add_key(returned,group_in(message,["hit"],true) == nil)`, input: "hit", key: "message", want: true},
		{name: "target_local", source: `result="other"; add_key(returned,group_in(message,["hit"],true,result) == nil); add_key(local,result)`, input: "hit", key: "local", want: "other"},
	})
}

func TestJITGroupInKeyFormsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "quoted_source", source: `group_in("message",["hit"],true,result)`, input: "hit", key: "result", want: true},
		{name: "source_alias", source: `group_in(_,["hit"],true,result)`, input: "hit", key: "result", want: true},
		{name: "quoted_source_alias", source: `group_in("_",["hit"],true,result)`, input: "hit", key: "result", want: true},
		{name: "target_alias", source: `group_in(message,["hit"],"changed",_)`, input: "hit", key: "message", want: "changed"},
		{name: "quoted_target_alias", source: `group_in(message,["hit"],"changed","_")`, input: "hit", key: "message", want: "changed"},
		{name: "overwrite_alias", source: `group_in(_,["hit"],"changed")`, input: "hit", key: "message", want: "changed"},
		{name: "nested_alias", source: `add_key(returned,group_in("_",["hit"],"changed","_")==nil)`, input: "hit", key: "message", want: "changed"},
		{name: "variable_alias", source: `choices=["hit"]; group_in("_",choices,"changed","_")`, input: "hit", key: "message", want: "changed"},
	})
}

func TestJITGroupInReplacementOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "variable", source: `replacement="yes"; group_in(message,["hit"],replacement,result)`, input: "hit", key: "result", want: nil},
		{name: "side_effect", source: `group_in(message,["hit"],pt_kvs_set("unexpected",1),result)`, input: "hit", key: "unexpected", want: nil},
		{name: "error", source: `group_in(message,["hit"],1/divisor,result)`, input: "hit", key: "result", want: nil},
		{name: "nested", source: `add_key(returned,group_in(message,["hit"],pt_kvs_set("unexpected",1),result)==nil)`, input: "hit", key: "unexpected", want: nil},
		{name: "parenthesized", source: `group_in(message,["hit"],("yes"),result)`, input: "hit", key: "result", want: nil},
		{name: "list", source: `group_in(message,["hit"],[pt_kvs_set("unexpected",1)],result)`, input: "hit", key: "unexpected", want: nil},
		{name: "negative", source: `group_in(message,["hit"],-2,result)`, input: "hit", key: "result", want: int64(-2)},
		{name: "variable_candidates", source: `choices=["hit"]; group_in(message,choices,1/divisor,result)`, input: "hit", key: "result", want: nil},
	})
}

func TestJITGroupInExpressionOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "computed_list", source: `item="hit"; group_in(message,[item],true,result)`, input: "hit", key: "result", want: true},
		{name: "side_effect", source: `group_in(message,[pt_kvs_set("side",1),"hit"],true,result)`, input: "hit", key: "side", want: int64(1)},
		{name: "missing_skips_list", source: `group_in(absent,[pt_kvs_set("unexpected",1)],true,result)`, key: "unexpected", want: nil},
		{name: "list_error", source: `group_in(message,[pt_kvs_set("side",1),1/divisor,pt_kvs_set("unexpected",1)],true,result)`, input: "hit", key: "side", want: int64(1)},
		{name: "nested", source: `item="hit"; add_key(returned,group_in(message,[item],true,result)==nil)`, input: "hit", key: "result", want: true},
		{name: "source_snapshot", source: `group_in(message,[pt_kvs_set("message","changed"),"hit"],true,result)`, input: "hit", key: "result", want: true},
	})
}

func TestJITGroupInIgnoredCandidateOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "call", source: `group_in(message,pt_kvs_set("unexpected",1),true,result)`, input: "hit", key: "unexpected", want: nil},
		{name: "index", source: `choices=[["hit"]]; group_in(message,choices[0],true,result)`, input: "hit", key: "result", want: nil},
		{name: "error", source: `group_in(message,1/divisor,true,result)`, input: "hit", key: "result", want: nil},
		{name: "parenthesized", source: `group_in(message,([pt_kvs_set("unexpected",1),"hit"]),true,result)`, input: "hit", key: "unexpected", want: nil},
		{name: "nil", source: `group_in(message,nil,true,result)`, input: "hit", key: "result", want: nil},
		{name: "nested", source: `add_key(returned,group_in(message,pt_kvs_set("unexpected",1),true,result)==nil)`, input: "hit", key: "unexpected", want: nil},
	})
}

func TestJITGroupInAttributeOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "source", source: `add_key("obj.value","hit"); group_in(obj.value,["hit"],true,result)`, key: "result", want: true},
		{name: "target", source: `group_in(message,["hit"],true,obj.result)`, input: "hit", key: "obj.result", want: true},
		{name: "local_object", source: `obj={"value":"hit"}; group_in(obj.value,["hit"],true,result)`, key: "result", want: nil},
		{name: "nested", source: `add_key("obj.value","hit"); add_key(returned,group_in(obj.value,["hit"],true,obj.result)==nil)`, key: "obj.result", want: true},
		{name: "variable", source: `choices=["hit"]; add_key("obj.value","hit"); group_in(obj.value,choices,true,obj.result)`, key: "obj.result", want: true},
	})
}

func TestJITGroupInNamedOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "candidate", source: `group_in(message,range=["hit"],true,result)`, input: "hit", key: "result", want: nil},
		{name: "replacement", source: `group_in(message,["hit"],new_value=true,result)`, input: "hit", key: "result", want: nil},
		{name: "candidate_effect", source: `group_in(message,range=[pt_kvs_set("unexpected",1),"hit"],true,result)`, input: "hit", key: "unexpected", want: nil},
		{name: "replacement_effect", source: `group_in(message,["hit"],new_value=pt_kvs_set("unexpected",1),result)`, input: "hit", key: "unexpected", want: nil},
		{name: "nested_candidate", source: `add_key(returned,group_in(message,range=[pt_kvs_set("unexpected",1),"hit"],true,result)==nil)`, input: "hit", key: "unexpected", want: nil},
		{name: "nested_replacement", source: `add_key(returned,group_in(message,["hit"],new_value=pt_kvs_set("unexpected",1),result)==nil)`, input: "hit", key: "unexpected", want: nil},
		{name: "both", source: `group_in(message,range=[pt_kvs_set("unexpected",1)],new_value=1/divisor,result)`, input: "hit", key: "unexpected", want: nil},
		{name: "arbitrary_names", source: `group_in(message,arbitrary=["hit"],another=true,result)`, input: "hit", key: "result", want: nil},
	})
}

func TestJITGroupInPointListOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "point_list", source: `group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: nil},
		{name: "local_shadows_point", source: `choices=["miss"]; group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: nil},
	}, func(p *point.Point) { p.AddKVs(point.NewKV("choices", []string{"hit"})) })
}

func TestJITGroupInVariableOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "list", source: `choices=["hit"]; group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: "yes"},
		{name: "changed", source: `choices=["miss"]; choices[0]="hit"; group_in(message,choices,true,result)`, input: "hit", key: "result", want: true},
		{name: "empty", source: `choices=[]; group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: nil},
		{name: "not_list", source: `choices="hit"; group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: nil},
		{name: "missing_choices", source: `group_in(message,choices,"yes",result)`, input: "hit", key: "result", want: nil},
		{name: "missing_source", source: `choices=[nil]; group_in(absent,choices,"yes",result)`, key: "result", want: nil},
	})
}

func TestJITGroupInMatchOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "missing_nil", source: `group_in(absent,[nil],"hit",result)`, key: "result", want: nil},
		{name: "existing_nil", source: `add_key(v,1); cast(v,"string"); group_in(v,[nil],"hit",result)`, key: "result", want: "hit"},
		{name: "int_float", source: `add_key(v,1); group_in(v,[1.0],"hit",result)`, key: "result", want: nil},
		{name: "float_int", source: `add_key(v,1.0); group_in(v,[1],"hit",result)`, key: "result", want: nil},
		{name: "int_int", source: `add_key(v,1); group_in(v,[1],"hit",result)`, key: "result", want: "hit"},
		{name: "float_float", source: `add_key(v,1.0); group_in(v,[1.0],"hit",result)`, key: "result", want: "hit"},
	})
}

// Supplemental cases from fn_group_in.go's accepted BoolLiteral replacement.
func TestJITGroupInBoolOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "true", source: `group_in(message,["hit"],true,result)`, input: "hit", key: "result", want: true},
		{name: "false", source: `group_in(message,["hit"],false,result)`, input: "hit", key: "result", want: false},
		{name: "miss", source: `group_in(message,["hit"],true,result)`, input: "miss", key: "result", want: nil},
		{name: "overwrite", source: `group_in(message,["hit"],true)`, input: "hit", key: "message", want: true},
		{name: "tag", source: `set_tag(result,"old"); group_in(message,["hit"],true,result)`, input: "hit", key: "result", want: "true"},
		{name: "missing", source: `group_in(absent,["hit"],true,result)`, key: "result", want: nil},
	})
}
