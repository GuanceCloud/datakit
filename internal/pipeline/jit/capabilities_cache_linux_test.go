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
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNativeCapabilitiesSnapshot(t *testing.T) {
	payload := []byte(`{"version":1,"source_hash":"` + strings.Repeat("a", 64) + `","execution_mode":"native_with_host","backend":"machine_code","execution_tier":"machine_code_helper","host_calls":["geoip"],"aggregate_events":false,"stateless":true,"input_projection":{"mode":"keys"},"static_output":{"mode":"slots"}}`)
	var calls, frees atomic.Int32
	rt := &nativeRuntime{
		capabilities: func(_ uintptr, out *nativeBuffer) int32 {
			calls.Add(1)
			*out = nativeBuffer{Pointer: &payload[0], Length: uintptr(len(payload)), Capacity: uintptr(len(payload))}
			return 0
		},
		free: func(_ *byte, _, _ uintptr) { frees.Add(1) },
	}
	p := &nativeProgram{runtime: rt, handle: 1}
	var workers sync.WaitGroup
	for range 32 {
		workers.Go(func() {
			capabilities, err := p.Capabilities()
			if err != nil {
				t.Error(err)
				return
			}
			if capabilities.HostCalls[0] != "geoip" || *capabilities.AggregateEvents || capabilities.Stateless == nil || !*capabilities.Stateless {
				t.Error("a caller changed the program-owned descriptor")
			}
			capabilities.HostCalls[0] = "changed"
			*capabilities.AggregateEvents = true
			if capabilities.Stateless != nil {
				*capabilities.Stateless = false
			}
		})
	}
	workers.Wait()
	if calls.Load() != 1 || frees.Load() != 1 {
		t.Fatalf("queries=%d frees=%d, want one of each", calls.Load(), frees.Load())
	}
	// A replacement program negotiates independently even in the same runtime.
	replacement := &nativeProgram{runtime: rt, handle: 2}
	if _, err := replacement.Capabilities(); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatal("new generation reused the old descriptor")
	}
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	if _, err := p.Capabilities(); err == nil {
		t.Fatal("closed program returned cached capabilities")
	}
}

func TestNativeCapabilitiesCachesInvalidDescriptor(t *testing.T) {
	payload := []byte(`{"version":999}`)
	var calls int
	p := &nativeProgram{handle: 1, runtime: &nativeRuntime{
		capabilities: func(_ uintptr, out *nativeBuffer) int32 {
			calls++
			*out = nativeBuffer{Pointer: &payload[0], Length: uintptr(len(payload)), Capacity: uintptr(len(payload))}
			return 0
		},
		free: func(_ *byte, _, _ uintptr) {},
	}}
	for range 2 {
		if _, err := p.Capabilities(); err == nil {
			t.Fatal("invalid capability descriptor was accepted")
		}
	}
	if calls != 1 {
		t.Fatalf("invalid immutable descriptor queried %d times", calls)
	}
}
