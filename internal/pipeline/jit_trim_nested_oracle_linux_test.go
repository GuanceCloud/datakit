// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"github.com/GuanceCloud/cliutils/point"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"os"
	"testing"
)

func TestJITTrimCheckingOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 64)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	for _, args := range []string{"", "value", `value,""`, `value,"xy"`, "value,123", "value,nil", `value,cutset="x"`, `value,"x",1`, "123", "obj.name", "obj[0]", "(value)", `"value"`} {
		for _, nested := range []bool{false, true} {
			source := "trim(" + args + ")"
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
	for _, source := range []string{
		`cutset="xy"; trim(value,cutset)`,
		`cutset="xy"; add_key(result,trim(value,cutset) == nil)`,
	} {
		if check := runner.Check(source); check.Route != pljit.RouteJITNative {
			t.Fatalf("dynamic trim extension rejected: source=%q check=%+v", source, check)
		}
	}
}

func TestJITNestedTrimOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "space", source: `add_key(result,trim(_) == nil)`, input: " \tvalue\n", key: "message", want: "value"},
		{name: "unicode_space", source: `add_key(result,trim(_) == nil)`, input: "\u2003value\u00a0", key: "message", want: "value"},
		{name: "cutset", source: `add_key(result,trim(_,"xy") == nil)`, input: "xyvalueyx", key: "message", want: "value"},
		{name: "empty_cutset", source: `add_key(result,trim(_,"") == nil)`, input: " value ", key: "message", want: "value"},
		{name: "tag", source: `set_tag(value," value "); add_key(result,trim(value) == nil)`, key: "value", want: "value"},
		{name: "missing", source: `add_key(result,trim(absent) == nil)`, key: "result", want: true},
		{name: "short_circuit", source: `result=false && trim(_) == nil`, input: " value ", key: "message", want: " value "},
		{name: "prefix", source: `result=(trim(_) == nil) && 1/divisor == 0`, input: " value ", key: "message", want: "value", fails: true},
		{name: "attribute", source: `add_key("obj.name"," value "); add_key(result,trim(obj.name) == nil)`, key: "obj.name", want: "value"},
		{name: "raw_preserved", source: `add_key(result,trim(_) == nil)`, input: " \xffvalue\xc3 ", key: "message", want: "\xffvalue\xc3"},
		{name: "raw_cutset", source: `add_key(result,trim(_,"�") == nil)`, input: "\xffvalue\xc3", key: "message", want: "value"},
		{name: "numeric", source: `value=123; add_key(result,trim(value) == nil)`, key: "value", want: "123"},
		{name: "list", source: `value=[" value ",1]; add_key(result,trim(value) == nil)`, key: "value", want: `[" value ",1]`},
		{name: "nil", source: `value=nil; add_key(result,trim(value) == nil)`, key: "result", want: true},
	})
}
