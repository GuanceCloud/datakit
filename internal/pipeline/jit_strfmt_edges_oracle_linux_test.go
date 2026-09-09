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

func TestJITStrfmtDiagnosticOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct {
		name, format, args string
		values             []any
	}{
		{"missing", "%s %d", `,"x"`, []any{"x"}},
		{"extra", "%s", `,"x",12`, []any{"x", int64(12)}},
		{"wrong_type", "%d", `,"x"`, []any{"x"}},
		{"bad_verb", "%z", `,12`, []any{int64(12)}},
		{"trailing_percent", "end%", ``, nil},
		{"literal_percent", "%%", ``, nil},
		{"indexed", "%[2]s %[1]d", `,12,"x"`, []any{int64(12), "x"}},
		{"bad_index", "%[3]s", `,"x"`, []any{"x"}},
		{"star_width", "%*s", `,4,"x"`, []any{int64(4), "x"}},
		{"star_precision", "%.*f", `,2,1.25`, []any{int64(2), 1.25}},
		{"negative_width", "%*s", `,-4,"x"`, []any{int64(-4), "x"}},
		{"bad_width", "%*s", `,"bad","x"`, []any{"bad", "x"}},
		{"signed_zero_pad", "%+08d", `,42`, []any{int64(42)}},
		{"alternate_hex", "%#x", `,42`, []any{int64(42)}},
		{"binary", "%b", `,42`, []any{int64(42)}},
		{"float_exponent", "%+10.3e", `,12.5`, []any{12.5}},
		{"float_compact", "%g", `,1200.5`, []any{1200.5}},
		{"quoted_string", "%q", `,"a\nb"`, []any{"a\nb"}},
		{"hex_string", "% x", `,"AB"`, []any{"AB"}},
		{"unicode_string_precision", "%.2s", `,"中文A"`, []any{"中文A"}},
		{"nil_value", "%v/%#v", `,nil,nil`, []any{nil, nil}},
		{"list_value", "%v", `,[1,"x",true]`, []any{[]any{int64(1), "x", true}}},
		{"map_value", "%v", `,{"b":2,"a":"x"}`, []any{map[string]any{"b": int64(2), "a": "x"}}},
		{"sharp_list", "%#v", `,[1,"x",true]`, []any{[]any{int64(1), "x", true}}},
		{"sharp_map", "%#v", `,{"b":2,"a":"x"}`, []any{map[string]any{"b": int64(2), "a": "x"}}},
		{"quoted_integer", "%q", `,42`, []any{int64(42)}},
		{"unicode_integer", "%U", `,20013`, []any{int64(20013)}},
		{"unicode_integer_sharp", "%#U", `,65`, []any{int64(65)}},
		{"modern_octal", "%O", `,42`, []any{int64(42)}},
		{"float_binary", "%b", `,12.5`, []any{12.5}},
		{"float_hex", "%x", `,12.5`, []any{12.5}},
		{"float_upper_hex", "%X", `,12.5`, []any{12.5}},
		{"float_upper_fixed", "%F", `,12.5`, []any{12.5}},
		{"float_hex_precision", "%.2x", `,12.5`, []any{12.5}},
		{"float_hex_sharp", "%#x", `,12.5`, []any{12.5}},
		{"float_hex_sharp_zero_precision", "%#.0x", `,12.5`, []any{12.5}},
		{"float_hex_zero", "%x", `,0.0`, []any{0.0}},
		{"float_hex_zero_precision", "%.2x", `,0.0`, []any{0.0}},
		{"float_hex_small", "%x", `,1e-320`, []any{1e-320}},
		{"float_hex_signed_width", "%+020.2x", `,-12.5`, []any{-12.5}},
		{"float_binary_precision", "%.3b", `,12.5`, []any{12.5}},
		{"indexed_width", "%[2]*[1]d", `,12,4`, []any{int64(12), int64(4)}},
		{"indexed_width_precision", "%[2]*.[1]*[3]f", `,2,8,1.25`, []any{int64(2), int64(8), 1.25}},
		{"index_resets_implicit", "%[2]d %d", `,11,22,33`, []any{int64(11), int64(22), int64(33)}},
		{"missing_index_end", "%[", `,1`, []any{int64(1)}},
		{"empty_index", "%[]s", `,"x"`, []any{"x"}},
		{"zero_index", "%[0]s", `,"x"`, []any{"x"}},
		{"huge_index", "%[999999999]s", `,"x"`, []any{"x"}},
		{"indexed_missing_verb", "%[2]", `,1,2`, []any{int64(1), int64(2)}},
		{"alternate_zero_hex", "%#08x", `,42`, []any{int64(42)}},
		{"plus_space_integer", "%+ d", `,42`, []any{int64(42)}},
		{"left_ignores_zero", "%-08d", `,42`, []any{int64(42)}},
		{"zero_integer_zero_precision", "%.0d", `,0`, []any{int64(0)}},
		{"zero_octal_zero_precision", "%#.0o", `,0`, []any{int64(0)}},
		{"alternate_modern_octal", "%#O", `,42`, []any{int64(42)}},
		{"ascii_quoted_string", "%+q", `,"é"`, []any{"é"}},
		{"alternate_spaced_upper_hex_string", "%# X", `,"AB"`, []any{"AB"}},
		{"single_rune_string_precision", "%.1s", `,"中文"`, []any{"中文"}},
		{"runtime_types", "%T/%T/%T/%T/%T/%T", `,1,1.2,true,"x",[1],{"a":1}`, []any{int64(1), 1.2, true, "x", []any{int64(1)}, map[string]any{"a": int64(1)}}},
		{"general_significant_large", "%.3g", `,12345.0`, []any{12345.0}},
		{"general_significant_small", "%.3g", `,0.0012345`, []any{0.0012345}},
		{"general_sharp_default", "%#g", `,1.2`, []any{1.2}},
		{"general_sharp_zero", "%#.3g", `,0.0`, []any{0.0}},
		{"general_sharp_upper_scientific", "%#.2G", `,12345.0`, []any{12345.0}},
		{"general_zero_width", "%+010.3g", `,12.5`, []any{12.5}},
		{"fixed_sharp_zero_precision", "%#.0f", `,12.5`, []any{12.5}},
		{"exponent_sharp_zero_precision", "%#.0e", `,12.5`, []any{12.5}},
		{"raw_quoted_string", "%#q", `,"plain text"`, []any{"plain text"}},
	} {
		cases = append(cases, jsonOracleCase{name: tc.name, source: fmt.Sprintf("strfmt(result,%q%s)", tc.format, tc.args), key: "result", want: fmt.Sprintf(tc.format, tc.values...)})
	}
	runJSONOracle(t, cases)
}
