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
	"math"
	"os"
	"testing"
)

func TestDurationPrecisionPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 24)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name     string
		source   string
		fields   map[string]any
		key      string
		expected int64
	}{
		{"upstream-ms-ns", `json(_,ts); cast(ts,"int"); duration_precision(ts,"ms","ns"); add_key(after,true)`, map[string]any{"message": `{"ts":12345}`}, "ts", 12_345_000_000},
		{"upstream-ms-s", `json(_,ts); cast(ts,"int"); duration_precision(ts,"ms","s"); add_key(after,true)`, map[string]any{"message": `{"ts":12345000}`}, "ts", 12_345},
		{"upstream-s-s", `json(_,ts); cast(ts,"int"); duration_precision(ts,"s","s"); add_key(after,true)`, map[string]any{"message": `{"ts":12345000}`}, "ts", 12_345_000},
		{"upstream-ns-us", `json(_,ts); cast(ts,"int"); duration_precision(ts,"ns","us"); add_key(after,true)`, map[string]any{"message": `{"ts":12345000}`}, "ts", 12_345},
		{"case-insensitive", `duration_precision(ts,"MS","NS"); add_key(after,true)`, map[string]any{"ts": int64(7)}, "ts", 7_000_000},
		{"unknown-old-noop", `duration_precision(ts,"bad","ns"); add_key(after,true)`, map[string]any{"ts": int64(7)}, "ts", 7},
		{"unknown-new-noop", `duration_precision(ts,"ns","bad"); add_key(after,true)`, map[string]any{"ts": int64(7)}, "ts", 7},
		{"micro-sign-noop", `duration_precision(ts,"µs","ns"); add_key(after,true)`, map[string]any{"ts": int64(7)}, "ts", 7},
		{"negative-truncates-zero", `duration_precision(ts,"ns","s"); add_key(after,true)`, map[string]any{"ts": int64(-1_999_999_999)}, "ts", -1},
		{"multiply-wraps", `duration_precision(ts,"s","ns"); add_key(after,true)`, map[string]any{"ts": int64(math.MaxInt64)}, "ts", -1_000_000_000},
		{"attribute-path", `duration_precision(a.value,"ms","us"); add_key(after,true)`, map[string]any{"a.value": int64(123)}, "a.value", 123_000},
	}

	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" ||
					check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				points := make([]Point, size)
				for i := range points {
					fields := cloneFields(tc.fields)
					fields["sequence"] = int64(i)
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name, Fields: fields}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					if got := points[i].Fields[tc.key]; got != tc.expected {
						t.Fatalf("%s record %d %s=%#v want %d", tc.name, i, tc.key, got, tc.expected)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}

	for _, tc := range []struct {
		name   string
		fields map[string]any
	}{
		{"missing", map[string]any{}},
		{"string", map[string]any{"ts": "7"}},
		{"float", map[string]any{"ts": float64(7)}},
		{"bool", map[string]any{"ts": true}},
	} {
		t.Run("terminal-"+tc.name, func(t *testing.T) {
			const source = `add_key(before,true); duration_precision(ts,"ms","ns"); add_key(after,true)`
			point := Point{Version: 1, Category: "logging", Measurement: tc.name, Fields: cloneFields(tc.fields)}
			input, err := EncodeFlatPoints([]Point{point})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(source, input)
			if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalError || !batch.Records[0].CommitPrefixError {
				t.Fatalf("batch=%+v error=%v", batch, err)
			}
			applyHostCompatRecord(t, batch, 0, &point)
			if point.Fields["before"] != true {
				t.Fatalf("prefix not committed: %#v", point)
			}
			if _, ok := point.Fields["after"]; ok {
				t.Fatalf("terminal error continued: %#v", point)
			}
		})
	}
}
