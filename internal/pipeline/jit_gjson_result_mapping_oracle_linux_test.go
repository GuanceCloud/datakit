// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"testing"

	"github.com/tidwall/gjson"
)

// TestJITGJSONResultMappingOracle checks the observable conversion performed by
// pipeline-go's gjson wrapper, rather than only checking that a path is admitted.
// The locked gjson v1.17.0 dependency is the path/result oracle; runJSONOracle
// then executes both pipeline-go and the real native runtime over full Points.
func TestJITGJSONResultMappingOracle(t *testing.T) {
	cases := []struct {
		name, input, path string
	}{
		{"integer_above_53_bits", `{"value":634866135153775564}`, "value"},
		{"integer_at_max_uint64", `{"value":18446744073709551615}`, "value"},
		{"integer_above_uint64", `{"value":18446744073709551616}`, "value"},
		{"integer_at_min_int64", `{"value":-9223372036854775808}`, "value"},
		{"integer_above_int64", `{"value":9223372036854775808}`, "value"},
		{"negative_above_53_bits", `{"value":-9007199254740993}`, "value"},
		{"fraction_above_53_bits", `{"value":634866135153775564.88172}`, "value"},
		{"positive_exponent", `{"value":1.25e40}`, "value"},
		{"negative_exponent", `{"value":-7.5e-40}`, "value"},
		{"negative_zero", `{"value":-0}`, "value"},
		{"escaped_surrogate_pair", `{"value":"open \\ud83d\\udd13"}`, "value"},
		{"escaped_nul", `{"value":"a\\u0000b"}`, "value"},
		{"escaped_control", `{"value":"line\\nnext\\tend"}`, "value"},
		{"unicode_key", `{"的情况下解":{"的情况":"匹配"}}`, "的情况下解.的情况"},
		{"escaped_dot_key", `{"lastly":{"end...ing":"soon"}}`, `lastly.end\.\.\.ing`},
		{"escaped_wildcard_key", `{"a*b":"literal","axb":"wild"}`, `a\*b`},
		{"array_json_preserves_token", `{"value":[1,true,null,{"x":"y"}]}`, "value"},
		{"object_json_preserves_token", `{"value":{"x":1,"nested":[2,3]}}`, "value"},
		{"duplicate_key_first_match", `{"value":"first","value":"second"}`, "value"},
		{"root_scalar", `123.5`, "@this"},
		{"valid_root_scalar", `123.5`, "@valid"},
		{"ugly_root_scalar", ` 123.5 `, "@ugly"},
		{"pretty_root_scalar", `123.5`, "@pretty"},
		{"flatten_root_scalar", `123.5`, "@flatten"},
		{"join_root_scalar", `123.5`, "@join"},
		{"root_boolean", `true`, "@this"},
		{"root_string", `"root text"`, "@this"},
		{"array_count", `[1,2,3,4]`, "#"},
		{"object_projection", `[{"name":"tom"},false,{"name":"janet"},null]`, "#.name"},
	}

	oracle := make([]jsonOracleCase, 0, len(cases))
	for _, tc := range cases {
		result := gjson.Get(tc.input, tc.path)
		var want any
		switch result.Type {
		case gjson.Number:
			want = result.Float()
		case gjson.True, gjson.False:
			want = result.Bool()
		case gjson.String, gjson.JSON:
			want = result.String()
		default:
			t.Fatalf("fixture %s unexpectedly produced non-observable type %s", tc.name, result.Type)
		}
		oracle = append(oracle, jsonOracleCase{
			name: tc.name, input: tc.input, key: "result", want: want,
			source: fmt.Sprintf("gjson(message,%q,\"result\")", tc.path),
		})
	}
	runJSONOracle(t, oracle)
}
