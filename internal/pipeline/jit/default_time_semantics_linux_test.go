// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"fmt"
	"os"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

func TestDefaultTimePositionAndSideEffects(t *testing.T) {
	runtimePath := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if runtimePath == "" {
		t.Skip("set PLATYPUS_JIT_RUNTIME to a trusted platypus_jit cdylib")
	}
	host, recorder := newObservedPipelineGoHost()
	runner, err := NewRunnerWithHost(runtimePath, "pipeline-go-1.4.3-datakit", 16, host)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	const pointTime = int64(1_700_000_000_000_000_123)
	const parsedTime = "2026-05-19T13:47:01.004+0800"
	wantTime, err := funcs.TimestampHandle(parsedTime, "Etc/UTC")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		source string
		fields map[string]any
		check  func(t *testing.T, point Point)
	}{
		{
			name:   "default-first",
			source: "default_time(time)\nadd_key(after, \"once\")\n",
			fields: map[string]any{"time": parsedTime, "foo": "value"},
			check: func(t *testing.T, p Point) {
				if p.Fields["time"] != wantTime || p.Fields["foo"] != "value" || p.Fields["after"] != "once" {
					t.Fatalf("unexpected first output: %#v", p)
				}
			},
		},
		{
			name:   "default-middle-json-regex",
			source: "add_key(before, \"once\")\ndefault_time(time)\nadd_key(after, \"once\")\n",
			fields: map[string]any{"time": parsedTime, "foo": "old"},
			check: func(t *testing.T, p Point) {
				if p.Fields["time"] != wantTime || p.Fields["before"] != "once" || p.Fields["after"] != "once" {
					t.Fatalf("unexpected middle output: %#v", p)
				}
			},
		},
		{
			name:   "default-last-drop-add",
			source: "add_key(before, \"once\")\ndrop_key(remove_me)\ndefault_time(time)\n",
			fields: map[string]any{"time": parsedTime, "remove_me": "gone"},
			check: func(t *testing.T, p Point) {
				if p.Fields["time"] != wantTime || p.Fields["before"] != "once" || p.Fields["remove_me"] != nil {
					t.Fatalf("unexpected last output: %#v", p)
				}
			},
		},
		{
			name:   "default-failure-preserves-point-time",
			source: "add_key(before, \"once\")\ndefault_time(time)\nadd_key(after, \"once\")\n",
			fields: map[string]any{"time": "not-a-time", "foo": "value"},
			check: func(t *testing.T, p Point) {
				_, timestampErr := funcs.TimestampHandle("not-a-time", "")
				wantMessage := fmt.Sprintf("time convert failed: %v", timestampErr)
				if p.Fields["time"] != pointTime || p.Fields["pl_msg"] != wantMessage ||
					p.Fields["before"] != "once" || p.Fields["after"] != "once" || p.Fields["foo"] != "value" {
					t.Fatalf("unexpected failed parse output: %#v", p)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder.reset()
			requireDefaultTimeNativeCapabilities(t, runner, test.source)
			input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: test.name, Fields: test.fields, TimeUnixNano: pointTime}})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.Process(test.source, input)
			if err != nil {
				t.Fatalf("default_time must compile at every script position: %v", err)
			}
			if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
				t.Fatalf("unexpected batch: %#v", batch)
			}
			requireDefaultTimeStaticBatch(t, batch)
			point := Point{Version: 1, Category: "logging", Measurement: test.name, Fields: cloneFields(test.fields), TimeUnixNano: pointTime}
			applyHostCompatRecord(t, batch, 0, &point)
			test.check(t, point)
			requireHostCompatCalls(t, recorder, hostCompatTimestampOp)
		})
	}
}

func cloneFields(fields map[string]any) map[string]any {
	clone := make(map[string]any, len(fields))
	for key, value := range fields {
		clone[key] = value
	}
	return clone
}
