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
	"reflect"
	"testing"
)

func TestPointWindowPipelineGoUpstreamCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	cases := []struct {
		name, source string
		input, want  []string
		filepaths    []string
	}{
		{"one", `point_window(3,3); drop(); if _=="4" { window_hit() }`,
			[]string{"1", "2", "3", "4", "5", "6"}, []string{"1", "2", "3", "4", "5", "6"}, nil},
		{"duplicate-registration", `point_window(3,3); point_window(2,2); if _=="4" { window_hit() }; if _!="6" { drop() }`,
			[]string{"1", "2", "3", "xx", "4", "5", "6", "7"}, []string{"1", "2", "3", "4", "5"}, []string{"b", "b", "b", "xx", "b", "b", "b", "b"}},
		{"explicit-stream-tags", `point_window(3,3,["host","filepath"]); if _=="4" { window_hit() }; if _!="3" { drop() }`,
			[]string{"1", "2", "3", "4", "5", "6", "7"}, []string{"1", "2", "4", "5", "6"}, nil},
		{"second-hit-after-drop", `point_window(3,3); if _=="4" { window_hit() }; if _!="6" { drop() }; if _=="7" { window_hit() }`,
			[]string{"1", "2", "3", "4", "5", "6", "7"}, []string{"1", "2", "3", "4", "5", "7"}, nil},
		{"ring-overwrite-order", `point_window(3,3); drop(); if _=="4" { window_hit() }`,
			[]string{"0", "1", "2", "3", "4", "5", "6"}, []string{"3", "1", "2", "4", "5", "6"}, nil},
		{"nested-value-calls", `add_key(registered,point_window(3,3)==nil); drop(); if _=="4" { add_key(hit,window_hit()==nil) }`,
			[]string{"1", "2", "3", "4", "5", "6"}, []string{"1", "2", "3", "4", "5", "6"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
			if err != nil {
				t.Fatal(err)
			}
			defer runner.Close()
			check := runner.Check(tc.source)
			if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
				check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
				t.Fatalf("route=%+v", check)
			}
			points := make([]Point, len(tc.input))
			for i, value := range tc.input {
				filepath := "b"
				if tc.filepaths != nil {
					filepath = tc.filepaths[i]
				}
				points[i] = Point{Version: 1, Category: "logging", Measurement: "test",
					Tags: map[string]string{"host": "a", "filepath": filepath}, Fields: map[string]any{"message": value}}
			}
			input, err := EncodeFlatPoints(points)
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(tc.source, input)
			if err != nil || len(batch.Records) != len(points) {
				t.Fatalf("records=%d error=%v", len(batch.Records), err)
			}
			var got []string
			for i, record := range batch.Records {
				if record.Status != TerminalOK && record.Status != TerminalDropped {
					t.Fatalf("record %d=%+v", i, record)
				}
				for childIndex, payload := range record.Emitted {
					var point struct {
						Fields map[string]any `json:"fields"`
					}
					if err := decodeEmittedTestPayload(payload, &point); err != nil {
						t.Fatalf("record %d emitted %d: %v", i, childIndex, err)
					}
					value, ok := point.Fields["message"].(string)
					if !ok {
						t.Fatalf("record %d emitted %d=%s", i, childIndex, payload)
					}
					got = append(got, value)
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("emitted=%v want=%v", got, tc.want)
			}
		})
	}
}
