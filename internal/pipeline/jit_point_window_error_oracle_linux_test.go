// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
)

func TestJITPointWindowRuntimeTypeErrorOracle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := pljit.NewRunner(path, "pipeline-go-1.4.3-datakit", 8)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	for _, tc := range []struct {
		name, source string
		fields       map[string]any
	}{
		{"string-before", `add_key(before,true); point_window(message,1); add_key(after,true)`, map[string]any{"message": "3"}},
		{"float-before", `add_key(before,true); point_window(message,1); add_key(after,true)`, map[string]any{"message": 3.0}},
		{"bool-after", `add_key(before,true); point_window(1,message); add_key(after,true)`, map[string]any{"message": true}},
		{"nil-after", `add_key(before,true); point_window(1,missing); add_key(after,true)`, map[string]any{"message": "x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script, err := NewPlScriptSimple(point.Logging, "window-type.p", tc.source)
			if err != nil {
				t.Fatal(err)
			}
			expected := newRealScriptPoint("window", tc.fields)
			wrapped := ptinput.PtWrap(point.Logging, expected)
			var goFailed bool
			func() {
				defer func() {
					if recover() != nil {
						// pipeline-go v1.4.3 asserts int64 before checking
						// the runtime type. The JIT deliberately turns that
						// process-level panic into a terminal record error.
						goFailed = true
					}
				}()
				goFailed = script.Run(wrapped, nil, nil) != nil
			}()
			if !goFailed {
				t.Fatal("pipeline-go unexpectedly accepted non-integer window size")
			}

			input, err := pljit.EncodeFlatPoints([]pljit.Point{{
				Version: 1, Category: "logging", Measurement: "window", Fields: tc.fields,
			}})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(tc.source, input)
			if err != nil {
				t.Fatal(err)
			}
			if len(batch.Records) != 1 || batch.Records[0].Status != pljit.TerminalError {
				t.Fatalf("native did not preserve terminal type error: %+v", batch)
			}
			actual := newRealScriptPoint("window", tc.fields)
			if batch.Static != nil {
				_, _, err = applyJITStatic(point.Logging, actual, batch.Static, 0, nil)
			} else {
				_, _, err = applyJITRecord(point.Logging, actual, batch.Records[0], 0, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"before", "after"} {
				if got, want := actual.Get(key), wrapped.Point().Get(key); got != want {
					t.Fatalf("%s: native=%#v pipeline-go=%#v", key, got, want)
				}
			}
		})
	}
}
