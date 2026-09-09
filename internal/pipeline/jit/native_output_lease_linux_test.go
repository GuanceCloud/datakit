// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNativeCloseWaitsForOutputCopyAndFree(t *testing.T) {
	testNativeOutputLease(t, false, false)
}

func TestNativeCloseWaitsForStaticOutputDecodeAndFree(t *testing.T) {
	testNativeOutputLease(t, true, false)
}

func TestNativeDecodeFailureStillFreesBeforeClose(t *testing.T) {
	t.Run("delta", func(t *testing.T) { testNativeOutputLease(t, false, true) })
	t.Run("static", func(t *testing.T) { testNativeOutputLease(t, true, true) })
}

func TestNativeReusableOutputLease(t *testing.T) {
	t.Run("valid", func(t *testing.T) { testNativeOutputLeaseMode(t, false, false, true) })
	t.Run("malformed", func(t *testing.T) { testNativeOutputLeaseMode(t, false, true, true) })
}

func testNativeOutputLease(t *testing.T, static, corrupt bool) {
	testNativeOutputLeaseMode(t, static, corrupt, false)
}

func testNativeOutputLeaseMode(t *testing.T, static, corrupt, reuse bool) {
	t.Helper()
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	n.manualCacheTicks = true
	// This test checks the owned allocation/free lease, including legacy runtimes.
	n.indexedInto = nil
	source := `add_key(result,cache_set("key","value") == nil)`
	if static {
		source = `add_key(result,"retained-output")`
	}
	compiled, err := rt.Compile(source, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	p := compiled.(*nativeProgram)
	if static {
		if p.static == nil || n.staticRun == nil {
			p.Close()
			t.Fatal("test requires actual static ABI")
		}
	} else {
		p.static = nil
	}
	returned, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer func() { unblock(); p.Close() }()
	var freed, destroyedBeforeFree atomic.Bool
	var injected atomic.Bool
	corruptHeader := func(out *nativeBuffer) {
		// Keep pointer/length/capacity valid for the real allocator. Only
		// corrupt the protocol magic in the owned output before decoding.
		if corrupt && out.Pointer != nil && out.Length > 0 {
			*out.Pointer ^= 0xff
			injected.Store(true)
		}
	}
	originalProcess, originalFree, originalDestroy := n.process, n.free, n.destroy
	n.process = func(handle uintptr, codec uint32, data *byte, size uintptr, opts *processOptions, out *nativeBuffer) int32 {
		status := originalProcess(handle, codec, data, size, opts, out)
		corruptHeader(out)
		close(returned)
		<-release
		return status
	}
	originalStatic := n.staticRun
	if static {
		n.staticRun = func(handle uintptr, data *byte, size uintptr, opts *processOptions, out *nativeBuffer) int32 {
			status := originalStatic(handle, data, size, opts, out)
			corruptHeader(out)
			close(returned)
			<-release
			return status
		}
	}
	n.free = func(data *byte, length, capacity uintptr) {
		originalFree(data, length, capacity)
		freed.Store(true)
	}
	n.destroy = func(handle uintptr) int32 {
		if !freed.Load() {
			destroyedBeforeFree.Store(true)
		}
		return originalDestroy(handle)
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "lease", Fields: map[string]any{"message": "input"}}})
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		batch Batch
		err   error
	}
	processed := make(chan result, 1)
	go func() {
		var batch Batch
		var err error
		if reuse {
			var scratch Batch
			batch, err = p.ProcessIndexedInto(input, &scratch)
		} else {
			batch, err = p.ProcessIndexed(input)
		}
		processed <- result{batch, err}
	}()
	select {
	case <-returned:
	case <-time.After(3 * time.Second):
		t.Fatal("native process did not return")
	}
	closing, closed := make(chan struct{}), make(chan error, 1)
	go func() { close(closing); closed <- p.Close() }()
	<-closing
	select {
	case err := <-closed:
		t.Fatalf("close completed before output release: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	if freed.Load() {
		t.Fatal("test failed to hold native output")
	}
	unblock()
	var got result
	select {
	case got = <-processed:
	case <-time.After(3 * time.Second):
		t.Fatal("process did not drain")
	}
	if got.err != nil && !corrupt {
		t.Fatal(got.err)
	}
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("close did not drain")
	}
	if !freed.Load() || destroyedBeforeFree.Load() {
		t.Fatal("program destroyed before native output was freed")
	}
	if corrupt {
		if !injected.Load() || got.err == nil {
			t.Fatal("corrupt native output was not rejected")
		}
		if len(got.batch.Records) != 0 || got.batch.Static != nil {
			t.Fatal("decode failure exposed a partial batch")
		}
		return
	}
	if len(got.batch.Records) != 1 || got.batch.Records[0].Status != TerminalOK {
		t.Fatalf("invalid copied output: %+v", got.batch)
	}
	if static {
		batch := got.batch.Static
		if batch == nil || !batch.Validated() {
			t.Fatal("static output missing or unvalidated")
		}
		valueIndex, found := 0, false
		for index, state := range batch.States {
			if state == StaticSetField || state == StaticSetTag {
				if batch.Schema.Keys[index] == "result" {
					found = true
					if state != StaticSetField || batch.Values[valueIndex] != "retained-output" {
						t.Fatalf("static result lost after close: %+v", batch)
					}
				}
				valueIndex++
			}
		}
		if !found {
			t.Fatal("static result absent after close")
		}
		return
	}
	deltas, err := got.batch.Records[0].MutationDeltas()
	if err != nil {
		t.Fatal(err)
	}
	point := Point{Fields: map[string]any{}}
	for _, delta := range deltas {
		if err := delta.Apply(&point); err != nil {
			t.Fatal(err)
		}
	}
	if point.Fields["result"] != true {
		t.Fatalf("copied output lost after close: %+v", point)
	}
}

func TestNativeInlineOutputOwnershipAndOversizeFallback(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	if n.staticInto == nil {
		t.Skip("runtime predates optional inline result ABI")
	}
	compiled, err := rt.Compile(`grok(_, "%{GREEDYDATA:out}")`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	var scratch Batch
	originalFree, originalInto := n.free, n.staticInto
	var calls, frees int
	n.free = func(data *byte, size, capacity uintptr) {
		frees++
		originalFree(data, size, capacity)
	}
	n.staticInto = func(handle uintptr, data *byte, size uintptr, options *processOptions, out *nativeBuffer, buffer *byte, capacity uintptr) int32 {
		calls++
		return originalInto(handle, data, size, options, out, buffer, capacity)
	}
	for _, size := range []int{16, 8192, 32} {
		message := strings.Repeat("x", size)
		input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Fields: map[string]any{"message": message}}})
		if err != nil {
			t.Fatal(err)
		}
		beforeCalls, beforeFrees := calls, frees
		result, err := p.ProcessIndexedInto(input, &scratch)
		if err != nil {
			t.Fatal(err)
		}
		if calls-beforeCalls != 1 {
			t.Fatal("native execution replayed for output sizing")
		}
		wantFrees := 0
		if size > 4096 {
			wantFrees = 1
		}
		if frees-beforeFrees != wantFrees {
			t.Fatalf("size %d: native frees=%d, want %d", size, frees-beforeFrees, wantFrees)
		}
		found := false
		for _, value := range result.Static.Values {
			if value == message {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing output of length %d", size)
		}
		clear(scratch.inline)
		for _, value := range result.Static.Values {
			if text, ok := value.(string); ok && len(text) == size && text != message {
				t.Fatal("decoded string aliases inline buffer")
			}
		}
	}
	// Malformed metadata must fail without passing a Go-owned buffer to Rust free.
	n.staticInto = func(_ uintptr, _ *byte, _ uintptr, _ *processOptions, out *nativeBuffer, buffer *byte, capacity uintptr) int32 {
		*out = nativeBuffer{Pointer: buffer, Length: capacity + 1, Capacity: capacity}
		return 0
	}
	before := frees
	if _, err := p.ProcessIndexedInto(nil, &scratch); err == nil {
		t.Fatal("oversized inline result accepted")
	}
	if frees != before {
		t.Fatal("freed caller-owned storage")
	}
}
