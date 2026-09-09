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
	"os"
	"testing"
	"time"
)

func TestTimestampPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	cases := []struct {
		name, source string
		unit         int64
	}{
		{"upstream-default", `add_key(time1,timestamp()); add_key(after,true)`, 1},
		{"upstream-ns", `add_key(time1,timestamp("ns")); add_key(after,true)`, 1},
		{"upstream-us", `add_key(time1,timestamp("us")*1000); add_key(after,true)`, 1},
		{"upstream-ms", `add_key(time1,timestamp("ms")*1000000); add_key(after,true)`, 1},
		{"upstream-s", `add_key(time1,timestamp("s")*1000000000); add_key(after,true)`, 1},
		{"named-ms", `add_key(raw,timestamp(precision="ms")); add_key(time1,raw*1000000); add_key(after,true)`, 1},
		{"unknown-is-ns", `add_key(time1,timestamp("unknown")); add_key(after,true)`, 1},
		{"micro-sign-is-ns", `add_key(time1,timestamp("µs")); add_key(after,true)`, 1},
		{"non-string-is-ns", `add_key(time1,timestamp(7)); add_key(after,true)`, 1},
	}
	for _, size := range []int{1, 2, 4, 8, 10, 128} {
		t.Run("batch-"+itoa(size), func(t *testing.T) {
			for _, tc := range cases {
				check := runner.Check(tc.source)
				if check.Route != RouteJITNative || check.Capabilities.ExecutionMode != "native" {
					t.Fatalf("%s route=%+v", tc.name, check)
				}
				before := time.Now().UnixNano()
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"sequence": int64(i)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := runner.Process(tc.source, input)
				after := time.Now().UnixNano()
				if err != nil || len(batch.Records) != size {
					t.Fatalf("%s records=%d error=%v", tc.name, len(batch.Records), err)
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("%s record %d terminal=%d error=%s", tc.name, i, record.Status, record.Error)
					}
					applyHostCompatRecord(t, batch, i, &points[i])
					got, ok := points[i].Fields["time1"].(int64)
					if !ok || got < before-1_000_000_000 || got > after+1_000_000_000 {
						t.Fatalf("%s record %d time1=%#v outside [%d,%d]", tc.name, i, points[i].Fields["time1"], before, after)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost input/continuation: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}

	const sequenceSource = `add_key(t1,timestamp()); add_key(t2,timestamp()); add_key(t3,timestamp()); add_key(after,true)`
	check := runner.Check(sequenceSource)
	if check.Route != RouteJITNative {
		t.Fatalf("sequence route=%+v", check)
	}
	points := make([]Point, 128)
	for i := range points {
		points[i] = Point{Version: 1, Category: "logging", Measurement: "timestamp-sequence", Fields: map[string]any{"sequence": int64(i)}}
	}
	input, err := EncodeFlatPoints(points)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := runner.Process(sequenceSource, input)
	if err != nil || len(batch.Records) != len(points) {
		t.Fatalf("sequence records=%d error=%v", len(batch.Records), err)
	}
	before := time.Now().UnixNano()
	for i, record := range batch.Records {
		if record.Status != TerminalOK {
			t.Fatalf("sequence record %d terminal=%d error=%s", i, record.Status, record.Error)
		}
		applyHostCompatRecord(t, batch, i, &points[i])
		t1, ok1 := points[i].Fields["t1"].(int64)
		t2, ok2 := points[i].Fields["t2"].(int64)
		t3, ok3 := points[i].Fields["t3"].(int64)
		after := time.Now().UnixNano()
		if !ok1 || !ok2 || !ok3 || t1 < before-1_000_000_000 || t1 > after+1_000_000_000 ||
			t2 < before-1_000_000_000 || t2 > after+1_000_000_000 ||
			t3 < before-1_000_000_000 || t3 > after+1_000_000_000 {
			t.Fatalf("sequence record %d values=%#v window=[%d,%d]", i, points[i].Fields, before, after)
		}
	}
}
