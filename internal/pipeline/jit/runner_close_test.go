// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

type closeErrorProgram struct {
	err       error
	closed    chan struct{}
	closeOnce sync.Once
}

func (program *closeErrorProgram) ProcessIndexed([]byte) (Batch, error) { return Batch{}, nil }
func (*closeErrorProgram) InputProjection() InputProjection             { return InputProjection{} }
func (*closeErrorProgram) Capabilities() (ProgramCapabilities, error) {
	return ProgramCapabilities{ExecutionMode: "native"}, nil
}
func (program *closeErrorProgram) Close() error {
	program.closeOnce.Do(func() { close(program.closed) })
	return program.err
}

type closeErrorRuntime struct {
	programs map[string]Program
	err      error
	closes   atomic.Int32
}

func (runtime *closeErrorRuntime) Compile(source, _ string) (Program, error) {
	return runtime.programs[source], nil
}
func (runtime *closeErrorRuntime) Close() error {
	runtime.closes.Add(1)
	return runtime.err
}

func newCloseErrorProgram(err error) *closeErrorProgram {
	return &closeErrorProgram{err: err, closed: make(chan struct{})}
}

func newTestRunner(runtime Runtime, max int) *Runner {
	return &Runner{
		runtime: runtime,
		profile: "test",
		max:     max,
		cache:   make(map[[32]byte]*cacheEntry),
	}
}

func TestCheckedRoutePinsInstanceUntilInvalidation(t *testing.T) {
	first := newCloseErrorProgram(nil)
	second := newCloseErrorProgram(nil)
	runtime := &closeErrorRuntime{programs: map[string]Program{"first": first, "second": second}}
	runner := newTestRunner(runtime, 1)
	defer runner.Close()
	if result := runner.Check("first"); result.Route != RouteJITNative {
		t.Fatalf("first route rejected: %#v", result)
	}
	if err := runner.Prepare("second"); err == nil {
		t.Fatal("capacity pressure evicted a selected native route")
	}
	assertNotSignaled(t, first.closed, "selected stateful instance was evicted")
	lease, err := runner.acquireProgram("first")
	if err != nil {
		t.Fatal(err)
	}
	if lease.program != first {
		t.Fatal("selected route changed instance")
	}
	runner.Invalidate("first")
	assertNotSignaled(t, first.closed, "in-flight selected instance closed early")
	lease.Release()
	awaitSignal(t, first.closed, "invalidated instance was not released")
	if result := runner.Check("second"); result.Route != RouteJITNative {
		t.Fatalf("capacity not released after invalidation: %#v", result)
	}
}

func TestRunnerCloseAggregatesInvalidatedProgramAndRuntimeErrors(t *testing.T) {
	programErr := errors.New("program close failed")
	runtimeErr := errors.New("runtime close failed")
	program := newCloseErrorProgram(programErr)
	runtime := &closeErrorRuntime{
		programs: map[string]Program{"source": program},
		err:      runtimeErr,
	}
	runner := newTestRunner(runtime, 1)

	if err := runner.Prepare("source"); err != nil {
		t.Fatalf("prepare source: %v", err)
	}
	runner.Invalidate("source")
	awaitSignal(t, program.closed, "invalidated program was not closed")

	err := runner.Close()
	if !errors.Is(err, programErr) || !errors.Is(err, runtimeErr) {
		t.Fatalf("Runner.Close error = %v, want program and runtime errors", err)
	}
	if got := runtime.closes.Load(); got != 1 {
		t.Fatalf("runtime Close calls = %d, want 1", got)
	}
}

func TestRunnerCloseAggregatesLRUProgramCloseError(t *testing.T) {
	programErr := errors.New("LRU program close failed")
	first := newCloseErrorProgram(programErr)
	second := newCloseErrorProgram(nil)
	runtime := &closeErrorRuntime{programs: map[string]Program{
		"first":  first,
		"second": second,
	}}
	runner := newTestRunner(runtime, 1)

	if err := runner.Prepare("first"); err != nil {
		t.Fatalf("prepare first: %v", err)
	}
	if err := runner.Prepare("second"); err != nil {
		t.Fatalf("prepare second: %v", err)
	}
	awaitSignal(t, first.closed, "LRU program was not closed")

	err := runner.Close()
	if !errors.Is(err, programErr) {
		t.Fatalf("Runner.Close error = %v, want LRU program error", err)
	}
}

func TestRunnerCloseAggregatesLeaseReleaseProgramCloseError(t *testing.T) {
	programErr := errors.New("lease-release program close failed")
	program := newCloseErrorProgram(programErr)
	runtime := &closeErrorRuntime{programs: map[string]Program{"source": program}}
	runner := newTestRunner(runtime, 1)

	lease, err := runner.acquireProgram("source")
	if err != nil {
		t.Fatalf("acquire source: %v", err)
	}
	runner.Invalidate("source")
	assertNotSignaled(t, program.closed, "in-flight program closed before lease release")
	lease.Release()
	awaitSignal(t, program.closed, "retired program was not closed after lease release")

	if err := runner.Close(); !errors.Is(err, programErr) {
		t.Fatalf("Runner.Close error = %v, want lease-release program error", err)
	}
}

type blockingCloseRuntime struct {
	program    Program
	err        error
	started    chan struct{}
	release    chan struct{}
	startOnce  sync.Once
	closeCount atomic.Int32
}

func (runtime *blockingCloseRuntime) Compile(string, string) (Program, error) {
	return runtime.program, nil
}
func (runtime *blockingCloseRuntime) Close() error {
	runtime.closeCount.Add(1)
	runtime.startOnce.Do(func() { close(runtime.started) })
	<-runtime.release
	return runtime.err
}

func TestRunnerConcurrentCloseWaitsAndReturnsSameResult(t *testing.T) {
	programErr := errors.New("program close failed")
	runtimeErr := errors.New("runtime close failed")
	runtime := &blockingCloseRuntime{
		program: newCloseErrorProgram(programErr),
		err:     runtimeErr,
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	runner := newTestRunner(runtime, 1)
	if err := runner.Prepare("source"); err != nil {
		t.Fatalf("prepare source: %v", err)
	}

	firstDone := make(chan error, 1)
	go func() { firstDone <- runner.Close() }()
	awaitSignal(t, runtime.started, "first Runner.Close did not reach runtime Close")

	secondDone := make(chan error, 1)
	go func() { secondDone <- runner.Close() }()
	assertNotSignaled(t, secondDone, "concurrent Runner.Close returned before the first close completed")

	close(runtime.release)
	firstErr := <-firstDone
	secondErr := <-secondDone
	if firstErr != secondErr {
		t.Fatalf("concurrent Close errors differ: first=%v second=%v", firstErr, secondErr)
	}
	if !errors.Is(firstErr, programErr) || !errors.Is(firstErr, runtimeErr) {
		t.Fatalf("concurrent Close error = %v, want program and runtime errors", firstErr)
	}
	if got := runtime.closeCount.Load(); got != 1 {
		t.Fatalf("runtime Close calls = %d, want 1", got)
	}

	// A caller arriving after completion observes the same stored result too.
	if err := runner.Close(); err != firstErr {
		t.Fatalf("post-completion Close error identity differs: got=%v want=%v", err, firstErr)
	}
}
