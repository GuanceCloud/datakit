// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && jitbench && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"errors"
	"runtime"
	"unsafe"
)

// BenchmarkEncodedBatch is the native wire result before Go decoding. It is
// compiled only for the opt-in jitbench harness and is not part of production.
type BenchmarkEncodedBatch struct {
	encoded []byte
	static  *StaticOutputSchema
}

func (output BenchmarkEncodedBatch) Decode() (Batch, error) {
	if output.static != nil {
		return decodeStaticBatch(output.encoded, *output.static)
	}
	return decodeIndexed(output.encoded)
}

// DecodeInto measures the same container reuse as the production static path.
// The encoded fixture is immutable and remains owned by output.
func (output BenchmarkEncodedBatch) DecodeInto(scratch *Batch) (Batch, error) {
	if output.static != nil {
		return decodeStaticBatchInto(output.encoded, *output.static, scratch)
	}
	return decodeIndexedInto(output.encoded, scratch)
}

func (output BenchmarkEncodedBatch) Len() int {
	return len(output.encoded)
}

// BenchmarkNative runs the cached native program and copies its wire output,
// deliberately leaving Go decoding to BenchmarkEncodedBatch.Decode.
func (r *Runner) BenchmarkNative(source string, input []byte) (BenchmarkEncodedBatch, error) {
	if r == nil {
		return BenchmarkEncodedBatch{}, errors.New("JIT runner is nil")
	}
	lease, err := r.acquireProgram(source)
	if err != nil {
		return BenchmarkEncodedBatch{}, err
	}
	defer lease.Release()
	program, ok := lease.program.(*nativeProgram)
	if !ok {
		return BenchmarkEncodedBatch{}, errors.New("JIT benchmark requires the Linux native runtime")
	}
	return program.benchmarkNative(input)
}

func (p *nativeProgram) benchmarkNative(input []byte) (BenchmarkEncodedBatch, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed || p.handle == 0 {
		return BenchmarkEncodedBatch{}, errors.New("JIT program is closed")
	}
	options := processOptions{StructSize: uint32(unsafe.Sizeof(processOptions{}))}
	var output nativeBuffer
	var status int32
	if p.static != nil && p.runtime.staticRun != nil {
		status = p.runtime.staticRun(p.handle, bytesPointer(input), uintptr(len(input)), &options, &output)
	} else {
		options.Flags = 4
		if p.runtime.aggregatePollV2 != nil {
			options.Flags |= 8
		}
		options.OutputCodec = PayloadMutationDelta
		status = p.runtime.process(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), &options, &output)
	}
	runtime.KeepAlive(input)
	encoded, err := p.runtime.copyAndFree(output)
	if err != nil {
		return BenchmarkEncodedBatch{}, err
	}
	if status != 0 {
		return BenchmarkEncodedBatch{}, nativeError("process indexed batch", status, encoded)
	}
	return BenchmarkEncodedBatch{encoded: encoded, static: p.static}, nil
}

// BenchmarkSetHostObserver exposes bounded callback counts to the opt-in
// benchmark without changing the production HostCompat API.
func BenchmarkSetHostObserver(host *HostCompat, observer func(operation string, recordIndex uintptr)) {
	if host == nil {
		return
	}
	host.observeInvoke = func(operation uint32, recordIndex uintptr) {
		name := "unknown"
		switch operation {
		case hostCompatGeoIPOp:
			name = "geoip"
		case hostCompatUserAgentOp:
			name = "user_agent"
		case hostCompatTimestampOp:
			name = "default_time"
		}
		observer(name, recordIndex)
	}
}
