// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

func TestJITGroupBetweenAssignmentOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "match", source: `add_key(v,2); group_between(v,[1,3],new_value=true,result)`, key: "result", want: nil, fails: true},
		{name: "miss", source: `add_key(v,4); group_between(v,[1,3],new_value=true,result)`, key: "result", want: nil},
		{name: "effect", source: `add_key(v,2); group_between(v,[1,3],new_value=pt_kvs_set("unexpected",1),result)`, key: "unexpected", want: nil, fails: true},
		{name: "nested", source: `add_key(v,2); add_key(returned,group_between(v,[1,3],new_value=pt_kvs_set("unexpected",1),result)==nil)`, key: "unexpected", want: nil, fails: true},
		{name: "missing", source: `group_between(absent,[1,3],anything=pt_kvs_set("unexpected",1),result)`, key: "unexpected", want: nil},
	})
}

func TestJITGroupBetweenSignedRangeOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "negative", source: `group_between(message,[-3,-1],true,result)`, input: "-2", key: "result", want: true},
		{name: "cross_zero", source: `group_between(message,[-1,1],true,result)`, input: "0", key: "result", want: true},
		{name: "fraction", source: `group_between(message,[-0.5,0.5],true,result)`, input: "0.25", key: "result", want: true},
		{name: "zero_width", source: `group_between(message,[-2,-2],true,result)`, input: "-2", key: "result", want: true},
		{name: "plus", source: `group_between(message,[+1,+3],true,result)`, input: "2", key: "result", want: true},
		{name: "nested", source: `add_key(returned,group_between(message,[-3,-1],true,result)==nil)`, input: "-2", key: "result", want: true},
	})
}

func TestJITGroupBetweenReplacementOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "variable_match", source: `add_key(v,2); replacement="yes"; group_between(v,[1,3],replacement,result)`, key: "result", want: nil, fails: true},
		{name: "variable_miss", source: `add_key(v,4); replacement="yes"; group_between(v,[1,3],replacement,result)`, key: "result", want: nil},
		{name: "effect_match", source: `add_key(v,2); group_between(v,[1,3],pt_kvs_set("unexpected",1),result)`, key: "unexpected", want: nil, fails: true},
		{name: "missing", source: `group_between(absent,[1,3],pt_kvs_set("unexpected",1),result)`, key: "unexpected", want: nil},
		{name: "nested", source: `add_key(v,2); add_key(returned,group_between(v,[1,3],pt_kvs_set("unexpected",1),result)==nil)`, key: "unexpected", want: nil, fails: true},
		{name: "parenthesized", source: `add_key(v,2); group_between(v,[1,3],("yes"),result)`, key: "result", want: nil, fails: true},
		{name: "negative", source: `add_key(v,2); group_between(v,[1,3],-2,result)`, key: "result", want: int64(-2)},
	})
}

func TestJITGroupBetweenKeyFormsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "quoted_source", source: `group_between("message",[1,3],true,result)`, input: "2", key: "result", want: true},
		{name: "quoted_target", source: `group_between(message,[1,3],true,"result")`, input: "2", key: "result", want: true},
		{name: "alias", source: `group_between(_,[1,3],"changed",_)`, input: "2", key: "message", want: "changed"},
		{name: "quoted_alias", source: `group_between("_",[1,3],"changed","_")`, input: "2", key: "message", want: "changed"},
		{name: "attribute", source: `add_key("obj.value",2); group_between(obj.value,[1,3],true,obj.result)`, key: "obj.result", want: true},
		{name: "nested", source: `add_key(returned,group_between("_",[1,3],"changed","_")==nil)`, input: "2", key: "message", want: "changed"},
	})
}

func TestJITGroupBetweenNestedOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "match", source: `add_key(v,2); add_key(returned,group_between(v,[1,3],true,result)==nil)`, key: "result", want: true},
		{name: "miss", source: `add_key(v,4); add_key(returned,group_between(v,[1,3],true,result)==nil)`, key: "result", want: nil},
		{name: "short_circuit", source: `add_key(v,2); r=false && group_between(v,[1,3],true,result)==nil; add_key(returned,r)`, key: "result", want: nil},
		{name: "overwrite", source: `add_key(v,2); add_key(returned,group_between(v,[1,3],true)==nil)`, key: "v", want: true},
		{name: "target_local", source: `result="other"; add_key(v,2); add_key(returned,group_between(v,[1,3],true,result)==nil); add_key(local,result)`, key: "local", want: "other"},
		{name: "missing", source: `add_key(returned,group_between(absent,[0,3],true,result)==nil)`, key: "result", want: nil},
	})
}

func TestJITGroupBetweenNumericOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "decimal", source: `group_between(message,[1,3],true,result)`, input: "2", key: "result", want: true},
		{name: "hex", source: `group_between(message,[1,3],true,result)`, input: "0x1p1", key: "result", want: true},
		{name: "underscore", source: `group_between(message,[10,13],true,result)`, input: "1_2", key: "result", want: true},
		{name: "invalid_zero", source: `group_between(message,[0,0],true,result)`, input: "invalid", key: "result", want: true},
		{name: "space_zero", source: `group_between(message,[0,0],true,result)`, input: " 2 ", key: "result", want: true},
		{name: "nan", source: `group_between(message,[0,3],true,result)`, input: "NaN", key: "result", want: nil},
		{name: "bool", source: `add_key(v,true); group_between(v,[1,1],true,result)`, key: "result", want: true},
		{name: "nil", source: `add_key(v,1); cast(v,"string"); group_between(v,[0,0],true,result)`, key: "result", want: true},
	})
}

// Supplemental Group/GroupHandle behavior from pipeline-go v1.4.3 fn_group.go.
func TestJITGroupBetweenBoolOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "true", source: `add_key(v,2); group_between(v,[1,3],true,result)`, key: "result", want: true},
		{name: "false", source: `add_key(v,2); group_between(v,[1,3],false,result)`, key: "result", want: false},
		{name: "lower", source: `add_key(v,1); group_between(v,[1,3],true,result)`, key: "result", want: true},
		{name: "upper", source: `add_key(v,3); group_between(v,[1,3],true,result)`, key: "result", want: true},
		{name: "miss", source: `add_key(v,4); group_between(v,[1,3],true,result)`, key: "result", want: nil},
		{name: "tag", source: `set_tag(result,"old"); add_key(v,2); group_between(v,[1,3],true,result)`, key: "result", want: "true"},
		{name: "missing", source: `group_between(absent,[0,3],true,result)`, key: "result", want: nil},
	})
}
