// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_pt_kvs_test.go.
// MIT License; Copyright 2021-present Guance, Inc.
// Six TestPtKvsSetMap execution tests and three checking assertions; only
// script whitespace and the clock are normalized. Raw dtype assertions remain.
import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITSetMapUpstreamOracle(t *testing.T) {
	table := []struct {
		name, source string
		want         map[string]any
		missing      []string
	}{
		{"TestPtKvsSetMapIncludeKeysAndWildcard", `fields={"service":"api","status":200,"trace_id":"abc","trace_span":"def","literal_*":"literal","drop":"x"}; count=pt_kvs_set_map(fields,include_keys=["literal_*","service"],key_patterns=["trace_*"]); pt_kvs_set("count",count)`, map[string]any{"service": "api", "trace_id": "abc", "literal_*": "literal", "count": int64(4), "trace_span": "def"}, []string{"drop"}},
		{"TestPtKvsSetMapWithoutKeysDoesNothing", `fields={"c":3,"a":1,"b":2}; count=pt_kvs_set_map(fields); pt_kvs_set("count",count)`, map[string]any{"count": int64(0)}, []string{"a", "b", "c"}},
		{"TestPtKvsSetMapEmptyIncludeKeys", `fields={"a":1}; count=pt_kvs_set_map(fields,include_keys=[]); pt_kvs_set("count",count)`, map[string]any{"count": int64(0)}, []string{"a"}},
		{"TestPtKvsSetMapDuplicateIncludeKeys", `fields={"a":1}; count=pt_kvs_set_map(fields,include_keys=["a","a"]); pt_kvs_set("count",count)`, map[string]any{"a": int64(1), "count": int64(1)}, nil},
		{"TestPtKvsSetMapDynamicIncludeKeys", `fields={"a":1,"b":2}; keys=["b"]; count=pt_kvs_set_map(fields,include_keys=keys); pt_kvs_set("count",count)`, map[string]any{"b": int64(2), "count": int64(1)}, []string{"a"}},
		{"TestPtKvsSetMapRawAndTag", `fields={"obj":{"a":1},"nums":[1,2],"tag":3}; pt_kvs_set_map(fields,include_keys=["obj","nums"],raw=true); pt_kvs_set_map(fields,include_keys=["tag"],as_tag=true)`, map[string]any{"tag": "3"}, nil},
	}
	var cases []jsonOracleCase
	for _, tc := range table {
		t.Run(tc.name+"/original_assertions", func(t *testing.T) {
			scripts, errs := engine.ParseScript(map[string]string{"upstream.p": tc.source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) != 0 {
				t.Fatal(errs)
			}
			pt := ptinput.NewPlPt(point.Logging, "test", nil, map[string]any{"message": "test"}, time.Unix(1700000000, 123))
			if err := scripts["upstream.p"].Run(pt, nil); err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.want {
				got, _, err := pt.Get(key)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("%s got=%#v err=%v want=%#v", key, got, err, want)
				}
			}
			for _, key := range tc.missing {
				if _, _, err := pt.Get(key); err == nil {
					t.Fatalf("%s must be absent", key)
				}
			}
			if tc.name == "TestPtKvsSetMapRawAndTag" {
				got, typ, err := pt.GetRaw("obj")
				if err != nil || typ != ast.Map || !reflect.DeepEqual(got, map[string]any{"a": int64(1)}) {
					t.Fatalf("raw obj=%#v type=%v err=%v", got, typ, err)
				}
				got, typ, err = pt.GetRaw("nums")
				if err != nil || typ != ast.List || !reflect.DeepEqual(got, []any{int64(1), int64(2)}) {
					t.Fatalf("raw nums=%#v type=%v err=%v", got, typ, err)
				}
				if pt.Tags()["tag"] != "3" {
					t.Fatalf("tag kind/value lost: %#v", pt.Tags())
				}
			}
		})
		cases = append(cases, jsonOracleCase{name: tc.name, source: tc.source, input: "test"})
	}
	runJSONOracle(t, cases)
}

// All eight original TestPtKvsTag table entries, including its duplicate set8
// label. Scripts only have whitespace normalized; expected values are retained.
func TestJITPointTagUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "set1", source: `pt_kvs_set("key0","abc",true); pt_kvs_set("key1",pt_kvs_get("key0")); pt_kvs_del("key0"); if pt_kvs_get("key0") == nil { for k in pt_kvs_keys() { if k == "key1" { pt_kvs_set("key2",pt_kvs_get("key1")) } } }`, key: "key2", want: "abc"},
		{name: "set2", source: `pt_kvs_set("key1",1,true); pt_kvs_set("key2",2); count=0; if "key1" in pt_kvs_keys(tags=true,fields=false) { count+=1 }; fields_key=pt_kvs_keys(false); if "key1" in fields_key { count=-1 }; if "key2" in fields_key { count+=2 }; if count == 3 { pt_kvs_set("test_ok",1,true) }`, key: "test_ok", want: "1"},
		{name: "set4", source: `pt_kvs_set("key1",as_tag=true,value=1.1)`, key: "key1", want: "1.1"},
		{name: "set5", source: `pt_kvs_set("key1",true,true)`, key: "key1", want: "true"},
		{name: "set6", source: `key_name="key1"; pt_kvs_set(key_name,[1,2],true)`, key: "key1", want: "[1,2]"},
		{name: "set7", source: `pt_kvs_set("key1",{"a":1,"b":2},true)`, key: "key1", want: `{"a":1,"b":2}`},
		{name: "set8_nil", source: `pt_kvs_set("key1",nil,true)`, key: "key1", want: ""},
		{name: "set8_underscore", source: `pt_kvs_set("_","1",true)`, key: "_", want: "1"},
	})
}

// All eight TestPtKvsSet entries, preserving the duplicate integer case and
// the original trailing-dot float syntax and expected Go types.
func TestJITPointSetUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "set1", source: `pt_kvs_set("key1","abc")`, key: "key1", want: "abc"},
		{name: "set2", source: `pt_kvs_set("key1",1)`, key: "key1", want: int64(1)},
		{name: "set3", source: `pt_kvs_set("key1",1)`, key: "key1", want: int64(1)},
		{name: "set4", source: `pt_kvs_set("key1",1.)`, key: "key1", want: float64(1)},
		{name: "set5", source: `pt_kvs_set("key1",true)`, key: "key1", want: true},
		{name: "set6", source: `pt_kvs_set("key1",[1,2])`, key: "key1", want: "[1,2]"},
		{name: "set7", source: `pt_kvs_set("key1",{"a":1,"b":2})`, key: "key1", want: `{"a":1,"b":2}`},
		{name: "set8", source: `pt_kvs_set("key1",nil)`, key: "key1", want: nil},
	})
}

func TestJITPointGetCompositeUpstreamOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{{name: "list", source: `pt_kvs_set("key2",pt_kvs_get("key1"))`, key: "key2", want: "[1,2]"}}, func(pt *point.Point) {
		pt.AddKVs(point.NewKV("key1", []int{1, 2}))
	})
}

// Script-visible extension of TestPtKvsKeysSkipsNonStringTag's input fixture.
// The upstream test calls an internal helper; these are supplementary cases,
// not a claim that its direct helper assertions have been migrated.
func TestJITPointNonStringTagKeysOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "tags_skip_numeric", source: `keys=pt_kvs_keys(tags=true,fields=false); add_key(found,"bad_tag" in keys)`, key: "found", want: false},
		{name: "tags_keep_string", source: `keys=pt_kvs_keys(tags=true,fields=false); add_key(found,"good_tag" in keys)`, key: "found", want: true},
		{name: "fields_include_numeric_tag", source: `keys=pt_kvs_keys(tags=false,fields=true); add_key(found,"bad_tag" in keys)`, key: "found", want: true},
		{name: "fields_keep_field", source: `keys=pt_kvs_keys(tags=false,fields=true); add_key(found,"field" in keys)`, key: "found", want: true},
		{name: "both_include_numeric_tag", source: `keys=pt_kvs_keys(); add_key(found,"bad_tag" in keys)`, key: "found", want: true},
	}, func(pt *point.Point) {
		pt.AddKVs(point.NewKV("bad_tag", int64(1), point.WithKVTagSet(true)), point.NewKV("good_tag", "x", point.WithKVTagSet(true)), point.NewKV("field", int64(1)))
		// WithKVTagSet does not manufacture an invalid numeric protobuf tag.
		// Pin the actual fixture instead of inferring it from the key's name.
		for _, kv := range pt.KVs() {
			if kv.Key == "bad_tag" && kv.IsTag {
				t.Fatal("fixture changed: numeric value remained a tag")
			}
		}
	})
}

func TestJITSetMapUpstreamCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, source := range []string{`pt_kvs_set_map({"a":1},include_keys=[1])`, `pt_kvs_set_map({"a":1},key_patterns=[1])`, `pt_kvs_set_map({"a":1},limit=1)`} {
		t.Run(source, func(t *testing.T) {
			_, errs := engine.ParseScript(map[string]string{"invalid.p": source}, funcs.FuncsMap, funcs.FuncsCheckMap)
			if len(errs) == 0 {
				t.Fatal("upstream accepted invalid source")
			}
			if check := runner.Check(source); check.Route == pljit.RouteJITNative {
				t.Fatalf("JIT accepted invalid source: %+v", check)
			}
		})
	}
}
