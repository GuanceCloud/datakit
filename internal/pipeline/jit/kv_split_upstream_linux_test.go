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

func TestKVSplitPipelineGoUpstreamAndBoundaryCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 32)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	for _, source := range []string{
		`kv_split(_,field_split_pattern="[")`,
		`kv_split(_,value_split_pattern="(")`,
		`kv_split(_,trim_key=1)`,
		`kv_split(_,trim_value=true)`,
		`kv_split(_,include_keys="a")`,
	} {
		check := runner.Check(source)
		if check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
			t.Fatalf("invalid script admitted: source=%q check=%+v", source, check)
		}
	}

	cases := []struct {
		name     string
		source   string
		fields   map[string]any
		expected map[string]any
		absent   []string
	}{
		{"upstream-empty-include", `kv_split(_); add_key(after,true)`, map[string]any{"message": "a=1 b=2 c=3"}, nil, []string{"a", "b", "c"}},
		{"upstream-default", `kv_split(_,include_keys=["a","b","c"]); add_key(after,true)`, map[string]any{"message": "a=1, b=2 c=3"}, map[string]any{"a": "1,", "b": "2", "c": "3"}, nil},
		{"upstream-trim-value", `kv_split(_,trim_value=",",include_keys=["a","b","c"]); add_key(after,true)`, map[string]any{"message": "a=1, b=2 c=3"}, map[string]any{"a": "1", "b": "2", "c": "3"}, nil},
		{"upstream-trim-key", `kv_split(_,trim_value=",",trim_key="a",include_keys=["a","b","c"]); add_key(after,true)`, map[string]any{"message": "a=1, b=2 c=3"}, map[string]any{"b": "2", "c": "3"}, []string{"", "a"}},
		{"upstream-prefix", `kv_split(_,prefix="prefix_",trim_value=",",trim_key="a",include_keys=["a","b","c"]); add_key(after,true)`, map[string]any{"message": "a=1, b=2 c=3"}, map[string]any{"prefix_b": "2", "prefix_c": "3"}, []string{"", "prefix_"}},
		{"upstream-filter", `kv_split(_,include_keys=["b"],prefix="prefix_",trim_value=",",trim_key="a"); add_key(after,true)`, map[string]any{"message": "a=1, b=2 c=3"}, map[string]any{"prefix_b": "2"}, []string{"prefix_", "prefix_c"}},
		{"upstream-custom-regex", `kv_split(_,include_keys=["b"],field_split_pattern="\\+",value_split_pattern="::",prefix="prefix_",trim_value=",",trim_key="a"); add_key(after,true)`, map[string]any{"message": "a::1,+b::2+c::3"}, map[string]any{"prefix_b": "2"}, []string{"prefix_", "prefix_c"}},
		{"dynamic-include", `include=["a","c"]; kv_split(_,include_keys=include); add_key(after,true)`, map[string]any{"message": "a=1 b=2 c=3"}, map[string]any{"a": "1", "c": "3"}, []string{"b"}},
		{"dynamic-prefix", `prefix="dynamic_"; kv_split(_,include_keys=["a"],prefix=prefix); add_key(after,true)`, map[string]any{"message": "a=1 b=2"}, map[string]any{"dynamic_a": "1"}, []string{"a", "b"}},
		{"duplicate-last-wins", `kv_split(_,field_split_pattern=";",include_keys=["a"]); add_key(after,true)`, map[string]any{"message": "a=1;a=2;a=3"}, map[string]any{"a": "3"}, nil},
		{"empty-and-malformed-fields", `kv_split(_,field_split_pattern=";",include_keys=["a","b"]); add_key(after,true)`, map[string]any{"message": ";a=;bad;b=x=y;=z"}, map[string]any{"a": "", "b": "x=y"}, []string{""}},
		{"scalar-int-source", `kv_split(value,include_keys=["123"]); add_key(after,true)`, map[string]any{"value": int64(123)}, nil, []string{"123"}},
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
					for key, want := range tc.expected {
						if got := points[i].Fields[key]; got != want {
							t.Fatalf("%s record %d %s=%#v want %#v", tc.name, i, key, got, want)
						}
					}
					for _, key := range tc.absent {
						if _, ok := points[i].Fields[key]; ok {
							t.Fatalf("%s record %d unexpectedly set %q: %#v", tc.name, i, key, points[i])
						}
					}
					if points[i].Fields["after"] != true || points[i].Fields["sequence"] != int64(i) {
						t.Fatalf("%s record %d lost continuation/input: %#v", tc.name, i, points[i])
					}
				}
			}
		})
	}
}
