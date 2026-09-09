// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Adapted from pipeline-go v1.4.3 ptinput/funcs/fn_parse_int_test.go.
// Copyright 2021-present Guance, Inc. MIT License; see testdata/PIPELINE_GO_TEST_LICENSE.txt.
package pipeline

import (
	"fmt"
	"strconv"
	"testing"
)

func TestJITUpstreamParseIntOracle(t *testing.T) {
	scripts := []string{
		`a = 17
b = format_int(a, 16)
if b != "11" { add_key(abc, b); exit() }
c = parse_int(b, 16)
if c != a { add_key(abc, c); exit() } else { add_key(abc, "ok") }`,
		`a = "11" # 0x11
b = parse_int(a, 16)
if b != 17 { add_key(abc, b); exit() }
c = format_int(b, 16)
if c != a { add_key(abc, c); exit() } else { add_key(abc, "ok") }`,
		`a = 7665324064912355185
b = format_int(a, 16)
if b != "6a60b39fd95aaf71" { add_key(abc, b); exit() }
c = parse_int(b, 16)
if c != a { add_key(abc, c); exit() } else { add_key(abc, "ok") }`,
		`a = "7665324064912355185"
b = format_int(parse_int(a, 10), 16)
if b != "6a60b39fd95aaf71" { add_key(abc, b); exit() } else { add_key(abc, "ok") }`,
		`a = "6a60b39fd95aaf71"
b = parse_int(a, 16)
if b != 7665324064912355185 { add_key(abc, b); exit() }
c = format_int(b, 16)
if c != a { add_key(abc, c) } else { add_key(abc, "ok") }`,
		`a = "0x6a60b39fd95aaf71"
b = parse_int(a, 0)
if b != 7665324064912355185 { add_key(abc, b); exit() }
c = format_int(b, 16)
if "0x"+c != a { add_key(abc, c) } else { add_key(abc, "ok") }`,
	}
	var cases []jsonOracleCase
	for i, script := range scripts {
		cases = append(cases, jsonOracleCase{name: fmt.Sprint(i), source: script, input: "test", key: "abc", want: "ok"})
	}
	runJSONOracle(t, cases)
}

func TestJITParseIntBoundaryOracle(t *testing.T) {
	var cases []jsonOracleCase
	for i, tc := range []struct {
		text string
		base int
	}{
		{"9223372036854775807", 10}, {"9223372036854775808", 10},
		{"-9223372036854775808", 10}, {"-9223372036854775809", 10},
		{"0b1010", 0}, {"0o17", 0}, {"077", 0}, {"08", 0},
		{"0x_ff", 0}, {"1_000", 0}, {"1_000", 10}, {"1__0", 0},
		{" 12", 10}, {"12 ", 10}, {"+12", 10}, {"-0", 10},
		{"z", 36}, {"11", 1}, {"11", 37}, {"", 10},
		{"999999999999999999999999x", 10},
		{"-999999999999999999999999x", 10}, {"9223372036854775808x", 10},
		{"18446744073709551615x", 10}, {"18446744073709551616x", 10},
		{"_12", 0}, {"12_", 0}, {"0x__ff", 0}, {"0x_", 0},
		{"0_7", 0}, {"0b_10", 0}, {"999999999999999999999__", 0},
	} {
		want, _ := strconv.ParseInt(tc.text, tc.base, 64)
		cases = append(cases, jsonOracleCase{name: fmt.Sprint(i), source: fmt.Sprintf(`v = parse_int(%q, %d); add_key(result, v)`, tc.text, tc.base), key: "result", want: want})
	}
	runJSONOracle(t, cases)
}

func TestJITIntegerArgumentEvaluationOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "parse-named-skips-error", source: `v = parse_int(base=1/divisor, val=123); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-named-skips-error", source: `v = format_int(base=1/divisor, val="123"); add_key(result, v)`, key: "result", want: ""},
		{name: "parse-top-level-skips-error", source: `parse_int(123, 1/divisor); add_key(result, 0)`, key: "result", want: int64(0)},
		{name: "format-top-level-skips-error", source: `format_int("123", 1/divisor); add_key(result, "")`, key: "result", want: ""},
		{name: "parse-wrong-type-skips-supported-effect", source: `v = parse_int(123, pt_kvs_set("unexpected", true)); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-wrong-type-skips-supported-effect", source: `v = format_int("123", pt_kvs_set("unexpected", true)); add_key(result, v)`, key: "result", want: ""},
		{name: "parse-valid-evaluates-supported-effect", source: `v = parse_int("123", pt_kvs_set("expected", true)); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-valid-evaluates-supported-effect", source: `v = format_int(123, pt_kvs_set("expected", true)); add_key(result, v)`, key: "result", want: ""},
		{name: "parse-wrong-type-skips-effect", source: `v = parse_int(123, add_key(unexpected, true)); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-wrong-type-skips-effect", source: `v = format_int("123", add_key(unexpected, true)); add_key(result, v)`, key: "result", want: ""},
		{name: "parse-missing-skips-error", source: `v = parse_int(missing, 1/divisor); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-missing-skips-error", source: `v = format_int(missing, 1/divisor); add_key(result, v)`, key: "result", want: ""},
		{name: "parse-valid-evaluates-effect", source: `v = parse_int("123", add_key(expected, true)); add_key(result, v)`, key: "result", want: int64(0)},
		{name: "format-valid-evaluates-effect", source: `v = format_int(123, add_key(expected, true)); add_key(result, v)`, key: "result", want: ""},
	})
}

func TestJITNestedAddKeyCompositionOracle(t *testing.T) {
	runJSONOracle(t, []jsonOracleCase{
		{name: "one-argument-local", source: `local = "kept"; v = add_key(local); add_key(result, value_type(v))`, key: "local", want: "kept"},
		{name: "container", source: `v = add_key(result, [1, "x", true]); add_key(return_type, value_type(v))`, key: "result", want: `[1,"x",true]`},
		{name: "origin", source: `v = add_key(_, "replaced"); add_key(return_type, value_type(v))`, key: "message", want: "replaced"},
		{name: "void-attribute", source: `obj = {"x": 123}; v = add_key(result, obj.x); add_key(return_type, value_type(v))`, key: "result", want: nil},
		{name: "assignment-value", source: `v = add_key(result, local="assigned"); add_key(local_copy, local)`, key: "result", want: "assigned"},
		{name: "short-circuit", source: `v = false && add_key(unexpected, true); add_key(result, v)`, key: "result", want: false},
	})
}
