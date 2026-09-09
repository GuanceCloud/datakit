// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITNestedCaseMutationOracle(t *testing.T) {
	for _, fn := range []string{"lowercase", "uppercase"} {
		t.Run(fn, func(t *testing.T) {
			want := "mixed"
			if fn == "uppercase" {
				want = "MIXED"
			}
			convert := strings.ToLower
			if fn == "uppercase" {
				convert = strings.ToUpper
			}
			runJSONOracle(t, []jsonOracleCase{
				{name: "field", source: fmt.Sprintf(`add_key(value,"MiXeD"); add_key(result,%s(value) == nil)`, fn), key: "value", want: want},
				{name: "tag", source: fmt.Sprintf(`set_tag(value,"MiXeD"); add_key(result,%s(value) == nil)`, fn), key: "value", want: want},
				{name: "missing", source: fmt.Sprintf(`add_key(result,%s(absent) == nil)`, fn), key: "result", want: true},
				{name: "short_circuit", source: fmt.Sprintf(`add_key(value,"MiXeD"); result=false && %s(value) == nil`, fn), key: "value", want: "MiXeD"},
				{name: "prefix", source: fmt.Sprintf(`add_key(value,"MiXeD"); result=(%s(value) == nil) && 1/divisor == 0`, fn), key: "value", want: want, fails: true},
				{name: "attribute", source: fmt.Sprintf(`add_key("obj.name","MiXeD"); add_key(result,%s(obj.name) == nil)`, fn), key: "obj.name", want: want},
				{name: "origin", source: fmt.Sprintf(`add_key(result,%s(_) == nil)`, fn), input: "MiXeD", key: "message", want: want},
				{name: "unicode", source: fmt.Sprintf(`add_key(result,%s(_) == nil)`, fn), input: "İıßΣςσKſ中", key: "message", want: convert("İıßΣςσKſ中")},
				{name: "invalid_utf8", source: fmt.Sprintf(`add_key(result,%s(_) == nil)`, fn), input: "A\xff\xc3z", key: "message", want: convert("A\xff\xc3z")},
				{name: "numeric", source: fmt.Sprintf(`add_key(value,123); add_key(result,%s(value) == nil)`, fn), key: "value", want: "123"},
				{name: "boolean", source: fmt.Sprintf(`add_key(value,true); add_key(result,%s(value) == nil)`, fn), key: "value", want: convert("true")},
				{name: "list", source: fmt.Sprintf(`value=["MiXeD",1,true]; add_key(result,%s(value) == nil)`, fn), key: "value", want: convert(`["MiXeD",1,true]`)},
				{name: "map", source: fmt.Sprintf(`value={"Key":"MiXeD"}; add_key(result,%s(value) == nil)`, fn), key: "value", want: convert(`{"Key":"MiXeD"}`)},
				{name: "nil", source: fmt.Sprintf(`value=nil; add_key(result,%s(value) == nil)`, fn), key: "result", want: true},
			})
		})
	}
}

func TestJITCaseMutationCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 64)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, fn := range []string{"lowercase", "uppercase"} {
		for _, arg := range []string{"", "value,other", "123", "true", "[]", "{}", `len("x")`, `value="x"`, "(value)", "value", `"value"`, "obj.name", "obj[0]"} {
			for _, nested := range []bool{false, true} {
				source := fmt.Sprintf("%s(%s)", fn, arg)
				if nested {
					source = "add_key(result," + source + " == nil)"
				}
				t.Run(source, func(t *testing.T) {
					_, goErr := NewPlScriptSimple(point.Logging, "checking.p", source)
					check := runner.Check(source)
					if (goErr == nil) != (check.Route == pljit.RouteJITNative) {
						t.Fatalf("Go error=%v JIT=%+v", goErr, check)
					}
				})
			}
		}
	}
}

func TestJITNestedCasePointRepresentationsOracle(t *testing.T) {
	for _, fn := range []string{"lowercase", "uppercase"} {
		convert := strings.ToLower
		if fn == "uppercase" {
			convert = strings.ToUpper
		}
		for _, fixture := range []struct {
			name  string
			value any
			tag   bool
			text  string
		}{
			{"raw_tag", "A\xff\xc3z", true, "A\xff\xc3z"},
			{"integer_array", []int{1, 2, 3}, false, "[1,2,3]"},
			{"string_array", []string{"MiXeD", "İ"}, false, `["MiXeD","İ"]`},
		} {
			t.Run(fn+"/"+fixture.name, func(t *testing.T) {
				runJSONOracle(t, []jsonOracleCase{{name: "convert", source: fmt.Sprintf(`add_key(result,%s(value) == nil)`, fn), key: "value", want: convert(fixture.text)}}, func(pt *point.Point) {
					kv := point.NewKV("value", fixture.value, point.WithKVTagSet(fixture.tag))
					if kv.IsTag != fixture.tag {
						t.Fatal("fixture lost tag/field identity")
					}
					pt.AddKVs(kv)
				})
			})
		}
	}
}
