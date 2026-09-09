// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import "testing"

func TestRetiredProgramCannotRegisterAggregateRetry(t *testing.T) {
	d := newCacheTickDriver()
	defer d.Close()
	p := &nativeProgram{handle: 1, runtime: &nativeRuntime{clockDriver: d, destroy: func(uintptr) int32 { return 0 }}}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	p.scheduleAggregateRetry()
	p.scheduledAggregateRetry()
	d.mu.Lock()
	n := len(d.entries)
	d.mu.Unlock()
	if n != 0 {
		t.Fatalf("retired program re-registered %d callbacks", n)
	}
}

func TestSelectedCacheTickCannotReregisterAfterClose(t *testing.T) {
	for round := 0; round < 50; round++ {
		d := newCacheTickDriver()
		entered, resume := make(chan struct{}), make(chan struct{})
		destroyed := 0
		p := &nativeProgram{handle: 1, pendingAggregates: []Point{{Version: 1}}, runtime: &nativeRuntime{
			free:        func(*byte, uintptr, uintptr) {},
			clockDriver: d, cacheTick: func(uintptr) int32 { return 0 },
			// Fail final drain deliberately: Close must still retire the program.
			aggregatePoll: func(uintptr, int64, uint32, *nativeBuffer) int32 { return 2 },
			aggregateSink: func([]Point) error { close(entered); <-resume; return nil },
			destroy:       func(uintptr) int32 { destroyed++; return 0 },
		}}
		done := make(chan struct{})
		go func() { p.scheduledCacheTick(); close(done) }()
		<-entered
		if err := p.Close(); err == nil {
			t.Error("final drain failure was hidden")
		}
		close(resume)
		<-done
		p.scheduledAggregateRetry()
		d.mu.Lock()
		n := len(d.entries)
		d.mu.Unlock()
		d.Close()
		if n != 0 || destroyed != 1 {
			t.Fatalf("round %d: retained=%d destroyed=%d", round, n, destroyed)
		}
	}
}
