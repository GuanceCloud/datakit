// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// Keep loading/negotiation unchanged, but avoid reflect-based argument packing
// for these fixed integer/pointer signatures. SyscallN's go:uintptrescapes
// contract pins pointer arguments on the Go heap through the native call;
// convert pointers in the call expression, never into stored uintptr values.
func configureNativeCalls(n *nativeRuntime) {
	if symbol, err := purego.Dlsym(n.handle, "pp_jit_process_raw_batch_flat_v3"); err == nil {
		n.process = func(program uintptr, codec uint32, input *byte, size uintptr, options *processOptions, output *nativeBuffer) int32 {
			result, _, _ := purego.SyscallN(symbol, program, uintptr(codec), uintptr(unsafe.Pointer(input)), size, uintptr(unsafe.Pointer(options)), uintptr(unsafe.Pointer(output)))
			return int32(result)
		}
	}
	if n.staticRun != nil {
		if symbol, err := purego.Dlsym(n.handle, "pp_jit_process_static_batch_flat_v1"); err == nil {
			n.staticRun = func(program uintptr, input *byte, size uintptr, options *processOptions, output *nativeBuffer) int32 {
				result, _, _ := purego.SyscallN(symbol, program, uintptr(unsafe.Pointer(input)), size, uintptr(unsafe.Pointer(options)), uintptr(unsafe.Pointer(output)))
				return int32(result)
			}
		}
	}
	if n.indexedInto != nil {
		if symbol, err := purego.Dlsym(n.handle, "pp_jit_process_indexed_into_v1"); err == nil {
			n.indexedInto = func(program uintptr, codec uint32, input *byte, size uintptr, options *processOptions, output *nativeBuffer, token uintptr, scratch *byte, capacity uintptr) int32 {
				result, _, _ := purego.SyscallN(symbol, program, uintptr(codec), uintptr(unsafe.Pointer(input)), size, uintptr(unsafe.Pointer(options)), uintptr(unsafe.Pointer(output)), token, uintptr(unsafe.Pointer(scratch)), capacity)
				return int32(result)
			}
		}
	}
	if n.staticInto != nil {
		if symbol, err := purego.Dlsym(n.handle, "pp_jit_process_static_batch_flat_into_v1"); err == nil {
			n.staticInto = func(program uintptr, input *byte, size uintptr, options *processOptions, output *nativeBuffer, scratch *byte, capacity uintptr) int32 {
				result, _, _ := purego.SyscallN(symbol, program, uintptr(unsafe.Pointer(input)), size, uintptr(unsafe.Pointer(options)), uintptr(unsafe.Pointer(output)), uintptr(unsafe.Pointer(scratch)), capacity)
				return int32(result)
			}
		}
	}
	if n.staticCancellableInto != nil {
		if symbol, err := purego.Dlsym(n.handle, "pp_jit_process_static_cancellable_into_v1"); err == nil {
			n.staticCancellableInto = func(program uintptr, input *byte, size uintptr, options *processOptions, output *nativeBuffer, token uintptr, scratch *byte, capacity uintptr) int32 {
				result, _, _ := purego.SyscallN(symbol, program, uintptr(unsafe.Pointer(input)), size, uintptr(unsafe.Pointer(options)), uintptr(unsafe.Pointer(output)), token, uintptr(unsafe.Pointer(scratch)), capacity)
				return int32(result)
			}
		}
	}

	if symbol, err := purego.Dlsym(n.handle, "pp_jit_buffer_free_flat_v1"); err == nil {
		n.free = func(data *byte, size, capacity uintptr) {
			purego.SyscallN(symbol, uintptr(unsafe.Pointer(data)), size, capacity)
		}
	}
}
