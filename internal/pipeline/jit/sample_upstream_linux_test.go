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
)

func TestSamplePipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
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
		name      string
		source    string
		mustHello bool
		noHello   bool
	}{
		{"upstream-fifty-fifty", `if sample(0.5) { add_key(hello,"world") }; add_key(after,true)`, false, false},
		{"upstream-definite", `if sample(1) { add_key(hello,"world") }; add_key(after,true)`, true, false},
		{"upstream-expression", `if sample(2*0.1) { add_key(hello,"world") }; add_key(after,true)`, false, false},
		{"upstream-negative", `if sample(-0.5) { add_key(hello,"world") }; add_key(after,true)`, false, true},
		{"upstream-out-of-range", `if sample(2) { add_key(hello,"world") }; add_key(after,true)`, false, true},
		{"zero-probability-go-boundary", `if sample(0) { add_key(hello,"world") }; add_key(after,true)`, false, false},
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
					points[i] = Point{Version: 1, Category: "logging", Measurement: tc.name,
						Fields: map[string]any{"message": "dummy input", "sequence": int64(i)}}
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
					hello, present := points[i].Fields["hello"]
					if tc.mustHello && hello != "world" {
						t.Fatalf("%s record %d did not take definite branch: %#v", tc.name, i, points[i])
					}
					if tc.noHello && present {
						t.Fatalf("%s record %d took nil probability branch: %#v", tc.name, i, points[i])
					}
					if present && hello != "world" {
						t.Fatalf("%s record %d invalid branch value: %#v", tc.name, i, points[i])
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}
}
