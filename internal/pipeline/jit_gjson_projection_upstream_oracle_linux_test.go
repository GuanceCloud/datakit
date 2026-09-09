// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import "testing"

func TestJITGJSONProjectionUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "object_names", source: `gjson(message,"items.#.name","result")`,
			input: `{"items":[{"name":"a"},{"name":"b"}]}`, key: "result", want: `["a","b"]`,
		},
		{
			name: "missing_members", source: `gjson(message,"items.#.name","result")`,
			input: `{"items":[{"name":"a"},{"other":1},{"name":null}]}`, key: "result", want: `["a",null]`,
		},
		{
			name: "nested_members", source: `gjson(message,"items.#.meta.id","result")`,
			input: `{"items":[{"meta":{"id":1}},{"meta":{"id":2}}]}`, key: "result", want: `[1,2]`,
		},
		{
			name: "scalar_members", source: `gjson(message,"items.#.name","result")`,
			input: `{"items":[1,2]}`, key: "result", want: `[]`,
		},
	})
}

func TestJITGJSONMalformedPathUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "static_double_dot", source: `gjson(message,"a..b","result"); add_key(after,true)`,
			input: `{"a":{"b":1}}`, key: "after", want: true,
		},
		{
			name: "dynamic_unclosed_group", source: `path="a("; gjson(message,path,"result"); add_key(after,true)`,
			input: `{"a":{"b":1}}`, key: "after", want: true,
		},
	})
}

func TestJITGJSONWildcardPathUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "star_key", source: `gjson(message,"child*.2","result")`,
			input: `{"children":["Sara","Alex","Jack"]}`, key: "result", want: "Jack",
		},
		{
			name: "question_key", source: `gjson(message,"c?ildren.0","result")`,
			input: `{"children":["Sara","Alex"]}`, key: "result", want: "Sara",
		},
		{
			name: "unicode_key", source: `gjson(message,"的?况.value","result")`,
			input: `{"的情况":{"value":"matched"}}`, key: "result", want: "matched",
		},
		{
			name: "escaped_wildcard_key", source: `gjson(message,"a\\\\*b","result")`,
			input: `{"axb":"wrong","a*b":"literal"}`, key: "result", want: "literal",
		},
		{
			name: "unknown_modifier_is_key", source: `gjson(message,"@key","result")`,
			input: `{"@key":"value"}`, key: "result", want: "value",
		},
		{
			name: "escaped_at_key", source: `gjson(message,"\\@key","result")`,
			input: `{"@key":"value"}`, key: "result", want: "value",
		},
		{
			name: "wildcard_query_prefix", source: `gjson(message,"i*.f*.#(age>=47).name","result")`,
			input: `{"info":{"friends":[{"name":"Dale","age":44},{"name":"Jane","age":47}]}}`, key: "result", want: "Jane",
		},
	})
}

func TestJITGJSONStaticUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "static_string", source: `gjson(message,"!\"bar\"","result")`,
			input: `{"ignored":true}`, key: "result", want: "bar",
		},
		{
			name: "static_object_chain", source: `gjson(message,"!{\"name\":{\"first\":\"Tom\"}}.name.first","result")`,
			input: `{"ignored":true}`, key: "result", want: "Tom",
		},
		{
			name: "static_object_multipath", source: `gjson(message,"{name.last,\"foo\":!\"bar\"}","result")`,
			input: `{"name":{"last":"Anderson"}}`, key: "result", want: `{"last":"Anderson","foo":"bar"}`,
		},
		{
			name: "static_array_matrix", source: `gjson(message,"[!true,!false,!null,!inf,!nan,!hello,{\"name\":!\"andy\",name.last},+inf,![\"any\",\"thing\"]]","result")`,
			input: `{"name":{"last":"Anderson"}}`, key: "result", want: `[true,false,null,inf,nan,{"name":"andy","last":"Anderson"},["any","thing"]]`,
		},
	})
}

func TestJITGJSONQueryUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "first_string_match", source: `gjson(message,"items.#(status==\"ok\").name","result")`,
			input: `{"items":[{"name":"a","status":"bad"},{"name":"b","status":"ok"}]}`, key: "result", want: "b",
		},
		{
			name: "all_string_matches", source: `gjson(message,"items.#(status==\"ok\")#.name","result")`,
			input: `{"items":[{"name":"a","status":"ok"},{"name":"b","status":"bad"},{"name":"c","status":"ok"}]}`, key: "result", want: `["a","c"]`,
		},
		{
			name: "numeric_comparison", source: `gjson(message,"items.#(latency>=20).name","result")`,
			input: `{"items":[{"name":"a","latency":10},{"name":"b","latency":20}]}`, key: "result", want: "b",
		},
		{
			name: "member_exists", source: `gjson(message,"items.#(meta).name","result")`,
			input: `{"items":[{"name":"a"},{"name":"b","meta":null}]}`, key: "result", want: "b",
		},
		{
			name: "scalar_array_query", source: `gjson(message,"items.#(==2)","result")`,
			input: `{"items":[1,2,3]}`, key: "result", want: float64(2),
		},
		{
			name: "empty_query_operand", source: `gjson(message,"#(!=)#","result")`,
			input: `["ig","","tw","fb","tw","ig","tw"]`, key: "result", want: `["ig","tw","fb","tw","ig","tw"]`,
		},
		{
			name: "wildcard_query", source: `gjson(message,"items.#(name%\"*phy\").name","result")`,
			input: `{"items":[{"name":"Craig"},{"name":"Murphy"}]}`, key: "result", want: "Murphy",
		},
		{
			name: "negative_wildcard_query", source: `gjson(message,"items.#(name!%\"*phy\").name","result")`,
			input: `{"items":[{"name":"Murphy"},{"name":"Craig"}]}`, key: "result", want: "Craig",
		},
		{
			name: "unicode_wildcard_rune", source: `gjson(message,"items.#(name%\"中?\").name","result")`,
			input: `{"items":[{"name":"中"},{"name":"中文"}]}`, key: "result", want: "中文",
		},
		{
			name: "tilde_true", source: `gjson(message,"vals.#(b==~true)#.a","result")`,
			input: `{"vals":[{"a":1,"b":"data"},{"a":2,"b":true},{"a":3,"b":false},{"a":4,"b":"0"},{"a":5,"b":0},{"a":6,"b":"1"},{"a":7,"b":1},{"a":8,"b":"true"},{"a":9,"b":false},{"a":10,"b":null},{"a":11}]}`, key: "result", want: `[2,6,7,8]`,
		},
		{
			name: "tilde_false", source: `gjson(message,"vals.#(b==~false)#.a","result")`,
			input: `{"vals":[{"a":1,"b":"data"},{"a":2,"b":true},{"a":3,"b":false},{"a":4,"b":"0"},{"a":5,"b":0},{"a":11}]}`, key: "result", want: `[3,4,5,11]`,
		},
		{
			name: "tilde_null", source: `gjson(message,"vals.#(b==~null)#.a","result")`,
			input: `{"vals":[{"a":1,"b":false},{"a":2,"b":null},{"a":3}]}`, key: "result", want: `[2,3]`,
		},
		{
			name: "tilde_missing", source: `gjson(message,"vals.#(b!=~*)#.a","result")`,
			input: `{"vals":[{"a":1,"b":null},{"a":2}]}`, key: "result", want: `[2]`,
		},
		{
			name: "nested_query", source: `gjson(message,"friends.#(nets.#(==\"fb\"))#.first","result")`,
			input: `{"friends":[{"first":"Dale","nets":["ig","fb"]},{"first":"Roger","nets":["fb","tw"]},{"first":"Jane","nets":["ig","tw"]}]}`, key: "result", want: `["Dale","Roger"]`,
		},
		{
			name: "recursive_nested_query", source: `gjson(message,"groups.#(members.#(roles.#(==\"admin\"))).name","result")`,
			input: `{"groups":[{"name":"viewer","members":[{"roles":["read"]}]},{"name":"owner","members":[{"roles":["read","admin"]}]}]}`, key: "result", want: "owner",
		},
	})
}

func TestJITGJSONTransformUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "root_this_preserves_raw", source: `gjson(message,"@this","result")`,
			input: `{ "hello" : "world", "arr" : [1, 2] }`, key: "result", want: `{ "hello" : "world", "arr" : [1, 2] }`,
		},
		{
			name: "nested_this_preserves_raw", source: `gjson(message,"other.@this","result")`,
			input: `{"other": { "hello" : "world" }}`, key: "result", want: `{ "hello" : "world" }`,
		},
		{
			name: "ugly", source: `gjson(message,"@ugly","result")`,
			input: `{ "hello" : "world", "arr" : [1, 2] }`, key: "result", want: `{"hello":"world","arr":[1,2]}`,
		},
		{
			name: "pipe_chain", source: `gjson(message,"info|friends|0|first","result")`,
			input: `{"info":{"friends":[{"first":"Dale"},{"first":"Roger"}]}}`, key: "result", want: "Dale",
		},
		{
			name: "reverse_chain", source: `gjson(message,"info.friends|@reverse|0|age","result")`,
			input: `{"info":{"friends":[{"age":44},{"age":68},{"age":47}]}}`, key: "result", want: float64(47),
		},
		{
			name: "valid", source: `gjson(message,"@valid","result")`,
			input: `[ 1, 2 ]`, key: "result", want: `[ 1, 2 ]`,
		},
		{
			name: "flatten", source: `gjson(message,"@flatten","result")`,
			input: `[1,[2],[3,4],[5,[6,[7]]],{"hi":"there"},8,[9]]`, key: "result", want: `[1,2,3,4,5,[6,[7]],{"hi":"there"},8,9]`,
		},
		{
			name: "flatten_deep", source: `gjson(message,"@flatten:{\"deep\":true}","result")`,
			input: `[1,[2],[3,4],[5,[6,[7]]],{"hi":"there"},8,[9]]`, key: "result", want: `[1,2,3,4,5,6,7,{"hi":"there"},8,9]`,
		},
		{
			name: "join", source: `gjson(message,"@join","result")`,
			input: `[{"a":1,"b":1},{"b":2},5,{"c":3}]`, key: "result", want: `{"a":1,"b":2,"c":3}`,
		},
		{
			name: "join_preserve_duplicates", source: `gjson(message,"@join:{\"preserve\":true}","result")`,
			input: `[{"a":1,"b":1},{"b":2},5,{"c":3}]`, key: "result", want: `{"a":1,"b":1,"b":2,"c":3}`,
		},
		{
			name: "join_preserve_then_query", source: `gjson(message,"@join:{\"preserve\":true}|b","result")`,
			input: `[{"a":1,"b":1},{"b":2}]`, key: "result", want: float64(1),
		},
		{
			name: "join_preserve_dot_query", source: `gjson(message,"@join:{\"preserve\":true}.b","result")`,
			input: `[{"a":1,"b":1},{"b":2}]`, key: "result", want: float64(1),
		},
		{
			name: "keys", source: `gjson(message,"@keys","result")`,
			input: `{"first":"Tom","last":"Smith"}`, key: "result", want: `["first","last"]`,
		},
		{
			name: "values_chain", source: `gjson(message,"@values|1|code","result")`,
			input: `{"a":{"code":"A"},"b":{"code":"B"}}`, key: "result", want: "B",
		},
		{
			name: "to_string", source: `gjson(message,"@tostr","result")`,
			input: `{"id":1023,"name":"alert"}`, key: "result", want: `{"id":1023,"name":"alert"}`,
		},
		{
			name: "from_string_chain", source: `gjson(message,"@fromstr|id","result")`,
			input: `"{\"id\":1023,\"name\":\"alert\"}"`, key: "result", want: float64(1023),
		},
		{
			name: "group", source: `gjson(message,"@group","result")`,
			input: `{"id":["123","456","789"],"val":[2,1]}`, key: "result", want: `[{"id":"123","val":2},{"id":"456","val":1},{"id":"789"}]`,
		},
		{
			name: "dig", source: `gjson(message,"@dig:name","result")`,
			input: `{"something":{"anything":{"abcdefg":{"finally":{"important":{"secret":"password","name":"jake"}},"name":"melinda"}}}}`, key: "result", want: `["melinda","jake"]`,
		},
		{
			name: "dig_nested_path", source: `gjson(message,"@dig:user.name","result")`,
			input: `{"root":{"user":{"name":"first"}},"children":[{"user":{"name":"second"}}]}`, key: "result", want: `["first","second"]`,
		},
		{
			name: "dig_query", source: `gjson(message,"group.@dig:#(refid=789)|0.fields.labels.0","result")`,
			input: `{"group":[[{"fields":{"labels":["milestone_1"]},"refid":123}],{"nested":[{"fields":{"labels":["milestone_3"]},"refid":789}]}]}`, key: "result", want: "milestone_3",
		},
		{
			name: "project_query_after_modifier", source: `gjson(message,"@ugly|#.c.#[a=10.11]","result")`,
			input: `[{"c":[{"a":10.11}]},{"c":[{"a":11.11}]}]`, key: "result", want: `[{"a":10.11}]`,
		},
		{
			name: "project_object_multipath", source: `gjson(message,"data.#.{q}|@ugly","result")`,
			input: `{"data":[{"q":11,"w":12},{"q":21,"w":22},{"q":31,"w":32}],"sql":"some stuff here"}`, key: "result", want: `[{"q":11},{"q":21},{"q":31}]`,
		},
		{
			name: "project_group_query", source: `gjson(message,"{\"id\":issues.#.id,\"plans\":issues.#.fields.labels.#(%\"plan:*\")#|#.#}|@group|#(plans>=2)#.id","result")`,
			input: `{"issues":[{"fields":{"labels":["milestone_1","group:foo","plan:a","plan:b"]},"id":"123"},{"fields":{"labels":["milestone_1","group:foo","plan:a","plan"]},"id":"456"}]}`, key: "result", want: `["123"]`,
		},
	})
}

// These cases exercise the token-preserving paths that cannot be represented
// faithfully by decoding into map[string]any: duplicate keys, multipaths,
// pretty's formatting options, and a completed selection before malformed
// unrelated input. The pipeline-go run is the oracle; key is intentionally
// empty because the complete Point comparison below is authoritative.
func TestJITGJSONTokenCompatibilityOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{
			name: "pretty_sorted", source: `gjson(message,"@pretty:{\"sortKeys\":true}","result")`,
			input: `{"z":2,"a":{"d":4,"c":3}}`,
		},
		{
			name: "pretty_indent_prefix", source: `gjson(message,"@pretty:{\"indent\":\"  \",\"prefix\":\"> \"}","result")`,
			input: `{"a":[1,2]}`,
		},
		{
			name: "array_multipath", source: `gjson(message,"[name.last,age,children.0]","result")`,
			input: `{"name":{"first":"Tom","last":"Anderson"},"age":37,"children":["Sara","Alex"]}`,
		},
		{
			name: "object_multipath", source: `gjson(message,"{surname:name.last,years:age}","result")`,
			input: `{"name":{"first":"Tom","last":"Anderson"},"age":37}`,
		},
		{
			name: "recursive_multipath_query", source: `gjson(message,"{names:groups.#(members.#(roles.#(==\"admin\")))#.name,count:groups.#}","result")`,
			input: `{"groups":[{"name":"viewer","members":[{"roles":["read"]}]},{"name":"owner","members":[{"roles":["read","admin"]}]}]}`,
		},
		{
			name: "duplicate_join_then_first", source: `gjson(message,"@join:{\"preserve\":true}|b","result")`,
			input: `[{"b":1},{"b":2}]`,
		},
		{
			name: "completed_query_before_malformed_suffix", source: `gjson(message,"items.#(id==2).name","result")`,
			input: `{"items":[{"id":1,"name":"one"},{"id":2,"name":"two"}],"broken":`,
		},
	})
}
