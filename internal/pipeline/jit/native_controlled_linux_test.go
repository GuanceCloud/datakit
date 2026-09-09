// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestControlledStaticOutputOwnership(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	if n.staticCancellableInto == nil {
		t.Fatal("missing controlled static ABI")
	}
	compiled, err := rt.Compile(`cast(n,"int")`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "control", Fields: map[string]any{"n": "42"}}})
	if err != nil {
		t.Fatal(err)
	}
	normal, err := p.ProcessIndexed(input)
	if err != nil || normal.Static == nil {
		t.Fatal("requires static output", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	oldCreate := n.cancelCreate
	creates := 0
	n.cancelCreate = func(out *uintptr) int32 { creates++; return oldCreate(out) }
	original, free := n.staticCancellableInto, n.free
	var scratch Batch
	n.free = func(data *byte, size, capacity uintptr) {
		if len(scratch.inline) > 0 && data == bytesPointer(scratch.inline) {
			t.Fatal("freed caller scratch")
		}
		free(data, size, capacity)
	}
	for _, mode := range []string{"deadline", "cancel", "cancel", "spill", "owned"} {
		n.staticCancellableInto = original
		current := ctx
		if mode == "deadline" {
			current = WithRecordTimeout(context.Background(), time.Second)
		}
		if mode == "spill" {
			n.staticCancellableInto = func(h uintptr, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
				return original(h, in, l, o, b, tok, nil, 0)
			}
		}
		outputScratch := &scratch
		if mode == "owned" {
			outputScratch = nil
		}
		got, err := p.ProcessIndexedContextInto(current, input, outputScratch)
		if err != nil || got.Static == nil {
			t.Fatal(mode, "lost static fast path", err)
		}
		if !reflect.DeepEqual(got.Static.Values, normal.Static.Values) || !reflect.DeepEqual(got.Records, normal.Records) {
			t.Fatal(mode, "changed output")
		}
		scratch.ResetForReuse()
	}
	if creates != 1 {
		t.Fatal("healthy cancellation did not reuse token", creates)
	}

}

func TestControlledStaticCancellationAndClockBudget(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	if n.staticCancellableInto == nil {
		t.Fatal("missing controlled static ABI")
	}
	compiled, err := rt.Compile("cast(n,\"int\")", "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	if p.static == nil {
		t.Fatal("loop lacks static schema")
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "control", Fields: map[string]any{"n": "42", "repeat": true}}})
	if err != nil {
		t.Fatal(err)
	}
	old := n.staticCancellableInto
	originalCancel := n.cancelRequest
	for _, mode := range []string{"cancel", "deadline"} {
		ctx := WithRecordTimeout(context.Background(), 0)
		var wait chan struct{}
		if mode == "cancel" {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(context.Background())
			defer cancel()
			entered := make(chan struct{})
			requested := make(chan struct{})
			n.cancelRequest = func(token uintptr) int32 { status := originalCancel(token); close(requested); return status }
			wait = make(chan struct{})
			n.staticCancellableInto = func(h uintptr, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
				close(entered)
				<-requested
				return old(h, in, l, o, b, tok, s, cap)
			}
			go func() { defer close(wait); <-entered; cancel() }()
		} else {
			n.staticCancellableInto = old
		}
		start := time.Now()
		b, err := p.ProcessIndexedContextInto(ctx, input, nil)
		if wait != nil {
			<-wait
		}
		n.cancelRequest = originalCancel
		if err != nil || len(b.Records) != 1 || b.Static == nil || b.Records[0].Status != TerminalError || !b.Records[0].CommitPrefixError {
			t.Fatal(mode, "lost interruption prefix", err, b.Records)
		}
		code := "E_CANCELLED"
		if mode == "deadline" {
			code = "E_TIMEOUT"
		}
		if !strings.Contains(string(b.Records[0].Error), code) {
			t.Fatal(mode, string(b.Records[0].Error))
		}
		if time.Since(start) > time.Second {
			t.Fatal(mode, "interruption too slow")
		}
	}
	n.staticCancellableInto = old
	fresh, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "control", Fields: map[string]any{"n": "42", "repeat": false}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := p.ProcessIndexedContextInto(ctx, fresh, nil)
	if err != nil || b.Records[0].Status != TerminalOK {
		t.Fatal("old cancellation poisoned reuse", err)
	}
}

func TestCancellationCompletionLatency(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	compiled, err := rt.Compile("for ; repeat; {};add_key(done,true)", "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "latency", Fields: map[string]any{"repeat": true}}})
	if err != nil {
		t.Fatal(err)
	}
	original := n.staticCancellableInto
	if original == nil {
		t.Fatal("missing dynamic cancellable ABI")
	}
	var latencies []time.Duration
	for i := 0; i < 10; i++ {
		entered := make(chan struct{})
		done := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		n.staticCancellableInto = func(h uintptr, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
			close(entered)
			return original(h, in, l, o, b, tok, s, cap)
		}
		var result Batch
		var processErr error
		go func() { defer close(done); result, processErr = p.ProcessIndexedContextInto(ctx, input, nil) }()
		select {
		case <-entered:
		case <-time.After(time.Second):
			cancel()
			t.Fatal("native did not enter")
		}
		time.Sleep(time.Millisecond)
		start := time.Now()
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("cancel did not finish")
		}
		latencies = append(latencies, time.Since(start))
		if processErr != nil || len(result.Records) != 1 || !strings.Contains(string(result.Records[0].Error), "E_CANCELLED") {
			t.Fatal("native failed to acknowledge cancel", processErr, result.Records)
		}
	}
	n.staticCancellableInto = original
	slices.Sort(latencies)
	t.Logf("cancel-to-return n=10 min=%s median=%s max=%s", latencies[0], latencies[5], latencies[9])
}
