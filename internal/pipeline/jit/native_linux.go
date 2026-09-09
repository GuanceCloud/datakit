// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

type nativeBuffer struct {
	Pointer  *byte
	Length   uintptr
	Capacity uintptr
}

type processOptions struct {
	StructSize  uint32
	Flags       uint32
	Fuel        uint64
	TimeoutMS   uint64
	RequestID   uint64
	OutputCodec uint32
	Reserved    uint32
}

// SyscallN requires stable heap addresses for pointer arguments. Reuse the
// descriptors rather than allocating options and output for every invocation.
// Buffers keep their existing ownership; a frame is cleared only after decode
// and native free finish, before the program's read lease is released.
type nativeCallState struct {
	options processOptions
	output  nativeBuffer
}

var nativeCallStates = sync.Pool{New: func() any { return new(nativeCallState) }}

func acquireNativeCallState(scratch *Batch) (*nativeCallState, bool) {
	if scratch != nil {
		if state, ok := scratch.nativeCall.(*nativeCallState); ok {
			return state, false
		}
		state := new(nativeCallState)
		scratch.nativeCall = state
		return state, false
	}
	return nativeCallStates.Get().(*nativeCallState), true
}

func releaseNativeCallState(state *nativeCallState, pooled bool) {
	*state = nativeCallState{}
	if pooled {
		nativeCallStates.Put(state)
	}
}

type hostCompatFlatV1 struct {
	StructSize uint32
	ABIVersion uint16
	Flags      uint16
	UserData   uintptr
	Invoke     uintptr
}

type nativeHostRegistration struct {
	token uintptr
	host  *HostCompat
}

const (
	hostCompatCallbackOK              int32 = 0
	hostCompatCallbackInvalidArgument int32 = 1
	hostCompatCallbackBufferTooSmall  int32 = 2
	hostCompatCallbackHostError       int32 = 3
)

var (
	hostCompatHosts     sync.Map
	hostCompatNextToken atomic.Uint64
	hostCompatCallback  = purego.NewCallback(invokeHostCompatCallback)
)

type nativeRuntime struct {
	cancelPoolMu          sync.Mutex
	cancelPool            []*nativeCancellation
	staticCancellableInto func(uintptr, *byte, uintptr, *processOptions, *nativeBuffer, uintptr, *byte, uintptr) int32
	services              uintptr
	servicesCreate        func(*byte, uintptr, *uintptr, *nativeBuffer) int32
	servicesDestroy       func(uintptr) int32
	compileWithServices   func(*byte, uintptr, *hostCompatFlatV1, uintptr, *uintptr, *nativeBuffer) int32
	clockDriver           *cacheTickDriver
	manualCacheTicks      bool // deterministic ABI tests only; normal runtimes auto-drive
	closeDone             chan struct{}
	closeErr              error
	cacheTicksEnable      func(uintptr, uint64) int32
	cacheTick             func(uintptr) int32
	aggregatePoll         func(uintptr, int64, uint32, *nativeBuffer) int32
	aggregatePollV2       func(uintptr, int64, uint32, *nativeBuffer) int32
	cancelCreate          func(*uintptr) int32
	cancelRequest         func(uintptr) int32
	cancelDestroy         func(uintptr) int32
	processCancellable    func(uintptr, uint32, *byte, uintptr, *processOptions, *nativeBuffer, uintptr) int32
	mu                    sync.Mutex
	handle                uintptr
	closed                bool
	compile               func(*byte, uintptr, *byte, uintptr, *uintptr, *nativeBuffer) int32
	compileV2             func(*byte, uintptr, *byte, uintptr, *hostCompatFlatV1, *uintptr, *nativeBuffer) int32
	compileModules        func(*byte, uintptr, *hostCompatFlatV1, *uintptr, *nativeBuffer) int32
	host                  *nativeHostRegistration
	projection            func(uintptr, *nativeBuffer) int32
	capabilities          func(uintptr, *nativeBuffer) int32
	staticDesc            func(uintptr, *nativeBuffer) int32
	process               func(uintptr, uint32, *byte, uintptr, *processOptions, *nativeBuffer) int32
	indexedInto           func(uintptr, uint32, *byte, uintptr, *processOptions, *nativeBuffer, uintptr, *byte, uintptr) int32
	staticInto            func(uintptr, *byte, uintptr, *processOptions, *nativeBuffer, *byte, uintptr) int32
	staticRun             func(uintptr, *byte, uintptr, *processOptions, *nativeBuffer) int32
	destroy               func(uintptr) int32
	free                  func(*byte, uintptr, uintptr)
	aggregateSinkMu       sync.RWMutex
	aggregateSink         func([]Point) error
}

type nativeProgram struct {
	capabilitiesOnce           sync.Once
	capabilitiesSnapshot       ProgramCapabilities
	capabilitiesErr            error
	clockRegistration          uint64
	aggregateRetryRegistration uint64
	aggregateRetryMu           sync.Mutex
	clockErr                   error
	clockErrMu                 sync.Mutex
	mu                         sync.RWMutex
	runtime                    *nativeRuntime
	handle                     uintptr
	projection                 InputProjection
	static                     *StaticOutputSchema
	closed                     bool
	aggregateMu                sync.Mutex
	pendingAggregates          []Point
}

func openRuntime(path string, host *HostCompat) (Runtime, error) {
	if path == "" {
		return nil, errors.New("JIT runtime path is empty")
	}
	handle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return nil, fmt.Errorf("open JIT runtime: %w", err)
	}
	native := &nativeRuntime{handle: handle}
	if err := register(&native.compile, handle, "pp_jit_compile_flat_v1"); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	registerOptional(&native.compileV2, handle, "pp_jit_compile_configured_flat_v2")
	registerOptional(&native.compileModules, handle, "pp_jit_compile_modules_flat_v1")
	registerOptional(&native.servicesCreate, handle, "pp_jit_services_create_flat_v1")
	registerOptional(&native.servicesDestroy, handle, "pp_jit_services_destroy_v1")
	registerOptional(&native.compileWithServices, handle, "pp_jit_compile_modules_with_services_flat_v2")
	registerOptional(&native.cacheTicksEnable, handle, "pp_jit_cache_ticks_enable_v1")
	registerOptional(&native.cacheTick, handle, "pp_jit_cache_tick_v1")
	registerOptional(&native.aggregatePoll, handle, "pp_jit_poll_aggregate_events_flat_v1")
	registerOptional(&native.aggregatePollV2, handle, "pp_jit_poll_aggregate_events_flat_v2")
	registerOptional(&native.cancelCreate, handle, "pp_jit_cancellation_create_v1")
	registerOptional(&native.cancelRequest, handle, "pp_jit_cancellation_cancel_v1")
	registerOptional(&native.cancelDestroy, handle, "pp_jit_cancellation_destroy_v1")
	registerOptional(&native.processCancellable, handle, "pp_jit_process_cancellable_flat_v1")
	if err := register(&native.process, handle, "pp_jit_process_raw_batch_flat_v3"); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	if err := register(&native.destroy, handle, "pp_jit_destroy_v1"); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	if err := register(&native.free, handle, "pp_jit_buffer_free_flat_v1"); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	registerOptional(&native.projection, handle, "pp_jit_input_projection_flat_v1")
	registerOptional(&native.capabilities, handle, "pp_jit_program_capabilities_flat_v1")
	hasStaticDesc := registerOptional(&native.staticDesc, handle, "pp_jit_static_output_schema_flat_v1")
	hasStaticRun := registerOptional(&native.staticRun, handle, "pp_jit_process_static_batch_flat_v1")
	registerOptional(&native.staticInto, handle, "pp_jit_process_static_batch_flat_into_v1")
	if err := register(&native.staticCancellableInto, handle, "pp_jit_process_static_cancellable_into_v1"); err != nil {
		_ = purego.Dlclose(handle)
		return nil, err
	}
	registerOptional(&native.indexedInto, handle, "pp_jit_process_indexed_into_v1")
	if hasStaticDesc != hasStaticRun {
		native.staticDesc = nil
		native.staticRun = nil
	}
	if (native.compileV2 != nil || native.compileWithServices != nil) && host != nil && host.flags != 0 {
		native.host = registerHostCompat(host)
	}
	configureNativeCalls(native)
	return native, nil
}

// Only called by the factory constructor before any Runner is published or
// any program is compiled. Host state remains immutable after this point.
func (n *nativeRuntime) bindUnpublishedHost(host *HostCompat) {
	if host != nil && host.flags != 0 {
		n.host = registerHostCompat(host)
	}
}

func (n *nativeRuntime) setAggregateEventHandler(handler func([]Point) error) {
	n.aggregateSinkMu.Lock()
	n.aggregateSink = handler
	n.aggregateSinkMu.Unlock()
}

func (n *nativeRuntime) deliverAggregateEvents(points []Point) error {
	if len(points) == 0 {
		return nil
	}
	n.aggregateSinkMu.RLock()
	handler := n.aggregateSink
	n.aggregateSinkMu.RUnlock()
	if handler == nil {
		return errors.New("JIT aggregate event handler is not configured")
	}
	return handler(points)
}

func register(target any, handle uintptr, name string) error {
	symbol, err := purego.Dlsym(handle, name)
	if err != nil {
		return fmt.Errorf("load JIT symbol %s: %w", name, err)
	}
	purego.RegisterFunc(target, symbol)
	return nil
}

func registerOptional(target any, handle uintptr, name string) bool {
	symbol, err := purego.Dlsym(handle, name)
	if err != nil {
		return false
	}
	purego.RegisterFunc(target, symbol)
	return true
}

func (n *nativeRuntime) configureServices(config ServiceConfig) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed || n.services != 0 {
		return errors.New("JIT services already configured or runtime closed")
	}
	if n.servicesCreate == nil || n.servicesDestroy == nil || n.compileWithServices == nil {
		return errors.New("JIT runtime lacks instance service ABI")
	}
	// JSON arrays, not null; marshal copies the caller's snapshot synchronously.
	if config.HTTP.HostWhitelist == nil {
		config.HTTP.HostWhitelist = []string{}
	}
	if config.HTTP.CIDRWhitelist == nil {
		config.HTTP.CIDRWhitelist = []string{}
	}
	payload, err := json.Marshal(config)
	if err != nil {
		return err
	}
	var handle uintptr
	var output nativeBuffer
	status := n.servicesCreate(bytesPointer(payload), uintptr(len(payload)), &handle, &output)
	runtime.KeepAlive(payload)
	detail, err := n.copyAndFree(output)
	if err != nil || status != 0 || handle == 0 {
		if handle != 0 {
			_ = n.servicesDestroy(handle)
		}
		if err != nil {
			return err
		}
		return nativeError("configure services", status, detail)
	}
	n.services = handle
	return nil
}

func (n *nativeRuntime) Compile(source, profile string) (Program, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return nil, errors.New("JIT runtime is closed")
	}
	if n.services != 0 {
		if profile != "pipeline-go-1.4.3-datakit" {
			return nil, errors.New("instance services require DataKit profile")
		}
		return n.compileModulesLocked("main.p", map[string]string{"main.p": source})
	}
	sourceBytes, profileBytes := []byte(source), []byte(profile)
	var handle uintptr
	var output nativeBuffer
	var status int32
	if n.compileV2 != nil && n.host != nil {
		configuration := hostCompatFlatV1{
			StructSize: uint32(unsafe.Sizeof(hostCompatFlatV1{})),
			ABIVersion: hostCompatABIVersion,
			Flags:      n.host.host.flags,
			UserData:   n.host.token,
			Invoke:     hostCompatCallback,
		}
		status = n.compileV2(bytesPointer(sourceBytes), uintptr(len(sourceBytes)), bytesPointer(profileBytes), uintptr(len(profileBytes)), &configuration, &handle, &output)
		runtime.KeepAlive(configuration)
	} else {
		status = n.compile(bytesPointer(sourceBytes), uintptr(len(sourceBytes)), bytesPointer(profileBytes), uintptr(len(profileBytes)), &handle, &output)
	}
	runtime.KeepAlive(sourceBytes)
	runtime.KeepAlive(profileBytes)
	return n.finishCompile(handle, status, output, profile == "pipeline-go-1.4.3-datakit")
}

// CompileModules borrows sources during compilation; callers must not mutate
// the map concurrently. The native program owns an immutable compiled snapshot.
func (n *nativeRuntime) CompileModules(entry string, sources map[string]string) (Program, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.compileModulesLocked(entry, sources)
}

func (n *nativeRuntime) compileModulesLocked(entry string, sources map[string]string) (Program, error) {
	if n.closed {
		return nil, errors.New("JIT runtime is closed")
	}
	if n.compileModules == nil && n.services == 0 {
		return nil, errors.New("JIT runtime lacks module snapshot compilation")
	}
	if len(sources) > 128 {
		return nil, errors.New("JIT module count exceeds 128")
	}
	type module struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	snapshot := struct {
		Entry   string   `json:"entry"`
		Modules []module `json:"modules"`
	}{Entry: entry}
	names := make([]string, 0, len(sources))
	total := 0
	for name, source := range sources {
		if len(source) > 1024*1024 {
			return nil, errors.New("JIT module source exceeds 1 MiB")
		}
		total += len(source)
		if total > 8*1024*1024 {
			return nil, errors.New("JIT module sources exceed 8 MiB")
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		snapshot.Modules = append(snapshot.Modules, module{name, sources[name]})
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var configuration *hostCompatFlatV1
	if n.host != nil {
		configuration = &hostCompatFlatV1{StructSize: uint32(unsafe.Sizeof(hostCompatFlatV1{})),
			ABIVersion: hostCompatABIVersion, Flags: n.host.host.flags,
			UserData: n.host.token, Invoke: hostCompatCallback}
	}
	var handle uintptr
	var output nativeBuffer
	var status int32
	if n.services != 0 {
		status = n.compileWithServices(bytesPointer(payload), uintptr(len(payload)), configuration, n.services, &handle, &output)
	} else {
		status = n.compileModules(bytesPointer(payload), uintptr(len(payload)), configuration, &handle, &output)
	}
	runtime.KeepAlive(payload)
	runtime.KeepAlive(configuration)
	return n.finishCompile(handle, status, output, true)
}

// Caller holds the runtime lock, including all cleanup on descriptor failures.
func (n *nativeRuntime) finishCompile(handle uintptr, status int32, output nativeBuffer, autoClock bool) (Program, error) {
	detail, err := n.copyAndFree(output)
	if err != nil {
		if handle != 0 {
			_ = n.destroy(handle)
		}
		return nil, err
	}
	if status != 0 || handle == 0 {
		if handle != 0 {
			_ = n.destroy(handle)
		}
		return nil, nativeError("compile", status, detail)
	}
	projection, err := n.inputProjection(handle)
	if err != nil {
		_ = n.destroy(handle)
		return nil, err
	}
	static, err := n.staticOutputSchema(handle)
	if err != nil {
		_ = n.destroy(handle)
		return nil, err
	}
	p := &nativeProgram{runtime: n, handle: handle, projection: projection, static: static}
	if autoClock && !n.manualCacheTicks && n.cacheTicksEnable != nil && n.cacheTick != nil {
		if status := n.cacheTicksEnable(handle, uint64(time.Second)); status != 0 {
			_ = n.destroy(handle)
			return nil, nativeError("enable automatic cache ticks", status, nil)
		}
		if n.clockDriver == nil {
			n.clockDriver = newCacheTickDriver()
		}
		id, ok := n.clockDriver.register(time.Second, p.scheduledCacheTick)
		if !ok {
			_ = n.destroy(handle)
			return nil, errors.New("JIT cache driver closed during compile")
		}
		p.clockRegistration = id
	}
	return p, nil
}

func (n *nativeRuntime) inputProjection(handle uintptr) (InputProjection, error) {
	if n.projection == nil {
		return InputProjection{}, nil
	}
	var output nativeBuffer
	status := n.projection(handle, &output)
	detail, err := n.copyAndFree(output)
	if err != nil {
		return InputProjection{}, err
	}
	if status != 0 {
		return InputProjection{}, nativeError("query input projection", status, detail)
	}
	var descriptor struct {
		Version int      `json:"version"`
		Mode    string   `json:"mode"`
		Keys    []string `json:"keys"`
	}
	if err := json.Unmarshal(detail, &descriptor); err != nil {
		return InputProjection{}, fmt.Errorf("decode JIT input projection: %w", err)
	}
	if descriptor.Version != 1 {
		return InputProjection{}, fmt.Errorf("unsupported JIT input projection version %d", descriptor.Version)
	}
	switch descriptor.Mode {
	case "all":
		return InputProjection{}, nil
	case "keys":
		if len(descriptor.Keys) == 0 {
			return InputProjection{}, errors.New("JIT key projection is empty")
		}
		return newKeyProjection(descriptor.Keys), nil
	default:
		return InputProjection{}, fmt.Errorf("unsupported JIT input projection mode %q", descriptor.Mode)
	}
}

func (n *nativeRuntime) staticOutputSchema(handle uintptr) (*StaticOutputSchema, error) {
	if n.staticDesc == nil || n.staticRun == nil {
		return nil, nil
	}
	var output nativeBuffer
	status := n.staticDesc(handle, &output)
	detail, err := n.copyAndFree(output)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, nativeError("query static output schema", status, detail)
	}
	var descriptor struct {
		Version    int      `json:"version"`
		Mode       string   `json:"mode"`
		SourceHash string   `json:"source_hash"`
		SchemaHash string   `json:"schema_hash"`
		Slots      []string `json:"slots"`
	}
	if err := json.Unmarshal(detail, &descriptor); err != nil {
		return nil, fmt.Errorf("decode JIT static output schema: %w", err)
	}
	if descriptor.Version != 1 {
		return nil, fmt.Errorf("unsupported JIT static output schema version %d", descriptor.Version)
	}
	if descriptor.Mode == "none" {
		return nil, nil
	}
	if (descriptor.Mode != "slots" || len(descriptor.Slots) == 0) && (descriptor.Mode != "dynamic" || len(descriptor.Slots) != 0) {
		return nil, fmt.Errorf("unsupported JIT static output schema mode %q", descriptor.Mode)
	}
	if len(descriptor.SourceHash) != 64 {
		return nil, errors.New("JIT static output source hash is invalid")
	}
	if _, err := hex.DecodeString(descriptor.SourceHash); err != nil {
		return nil, fmt.Errorf("decode JIT static output source hash: %w", err)
	}
	hash, err := hex.DecodeString(descriptor.SchemaHash)
	if err != nil || len(hash) != 32 {
		return nil, errors.New("JIT static output schema hash is invalid")
	}
	seen := make(map[string]struct{}, len(descriptor.Slots))
	for _, key := range descriptor.Slots {
		if key == "" {
			return nil, errors.New("JIT static output schema contains an empty key")
		}
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("JIT static output schema contains duplicate key %q", key)
		}
		seen[key] = struct{}{}
	}
	schema := &StaticOutputSchema{
		Dynamic:    descriptor.Mode == "dynamic",
		SourceHash: descriptor.SourceHash,
		Keys:       append([]string(nil), descriptor.Slots...),
	}
	copy(schema.Hash[:], hash)
	return schema, nil
}

func (n *nativeRuntime) Close() error {
	n.mu.Lock()
	if n.closed {
		done := n.closeDone
		n.mu.Unlock()
		if done != nil {
			<-done
		}
		n.mu.Lock()
		defer n.mu.Unlock()
		return n.closeErr
	}
	n.closed = true
	n.closeDone = make(chan struct{})
	driver := n.clockDriver
	n.mu.Unlock()
	// A dispatched callback takes n.mu; never wait for it while holding n.mu.
	if driver != nil {
		driver.Close()
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closeErr = errors.Join(n.closeErr, n.closeCancellationPool())
	if n.services != 0 {
		if status := n.servicesDestroy(n.services); status != 0 {
			n.closeErr = fmt.Errorf("destroy JIT services: status %d", status)
		}
		n.services = 0
	}
	err := purego.Dlclose(n.handle)
	n.handle = 0
	if n.host != nil {
		n.host.close()
		n.host = nil
	}
	if err != nil {
		n.closeErr = errors.Join(n.closeErr, fmt.Errorf("close JIT runtime: %w", err))
	}
	close(n.closeDone)
	return n.closeErr
}

func (p *nativeProgram) ProcessIndexed(input []byte) (Batch, error) {
	return p.ProcessIndexedInto(input, nil)
}

// ProcessIndexedInto borrows caller-owned result scratch for this synchronous
// call. Ordinary ProcessIndexed results retain their independent ownership.
func (p *nativeProgram) ProcessIndexedInto(input []byte, scratch *Batch) (result Batch, processErr error) {
	defer func() { processErr = classifyProcessError(processErr) }()
	p.mu.RLock()
	defer p.mu.RUnlock()
	if err := p.cacheClockError(); err != nil {
		return Batch{}, err
	}
	if p.closed || p.handle == 0 {
		return Batch{}, errors.New("JIT program is closed")
	}
	call, pooled := acquireNativeCallState(scratch)
	defer releaseNativeCallState(call, pooled)
	options, output := &call.options, &call.output
	options.StructSize = uint32(unsafe.Sizeof(processOptions{}))
	var status int32
	var inline []byte
	if p.static != nil && p.runtime.staticInto != nil && scratch != nil {
		if cap(scratch.inline) < 4096 {
			scratch.inline = make([]byte, 4096)
		}
		inline = scratch.inline[:4096]
		status = p.runtime.staticInto(p.handle, bytesPointer(input), uintptr(len(input)), options, output, bytesPointer(inline), uintptr(len(inline)))
	} else if p.static != nil && p.runtime.staticRun != nil {
		status = p.runtime.staticRun(p.handle, bytesPointer(input), uintptr(len(input)), options, output)
	} else {
		options.Flags = 4
		if p.runtime.aggregatePollV2 != nil {
			options.Flags |= 8
		}
		options.OutputCodec = PayloadMutationDelta
		if p.runtime.indexedInto != nil && scratch != nil {
			if cap(scratch.inline) < 4096 {
				scratch.inline = make([]byte, 4096)
			}
			inline = scratch.inline[:4096]
			status = p.runtime.indexedInto(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), options, output, 0, bytesPointer(inline), uintptr(len(inline)))
		} else {
			status = p.runtime.process(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), options, output)
		}
	}
	runtime.KeepAlive(input)
	if len(inline) != 0 && output.Pointer == bytesPointer(inline) {
		if output.Length > uintptr(len(inline)) || output.Capacity != uintptr(len(inline)) {
			return Batch{}, errors.New("invalid JIT inline result buffer")
		}
		view := inline[:output.Length]
		if status != 0 {
			return Batch{}, nativeError("process indexed batch", status, view)
		}
		var batch Batch
		var err error
		if p.static != nil {
			batch, err = decodeStaticBatchInto(view, *p.static, scratch)
		} else {
			batch, err = decodeIndexedInto(view, scratch)
		}
		runtime.KeepAlive(inline)
		return batch, err
	}
	if status != 0 {
		encoded, err := p.runtime.copyAndFree(*output)
		if err != nil {
			return Batch{}, err
		}
		return Batch{}, nativeError("process indexed batch", status, encoded)
	}
	if p.static != nil {
		return p.runtime.decodeStaticAndFreeInto(*output, *p.static, scratch)
	}
	if scratch != nil {
		defer p.runtime.free(output.Pointer, output.Length, output.Capacity)
		view, err := nativeBufferView(*output)
		if err != nil {
			return Batch{}, err
		}
		scratch.encoded = append(scratch.encoded[:0], view...)
		return decodeIndexedInto(scratch.encoded, scratch)
	}
	encoded, err := p.runtime.copyAndFree(*output)
	if err != nil {
		return Batch{}, err
	}
	return decodeIndexed(encoded)
}

// ProcessIndexedContext uses indexed delta output even for a static program.
// Cancellation is cooperative; keep the returned native terminal records rather
// than replacing them with ctx.Err(), which would discard partial-work evidence.
func (p *nativeProgram) ProcessIndexedContext(ctx context.Context, input []byte) (Batch, error) {
	return p.ProcessIndexedContextInto(ctx, input, nil)
}

func (p *nativeProgram) ProcessIndexedContextInto(ctx context.Context, input []byte, scratch *Batch) (result Batch, processErr error) {
	defer func() { processErr = classifyProcessError(processErr) }()
	if ctx == nil {
		return Batch{}, errors.New("nil JIT process context")
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	timeout, limited := RecordTimeout(ctx)
	cancellable := ctx.Done() != nil
	if !cancellable && !limited {
		return p.ProcessIndexedInto(input, scratch)
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if err := p.cacheClockError(); err != nil {
		return Batch{}, err
	}
	if p.closed || p.handle == 0 {
		return Batch{}, errors.New("JIT program is closed")
	}
	n := p.runtime
	var token uintptr
	if cancellable {
		if n.cancelCreate == nil || n.cancelRequest == nil || n.cancelDestroy == nil || n.processCancellable == nil {
			return Batch{}, errors.New("JIT runtime lacks cooperative cancellation")
		}
		var release func()
		var err error
		token, release, err = n.watchCancellation(ctx)
		if err != nil {
			return Batch{}, err
		}
		defer release()
	}
	call, pooled := acquireNativeCallState(scratch)
	defer releaseNativeCallState(call, pooled)
	options, output := &call.options, &call.output
	options.StructSize = uint32(unsafe.Sizeof(processOptions{}))
	if limited {
		options.Flags = 2 // HAS_TIMEOUT: per-record budget, not the request deadline.
		if timeout > 0 {
			options.TimeoutMS = uint64(timeout / time.Millisecond)
			if timeout%time.Millisecond != 0 {
				options.TimeoutMS++
			}
		}
	}
	static := p.static != nil
	var inline []byte
	if scratch != nil && (static || n.indexedInto != nil) {
		if cap(scratch.inline) < 4096 {
			scratch.inline = make([]byte, 4096)
		}
		inline = scratch.inline[:4096]
	}
	var status int32
	if static {
		status = n.staticCancellableInto(p.handle, bytesPointer(input), uintptr(len(input)), options, output, token, bytesPointer(inline), uintptr(len(inline)))
	} else {
		options.Flags |= 4
		if n.aggregatePollV2 != nil {
			options.Flags |= 8
		}
		options.OutputCodec = PayloadMutationDelta
		if n.indexedInto != nil {
			status = n.indexedInto(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), options, output, token, bytesPointer(inline), uintptr(len(inline)))
		} else if cancellable {
			status = n.processCancellable(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), options, output, token)
		} else {
			status = n.process(p.handle, CodecFlatPoint, bytesPointer(input), uintptr(len(input)), options, output)
		}
	}
	runtime.KeepAlive(input)
	borrowed := len(inline) != 0 && output.Pointer == bytesPointer(inline)
	var view []byte
	var err error
	if borrowed {
		if output.Length > uintptr(len(inline)) || output.Capacity != uintptr(len(inline)) {
			return Batch{}, errors.New("invalid JIT inline controlled buffer")
		}
		view = inline[:output.Length]
	} else {
		// Decode synchronously while the native output is alive. Indexed records
		// borrow their input wire, so retain an owned copy before freeing native.
		defer n.free(output.Pointer, output.Length, output.Capacity)
		view, err = nativeBufferView(*output)
		if err != nil {
			return Batch{}, err
		}
		if !static && status == 0 {
			if scratch != nil {
				scratch.encoded = append(scratch.encoded[:0], view...)
				view = scratch.encoded
			} else {
				view = append([]byte(nil), view...)
			}
		}
	}
	if status != 0 {
		return Batch{}, nativeError("process controlled batch", status, view)
	}
	if static {
		result, err = decodeStaticBatchInto(view, *p.static, scratch)
	} else {
		result, err = decodeIndexedInto(view, scratch)
	}
	runtime.KeepAlive(inline)
	return result, err
}

func (p *nativeProgram) InputProjection() InputProjection {
	return p.projection
}

func (p *nativeProgram) Capabilities() (ProgramCapabilities, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed || p.handle == 0 {
		return ProgramCapabilities{}, errors.New("JIT program is closed")
	}
	p.capabilitiesOnce.Do(func() {
		p.capabilitiesSnapshot, p.capabilitiesErr = p.readCapabilitiesLocked()
	})
	// Callers may modify descriptors; the program-owned snapshot is immutable.
	result := p.capabilitiesSnapshot
	result.HostCalls = append([]string(nil), result.HostCalls...)
	if result.AggregateEvents != nil {
		aggregate := *result.AggregateEvents
		result.AggregateEvents = &aggregate
	}
	if result.Stateless != nil {
		stateless := *result.Stateless
		result.Stateless = &stateless
	}
	return result, p.capabilitiesErr
}

// The caller holds p.mu for the complete native read and output decode.
func (p *nativeProgram) readCapabilitiesLocked() (ProgramCapabilities, error) {
	if p.runtime.capabilities == nil {
		return ProgramCapabilities{
			ExecutionMode: "native",
			Backend:       ExecutionBackendInstructionEngine,
			ExecutionTier: ExecutionTierInstructionEngine,
			InputProjection: CapabilityMode{
				Mode: "unknown",
			},
			StaticOutput: CapabilityMode{Mode: "unknown"},
		}, nil
	}
	var output nativeBuffer
	status := p.runtime.capabilities(p.handle, &output)
	detail, err := p.runtime.copyAndFree(output)
	if err != nil {
		return ProgramCapabilities{}, err
	}
	if status != 0 {
		return ProgramCapabilities{}, nativeError("query program capabilities", status, detail)
	}
	var capabilities ProgramCapabilities
	if err := json.Unmarshal(detail, &capabilities); err != nil {
		return ProgramCapabilities{}, fmt.Errorf("decode JIT program capabilities: %w", err)
	}
	if capabilities.Version != 1 {
		return ProgramCapabilities{}, fmt.Errorf("unsupported JIT program capabilities version %d", capabilities.Version)
	}
	if capabilities.AggregateEvents == nil {
		return ProgramCapabilities{}, errors.New("JIT capability descriptor lacks aggregate event contract")
	}
	if *capabilities.AggregateEvents && p.runtime.aggregatePoll == nil && p.runtime.aggregatePollV2 == nil {
		return ProgramCapabilities{}, errors.New("JIT aggregation program lacks aggregate event drain ABI")
	}
	if len(capabilities.SourceHash) != 64 {
		return ProgramCapabilities{}, errors.New("JIT program capability source hash is invalid")
	}
	if _, err := hex.DecodeString(capabilities.SourceHash); err != nil {
		return ProgramCapabilities{}, fmt.Errorf("decode JIT program capability source hash: %w", err)
	}
	if capabilities.ExecutionMode != "native" && capabilities.ExecutionMode != "native_with_host" {
		return ProgramCapabilities{}, fmt.Errorf("unsupported JIT execution mode %q", capabilities.ExecutionMode)
	}
	if err := normalizeExecutionTier(&capabilities); err != nil {
		return ProgramCapabilities{}, err
	}
	if capabilities.InputProjection.Mode != "all" && capabilities.InputProjection.Mode != "keys" {
		return ProgramCapabilities{}, fmt.Errorf("unsupported JIT capability projection mode %q", capabilities.InputProjection.Mode)
	}
	if capabilities.StaticOutput.Mode != "none" && capabilities.StaticOutput.Mode != "slots" {
		return ProgramCapabilities{}, fmt.Errorf("unsupported JIT capability static output mode %q", capabilities.StaticOutput.Mode)
	}
	capabilities.DescriptorPresent = true
	return capabilities, nil
}

func (p *nativeProgram) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	if p.clockRegistration != 0 {
		p.runtime.clockDriver.unregister(p.clockRegistration)
	}
	p.aggregateRetryMu.Lock()
	if p.aggregateRetryRegistration != 0 {
		p.runtime.clockDriver.unregister(p.aggregateRetryRegistration)
		p.aggregateRetryRegistration = 0
	}
	p.aggregateRetryMu.Unlock()
	var flushErr error
	if p.runtime.aggregatePoll != nil || p.runtime.aggregatePollV2 != nil {
		p.runtime.mu.Lock()
		points, err := p.drainAggregateEventsLocked(time.Now(), true)
		p.runtime.mu.Unlock()
		if err == nil {
			err = p.deliverAggregateEvents(points)
		}
		flushErr = err
	}
	status := p.runtime.destroy(p.handle)
	p.handle = 0
	if status != 0 {
		return errors.Join(flushErr, nativeError("destroy", status, nil))
	}
	return flushErr
}

func (p *nativeProgram) scheduledCacheTick() {
	p.mu.RLock()
	n := p.runtime
	n.mu.Lock()
	p.clockErrMu.Lock()
	if p.closed || n.closed || p.clockErr != nil {
		p.clockErrMu.Unlock()
		n.mu.Unlock()
		p.mu.RUnlock()
		return
	}
	if status := n.cacheTick(p.handle); status != 0 {
		// Fail subsequent execution rather than silently keeping stale cache
		// entries forever. The pipeline's existing native error path logs this.
		p.clockErr = nativeError("scheduled cache tick", status, nil)
	}
	p.clockErrMu.Unlock()
	n.mu.Unlock()
	p.mu.RUnlock()
	if p.cacheClockError() != nil || (n.aggregatePoll == nil && n.aggregatePollV2 == nil) {
		return
	}
	// Retry already-drained events before removing more buckets from Rust. A
	// persistent upload outage therefore cannot grow the Go pending queue on
	// every tick; not-yet-drained groups remain under Rust state budgets.
	if err := p.deliverAggregateEvents(nil); err != nil {
		p.scheduleAggregateRetry()
		return
	}
	points, err := p.DrainAggregateEvents(time.Now(), false)
	if err == nil {
		err = p.deliverAggregateEvents(points)
	}
	if err != nil {
		// Retain failed deliveries and retry on the next shared-clock tick.
		// Native execution remains available; setting clockErr here would stop
		// both retries and new aggregation input after a transient upload error.
		p.scheduleAggregateRetry()
		return
	}
}

const aggregateRetryInterval = 250 * time.Millisecond

// Failed uploads are Go-owned pending data and must not depend solely on the
// one-second Rust state clock for another attempt. A separate registration on
// the shared driver retries only the pending queue; it never advances cache or
// aggregation time. aggregateMu serializes it with normal ticks, preventing a
// successful batch from being delivered twice.
func (p *nativeProgram) scheduleAggregateRetry() {
	// Hold the lifecycle lease through registration. Close must either remove
	// this registration or finish first and make this a no-op.
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return
	}
	p.runtime.mu.Lock()
	closed := p.runtime.closed
	p.runtime.mu.Unlock()
	if closed {
		return
	}
	p.aggregateRetryMu.Lock()
	defer p.aggregateRetryMu.Unlock()
	if p.aggregateRetryRegistration != 0 || p.runtime.clockDriver == nil {
		return
	}
	id, ok := p.runtime.clockDriver.register(aggregateRetryInterval, p.scheduledAggregateRetry)
	if ok {
		p.aggregateRetryRegistration = id
	}
}

func (p *nativeProgram) scheduledAggregateRetry() {
	p.mu.RLock()
	p.runtime.mu.Lock()
	closed := p.closed || p.runtime.closed
	p.runtime.mu.Unlock()
	p.mu.RUnlock()
	if closed {
		return
	}
	_ = p.deliverAggregateEvents(nil)

	// Keep the registration if a normal clock tick raced with this successful
	// attempt and installed a new failed batch before we examined the queue.
	p.aggregateRetryMu.Lock()
	p.aggregateMu.Lock()
	empty := len(p.pendingAggregates) == 0
	p.aggregateMu.Unlock()
	if empty && p.aggregateRetryRegistration != 0 {
		id := p.aggregateRetryRegistration
		p.aggregateRetryRegistration = 0
		p.runtime.clockDriver.unregister(id)
	}
	p.aggregateRetryMu.Unlock()
}

func (p *nativeProgram) deliverAggregateEvents(points []Point) error {
	p.aggregateMu.Lock()
	defer p.aggregateMu.Unlock()
	if len(points) != 0 {
		if len(p.pendingAggregates) > maxFrames-len(points) {
			return fmt.Errorf("JIT aggregate pending queue exceeds %d points", maxFrames)
		}
		p.pendingAggregates = append(p.pendingAggregates, points...)
	}
	if len(p.pendingAggregates) == 0 {
		return nil
	}
	if err := p.runtime.deliverAggregateEvents(p.pendingAggregates); err != nil {
		return err
	}
	p.pendingAggregates = nil
	return nil
}

func (p *nativeProgram) cacheClockError() error {
	p.clockErrMu.Lock()
	defer p.clockErrMu.Unlock()
	return p.clockErr
}

// Explicit opt-in until the shared driver is installed. Lock the runtime too:
// even a late scheduler call must not enter an already dlclosed library.
func (p *nativeProgram) enableCacheTicks(intervalNS uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.runtime.mu.Lock()
	defer p.runtime.mu.Unlock()
	if p.closed || p.runtime.closed {
		return errors.New("JIT cache clock is closed")
	}
	if p.runtime.cacheTicksEnable == nil || p.runtime.cacheTick == nil {
		return errors.New("JIT runtime lacks cache tick ABI")
	}
	if status := p.runtime.cacheTicksEnable(p.handle, intervalNS); status != 0 {
		return nativeError("enable cache ticks", status, nil)
	}
	return nil
}

func (p *nativeProgram) advanceCacheTick() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	p.runtime.mu.Lock()
	defer p.runtime.mu.Unlock()
	if p.closed || p.runtime.closed {
		return errors.New("JIT cache clock is closed")
	}
	if p.runtime.cacheTick == nil {
		return errors.New("JIT runtime lacks cache tick ABI")
	}
	if status := p.runtime.cacheTick(p.handle); status != 0 {
		return nativeError("advance cache tick", status, nil)
	}
	return nil
}

func (p *nativeProgram) DrainAggregateEvents(now time.Time, force bool) ([]Point, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	p.runtime.mu.Lock()
	defer p.runtime.mu.Unlock()
	if p.closed || p.runtime.closed || p.handle == 0 {
		return nil, errors.New("JIT aggregate event drain is closed")
	}
	if p.runtime.aggregatePoll == nil && p.runtime.aggregatePollV2 == nil {
		return nil, errors.New("JIT runtime lacks aggregate event drain ABI")
	}
	return p.drainAggregateEventsLocked(now, force)
}

// Caller holds the program read/write lock and the runtime lock, preventing
// process, destroy and dlclose from racing the borrowed native buffer.
func (p *nativeProgram) drainAggregateEventsLocked(now time.Time, force bool) ([]Point, error) {
	var flags uint32
	if force {
		flags = 1
	}
	var output nativeBuffer
	poll := p.runtime.aggregatePollV2
	if poll == nil {
		poll = p.runtime.aggregatePoll
	}
	status := poll(p.handle, now.UnixNano(), flags, &output)
	encoded, err := p.runtime.copyAndFree(output)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, nativeError("drain aggregate events", status, encoded)
	}
	var points []Point
	if p.runtime.aggregatePollV2 != nil {
		points, err = DecodeFlatPoints(encoded)
	} else {
		err = json.Unmarshal(encoded, &points)
	}
	if err != nil {
		return nil, fmt.Errorf("decode aggregate event points: %w", err)
	}
	for index := range points {
		if points[index].Time != "" {
			parsed, err := time.Parse(time.RFC3339Nano, points[index].Time)
			if err != nil {
				return nil, fmt.Errorf("decode aggregate event point %d time: %w", index, err)
			}
			points[index].TimeUnixNano = parsed.UnixNano()
		}
	}
	return points, nil
}

func (n *nativeRuntime) copyAndFree(buffer nativeBuffer) ([]byte, error) {
	defer n.free(buffer.Pointer, buffer.Length, buffer.Capacity)
	view, err := nativeBufferView(buffer)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), view...), nil
}

func (n *nativeRuntime) decodeStaticAndFree(buffer nativeBuffer, schema StaticOutputSchema) (Batch, error) {
	return n.decodeStaticAndFreeInto(buffer, schema, nil)
}

func (n *nativeRuntime) decodeStaticAndFreeInto(buffer nativeBuffer, schema StaticOutputSchema, scratch *Batch) (Batch, error) {
	defer n.free(buffer.Pointer, buffer.Length, buffer.Capacity)
	view, err := nativeBufferView(buffer)
	if err != nil {
		return Batch{}, err
	}
	return decodeStaticBatchInto(view, schema, scratch)
}

func nativeBufferView(buffer nativeBuffer) ([]byte, error) {
	if buffer.Length > buffer.Capacity || buffer.Length > staticBatchHeaderSize+maxResultBytes {
		return nil, errors.New("invalid JIT native buffer")
	}
	if buffer.Length == 0 {
		return nil, nil
	}
	if buffer.Pointer == nil {
		return nil, errors.New("nil JIT native buffer")
	}
	return unsafe.Slice(buffer.Pointer, buffer.Length), nil
}

func bytesPointer(value []byte) *byte {
	if len(value) == 0 {
		return nil
	}
	return unsafe.SliceData(value)
}

func nativeError(operation string, status int32, detail []byte) error {
	return &NativeCallError{Operation: operation, Status: status, Detail: string(detail)}
}

func registerHostCompat(host *HostCompat) *nativeHostRegistration {
	for {
		token := uintptr(hostCompatNextToken.Add(1))
		if token == 0 {
			continue
		}
		registration := &nativeHostRegistration{token: token, host: host}
		if _, loaded := hostCompatHosts.LoadOrStore(token, registration); !loaded {
			return registration
		}
	}
}

func (registration *nativeHostRegistration) close() {
	if registration == nil || registration.token == 0 {
		return
	}
	hostCompatHosts.Delete(registration.token)
}

// Rust owns every pointer passed to this callback, so checkptr cannot associate
// the integer C ABI addresses with Go allocations.
//
//go:nocheckptr
func invokeHostCompatCallback(userData, operation, requestID, recordIndex, inputPointer, inputLength,
	outputPointer, outputCapacity, outputLengthPointer uintptr,
) int32 {
	_ = requestID
	value, hostFound := hostCompatHosts.Load(userData)
	registration, registrationValid := value.(*nativeHostRegistration)
	if registrationValid && registration.host != nil && registration.host.observeInvoke != nil {
		// Count every Rust-to-Go ABI attempt for a live host. In particular, a
		// sizing call followed by a buffer retry is two callback invocations.
		registration.host.observeInvoke(uint32(operation), recordIndex)
	}
	if outputLengthPointer == 0 || inputLength > hostCompatMaximumPayload || inputLength != 0 && inputPointer == 0 {
		return hostCompatCallbackInvalidArgument
	}
	*(*uintptr)(unsafe.Pointer(outputLengthPointer)) = 0
	if !hostFound {
		return hostCompatCallbackHostError
	}
	if !registrationValid || registration.host == nil {
		return hostCompatCallbackHostError
	}
	var input []byte
	if inputLength != 0 {
		input = unsafe.Slice((*byte)(unsafe.Pointer(inputPointer)), inputLength)
	}
	output, err := registration.host.invoke(uint32(operation), input)
	if err != nil {
		if isHostCompatProtocolError(err) {
			return hostCompatCallbackInvalidArgument
		}
		return hostCompatCallbackHostError
	}
	*(*uintptr)(unsafe.Pointer(outputLengthPointer)) = uintptr(len(output))
	if uintptr(len(output)) > outputCapacity {
		return hostCompatCallbackBufferTooSmall
	}
	if len(output) != 0 {
		if outputPointer == 0 {
			return hostCompatCallbackInvalidArgument
		}
		copy(unsafe.Slice((*byte)(unsafe.Pointer(outputPointer)), len(output)), output)
	}
	return hostCompatCallbackOK
}
