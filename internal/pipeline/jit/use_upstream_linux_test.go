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

func TestUsePipelineGoUpstreamGraphAndExecutionCorpus(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	validSources := map[string]string{
		"a.p": `if true { use("b.p") }`,
		"b.p": `add_key(b, 1)`,
	}
	for _, entry := range []string{"a.p", "b.p"} {
		snapshot, err := NewModuleSnapshot(entry, validSources)
		if err != nil {
			t.Fatal(err)
		}
		check := runner.CheckModules(snapshot)
		if check.Route != RouteJITNative || check.Capabilities.Backend != ExecutionBackendMachineCode ||
			check.Capabilities.RequiredHostFlags != 0 || len(check.Capabilities.HostCalls) != 0 {
			t.Fatalf("entry %s route=%+v", entry, check)
		}
		for _, size := range []int{1, 2, 4, 8, 10, 128} {
			points := make([]Point, size)
			for i := range points {
				points[i] = Point{Version: 1, Category: "network", Measurement: "default",
					Tags: map[string]string{"ax": "1"}, Fields: map[string]any{"sequence": int64(i)}}
			}
			input, err := EncodeFlatPoints(points)
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.ProcessModules(snapshot, input)
			if err != nil || len(batch.Records) != size {
				t.Fatalf("entry %s batch %d records=%d error=%v", entry, size, len(batch.Records), err)
			}
			for i := range batch.Records {
				if batch.Records[i].Status != TerminalOK {
					t.Fatalf("entry %s batch %d record %d=%+v", entry, size, i, batch.Records[i])
				}
				applyHostCompatRecord(t, batch, i, &points[i])
				if points[i].Category != "network" || points[i].Measurement != "default" ||
					points[i].Tags["ax"] != "1" || points[i].Fields["b"] != int64(1) ||
					points[i].Fields["sequence"] != int64(i) || points[i].Dropped {
					t.Fatalf("entry %s batch %d record %d=%#v", entry, size, i, points[i])
				}
			}
		}
	}

	invalid := []struct {
		name, entry string
		sources     map[string]string
	}{
		{"cycle", "d.p", map[string]string{
			"a.p": `if true { use("b.p") }`, "b.p": `add_key(b, 1)`,
			"d.p": `use("c.p")`, "c.p": `use("a.p"); use("d.p"); use("fcName.p")`,
		}},
		{"missing", "main.p", map[string]string{"main.p": `use("missing.p")`}},
		{"dynamic", "main.p", map[string]string{"main.p": `use(script_name)`}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := NewModuleSnapshot(tc.entry, tc.sources)
			if err != nil {
				t.Fatal(err)
			}
			check := runner.CheckModules(snapshot)
			if check.Route != RoutePipelineGo || check.Reason != CheckReasonCompileError {
				t.Fatalf("route=%+v", check)
			}
		})
	}
}

func TestUsePipelineGoNestedExitErrorAndLocalIsolation(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 16)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()

	for _, tc := range []struct {
		name, child string
		terminal    TerminalStatus
		wantChild   int64
		wantParent  bool
		wantAfter   bool
	}{
		{"normal", `local=9; add_key(child,local)`, TerminalOK, 9, true, true},
		{"exit", `add_key(child,1); exit(); add_key(unreachable,true)`, TerminalOK, 1, true, true},
		{"error", `add_key(child,2); value=1/divisor; add_key(unreachable,true)`, TerminalError, 2, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, err := NewModuleSnapshot("main.p", map[string]string{
				"main.p":  `local=7; use("child.p"); add_key(parent_local,local); add_key(after,true)`,
				"child.p": tc.child,
			})
			if err != nil {
				t.Fatal(err)
			}
			if check := runner.CheckModules(snapshot); check.Route != RouteJITNative {
				t.Fatalf("route=%+v", check)
			}
			point := Point{Version: 1, Category: "logging", Measurement: "use", Fields: map[string]any{"divisor": int64(0)}}
			input, err := EncodeFlatPoints([]Point{point})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := runner.ProcessModules(snapshot, input)
			if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != tc.terminal {
				t.Fatalf("batch=%+v error=%v", batch, err)
			}
			applyHostCompatRecord(t, batch, 0, &point)
			if point.Fields["child"] != tc.wantChild || (point.Fields["parent_local"] == int64(7)) != tc.wantParent {
				t.Fatalf("local/call output=%#v", point)
			}
			if (point.Fields["after"] == true) != tc.wantAfter {
				t.Fatalf("continuation output=%#v", point)
			}
			if _, ok := point.Fields["unreachable"]; ok {
				t.Fatalf("child continued after exit/error: %#v", point)
			}
		})
	}
}
