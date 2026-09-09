// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDecodeIndexedRejectsOversizedRecordCountBeforeAllocation(t *testing.T) {
	input := make([]byte, headerBytes)
	copy(input[:4], "PPR3")
	binary.LittleEndian.PutUint16(input[4:6], 3)
	binary.LittleEndian.PutUint32(input[8:12], pointWireMaxItems+1)

	if _, err := decodeIndexed(input); err == nil || err.Error() != "invalid PPR3 counts" {
		t.Fatalf("decodeIndexed oversized record count error = %v", err)
	}
}

type indexedTestFrame struct {
	recordIndex uint64
	kind        byte
	status      TerminalStatus
	category    byte
	codec       byte
	flags       uint16
	payload     []byte
}

//nolint:makezero // The initialized wire header must precede appended payloads.
func encodeIndexedTestBatch(tb testing.TB, frames ...indexedTestFrame) []byte {
	tb.Helper()
	encoded := make([]byte, headerBytes)
	copy(encoded[:4], "PPR3")
	binary.LittleEndian.PutUint16(encoded[4:6], 3)
	binary.LittleEndian.PutUint16(encoded[6:8], 1)
	binary.LittleEndian.PutUint32(encoded[8:12], 1)
	binary.LittleEndian.PutUint32(encoded[12:16], uint32(len(frames)))
	for _, frame := range frames {
		header := make([]byte, frameHeaderBytes)
		binary.LittleEndian.PutUint64(header[:8], frame.recordIndex)
		header[8] = frame.kind
		header[9] = byte(frame.status)
		header[10] = frame.category
		header[11] = frame.codec
		binary.LittleEndian.PutUint16(header[12:14], frame.flags)
		binary.LittleEndian.PutUint32(header[16:20], uint32(len(frame.payload)))
		encoded = append(encoded, header...)
		encoded = append(encoded, frame.payload...)
	}
	binary.LittleEndian.PutUint32(encoded[16:20], uint32(len(encoded)-headerBytes))
	return encoded
}

func TestDecodeIndexedGlobalEventShapeIsStrict(t *testing.T) {
	event := indexedTestFrame{
		recordIndex: ^uint64(0),
		kind:        3,
		category:    4,
		codec:       5,
		payload:     []byte(`{"kind":"aggregate"}`),
	}
	errorFrame := indexedTestFrame{
		kind:     5,
		category: 3,
		codec:    5,
		payload:  []byte(`{"code":"E_TEST"}`),
	}
	terminal := indexedTestFrame{kind: 6, status: TerminalError}
	batch, err := decodeIndexed(encodeIndexedTestBatch(t, event, errorFrame, terminal))
	if err != nil {
		t.Fatalf("decode valid global event: %v", err)
	}
	if len(batch.Diagnostics) != 1 || string(batch.Diagnostics[0]) != string(event.payload) {
		t.Fatalf("global event diagnostic = %#v", batch.Diagnostics)
	}

	tests := map[string]func(*indexedTestFrame){
		"terminal-status": func(frame *indexedTestFrame) { frame.status = TerminalOK },
		"point-category":  func(frame *indexedTestFrame) { frame.category = 1 },
		"utf8-codec":      func(frame *indexedTestFrame) { frame.codec = 4 },
		"terminal-flag":   func(frame *indexedTestFrame) { frame.flags = indexedTerminalCommitPrefixErrorFlag },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			malformed := event
			mutate(&malformed)
			if _, err := decodeIndexed(encodeIndexedTestBatch(t, malformed, errorFrame, terminal)); err == nil {
				t.Fatal("malformed global event was accepted")
			}
		})
	}
}

func TestDecodeIndexedCommittedPrefixError(t *testing.T) {
	delta, err := EncodeMutationDeltas([]MutationDelta{{RecordIndex: 0}})
	if err != nil {
		t.Fatal(err)
	}
	errorFrame := indexedTestFrame{kind: 5, category: 3, codec: 5, payload: []byte(`{"code":"E_PIPELINE_RUNTIME","message":"key not found"}`)}
	deltaFrame := indexedTestFrame{kind: 7, category: 1, codec: PayloadMutationDelta, payload: delta}
	committedTerminal := indexedTestFrame{kind: 6, status: TerminalError, flags: indexedTerminalCommitPrefixErrorFlag}

	batch, err := decodeIndexed(encodeIndexedTestBatch(t, deltaFrame, errorFrame, committedTerminal))
	if err != nil {
		t.Fatalf("decode committed prefix error: %v", err)
	}
	if len(batch.Records) != 1 || !batch.Records[0].CommitPrefixError || batch.Records[0].Status != TerminalError {
		t.Fatalf("unexpected committed prefix record: %#v", batch.Records)
	}
	deltas, err := batch.Records[0].MutationDeltas()
	if err != nil || len(deltas) != 1 || len(deltas[0].Operations) != 0 {
		t.Fatalf("empty committed delta = %#v, err=%v", deltas, err)
	}
	ordinary, err := decodeIndexed(encodeIndexedTestBatch(t,
		errorFrame, indexedTestFrame{kind: 6, status: TerminalError}))
	if err != nil || len(ordinary.Records) != 1 || ordinary.Records[0].CommitPrefixError ||
		ordinary.Records[0].Status != TerminalError {
		t.Fatalf("ordinary terminal error compatibility changed: batch=%#v err=%v", ordinary, err)
	}

	primaryFrame := indexedTestFrame{kind: 1, category: 1, codec: PayloadPointJSON, payload: []byte(`{}`)}
	tests := map[string][]indexedTestFrame{
		"missing-error":           {deltaFrame, committedTerminal},
		"missing-delta":           {errorFrame, committedTerminal},
		"point-and-delta":         {primaryFrame, deltaFrame, errorFrame, committedTerminal},
		"ordinary-error-delta":    {deltaFrame, errorFrame, {kind: 6, status: TerminalError}},
		"flag-on-ok":              {deltaFrame, {kind: 6, status: TerminalOK, flags: indexedTerminalCommitPrefixErrorFlag}},
		"flag-on-canceled":        {errorFrame, {kind: 6, status: TerminalCancelled, flags: indexedTerminalCommitPrefixErrorFlag}},
		"unknown-terminal-flag":   {deltaFrame, errorFrame, {kind: 6, status: TerminalError, flags: 2}},
		"flag-on-non-terminal":    {{kind: 5, category: 3, codec: 5, flags: 1, payload: errorFrame.payload}, committedTerminal},
		"duplicate-error":         {deltaFrame, errorFrame, errorFrame, committedTerminal},
		"duplicate-delta":         {deltaFrame, deltaFrame, errorFrame, committedTerminal},
		"committed-cancel-output": {deltaFrame, errorFrame, {kind: 6, status: TerminalCancelled}},
	}
	for name, frames := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeIndexed(encodeIndexedTestBatch(t, frames...)); err == nil {
				t.Fatal("malformed committed error was accepted")
			}
		})
	}
}

func TestDecodeIndexedErrorFrameMatchesTerminal(t *testing.T) {
	errorFrame := indexedTestFrame{
		kind:     5,
		category: 3,
		codec:    5,
		payload:  []byte(`{"code":"E_TEST"}`),
	}
	primaryFrame := indexedTestFrame{
		kind:     1,
		category: 1,
		codec:    PayloadPointJSON,
		payload:  []byte(`{}`),
	}

	valid := map[string]struct {
		flags  uint16
		frames []indexedTestFrame
	}{
		"ok without error": {
			frames: []indexedTestFrame{primaryFrame, {kind: 6, status: TerminalOK}},
		},
		"dropped without error": {
			frames: []indexedTestFrame{{kind: 6, status: TerminalDropped}},
		},
		"error with diagnostic": {
			flags:  1,
			frames: []indexedTestFrame{errorFrame, {kind: 6, status: TerminalError}},
		},
		"canceled with diagnostic": {
			flags:  1,
			frames: []indexedTestFrame{errorFrame, {kind: 6, status: TerminalCancelled}},
		},
	}
	for name, test := range valid {
		t.Run(name, func(t *testing.T) {
			encoded := encodeIndexedTestBatch(t, test.frames...)
			binary.LittleEndian.PutUint16(encoded[6:8], test.flags)
			if _, err := decodeIndexed(encoded); err != nil {
				t.Fatalf("valid terminal/error association rejected: %v", err)
			}
		})
	}

	malformed := map[string]struct {
		flags  uint16
		frames []indexedTestFrame
	}{
		"ok with diagnostic": {
			frames: []indexedTestFrame{primaryFrame, errorFrame, {kind: 6, status: TerminalOK}},
		},
		"dropped with diagnostic": {
			frames: []indexedTestFrame{errorFrame, {kind: 6, status: TerminalDropped}},
		},
		"error without diagnostic": {
			flags:  1,
			frames: []indexedTestFrame{{kind: 6, status: TerminalError}},
		},
		"canceled without diagnostic": {
			flags:  1,
			frames: []indexedTestFrame{{kind: 6, status: TerminalCancelled}},
		},
	}
	for name, test := range malformed {
		t.Run(name, func(t *testing.T) {
			encoded := encodeIndexedTestBatch(t, test.frames...)
			binary.LittleEndian.PutUint16(encoded[6:8], test.flags)
			if _, err := decodeIndexed(encoded); err == nil {
				t.Fatal("malformed terminal/error association was accepted")
			}
		})
	}
}

type checkRuntime struct {
	compileCount int
	program      Program
	err          error
}

type leaseRuntime struct {
	programs map[string]*leaseProgram
	closed   chan struct{}
}

func (runtime *leaseRuntime) Compile(source, _ string) (Program, error) {
	return runtime.programs[source], nil
}

func (runtime *leaseRuntime) Close() error {
	close(runtime.closed)
	return nil
}

type leaseProgram struct {
	started   chan struct{}
	unblock   chan struct{}
	closed    chan struct{}
	startOnce sync.Once
	closeOnce sync.Once
}

type invalidateRuntime struct {
	mu       sync.Mutex
	programs []*leaseProgram
	block    bool
}

func (runtime *invalidateRuntime) Compile(string, string) (Program, error) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	program := newLeaseProgram()
	if !runtime.block {
		close(program.unblock)
	}
	runtime.programs = append(runtime.programs, program)
	return program, nil
}

func (*invalidateRuntime) Close() error { return nil }

func (runtime *invalidateRuntime) compiled() []*leaseProgram {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return append([]*leaseProgram(nil), runtime.programs...)
}

func (program *leaseProgram) ProcessIndexed([]byte) (Batch, error) {
	program.startOnce.Do(func() { close(program.started) })
	<-program.unblock
	return Batch{}, nil
}

func (*leaseProgram) InputProjection() InputProjection { return InputProjection{} }
func (*leaseProgram) Capabilities() (ProgramCapabilities, error) {
	return ProgramCapabilities{ExecutionMode: "native"}, nil
}
func (program *leaseProgram) Close() error {
	program.closeOnce.Do(func() { close(program.closed) })
	return nil
}

func newLeaseProgram() *leaseProgram {
	return &leaseProgram{
		started: make(chan struct{}),
		unblock: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

func TestRunnerDefersProgramAndRuntimeCloseUntilLeasesDrain(t *testing.T) {
	first := newLeaseProgram()
	second := newLeaseProgram()
	runtime := &leaseRuntime{
		programs: map[string]*leaseProgram{"first": first, "second": second},
		closed:   make(chan struct{}),
	}
	runner := &Runner{
		runtime: runtime,
		profile: "test",
		max:     1,
		cache:   make(map[[32]byte]*cacheEntry),
	}

	firstDone := make(chan error, 1)
	go func() {
		_, err := runner.Process("first", nil)
		firstDone <- err
	}()
	awaitSignal(t, first.started, "first process did not start")

	if err := runner.Prepare("second"); err != nil {
		t.Fatalf("prepare second program: %v", err)
	}
	select {
	case <-first.closed:
		t.Fatal("LRU eviction closed a program with an active Process lease")
	default:
	}

	secondDone := make(chan error, 1)
	go func() {
		_, err := runner.Process("second", nil)
		secondDone <- err
	}()
	awaitSignal(t, second.started, "second process did not start")

	closeDone := make(chan error, 1)
	go func() { closeDone <- runner.Close() }()
	assertNotSignaled(t, closeDone, "Runner.Close returned with active program leases")
	assertNotSignaled(t, runtime.closed, "runtime closed with active program leases")

	close(first.unblock)
	if err := <-firstDone; err != nil {
		t.Fatalf("first process: %v", err)
	}
	awaitSignal(t, first.closed, "evicted program was not closed after its lease drained")
	assertNotSignaled(t, closeDone, "Runner.Close ignored the second active lease")

	close(second.unblock)
	if err := <-secondDone; err != nil {
		t.Fatalf("second process: %v", err)
	}
	if err := <-closeDone; err != nil {
		t.Fatalf("close runner: %v", err)
	}
	awaitSignal(t, second.closed, "current program was not closed")
	awaitSignal(t, runtime.closed, "runtime was not closed after programs drained")
}

func TestRunnerInvalidateClosesIdleProgramImmediately(t *testing.T) {
	runtime := &invalidateRuntime{}
	runner := &Runner{
		runtime: runtime,
		profile: "test",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}
	if err := runner.Prepare("source"); err != nil {
		t.Fatalf("prepare source: %v", err)
	}
	program := runtime.compiled()[0]
	runner.Invalidate("source")
	awaitSignal(t, program.closed, "Invalidate did not close an idle program before returning")
}

func TestRunnerInvalidateDefersCloseForInflightProcess(t *testing.T) {
	runtime := &invalidateRuntime{block: true}
	runner := &Runner{
		runtime: runtime,
		profile: "test",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}
	processed := make(chan error, 1)
	go func() {
		_, err := runner.Process("source", nil)
		processed <- err
	}()
	for len(runtime.compiled()) == 0 {
		time.Sleep(time.Millisecond)
	}
	program := runtime.compiled()[0]
	awaitSignal(t, program.started, "Process did not start")
	runner.Invalidate("source")
	assertNotSignaled(t, program.closed, "Invalidate closed an in-flight program")
	close(program.unblock)
	if err := <-processed; err != nil {
		t.Fatalf("process source: %v", err)
	}
	awaitSignal(t, program.closed, "retired program did not close after Process released its lease")
}

func TestRunnerInvalidateIsIdempotentAndAllowsRecompile(t *testing.T) {
	runtime := &invalidateRuntime{}
	runner := &Runner{
		runtime: runtime,
		profile: "test",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}
	if err := runner.Prepare("source"); err != nil {
		t.Fatalf("prepare first source: %v", err)
	}
	first := runtime.compiled()[0]
	runner.Invalidate("source")
	runner.Invalidate("source")
	awaitSignal(t, first.closed, "first program was not closed")

	if err := runner.Prepare("source"); err != nil {
		t.Fatalf("recompile source: %v", err)
	}
	programs := runtime.compiled()
	if len(programs) != 2 || programs[1] == first {
		t.Fatalf("compile count = %d, want a new second program", len(programs))
	}
	assertNotSignaled(t, programs[1].closed, "newly recompiled program was already closed")
	if err := runner.Close(); err != nil {
		t.Fatalf("close runner: %v", err)
	}
	awaitSignal(t, programs[1].closed, "recompiled program was not closed with runner")
}

func awaitSignal(t *testing.T, signal <-chan struct{}, message string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(5 * time.Second):
		t.Fatal(message)
	}
}

func assertNotSignaled[T any](t *testing.T, signal <-chan T, message string) {
	t.Helper()
	select {
	case <-signal:
		t.Fatal(message)
	case <-time.After(50 * time.Millisecond):
	}
}

func (runtime *checkRuntime) Compile(string, string) (Program, error) {
	runtime.compileCount++
	return runtime.program, runtime.err
}

func (*checkRuntime) Close() error { return nil }

type checkProgram struct {
	capabilities ProgramCapabilities
	err          error
}

func (*checkProgram) ProcessIndexed([]byte) (Batch, error) { return Batch{}, nil }
func (*checkProgram) InputProjection() InputProjection     { return InputProjection{} }
func (program *checkProgram) Capabilities() (ProgramCapabilities, error) {
	return program.capabilities, program.err
}
func (*checkProgram) Close() error { return nil }

func TestRunnerCheckReusesCompiledProgram(t *testing.T) {
	runtime := &checkRuntime{program: &checkProgram{capabilities: ProgramCapabilities{
		Version:           1,
		ExecutionMode:     "native_with_host",
		HostCalls:         []string{"default_time"},
		RequiredHostFlags: 4,
		InputProjection:   CapabilityMode{Mode: "all"},
		StaticOutput:      CapabilityMode{Mode: "none"},
		DescriptorPresent: true,
	}}}
	runner := &Runner{
		runtime: runtime,
		profile: "pipeline-go-1.4.3-datakit",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}

	first := runner.Check("default_time(ts)\n")
	second := runner.Check("default_time(ts)\n")
	if runtime.compileCount != 1 {
		t.Fatalf("compile count = %d, want 1", runtime.compileCount)
	}
	for _, result := range []CheckResult{first, second} {
		if result.Route != RouteJITWithHost || result.Reason != "jit_ready" {
			t.Fatalf("unexpected route result: %#v", result)
		}
		if result.Capabilities.ExecutionMode != "native_with_host" ||
			result.Capabilities.RequiredHostFlags != 4 ||
			len(result.Capabilities.HostCalls) != 1 || result.Capabilities.HostCalls[0] != "default_time" {
			t.Fatalf("unexpected capabilities: %#v", result.Capabilities)
		}
		if len(result.Capabilities.SourceHash) != 64 {
			t.Fatalf("source hash = %q, want SHA-256", result.Capabilities.SourceHash)
		}
	}
}

func TestRunnerCheckUsesConservativeNativeRouteWithoutDescriptor(t *testing.T) {
	runtime := &checkRuntime{program: &checkProgram{capabilities: ProgramCapabilities{
		ExecutionMode:   "native",
		InputProjection: CapabilityMode{Mode: "unknown"},
		StaticOutput:    CapabilityMode{Mode: "unknown"},
	}}}
	runner := &Runner{
		runtime: runtime,
		profile: "pipeline-go-1.4.3-datakit",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}

	result := runner.Check("drop_key(foo)\n")
	if result.Route != RouteJITNative || result.Reason != "jit_ready" {
		t.Fatalf("unexpected route result: %#v", result)
	}
	if result.Capabilities.DescriptorPresent {
		t.Fatalf("old runtime unexpectedly reported a descriptor: %#v", result.Capabilities)
	}
}

func TestRunnerCheckNegativelyCachesCompileFailure(t *testing.T) {
	runtime := &checkRuntime{err: errors.New("unsupported builtin")}
	runner := &Runner{
		runtime: runtime,
		profile: "pipeline-go-1.4.3-datakit",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}

	for range 2 {
		result := runner.Check("unsupported()\n")
		if result.Route != RoutePipelineGo || result.Reason != "compile_error" || result.Detail == "" {
			t.Fatalf("unexpected fallback result: %#v", result)
		}
	}
	if runtime.compileCount != 1 {
		t.Fatalf("compile count = %d, want 1", runtime.compileCount)
	}
}

func TestRunnerCheckRoutesCapabilityFailureToPipelineGo(t *testing.T) {
	runtime := &checkRuntime{program: &checkProgram{err: errors.New("invalid descriptor")}}
	runner := &Runner{
		runtime: runtime,
		profile: "pipeline-go-1.4.3-datakit",
		max:     8,
		cache:   make(map[[32]byte]*cacheEntry),
	}

	result := runner.Check("drop_key(foo)\n")
	if result.Route != RoutePipelineGo || result.Reason != "capabilities_error" {
		t.Fatalf("unexpected route result: %#v", result)
	}
}

func TestCheckResultJSONContract(t *testing.T) {
	encoded, err := json.Marshal(CheckResult{
		Route:  RouteJITNative,
		Reason: "jit_ready",
		Capabilities: ProgramCapabilities{
			ExecutionMode:     "native",
			Backend:           ExecutionBackendMachineCode,
			ExecutionTier:     ExecutionTierMachineCodeHelper,
			InputProjection:   CapabilityMode{Mode: "keys"},
			StaticOutput:      CapabilityMode{Mode: "slots"},
			DescriptorPresent: true,
		},
	})
	if err != nil {
		t.Fatalf("marshal route result: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode route result: %v", err)
	}
	if decoded["route"] != "jit_native" || decoded["reason"] != "jit_ready" {
		t.Fatalf("unexpected route JSON: %s", encoded)
	}
	capabilities, ok := decoded["capabilities"].(map[string]any)
	if !ok || capabilities["descriptor_present"] != true || capabilities["execution_mode"] != "native" ||
		capabilities["backend"] != "machine_code" || capabilities["execution_tier"] != "machine_code_helper" {
		t.Fatalf("unexpected capabilities JSON: %s", encoded)
	}
}

func TestNormalizeExecutionTierCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		capability  ProgramCapabilities
		wantBackend string
		wantTier    string
		wantError   bool
	}{
		{
			name:        "descriptor-before-backend",
			capability:  ProgramCapabilities{},
			wantBackend: ExecutionBackendInstructionEngine,
			wantTier:    ExecutionTierInstructionEngine,
		},
		{
			name:        "first-cranelift-runtime",
			capability:  ProgramCapabilities{Backend: ExecutionBackendMachineCode},
			wantBackend: ExecutionBackendMachineCode,
			wantTier:    ExecutionTierMachineCodeHelper,
		},
		{
			name: "explicit-helper",
			capability: ProgramCapabilities{
				Backend:       ExecutionBackendMachineCode,
				ExecutionTier: ExecutionTierMachineCodeHelper,
			},
			wantBackend: ExecutionBackendMachineCode,
			wantTier:    ExecutionTierMachineCodeHelper,
		},
		{
			name: "future-slots",
			capability: ProgramCapabilities{
				Backend:       ExecutionBackendMachineCode,
				ExecutionTier: ExecutionTierMachineCodeSlots,
			},
			wantBackend: ExecutionBackendMachineCode,
			wantTier:    ExecutionTierMachineCodeSlots,
		},
		{
			name: "instruction-engine-cannot-claim-slots",
			capability: ProgramCapabilities{
				Backend:       ExecutionBackendInstructionEngine,
				ExecutionTier: ExecutionTierMachineCodeSlots,
			},
			wantError: true,
		},
		{
			name:       "unknown-backend",
			capability: ProgramCapabilities{Backend: "llvm"},
			wantError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capability := test.capability
			err := normalizeExecutionTier(&capability)
			if (err != nil) != test.wantError {
				t.Fatalf("normalize error = %v, wantError=%v", err, test.wantError)
			}
			if test.wantError {
				return
			}
			if capability.Backend != test.wantBackend || capability.ExecutionTier != test.wantTier {
				t.Fatalf("normalized capability = %#v, want backend=%q tier=%q",
					capability, test.wantBackend, test.wantTier)
			}
		})
	}
}

func TestDecodeIndexedReuseResetsTerminalsAndProtocol(t *testing.T) {
	errorFrame := indexedTestFrame{kind: 5, category: 3, codec: 5, payload: []byte(`{"code":"E_TEST"}`)}
	frames := [][]indexedTestFrame{
		{
			{kind: 1, category: 1, codec: PayloadPointJSON, payload: []byte(`{"result":1}`)},
			{kind: 2, category: 1, codec: PayloadPointJSON, payload: []byte(`{"child":1}`)},
			{kind: 4, category: 3, codec: 4, payload: []byte("stdout")},
			{recordIndex: ^uint64(0), kind: 3, category: 4, codec: 5, payload: []byte(`{"event":1}`)},
			{kind: 6, status: TerminalOK},
		},
		{{kind: 6, status: TerminalDropped}},
		{{kind: 7, category: 1, codec: PayloadMutationDelta, payload: []byte("delta")}, errorFrame, {kind: 6, status: TerminalError, flags: indexedTerminalCommitPrefixErrorFlag}},
		{errorFrame, {kind: 6, status: TerminalCancelled}},
	}
	scratch := Batch{Static: &StaticBatch{validated: true}}
	for range 5 {
		for _, frames := range frames {
			input := encodeIndexedTestBatch(t, frames...)
			last := frames[len(frames)-1].status
			if last == TerminalOK || last == TerminalDropped {
				binary.LittleEndian.PutUint16(input[6:8], 0)
			}
			expected, err := decodeIndexed(input)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := decodeIndexedInto(input, &scratch)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual.Records, expected.Records) || actual.Static != nil || actual.Protocol != expected.Protocol || len(actual.Diagnostics) != len(expected.Diagnostics) {
				t.Fatalf("reused result differs: %#v, want %#v", actual, expected)
			}
			// A terminal already seen in the current call must still be rejected.
			duplicate := append(append([]byte(nil), input...), input[len(input)-frameHeaderBytes:]...)
			binary.LittleEndian.PutUint32(duplicate[12:16], uint32(len(frames)+1))
			binary.LittleEndian.PutUint32(duplicate[16:20], uint32(len(duplicate)-headerBytes))
			if _, err := decodeIndexedInto(duplicate, &scratch); err == nil {
				t.Fatal("duplicate terminal accepted")
			}
		}
	}
	scratch.encoded = make([]byte, 1<<20+1)
	scratch.ResetForReuse()
	if cap(scratch.encoded) != 0 || cap(scratch.Records) != 0 {
		t.Fatal("oversized result storage retained")
	}
}
