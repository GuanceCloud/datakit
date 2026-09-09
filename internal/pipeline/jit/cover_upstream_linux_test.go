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

func TestCoverPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	// These scripts are structurally invalid. A variable range is intentionally
	// accepted by the Rust language extension and validated at runtime.
	for _, source := range []string{
		"json(_,`str`); cover(`str`)",
		"json(_,`str`); cover(`str`,[\"刘\",\"波\"])",
		"cover(`str`,[1])",
		"cover(`str`,[1,2,3])",
	} {
		check := runner.Check(source)
		if check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
			t.Fatalf("invalid script admitted: source=%q check=%+v", source, check)
		}
	}

	cases := []struct {
		name     string
		source   string
		message  string
		fields   map[string]any
		key      string
		expected any
	}{
		{"upstream-8-13-clamped", "json(_,`str`); cover(`str`,[8,13]); add_key(after,true)", `{"str":"13838130517"}`, nil, "str", "1383813****"},
		{"upstream-8-11", "json(_,`str`); cover(`str`,[8,11]); add_key(after,true)", `{"str":"13838130517"}`, nil, "str", "1383813****"},
		{"upstream-2-4", "json(_,`str`); cover(`str`,[2,4]); add_key(after,true)", `{"str":"13838130517"}`, nil, "str", "1***8130517"},
		{"upstream-1-1", "json(_,`str`); cover(`str`,[1,1]); add_key(after,true)", `{"str":"13838130517"}`, nil, "str", "*3838130517"},
		{"upstream-han", "json(_,`str`); cover(`str`,[1,100]); add_key(after,true)", `{"str":"刘少波"}`, nil, "str", "＊＊＊"},
		{"upstream-start-after-end", "json(_,`str`); cover(`str`,[3,2]); add_key(after,true)", `{"str":"刘少波"}`, nil, "str", "刘少波"},
		{"upstream-number-noop", "json(_,`str`); cover(`str`,[1,2]); add_key(after,true)", `{"str":123456}`, nil, "str", float64(123456)},
		{"upstream-negative-number-noop", "json(_,`str`); cover(`str`,[-1,-2]); add_key(after,true)", `{"str":123456}`, nil, "str", float64(123456)},
		{"upstream-cast-number-noop", "json(_,`str`); cast(`str`,\"int\"); cover(`str`,[-2,10000]); add_key(after,true)", `{"str":123456}`, nil, "str", int64(123456)},
		{"missing-noop", "cover(`str`,[1,2]); add_key(after,true)", `{}`, nil, "str", nil},
		{"negative-last-two", "cover(`str`,[-2,-1]); add_key(after,true)", `{}`, map[string]any{"str": "abcdef"}, "str", "abcd**"},
		{"zero-start", "cover(`str`,[0,2]); add_key(after,true)", `{}`, map[string]any{"str": "abcdef"}, "str", "abcdef"},
		{"float-truncation", "cover(`str`,[2.9,4.9]); add_key(after,true)", `{}`, map[string]any{"str": "abcdef"}, "str", "a***ef"},
		{"mixed-han-ascii", "cover(`str`,[1,4]); add_key(after,true)", `{}`, map[string]any{"str": "刘A𠀀e\u0301"}, "str", "＊*＊*́"},
		{"attribute-path", `cover(a.secret,[2,3]); add_key(after,true)`, `{}`, map[string]any{"a.secret": "甲b乙d"}, "a.secret", "甲*＊d"},
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
					fields := map[string]any{"message": tc.message, "sequence": int64(i)}
					for key, value := range tc.fields {
						fields[key] = value
					}
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
					if tc.expected == nil {
						if _, ok := points[i].Fields[tc.key]; ok {
							t.Fatalf("%s record %d unexpectedly created %s: %#v", tc.name, i, tc.key, points[i])
						}
					} else if got := points[i].Fields[tc.key]; got != tc.expected {
						t.Fatalf("%s record %d %s=%#v want %#v", tc.name, i, tc.key, got, tc.expected)
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}
}
