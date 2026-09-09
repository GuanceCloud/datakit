// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"testing"
)

func TestJITCastBoolPathsOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, input := range []string{"1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False", "TrUe", "yes", " true ", "", "2"} {
		want := input == "1" || input == "t" || input == "T" || input == "true" || input == "TRUE" || input == "True"
		for _, nested := range []bool{false, true} {
			source := `cast(message,"bool")`
			if nested {
				source = `add_key(result,cast(message,"bool") == nil)`
			}
			cases = append(cases, jsonOracleCase{name: fmt.Sprintf("%q/nested=%t", input, nested), source: source, input: input, key: "message", want: want})
		}
	}
	runJSONOracle(t, cases)
}

// Supplemental compositions of pipeline-go v1.4.3 Cast semantics.
func TestJITNestedCastOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "integer", source: `add_key(v,"42"); add_key(result,cast(v,"int") == nil)`, key: "v", want: int64(42)},
		{name: "float", source: `add_key(v,"1.25"); add_key(result,cast(v,"float") == nil)`, key: "v", want: float64(1.25)},
		{name: "bool", source: `add_key(v,"true"); add_key(result,cast(v,"bool") == nil)`, key: "v", want: true},
		{name: "missing", source: `add_key(result,cast(absent,"int") == nil)`, key: "result", want: true},
		{name: "short_circuit", source: `add_key(v,"42"); r=false && cast(v,"int") == nil; add_key(result,r)`, key: "v", want: "42"},
		{name: "local_shadow", source: `v="42"; add_key(result,cast(v,"int") == nil); add_key(local,v)`, key: "local", want: "42"},
		{name: "existing_nil", source: `add_key(v,42); cast(v,"string"); add_key(result,cast(v,"int") == nil)`, key: "v", want: int64(0)},
		{name: "existing_nil_float", source: `add_key(v,42); cast(v,"string"); add_key(result,cast(v,"float") == nil)`, key: "v", want: float64(0)},
		{name: "existing_nil_bool", source: `add_key(v,42); cast(v,"string"); add_key(result,cast(v,"bool") == nil)`, key: "v", want: false},
		{name: "existing_nil_str", source: `add_key(v,42); cast(v,"string"); add_key(result,cast(v,"str") == nil)`, key: "v", want: ""},
	})
}

// The Go checker accepts "string" as well as "str"; execution must be
// compared separately rather than assuming the two spellings are aliases.
func TestJITCastStringTargetOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "text", source: `add_key(v,"42"); cast(v,"string")`, key: "v", want: nil},
		{name: "integer", source: `add_key(v,42); cast(v,"string")`, key: "v", want: nil},
		{name: "tag", source: `set_tag(v,"42"); cast(v,"string")`, key: "v", want: ""},
		{name: "missing", source: `cast(absent,"string")`, key: "absent", want: nil},
		{name: "nested", source: `add_key(v,42); add_key(result,cast(v,"string") == nil)`, key: "v", want: nil},
		{name: "nested_tag", source: `set_tag(v,"42"); add_key(result,cast(v,"string") == nil)`, key: "v", want: ""},
		{name: "short_circuit", source: `add_key(v,42); r=false && cast(v,"string") == nil; add_key(result,r)`, key: "v", want: int64(42)},
		{name: "local", source: `v=42; add_key(result,cast(v,"string") == nil); add_key(local,v)`, key: "local", want: int64(42)},
	})
}
