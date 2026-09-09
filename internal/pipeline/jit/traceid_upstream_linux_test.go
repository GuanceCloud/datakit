// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"strconv"
	"testing"
)

func goTraceIDOracle(input string) string {
	original := input
	if len(input) == 32 {
		input = input[16:]
	} else if len(input) != 16 {
		return original
	}
	decoded, err := hex.DecodeString(input)
	if err != nil || len(decoded) != 8 {
		return original
	}
	return strconv.FormatUint(binary.BigEndian.Uint64(decoded), 10)
}

func TestConvTraceIDPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	upstream := []struct{ input, want string }{
		{"18962fdd9eea517f2ae0771ea69d6e16", "3089600317904219670"},
		{"2f7cdb2b45447bcb43820ddf56d9a654", "4864465800399529556"},
		{"43820ddf56d9a654", "4864465800399529556"},
		{"03820ddf56d9a654", "252779781972141652"},
		{"0f7cdb2b45447bcb43820ddf56d9a654", "4864465800399529556"},
		{"10f7cdb2b45447bcb43820ddf56d9a654", "10f7cdb2b45447bcb43820ddf56d9a654"},
	}
	for _, tc := range upstream {
		if got := goTraceIDOracle(tc.input); got != tc.want {
			t.Fatalf("bad copied upstream fixture %q: got=%q want=%q", tc.input, got, tc.want)
		}
	}
	const composed = `grok(_, "%{NOTSPACE:trace_id}"); conv_traceid_w3c_to_dd(trace_id); add_key(after,true)`
	check := runner.Check(composed)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("composed route=%+v", check)
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("upstream-batch-"+itoa(size), func(t *testing.T) {
			points := make([]Point, size)
			for i := range points {
				tc := upstream[i%len(upstream)]
				points[i] = Point{Version: 1, Category: "logging", Measurement: "upstream-trace", Fields: map[string]any{"message": tc.input, "sequence": int64(i)}}
			}
			assertTraceIDBatch(t, runner, composed, points, func(i int) any { return upstream[i%len(upstream)].want })
		})
	}

	boundaries := []any{
		"0000000000000000", "000000000000000f", "ffffffffffffffff",
		"FFFFFFFFFFFFFFFF", "ffffffffffffffff000000000000000f",
		"00000000000000fg", "00000000000000f", "0000000000000000f",
		"ééééééééa", "", " 000000000000000f", int64(15), nil,
	}
	const direct = `conv_traceid_w3c_to_dd(trace_id); add_key(after,true)`
	check = runner.Check(direct)
	if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
		t.Fatalf("direct route=%+v", check)
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("boundary-batch-"+itoa(size), func(t *testing.T) {
			points := make([]Point, size)
			for i := range points {
				value := boundaries[i%len(boundaries)]
				fields := map[string]any{"sequence": int64(i)}
				if value != nil {
					fields["trace_id"] = value
				}
				points[i] = Point{Version: 1, Category: "logging", Measurement: "boundary-trace", Fields: fields}
			}
			assertTraceIDBatch(t, runner, direct, points, func(i int) any {
				value := boundaries[i%len(boundaries)]
				if text, ok := value.(string); ok {
					return goTraceIDOracle(text)
				}
				return value
			})
		})
	}
}

func assertTraceIDBatch(t *testing.T, runner *Runner, source string, points []Point, want func(int) any) {
	t.Helper()
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(source, input)
	if err != nil || len(batch.Records) != len(points) {
		t.Fatalf("records=%d error=%v", len(batch.Records), err)
	}
	for i, record := range batch.Records {
		if record.Status != TerminalOK {
			t.Fatalf("record %d terminal=%d error=%s", i, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, i, &points[i])
		if points[i].Fields["trace_id"] != want(i) || points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
			t.Fatalf("record %d output=%#v want trace_id=%#v", i, points[i], want(i))
		}
	}
}
