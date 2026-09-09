// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"fmt"
	"math"
	"runtime"
	"strconv"
	"testing"
)

func TestJITCastHexFloatOracle(t *testing.T) {
	inputs := []string{"0x1p0", "-0X1.8P+2", "0x_1_2.8p1_0", "0x.8p0", "0x1.p0",
		"0x1", "0x", "0x.p0", "0x1p", "0x1p+", "0x1p_0", "0x1_p0", "0x1._2p0", "0x__1p0", "0x1p0!",
		"0x1p1024", "-0x1p999999999999999999999", "0x1p-999999999999999999999", "-0x0p999999999999999999999",
		"0x1.fffffffffffffp1023", "0x1.fffffffffffff8p1023", "0x1.fffffffffffff7p1023",
		"0x1p-1074", "0x1p-1075", "0x1.0000000000000001p-1075", "0x1.fffffffffffffp-1023",
		"0x1.00000000000008p0", "0x1.00000000000008000000001p0", "0x1.00000000000018p0"}
	for exponent := -1100; exponent <= 1050; exponent += 17 {
		for _, mantissa := range []string{"1", "1.fffffffffffff8", "0.0000123456789abcdef0123456789"} {
			inputs = append(inputs, fmt.Sprintf("0x%sp%d", mantissa, exponent))
		}
	}
	var cases []jsonOracleCase
	for _, input := range inputs {
		want, err := strconv.ParseFloat(input, 64)
		if err != nil {
			want = 0
		}
		if math.IsInf(want, 0) || math.IsNaN(want) {
			t.Fatalf("unexpected oracle result for %q", input)
		}
		cases = append(cases, jsonOracleCase{name: input, source: `cast(message,"float")`, input: input, key: "message", want: want})
	}
	runJSONOracle(t, cases)
}

func TestJITCastFiniteRangeOracle(t *testing.T) {
	// Go's ordinary amd64 float64->int64 lowering uses CVTTSD2SQ;
	// out-of-range conversion produces the integer-indefinite value.
	// Other architectures need their own execution evidence.
	if runtime.GOARCH != "amd64" {
		t.Skip("amd64 conversion oracle")
	}
	runJSONOracle(t, []jsonOracleCase{
		{name: "max_integer_rounded", source: `cast(message,"int")`, input: "9223372036854775807", key: "message", want: int64(-9223372036854775808)},
		{name: "upper_boundary", source: `cast(message,"int")`, input: "9223372036854775808", key: "message", want: int64(-9223372036854775808)},
		{name: "positive_overflow", source: `cast(message,"int")`, input: "1e30", key: "message", want: int64(-9223372036854775808)},
		{name: "negative_overflow", source: `cast(message,"int")`, input: "-1e30", key: "message", want: int64(-9223372036854775808)},
		{name: "last_in_range", source: `cast(message,"int")`, input: "9223372036854774784", key: "message", want: int64(9223372036854774784)},
	})
}

func TestJITCastInvalidDecimalOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, input := range []string{"+NaN", "-NaN", " NaN", "NaN ", " inf", "inf!", "", " ", "1e9999", "-1e9999", "12junk", " 12", "12 "} {
		cases = append(cases, jsonOracleCase{name: input, source: `cast(message,"int")`, input: input, key: "message", want: int64(0)})
	}
	runJSONOracle(t, cases)
}

func TestJITCastDecimalUnderscoreOracle(t *testing.T) {
	var cases []jsonOracleCase
	for _, tc := range []struct {
		input string
		want  int64
	}{
		{"1_234", 1234}, {"-1_234", -1234}, {"1_2.5_0e+1_0", 125000000000},
		{"1_2e-1", 1}, {"_12", 0}, {"12_", 0}, {"1__2", 0},
		{"1_.2", 0}, {"1._2", 0}, {"1_e2", 0}, {"1e_2", 0}, {"1e+_2", 0},
	} {
		cases = append(cases, jsonOracleCase{name: tc.input, source: `cast(message,"int")`, input: tc.input, key: "message", want: tc.want})
	}
	cases = append(cases, jsonOracleCase{name: "float_fraction", source: `cast(message,"float")`, input: "1_2.5_0e-1", key: "message", want: float64(1.25)})
	runJSONOracle(t, cases)
}

func TestJITCastNonFiniteOracle(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("amd64 conversion oracle")
	}
	var cases []jsonOracleCase
	for _, input := range []string{"NaN", "Inf", "+Inf", "-Inf", "Infinity", "+Infinity", "-Infinity", "nan", "NAN", "inf", "+iNf", "-iNf", "infinity", "-INFINITY"} {
		cases = append(cases, jsonOracleCase{name: input, source: `cast(message,"int")`, input: input, key: "message", want: int64(-9223372036854775808)})
	}
	runJSONOracle(t, cases)
}
