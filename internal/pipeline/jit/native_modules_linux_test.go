// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNativeNestedCreatePointAfterUse(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	program, err := rt.(ModuleRuntime).CompileModules("main.p", map[string]string{
		"main.p":  `add_key(returned,create_point("child",{}, {"value":1},after_use="child.p")==nil); add_key(parent,true)`,
		"child.p": `add_key(child_processed,true)`,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer program.Close()
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "parent"}})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := program.ProcessIndexed(input)
	if err != nil || len(batch.Records) != 1 {
		t.Fatalf("process nested create_point: records=%d err=%v", len(batch.Records), err)
	}
	record := batch.Records[0]
	if record.Status != TerminalOK || len(record.Emitted) != 0 {
		t.Fatalf("nested create_point record=%+v", record)
	}
	point := Point{Version: 1, Category: "logging", Measurement: "parent"}
	deltas, err := record.MutationDeltas()
	if err != nil {
		t.Fatal(err)
	}
	for _, delta := range deltas {
		if err := delta.Apply(&point); err != nil {
			t.Fatal(err)
		}
	}
	if point.Fields["returned"] != true || point.Fields["parent"] != true || len(point.Subpoints) != 1 {
		t.Fatalf("nested parent=%#v", point)
	}
	child := point.Subpoints[0]
	if child.Measurement != "child" || child.Fields["value"] != int64(1) || child.Fields["child_processed"] != true {
		t.Fatalf("nested child=%#v", child)
	}
}

func TestRunnerCloseWaitsForCancellationWatcher(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := NewModuleSnapshot("main.p", map[string]string{"main.p": "for ; repeat; {}\nadd_key(done, true)"})
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.PrepareModules(snapshot); err != nil {
		t.Fatal(err)
	}
	native := runner.runtime.(*nativeRuntime)
	entered, cancelling, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); unblock(); runner.Close() }()
	originalProcess, originalCancel := native.processCancellable, native.cancelRequest
	if original := native.staticCancellableInto; original != nil {
		native.staticCancellableInto = func(h uintptr, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
			close(entered)
			return original(h, in, l, o, b, tok, s, cap)
		}
	}
	if originalInto := native.indexedInto; originalInto != nil {
		native.indexedInto = func(h uintptr, c uint32, in *byte, l uintptr, o *processOptions, b *nativeBuffer, t uintptr, s *byte, cap uintptr) int32 {
			close(entered)
			return originalInto(h, c, in, l, o, b, t, s, cap)
		}
	}
	native.processCancellable = func(handle uintptr, codec uint32, data *byte, size uintptr, options *processOptions, out *nativeBuffer, token uintptr) int32 {
		close(entered)
		return originalProcess(handle, codec, data, size, options, out, token)
	}
	native.cancelRequest = func(token uintptr) int32 {
		close(cancelling)
		<-release
		return originalCancel(token)
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "close", Fields: map[string]any{"repeat": true}}})
	if err != nil {
		t.Fatal(err)
	}
	processed := make(chan struct{})
	go func() { defer close(processed); runner.ProcessModulesContext(ctx, snapshot, input) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not enter ABI")
	}
	cancel()
	select {
	case <-cancelling:
	case <-time.After(3 * time.Second):
		t.Fatal("watcher did not cancel")
	}
	closed, closeStarted := make(chan error, 1), make(chan struct{})
	go func() { close(closeStarted); closed <- runner.Close() }()
	<-closeStarted
	select {
	case err := <-closed:
		t.Fatalf("runner closed while cancel callback active: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	select {
	case <-processed:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not finish")
	}
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runner close did not finish")
	}
}

func TestNativeContextCancellationAndReuse(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	p, err := rt.(ModuleRuntime).CompileModules("main.p", map[string]string{
		"main.p":  `create_point("child", {}, {"repeat": repeat}, after_use="child.p"); add_key(after, true)`,
		"child.p": "for ; repeat; {}\nadd_key(done, true)",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	native := rt.(*nativeRuntime)
	if native.processCancellable == nil {
		t.Fatal("missing cancellation ABI")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	original := native.processCancellable
	originalInto := native.staticCancellableInto
	if originalInto != nil {
		native.staticCancellableInto = func(h uintptr, in *byte, l uintptr, o *processOptions, b *nativeBuffer, t uintptr, s *byte, cap uintptr) int32 {
			cancel()
			return originalInto(h, in, l, o, b, t, s, cap)
		}
	}
	// Trigger only after Go has created the token and installed its watcher.
	// The actual token cancellation and native execution still cross the ABI.
	native.processCancellable = func(handle uintptr, codec uint32, data *byte, size uintptr, options *processOptions, out *nativeBuffer, token uintptr) int32 {
		cancel()
		return original(handle, codec, data, size, options, out, token)
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "cancel", Fields: map[string]any{"repeat": true}}})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := p.(ContextProgram).ProcessIndexedContext(ctx, input)
	native.processCancellable = original
	native.staticCancellableInto = originalInto
	cancelled := err != nil && strings.Contains(err.Error(), "E_CANCELLED")
	for _, record := range batch.Records {
		cancelled = cancelled || strings.Contains(string(record.Error), "E_CANCELLED")
	}
	if !cancelled {
		t.Fatalf("expected native cancellation, got batch=%+v err=%v", batch, err)
	}
	input, err = EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "next", Fields: map[string]any{"repeat": false}}})
	if err != nil {
		t.Fatal(err)
	}
	next, stop := context.WithCancel(context.Background())
	defer stop()
	batch, err = p.(ContextProgram).ProcessIndexedContext(next, input)
	if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
		t.Fatalf("cancel contaminated next request: %+v %v", batch, err)
	}
}

func TestNativeModuleSnapshotExecution(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	modules := rt.(ModuleRuntime)
	for _, tc := range []struct {
		name, child string
		fail        bool
		want        int64
	}{
		{"normal", `local = 9; add_key(child, local)`, false, 9},
		{"exit", `add_key(child, 1); exit(); add_key(unreachable, true)`, false, 1},
		{"error", `add_key(child, 2); value = 1 / divisor; add_key(unreachable, true)`, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			program, err := modules.CompileModules("main.p", map[string]string{
				"main.p":  `local = 7; use("child.p"); add_key(parent_local, local); add_key(after, true)`,
				"child.p": tc.child,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer program.Close()
			for _, size := range []int{1, 8, 128} {
				points := make([]Point, size)
				for i := range points {
					points[i] = Point{Version: 1, Category: "logging", Measurement: "modules", Fields: map[string]any{"divisor": int64(0)}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := program.ProcessIndexed(input)
				if err != nil {
					t.Fatal(err)
				}
				if len(batch.Records) != size {
					t.Fatalf("record count: %d", len(batch.Records))
				}
				for i, record := range batch.Records {
					if tc.fail {
						if record.Status == TerminalOK || !record.CommitPrefixError {
							t.Fatalf("missing committed-prefix error: %+v", record)
						}
					} else if record.Status != TerminalOK {
						t.Fatalf("failed: %s", record.Error)
					}
					deltas, err := record.MutationDeltas()
					if err != nil {
						t.Fatal(err)
					}
					for _, delta := range deltas {
						if err := delta.Apply(&points[i]); err != nil {
							t.Fatal(err)
						}
					}
					fields := points[i].Fields
					if fields["child"] != tc.want || fields["unreachable"] != nil {
						t.Fatalf("fields: %#v", fields)
					}
					if tc.fail {
						if fields["after"] != nil {
							t.Fatal("parent continued")
						}
					} else if fields["parent_local"] != int64(7) || fields["after"] != true {
						t.Fatalf("parent state: %#v", fields)
					}
				}
			}
		})
	}
}

func TestModuleCompileMissingSymbolIsExplicit(t *testing.T) {
	native := &nativeRuntime{}
	if _, err := native.CompileModules("main.p", map[string]string{"main.p": "exit()"}); err == nil {
		t.Fatal("missing optional symbol accepted")
	}
}

func TestRunnerModuleNamespaceStateIsolation(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	sources := map[string]string{"main.p": `use("child.p")`, "child.p": `if seed { cache_set("key", message) }; previous = cache_get("key"); add_key(previous, previous)`}
	first, err := NewScopedModuleSnapshot("local", "logging", "main.p", sources)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewScopedModuleSnapshot("remote", "logging", "main.p", sources)
	if err != nil {
		t.Fatal(err)
	}
	for _, snapshot := range []*ModuleSnapshot{first, second} {
		if check := runner.CheckModules(snapshot); check.Route != RouteJITNative {
			t.Fatalf("route: %+v", check)
		}
	}
	for _, step := range []struct {
		snapshot      *ModuleSnapshot
		seed          bool
		message, want string
	}{
		{first, true, "local-value", "local-value"},
		{second, true, "remote-value", "remote-value"},
		{first, false, "ignored", "local-value"},
		{second, false, "ignored", "remote-value"},
	} {
		input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "isolation", Fields: map[string]any{"seed": step.seed, "message": step.message}}})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := runner.ProcessModules(step.snapshot, input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("batch: %+v", batch)
		}
		deltas, err := batch.Records[0].MutationDeltas()
		if err != nil {
			t.Fatal(err)
		}
		point := Point{}
		for _, delta := range deltas {
			if err := delta.Apply(&point); err != nil {
				t.Fatal(err)
			}
		}
		if point.Fields["previous"] != step.want {
			t.Fatalf("shared namespace state: got %v want %s", point.Fields["previous"], step.want)
		}
	}
}

func TestRunnerModuleSnapshotCacheAndRetirement(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	sources := map[string]string{"main.p": `use("child.p")`, "child.p": `add_key(version, 1)`}
	old, err := NewModuleSnapshot("main.p", sources)
	if err != nil {
		t.Fatal(err)
	}
	same, err := NewModuleSnapshot("main.p", map[string]string{"child.p": sources["child.p"], "main.p": sources["main.p"]})
	if err != nil {
		t.Fatal(err)
	}
	if old.key != same.key {
		t.Fatal("map ordering changed identity")
	}
	sources["child.p"] = `add_key(version, 2)`
	next, err := NewModuleSnapshot("main.p", sources)
	if err != nil {
		t.Fatal(err)
	}
	if old.key == next.key || old.sources["child.p"] == sources["child.p"] {
		t.Fatal("snapshot not isolated")
	}
	first, err := runner.acquireModules(old)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	again, err := runner.acquireModules(same)
	if err != nil {
		t.Fatal(err)
	}
	if first.program != again.program {
		t.Fatal("identical snapshot recompiled")
	}
	again.Release()
	runner.InvalidateModules(old)
	oldProgram := first.program.(*nativeProgram)
	if oldProgram.closed {
		t.Fatal("closed leased program")
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "cache"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		process func([]byte) (Batch, error)
		want    int64
	}{
		{first.program.ProcessIndexed, 1},
		{func(input []byte) (Batch, error) { return runner.ProcessModules(next, input) }, 2},
	} {
		batch, err := tc.process(input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("bad batch: %+v", batch)
		}
		deltas, err := batch.Records[0].MutationDeltas()
		if err != nil {
			t.Fatal(err)
		}
		point := Point{}
		for _, delta := range deltas {
			if err := delta.Apply(&point); err != nil {
				t.Fatal(err)
			}
		}
		if point.Fields["version"] != tc.want {
			t.Fatalf("version: %+v", point.Fields)
		}
	}
	first.Release()
	if !oldProgram.closed {
		t.Fatal("retired program not closed by final lease")
	}
}
