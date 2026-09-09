// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_jsonall_test.go.
// MIT License; Copyright 2021-present Guance, Inc.
// All eight execution tests and eight TestJSONAllChecking assertions are
// retained. JSON whitespace and script layout are normalized, not values.
import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/engine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITJSONAllUpstreamOracle(t *testing.T) {
	table := []struct {
		name, source, input string
		want                map[string]any
		missing             []string
	}{
		{"TestJSONAllIncludeKeys", `json_all(_, include_keys=["age","active","service","name.first"])`,
			`{"service":"api","name":{"first":"Tom","last":"Anderson"},"age":37,"active":true,"children":["Sara","Alex","Jack"],"friends":[{"first":"Dale","age":44},{"first":"Roger","age":68}]}`,
			map[string]any{"service": "api", "age": float64(37), "active": true}, []string{"name", "name.first", "children", "children[0]", "friends[0].first"}},
		{"TestJSONAllIncludeKeysPositional", `json_all(_, ["age"])`, `{"name":"Tom","age":37}`,
			map[string]any{"age": float64(37)}, []string{"name"}},
		{"TestJSONAllKeyPatterns", `json_all(_, include_keys=["trace_*"], key_patterns=["trace_?d"])`, `{"trace_*":"literal","trace_id":"abc","trace_span":"def","service":"api"}`,
			map[string]any{"trace_*": "literal", "trace_id": "abc"}, []string{"trace_span", "service"}},
		{"TestJSONAllWithoutKeysDoesNothing", `json_all(_)`, `{"name":"Tom","age":37}`,
			nil, []string{"name", "age"}},
		{"TestJSONAllDynamicIncludeKeys", `keys=["service","status","name.first"]; json_all(_, include_keys=keys)`, `{"service":"api","status":"ok","name":{"first":"Tom","last":"Anderson"},"age":37}`,
			map[string]any{"service": "api", "status": "ok"}, []string{"name.first", "age"}},
		{"TestJSONAllTopLevelArray", `json_all(_, include_keys=["[0]","[1]","[2].name"])`, `["first",2,{"name":"nested"}]`,
			map[string]any{"[0]": "first", "[1]": float64(2)}, []string{"[2]", "[2].name"}},
		{"TestJSONAllEmptyIncludeKeys", `json_all(_, include_keys=[])`, `{"name":"Tom","age":37}`,
			nil, []string{"name", "age"}},
		{"TestJSONAllInvalidJSONDoesNothing", `json_all(_, include_keys=["service"])`, `{"service":"api"`,
			nil, []string{"service"}},
	}
	var cases []jsonOracleCase
	for _, tc := range table {
		t.Run(tc.name+"/original_assertions", func(t *testing.T) {
			// Match upstream NewTestingRunner/runScript, without DataKit's
			// successful-run status/time finalization. The production oracle
			// below keeps finalization enabled on both engines.
			scripts, errors := engine.ParseScript(map[string]string{"jsonall-upstream.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errors) != 0 {
				t.Fatal(errors)
			}
			script := scripts["jsonall-upstream.p"]
			if script == nil {
				t.Fatal("upstream compiler returned no script")
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": tc.input}, time.Unix(1700000000, 123))
			if err := script.Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.want {
				got, _, err := pt.Get(key)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("%s got=%#v err=%v want=%#v", key, got, err, want)
				}
			}
			for _, key := range tc.missing {
				if got, _, err := pt.Get(key); err == nil {
					t.Fatalf("%s must be absent, got=%#v", key, got)
				}
			}
		})
		cases = append(cases, jsonOracleCase{name: tc.name, source: tc.source, input: tc.input})
	}
	runJSONOracle(t, cases)
}

func TestJITJSONAllUpstreamCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for i, source := range []string{
		`json_all()`, `json_all(include_keys=["age"])`, `json_all(["age"])`, `json_all(_, limit=3)`,
		`json_all(_, include_keys=[1])`, `json_all(_, key_patterns=[1])`,
		`json_all(_, include_keys="age")`, `json_all(_, key_patterns="trace_*")`,
	} {
		t.Run(fmt.Sprintf("TestJSONAllChecking/%d", i), func(t *testing.T) {
			if _, err := NewPlScriptSimple(point.Logging, "jsonall-check.p", source); err == nil {
				t.Fatal("Go accepted upstream invalid source")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative {
				t.Fatalf("JIT accepted invalid source: %+v", check)
			}
		})
	}
}

func TestJITJSONAllDynamicFilterOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, filter := range []struct {
		name, expression string
		fails            bool
	}{
		{"nil", "nil", false}, {"missing", "missing", false}, {"empty", "[]", false},
		{"string", `"a"`, true}, {"integer", "1", true}, {"map", `{"a":1}`, true},
		{"bad_element", `["a",1]`, true}, {"nil_element", `[nil]`, true},
	} {
		for _, parameter := range []string{"include_keys", "key_patterns"} {
			for _, input := range []struct{ name, value string }{{"valid", `{"a":1}`}, {"malformed", `{"a":`}} {
				cases = append(cases, jsonOracleCase{
					name:   parameter + "/" + filter.name + "/" + input.name,
					source: "filters=" + filter.expression + "; json_all(_, " + parameter + "=filters)",
					input:  input.value, fails: filter.fails,
				})
			}
		}
	}
	cases = append(cases,
		jsonOracleCase{"mutation_between_calls", `filters=["a"]; json_all(_,include_keys=filters); filters[0]="b"; json_all(_,include_keys=filters)`, `{"a":1,"b":2}`, "b", float64(2), false},
		jsonOracleCase{"failure_after_first_call", `filters=["a"]; json_all(_,include_keys=filters); filters[0]=1; json_all(_,include_keys=filters)`, `{"a":1}`, "a", float64(1), true},
	)
	runJSONOracle(t, cases)
}

func TestJITJSONAllWildcardBoundaryOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"unicode_question", `p=["?"]; json_all(_,key_patterns=p)`, `{"中":1,"😀":2,"ab":3}`, "中", float64(1), false},
		{"newline_star", `p=["*"]; json_all(_,key_patterns=p)`, "{\"a\\nb\":1,\"normal\":2}", "a\nb", nil, false},
		{"literal_regex", `p=["a.b"]; json_all(_,key_patterns=p)`, `{"a.b":1,"axb":2}`, "a.b", float64(1), false},
		{"literal_brackets", `p=["[a]"]; json_all(_,key_patterns=p)`, `{"[a]":1,"a":2}`, "[a]", float64(1), false},
		{"literal_backslash", `p=["a\\b"]; json_all(_,key_patterns=p)`, `{"a\\b":1,"ab":2}`, "a\\b", float64(1), false},
		{"empty_key", `p=["*"]; json_all(_,key_patterns=p)`, `{"":1,"a":2}`, "a", float64(2), false},
	})
}

func TestJITJSONAllArgumentEffectsOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "bad_include_type", source: `f=1; json_all(_,include_keys=f,key_patterns=pt_kvs_keys(fields=pt_kvs_set("later",true)))`, input: `{"a":7}`, key: "later", want: nil, fails: true},
		{name: "bad_include_element", source: `f=[1]; json_all(_,include_keys=f,key_patterns=pt_kvs_keys(fields=pt_kvs_set("later",true)))`, input: `{"a":7}`, key: "later", want: nil, fails: true},
		{name: "reversed_bad_include", source: `f=[1]; json_all(_,key_patterns=pt_kvs_keys(fields=pt_kvs_set("later",true)),include_keys=f)`, input: `{"a":7}`, key: "later", want: nil, fails: true},
		{name: "ordered_filters", source: `json_all(_,key_patterns=pt_kvs_keys(fields=pt_kvs_set("order","patterns")),include_keys=pt_kvs_keys(fields=pt_kvs_set("order","include")))`, input: `{"a":7}`, key: "order", want: "patterns"},
		{name: "malformed_after_filters", source: `json_all(_,include_keys=pt_kvs_keys(fields=pt_kvs_set("evaluated",true)))`, input: `bad`, key: "evaluated", want: true},
		{name: "append_result_independence", source: `f=["a"]; json_all(_,include_keys=f,key_patterns=pt_kvs_keys(fields=pt_kvs_set("mutation",append(f,"b"))))`, input: `{"a":7,"b":8}`, key: "b", want: nil},
	})
}

func TestJITJSONAllFilterBindingOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "later_rebind", source: `f=["a"]; alias=f; json_all(_,include_keys=f,key_patterns=pt_kvs_keys(fields=pt_kvs_set("changed",load_json(f="[\"b\"]")))); add_key(binding,f); add_key(original,alias)`, input: `{"a":7,"b":8}`, key: "a", want: float64(7)},
		{name: "reverse_named_rebind", source: `f=["a"]; json_all(_,key_patterns=pt_kvs_keys(fields=pt_kvs_set("changed",load_json(f="[\"b\"]"))),include_keys=f)`, input: `{"a":7,"b":8}`, key: "b", want: nil},
		{name: "shared_filters", source: `f=["a"]; alias=f; json_all(_,include_keys=f,key_patterns=alias); add_key(original,alias)`, input: `{"a":7,"b":8}`, key: "a", want: float64(7)},
		{name: "source_changed_by_patterns", source: `json_all(_,include_keys=["a"],key_patterns=pt_kvs_keys(fields=pt_kvs_set("message","{\"a\":9}")))`, input: `{"a":7}`, key: "a", want: float64(9)},
		{name: "source_invalidated_by_patterns", source: `json_all(_,include_keys=["a"],key_patterns=pt_kvs_keys(fields=pt_kvs_set("message","bad")))`, input: `{"a":7}`, key: "a", want: nil},
	})
}

func TestJITJSONAllRawFilterOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"raw_include", `filters=[message]; payload='{"a":1}'; json_all(payload,include_keys=filters)`, "\xff", "a", nil, false},
		{"mixed_include", `filters=[message,"a"]; payload='{"a":1}'; json_all(payload,include_keys=filters)`, "\xff", "a", float64(1), false},
		{"raw_pattern", `filters=[message]; payload='{"a":1}'; json_all(payload,key_patterns=filters)`, "\xff", "a", nil, false},
		{"raw_pattern_invalid_json", `filters=[message]; payload='bad'; json_all(payload,key_patterns=filters)`, "\xff", "a", nil, false},
		{"raw_pattern_rune", `filters=[message]; payload='{"�":1}'; json_all(payload,key_patterns=filters)`, "\xff", "�", float64(1), false},
		{"raw_pattern_truncated", `filters=[message]; payload='{"��":1,"�":2}'; json_all(payload,key_patterns=filters)`, "\xe2\x82", "��", float64(1), false},
		{"raw_include_not_rune", `filters=[message,"a"]; payload='{"�":2,"a":1}'; json_all(payload,include_keys=filters)`, "\xff", "�", nil, false},
		{"unselected_raw_key", `json_all(_,include_keys=["a"])`, "{\"\xff\":2,\"a\":1}", "a", float64(1), false},
		{"unmatched_raw_key_pattern", `json_all(_,key_patterns=["a"])`, "{\"\xff\":2,\"a\":1}", "a", float64(1), false},
	})
}

func TestGoJSONAllSelectedByteKeyContract(t *testing.T) {
	for _, key := range []string{"\xff", "\xe2\x82", "a\x00b"} {
		t.Run(fmt.Sprintf("%x", key), func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, "byte-key.p", `json_all(_,key_patterns=["*"]); value=pt_kvs_get(selector); add_key(observed,value)`)
			if err != nil {
				t.Fatal(err)
			}
			input := "{\"" + key + "\":7}"
			if key == "a\x00b" {
				input = `{"a\u0000b":7}`
			}
			pt := newRealScriptPoint("bytes", map[string]any{"message": input, "selector": key})
			if err := script.Run(ptinput.PtWrap(point.Logging, pt), nil, nil); err != nil {
				t.Fatal(err)
			}
			if pt.Get(key) != float64(7) || pt.Get("observed") != float64(7) {
				t.Fatalf("byte key or subsequent lookup lost: %#v", pt.KVMap())
			}
		})
	}
}

// These scripts remove their temporary byte key before returning. They verify
// real Point storage and dynamic key operations, not raw-key output wire support.
func TestJITBytePointKeyLifecycleOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, key := range []string{"\xff", "\xe2\x82", "a\x00b"} {
		for _, scenario := range []struct{ name, source string }{
			{"field", `raw=message; pt_kvs_set(raw,7); add_key(observed,pt_kvs_get(raw)); pt_kvs_del(raw)`},
			{"tag", `raw=message; pt_kvs_set(raw,"tag",true); add_key(observed,pt_kvs_get(raw)); pt_kvs_del(raw)`},
			{"coexist", `raw=message; pt_kvs_set(raw,7); pt_kvs_set("�",8); add_key(observed,pt_kvs_get(raw)); add_key(other,pt_kvs_get("�")); pt_kvs_del(raw)`},
			{"enumerate", `raw=message; pt_kvs_set(raw,7); names=pt_kvs_keys(); for k in names { if k == raw { add_key(found,true) } }; pt_kvs_del(raw)`},
		} {
			cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%s/%x", scenario.name, key), source: scenario.source, input: key})
		}
	}
	runJSONOracle(t, cases)
}

func TestJITBytePointKeyOutputOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, key := range []string{"\xff", "\xfe", "\xe2\x82"} {
		cases = append(cases,
			jsonOracleCase{name: fmt.Sprintf("json/%x", key), source: `json_all(_,key_patterns=["*"])`, input: "{\"" + key + "\":7,\"�\":8}", key: key, want: float64(7)},
			jsonOracleCase{name: fmt.Sprintf("field/%x", key), source: `raw=message; pt_kvs_set(raw,7); add_key(observed,pt_kvs_get(raw))`, input: key, key: key, want: int64(7)},
			jsonOracleCase{name: fmt.Sprintf("tag/%x", key), source: `raw=message; pt_kvs_set(raw,"tag",true)`, input: key, key: key, want: "tag"},
			jsonOracleCase{name: fmt.Sprintf("raw_tag/%x", key), source: `raw=message; pt_kvs_set(raw,raw,true)`, input: key, key: key, want: key},
			jsonOracleCase{name: fmt.Sprintf("error_prefix/%x", key), source: `raw=message; pt_kvs_set(raw,7); fail=1/divisor`, input: key, key: key, want: int64(7), fails: true},
		)
	}
	runJSONOracle(t, cases)
}

func TestJITBytePointKeyInputOracle(t *testing.T) {
	for _, tag := range []bool{false, true} {
		t.Run(fmt.Sprintf("tag=%t", tag), func(t *testing.T) {
			runJSONOracle(t, []jsonOracleCase{
				{name: "read", source: `add_key(observed,pt_kvs_get(message))`, input: "\xff"},
				{name: "delete", source: `pt_kvs_del(message)`, input: "\xff"},
				{name: "update", source: `pt_kvs_set(message,9)`, input: "\xff"},
			}, func(pt *point.Point) {
				if tag {
					pt.AddKVs(point.NewKV("\xff", "old", point.WithKVTagSet(true)))
				} else {
					pt.AddKVs(point.NewKV("\xff", int64(7)))
				}
			})
		})
	}
}

func TestJITJSONAllOriginAliasOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{"include_alias", `json_all(_,include_keys=["_"]); add_key(observed,_)`, `{"_":"replacement"}`, "message", "replacement", false},
		{"pattern_alias", `json_all(_,key_patterns=["*"]); add_key(observed,_)`, `{"_":"replacement","a":1}`, "message", "replacement", false},
		{"alias_numeric", `json_all(_,include_keys=["_"])`, `{"_":7}`, "message", float64(7), false},
		{"alias_before_message", `json_all(_,key_patterns=["*"])`, `{"_":"first","message":"last"}`, "message", "last", false},
		{"alias_after_message", `json_all(_,key_patterns=["*"])`, `{"message":"first","_":"last"}`, "message", "last", false},
		{"filter_before_alias", `json_all(_,include_keys=["message"])`, `{"_":"not-selected"}`, "message", `{"_":"not-selected"}`, false},
		{"nul_key", `json_all(_,key_patterns=["*"])`, `{"a\u0000b":7}`, "a\x00b", float64(7), false},
	})
}
