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
	"testing"
)

func TestNativeCallScratchClearsMetadataAcrossModesAndFailures(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	n.manualCacheTicks = true
	compiled, err := rt.Compile(`cast(n,"int")`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	if p.static == nil || n.staticInto == nil {
		t.Fatal("requires static inline ABI")
	}
	originalStatic, originalInto := p.static, n.staticInto
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "reuse", Fields: map[string]any{"n": "42"}}})
	if err != nil {
		t.Fatal(err)
	}
	var scratch Batch
	var retained *nativeCallState
	for _, mode := range []string{"inline", "general", "owned", "bad-input", "bad-output", "inline"} {
		p.static, n.staticInto = originalStatic, originalInto
		data := input
		switch mode {
		case "general":
			p.static = nil
		case "owned":
			n.staticInto = nil
		case "bad-input":
			data = []byte("invalid wire")
		case "bad-output":
			n.staticInto = func(handle uintptr, data *byte, size uintptr, opts *processOptions, out *nativeBuffer, buffer *byte, capacity uintptr) int32 {
				status := originalInto(handle, data, size, opts, out, buffer, capacity)
				if status == 0 && out.Length > 0 {
					*out.Pointer ^= 0xff
				}
				return status
			}
		}
		batch, err := p.ProcessIndexedInto(data, &scratch)
		if mode == "bad-input" || mode == "bad-output" {
			if err == nil {
				t.Fatalf("%s unexpectedly succeeded", mode)
			}
		} else if err != nil || len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("%s: records=%v error=%v", mode, batch.Records, err)
		}
		state, ok := scratch.nativeCall.(*nativeCallState)
		if !ok {
			t.Fatal("caller scratch lost call state")
		}
		if retained != nil && retained != state {
			t.Fatal("reallocated reusable call descriptors")
		}
		retained = state
		if state.options != (processOptions{}) || state.output != (nativeBuffer{}) {
			t.Fatalf("%s retained options or native output after return", mode)
		}
		scratch.ResetForReuse()
	}
}

func TestNativeCallPoolClearsMetadataAfterOwnedOutput(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	n.manualCacheTicks = true
	compiled, err := rt.Compile(`cast(n,"int")`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	if p.static == nil || n.staticRun == nil {
		t.Fatal("requires static ABI")
	}
	var observedOptions *processOptions
	var observedOutput *nativeBuffer
	original := n.staticRun
	n.staticRun = func(handle uintptr, data *byte, size uintptr, opts *processOptions, out *nativeBuffer) int32 {
		observedOptions, observedOutput = opts, out
		return original(handle, data, size, opts, out)
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "pool", Fields: map[string]any{"n": "42"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.ProcessIndexed(input); err != nil {
		t.Fatal(err)
	}
	if observedOptions == nil || observedOutput == nil {
		t.Fatal("native descriptors not observed")
	}
	if *observedOptions != (processOptions{}) || *observedOutput != (nativeBuffer{}) {
		t.Fatal("pooled call retained metadata after native free")
	}
}

func TestIndexedScratchCancellationOwnershipAndFallback(t *testing.T) {
	rt, err := openRuntime(os.Getenv("PLATYPUS_JIT_RUNTIME"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	n.manualCacheTicks = true
	compiled, err := rt.Compile(`a=load_json(_); add_key(result,a["v"]); pt_name("new")`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	p := compiled.(*nativeProgram)
	p.static = nil // Exercise indexed buffer ownership independently of static output.
	if n.indexedInto == nil {
		t.Fatal("requires indexed into ABI")
	}
	original, free := n.indexedInto, n.free
	borrowed := 0
	n.indexedInto = func(h uintptr, c uint32, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
		status := original(h, c, in, l, o, b, tok, s, cap)
		if b.Pointer == s {
			borrowed++
		}
		return status
	}
	var scratch Batch
	n.free = func(p *byte, l, c uintptr) {
		if len(scratch.inline) > 0 && p == bytesPointer(scratch.inline) {
			t.Fatal("freed Go scratch")
		}
		free(p, l, c)
	}
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "old", Fields: map[string]any{"message": `{"v":3}`}}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, legacy := range []bool{false, true, false} {
		if legacy {
			n.indexedInto = nil
		} else {
			n.indexedInto = original
		}
		for i := 0; i < 4; i++ {
			b, err := p.ProcessIndexedContextInto(ctx, input, &scratch)
			if err != nil || len(b.Records) != 1 || b.Records[0].Status != TerminalOK {
				t.Fatalf("legacy=%v batch=%v err=%v", legacy, b, err)
			}
			state := scratch.nativeCall.(*nativeCallState)
			if state.options != (processOptions{}) || state.output != (nativeBuffer{}) {
				t.Fatal("retained native descriptors")
			}
			scratch.ResetForReuse()
		}
	}
	// Observe a direct borrowed result separately from legacy calls.
	n.indexedInto = func(h uintptr, c uint32, in *byte, l uintptr, o *processOptions, b *nativeBuffer, tok uintptr, s *byte, cap uintptr) int32 {
		status := original(h, c, in, l, o, b, tok, s, cap)
		if b.Pointer == s {
			borrowed++
		}
		return status
	}
	if _, err := p.ProcessIndexedContextInto(ctx, input, &scratch); err != nil {
		t.Fatal(err)
	}
	if borrowed == 0 {
		t.Fatal("indexed call did not borrow scratch")
	}
	cancel()
	if _, err := p.ProcessIndexedContextInto(ctx, input, &scratch); err == nil {
		t.Fatal("cancelled input executed")
	}
}
