// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

const (
	CodecPointJSON       = 2
	CodecFlatPoint       = 5
	PayloadPointJSON     = 1
	PayloadFlatPoint     = 6
	PayloadMutationDelta = 7

	headerBytes      = 20
	frameHeaderBytes = 20
	maxResultBytes   = 64 << 20
	maxInputBytes    = 64 << 20
	maxFrames        = 1_000_000
	minimumBatchSize = 128

	indexedTerminalCommitPrefixErrorFlag = 1 << 0
)

type TerminalStatus uint8

const (
	TerminalNone TerminalStatus = iota
	TerminalOK
	TerminalDropped
	TerminalError
	TerminalCancelled
)

type Record struct {
	Mutations     []MutationOp // owned decoded operations for dynamic PPS2
	mutationIndex uint64

	Status            TerminalStatus
	Point             []byte
	Emitted           [][]byte
	Delta             []byte
	Error             []byte
	CommitPrefixError bool
}

// MutationDeltas returns already decoded dynamic operations or decodes legacy PPD1.
func (record Record) MutationDeltas() ([]MutationDelta, error) {
	if record.Mutations != nil {
		return []MutationDelta{{RecordIndex: record.mutationIndex, Operations: record.Mutations}}, nil
	}
	return DecodeMutationDeltas(record.Delta)
}

func (record Record) HasMutations() bool { return record.Mutations != nil || record.Delta != nil }

type Batch struct {
	operations []MutationOp // bounded scratch; Records borrow it until reuse

	preparePointFields bool   // adapter-only: transfer owned fields directly into Point
	nativeCall         any    //nolint:unused // Used by the pipeline_jit build on Linux.
	inline             []byte //nolint:unused // Used by the pipeline_jit build on Linux.
	encoded            []byte // owned indexed wire storage; payloads borrow it until scratch reuse
	Records            []Record
	Diagnostics        [][]byte
	Static             *StaticBatch
	Protocol           string
	EncodedBytes       int
}

// PreparePointFields opts synchronous adapter scratch into owned Point fields.
// Ordinary Process callers continue to receive Static.Values.
func (batch *Batch) PreparePointFields() { batch.preparePointFields = true }

// ResetForReuse clears all result references while retaining bounded scratch.
// The caller must own the batch exclusively and finish applying it first.
func (batch *Batch) ResetForReuse() {
	if cap(batch.operations) > 65536 || cap(batch.encoded) > 1<<20 || cap(batch.Records) > 4096 || cap(batch.Diagnostics) > 4096 ||
		(batch.Static != nil && (cap(batch.Static.States) > 65536 || cap(batch.Static.Values) > 65536 ||
			cap(batch.Static.PointFields) > 65536 || cap(batch.Static.StateOffsets) > 4097 || cap(batch.Static.ValueOffsets) > 4097)) {
		*batch = Batch{preparePointFields: batch.preparePointFields}
		return
	}
	clear(batch.operations[:cap(batch.operations)])
	batch.operations = batch.operations[:0]
	clear(batch.Records[:cap(batch.Records)])
	clear(batch.Diagnostics[:cap(batch.Diagnostics)])
	batch.Records = batch.Records[:0]
	batch.Diagnostics = batch.Diagnostics[:0]
	batch.encoded = batch.encoded[:0]
	batch.Protocol = ""
	batch.EncodedBytes = 0
	if static := batch.Static; static != nil {
		clear(static.PointFields[:cap(static.PointFields)])
		static.PointFields = static.PointFields[:0]
		clear(static.Values[:cap(static.Values)])
		static.Values = static.Values[:0]
		static.States = static.States[:0]
		static.StateOffsets = static.StateOffsets[:0]
		static.ValueOffsets = static.ValueOffsets[:0]
		static.Schema = StaticOutputSchema{}
		static.validated = false
	}
}

type StaticOutputSchema struct {
	Dynamic    bool
	SourceHash string
	Hash       [32]byte
	Keys       []string
}

type StaticMutationState uint8

const (
	StaticNoop StaticMutationState = iota
	StaticSetField
	StaticDeleteField
	StaticSetTag
	StaticDeleteTag
)

type StaticBatch struct {
	Schema       StaticOutputSchema
	States       []StaticMutationState
	StateOffsets []uint32
	ValueOffsets []uint32
	Values       []any
	// PointFields replaces Values only for adapter scratch that opted in.
	// Fields and their strings are owned independently of reusable wire buffers.
	PointFields []*point.Field
	validated   bool
}

// Validated reports whether the batch was decoded and fully validated by the
// static wire decoder. Callers may then apply its normalized values without a
// second allocation-heavy validation pass.
func (batch *StaticBatch) Validated() bool {
	return batch != nil && batch.validated
}

type CapabilityMode struct {
	Mode string `json:"mode"`
}

type ProgramCapabilities struct {
	Version           int      `json:"version"`
	SourceHash        string   `json:"source_hash"`
	ExecutionMode     string   `json:"execution_mode"`
	Backend           string   `json:"backend"`
	ExecutionTier     string   `json:"execution_tier"`
	HostCalls         []string `json:"host_calls"`
	RequiredHostFlags uint16   `json:"required_host_flags"`
	AggregateEvents   *bool    `json:"aggregate_events"`
	// Missing on older libraries: shared-state safety has not been established.
	Stateless *bool `json:"stateless"`
	// RawStringValues permits wire value tag 8 for field values, not metadata.
	// Absent on older runtimes: callers must keep strict UTF-8 encoding.
	RawStringValues   bool           `json:"raw_string_values"`
	RawTagValues      bool           `json:"raw_tag_values"`
	RawPointKeys      bool           `json:"raw_point_keys"`
	InputProjection   CapabilityMode `json:"input_projection"`
	StaticOutput      CapabilityMode `json:"static_output"`
	DescriptorPresent bool           `json:"descriptor_present"`
}

const (
	ExecutionBackendInstructionEngine = "instruction_engine"
	ExecutionBackendMachineCode       = "machine_code"

	ExecutionTierInstructionEngine = "instruction_engine"
	ExecutionTierMachineCodeHelper = "machine_code_helper"
	ExecutionTierMachineCodeSlots  = "machine_code_slots"
)

// normalizeExecutionTier keeps capability descriptors from older runtimes
// useful without overstating what their native entry point does. The first
// Cranelift backend compiles dispatch and control flow to machine code but
// still executes builtins through Rust helpers; only a runtime that explicitly
// advertises machine_code_slots may claim direct slot execution.
func normalizeExecutionTier(capabilities *ProgramCapabilities) error {
	if capabilities.Backend == "" {
		capabilities.Backend = ExecutionBackendInstructionEngine
	}
	if capabilities.ExecutionTier == "" {
		switch capabilities.Backend {
		case ExecutionBackendInstructionEngine:
			capabilities.ExecutionTier = ExecutionTierInstructionEngine
		case ExecutionBackendMachineCode:
			capabilities.ExecutionTier = ExecutionTierMachineCodeHelper
		default:
			return fmt.Errorf("unsupported JIT backend %q", capabilities.Backend)
		}
	}

	switch capabilities.Backend {
	case ExecutionBackendInstructionEngine:
		if capabilities.ExecutionTier != ExecutionTierInstructionEngine {
			return fmt.Errorf("JIT backend %q cannot advertise execution tier %q",
				capabilities.Backend, capabilities.ExecutionTier)
		}
	case ExecutionBackendMachineCode:
		if capabilities.ExecutionTier != ExecutionTierMachineCodeHelper &&
			capabilities.ExecutionTier != ExecutionTierMachineCodeSlots {
			return fmt.Errorf("JIT backend %q cannot advertise execution tier %q",
				capabilities.Backend, capabilities.ExecutionTier)
		}
	default:
		return fmt.Errorf("unsupported JIT backend %q", capabilities.Backend)
	}
	return nil
}

type Route string

const (
	RouteJITNative   Route = "jit_native"
	RouteJITWithHost Route = "jit_with_host"
	RoutePipelineGo  Route = "pipeline_go"

	CheckReasonJITReady          = "jit_ready"
	CheckReasonRunnerUnavailable = "runner_unavailable"
	CheckReasonCompileError      = "compile_error"
	CheckReasonCapabilitiesError = "capabilities_error"
)

type CheckResult struct {
	Route        Route               `json:"route"`
	Reason       string              `json:"reason"`
	Detail       string              `json:"detail,omitempty"`
	Capabilities ProgramCapabilities `json:"capabilities"`
}

// InputProjection describes the Point keys a compiled program can observe.
// The zero value means all keys, which keeps old and fallback paths conservative.
type InputProjection struct {
	keys            map[string]struct{}
	ordered         []string
	rawStringValues bool
	rawTagValues    bool
	rawPointKeys    bool
}

// AllowsRawStringValues is negotiated per program, never inferred from ABI version.
func (projection InputProjection) AllowsRawStringValues() bool { return projection.rawStringValues }
func (projection InputProjection) AllowsRawTagValues() bool    { return projection.rawTagValues }
func (projection InputProjection) AllowsRawPointKeys() bool    { return projection.rawPointKeys }

func programProjection(program Program) (InputProjection, error) {
	projection := program.InputProjection()
	capabilities, err := program.Capabilities()
	if err != nil {
		return InputProjection{}, err
	}
	projection.rawStringValues = capabilities.DescriptorPresent && capabilities.RawStringValues
	projection.rawTagValues = capabilities.DescriptorPresent && capabilities.RawTagValues
	projection.rawPointKeys = capabilities.DescriptorPresent && capabilities.RawPointKeys
	return projection, nil
}

func newKeyProjection(keys []string) InputProjection {
	projection := InputProjection{
		keys:    make(map[string]struct{}, len(keys)),
		ordered: append([]string(nil), keys...),
	}
	for _, key := range keys {
		projection.keys[key] = struct{}{}
	}
	return projection
}

func (projection InputProjection) IsAll() bool {
	return projection.keys == nil
}

func (projection InputProjection) Includes(key string) bool {
	if projection.keys == nil {
		return true
	}
	_, ok := projection.keys[key]
	return ok
}

func (projection InputProjection) Keys() []string {
	return append([]string(nil), projection.ordered...)
}

// KeyCount returns the number of statically projected point keys.
func (projection InputProjection) KeyCount() int {
	return len(projection.ordered)
}

type Program interface {
	ProcessIndexed([]byte) (Batch, error)
	InputProjection() InputProjection
	Capabilities() (ProgramCapabilities, error)
	Close() error
}

type ContextProgram interface {
	Program
	ProcessIndexedContext(context.Context, []byte) (Batch, error)
}

// AggregateEventProgram exposes interval and retirement drains for stateful
// aggregation programs. force is reserved for a single owner after new
// executions have stopped for that program generation.
type AggregateEventProgram interface {
	Program
	DrainAggregateEvents(time.Time, bool) ([]Point, error)
}

// SetAggregateEventHandler installs the process-level owner for interval and
// retirement aggregation points. Production installs this before publishing
// the Runner; tests may install an isolated collector.
func (r *Runner) SetAggregateEventHandler(handler func([]Point) error) error {
	if r == nil || handler == nil {
		return errors.New("nil JIT aggregate event handler")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return errors.New("JIT runner is closed")
	}
	setter, ok := r.runtime.(interface{ setAggregateEventHandler(func([]Point) error) })
	if !ok {
		return errors.New("JIT runtime lacks aggregate event handler support")
	}
	setter.setAggregateEventHandler(handler)
	return nil
}

type Runtime interface {
	Compile(source, profile string) (Program, error)
	Close() error
}

// ModuleRuntime is optional: older libraries continue to support single sources.
// The module entry uses the pipeline-go-1.4.3-datakit profile.
type ModuleRuntime interface {
	Runtime
	CompileModules(entry string, sources map[string]string) (Program, error)
}

type cacheEntry struct {
	program Program
	err     error
	used    uint64
	refs    int
	retired bool
	closing bool
	pinned  bool // selected production route owns this instance until invalidation
}

type Runner struct {
	mu      sync.Mutex
	runtime Runtime
	profile string
	max     int
	tick    uint64
	cache   map[[32]byte]*cacheEntry
	closed  bool
	active  int
	closing int
	cond    *sync.Cond
	// closeDone is closed after the one Runner.Close operation has finished.
	// Callers racing with Close wait on this channel and observe closeErr,
	// rather than returning before the underlying runtime has been closed.
	closeDone chan struct{}
	closeErr  error
}

type programLease struct {
	runner  *Runner
	entry   *cacheEntry
	program Program
	once    sync.Once
}

func NewRunner(runtimePath, profile string, maxPrograms int) (*Runner, error) {
	return NewRunnerWithHost(runtimePath, profile, maxPrograms, nil)
}

// ServiceConfig is an immutable instance snapshot, shared by this runner's scripts.
// Replace the runner to change policy; existing leases retain the old snapshot.
type ServiceConfig struct {
	DisableHTTPRequestFunc bool              `json:"disable_http_request_func,omitempty"`
	EnableMixedArrayField  bool              `json:"enable_mixed_array_field,omitempty"`
	Version                uint32            `json:"version"`
	HTTP                   HTTPServiceConfig `json:"http"`
}

type HTTPServiceConfig struct {
	DisableInternalNetwork bool     `json:"disable_internal_network"`
	HostWhitelist          []string `json:"host_whitelist"`
	CIDRWhitelist          []string `json:"cidr_whitelist"`
}

func NewRunnerWithServices(path, profile string, maxPrograms int, host *HostCompat, config ServiceConfig) (*Runner, error) {
	if profile != "pipeline-go-1.4.3-datakit" {
		return nil, errors.New("JIT instance services require DataKit profile")
	}
	runner, err := NewRunnerWithHost(path, profile, maxPrograms, host)
	if err != nil {
		return nil, err
	}
	configured, ok := runner.runtime.(interface{ configureServices(ServiceConfig) error })
	if !ok {
		_ = runner.Close()
		return nil, errors.New("JIT runtime lacks instance services")
	}
	if err := configured.configureServices(config); err != nil {
		_ = runner.Close()
		return nil, err
	}
	return runner, nil
}

// CanBindUnpublishedHost verifies the loader contract before callers publish
// any DataKit configuration used to construct the immutable host snapshot.
func (r *Runner) CanBindUnpublishedHost() bool {
	if r == nil || r.runtime == nil {
		return false
	}
	_, ok := r.runtime.(interface{ bindUnpublishedHost(*HostCompat) })
	return ok
}

// BindUnpublishedHost attaches the generation-scoped compatibility services.
// Callers must preflight CanBindUnpublishedHost and invoke this exactly once,
// before compiling programs or publishing the Runner.
func (r *Runner) BindUnpublishedHost(host *HostCompat) error {
	if !r.CanBindUnpublishedHost() {
		return errors.New("JIT runtime cannot bind unpublished host")
	}
	r.runtime.(interface{ bindUnpublishedHost(*HostCompat) }).bindUnpublishedHost(host)
	return nil
}

// NewRunnerWithServicesHostFactory loads and configures the runtime before
// invoking factory. The runner cannot compile or execute until factory returns
// and its immutable host is attached. A failed factory closes the candidate.
func NewRunnerWithServicesHostFactory(path, profile string, maxPrograms int, config ServiceConfig, factory func() (*HostCompat, error)) (*Runner, error) {
	if factory == nil {
		return nil, errors.New("nil JIT host factory")
	}
	runner, err := NewRunnerWithServices(path, profile, maxPrograms, nil, config)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = runner.Close()
		}
	}()
	if !runner.CanBindUnpublishedHost() {
		return nil, errors.New("JIT runtime cannot bind unpublished host")
	}
	host, err := factory()
	if err != nil {
		return nil, err
	}
	if err := runner.BindUnpublishedHost(host); err != nil {
		return nil, err
	}
	committed = true
	return runner, nil
}

// NewRunnerWithHost opens a runtime with optional pipeline-go compatibility
// callbacks. Runtimes predating the configured compile ABI continue to compile
// through pp_jit_compile_flat_v1 without callbacks.
func NewRunnerWithHost(runtimePath, profile string, maxPrograms int, host *HostCompat) (*Runner, error) {
	if profile == "" {
		profile = "pipeline-go-1.4.3"
	}
	if maxPrograms <= 0 {
		maxPrograms = 256
	}
	runtime, err := openRuntime(runtimePath, host)
	if err != nil {
		return nil, err
	}
	runner := &Runner{
		runtime: runtime,
		profile: profile,
		max:     maxPrograms,
		cache:   make(map[[32]byte]*cacheEntry),
	}
	runner.cond = sync.NewCond(&runner.mu)
	runner.closeDone = make(chan struct{})
	return runner, nil
}

func (r *Runner) Process(source string, input []byte) (Batch, error) {
	return r.ProcessContext(context.Background(), source, input)
}

func (r *Runner) ProcessContext(ctx context.Context, source string, input []byte) (Batch, error) {
	if ctx == nil {
		return Batch{}, errors.New("nil JIT process context")
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	if r == nil {
		return Batch{}, errors.New("JIT runner is nil")
	}
	if len(input) > maxInputBytes {
		return Batch{}, fmt.Errorf("JIT input exceeds %d bytes", maxInputBytes)
	}
	lease, err := r.acquireProgram(source)
	if err != nil {
		return Batch{}, err
	}
	defer lease.Release()
	return processProgramContext(ctx, lease.program, input)
}

func processProgramContext(ctx context.Context, program Program, input []byte) (Batch, error) {
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	if !HasExecutionControl(ctx) {
		return program.ProcessIndexed(input)
	}
	if cancellable, ok := program.(ContextProgram); ok {
		return cancellable.ProcessIndexedContext(ctx, input)
	}
	return Batch{}, errors.New("JIT program lacks cooperative cancellation")
}

// Projection returns the cached program's proven input-key projection.
func (r *Runner) Projection(source string) (InputProjection, error) {
	if r == nil {
		return InputProjection{}, errors.New("JIT runner is nil")
	}
	lease, err := r.acquireProgram(source)
	if err != nil {
		return InputProjection{}, err
	}
	defer lease.Release()
	return programProjection(lease.program)
}

// Prepare compiles and caches a program before the caller encodes its input.
// Unsupported programs are negatively cached so the fallback path stays cheap.
func (r *Runner) Prepare(source string) error {
	if r == nil {
		return errors.New("JIT runner is nil")
	}
	lease, err := r.acquireProgram(source)
	if lease != nil {
		lease.Release()
	}
	return err
}

// Invalidate removes source from the compile cache. An in-flight program is
// retired and closed by its final lease; an idle program is closed before this
// call returns. A later operation can compile the same source again.
func (r *Runner) Invalidate(source string) {
	if r == nil {
		return
	}
	key := sha256.Sum256([]byte(source))
	r.invalidateKey(key)
}

func (r *Runner) invalidateKey(key [32]byte) {
	r.mu.Lock()
	r.initCondLocked()
	if r.closed {
		r.mu.Unlock()
		return
	}
	entry, ok := r.cache[key]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.cache, key)
	entry.retired = true
	var retired *cacheEntry
	if entry.program != nil && entry.refs == 0 && !entry.closing {
		entry.closing = true
		r.closing++
		retired = entry
	}
	r.mu.Unlock()
	r.closeRetired(retired)
}

// Check compiles at most once and returns the route decision cached for source.
// Older runtimes without the capability descriptor remain usable with a
// conservative native capability result.
func (r *Runner) Check(source string) CheckResult {
	if r == nil {
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonRunnerUnavailable, Detail: "JIT runner is nil"}
	}
	lease, err := r.acquireProgram(source)
	if err != nil {
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonCompileError, Detail: err.Error()}
	}
	return r.checkLease(lease, sha256.Sum256([]byte(source)))
}

func (r *Runner) checkLease(lease *programLease, key [32]byte) CheckResult {
	defer lease.Release()
	capabilities, err := lease.program.Capabilities()
	if err != nil {
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonCapabilitiesError, Detail: err.Error()}
	}
	if capabilities.SourceHash == "" {
		capabilities.SourceHash = fmt.Sprintf("%x", key)
	}
	// A successful route decision owns the stateful instance, not just its
	// compiled code. LRU pressure must not silently reset that state.
	r.mu.Lock()
	if lease.entry.retired || r.closed {
		r.mu.Unlock()
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonRunnerUnavailable, Detail: "program retired during route check"}
	}
	lease.entry.pinned = true
	r.mu.Unlock()
	route := RouteJITNative
	if capabilities.ExecutionMode == "native_with_host" {
		route = RouteJITWithHost
	}
	return CheckResult{
		Route:        route,
		Reason:       CheckReasonJITReady,
		Capabilities: capabilities,
	}
}

// MinBatchSize is the benchmarked crossover where the native path repays its
// fixed FlatPoint and mutation-delta adapter costs.
func (r *Runner) MinBatchSize() int {
	return minimumBatchSize
}

func (r *Runner) acquireProgram(source string) (*programLease, error) {
	key := sha256.Sum256([]byte(source))
	return r.acquireCompiled(key, func() (Program, error) { return r.runtime.Compile(source, r.profile) })
}

func (r *Runner) acquireCompiled(key [32]byte, compile func() (Program, error)) (*programLease, error) {
	r.mu.Lock()
	r.initCondLocked()
	if r.closed {
		r.mu.Unlock()
		return nil, errors.New("JIT runner is closed")
	}
	r.tick++
	if entry, ok := r.cache[key]; ok {
		entry.used = r.tick
		if entry.err != nil {
			r.mu.Unlock()
			return nil, entry.err
		}
		entry.refs++
		r.active++
		lease := &programLease{runner: r, entry: entry, program: entry.program}
		r.mu.Unlock()
		return lease, nil
	}
	if len(r.cache) >= r.max {
		evictable := false
		for _, entry := range r.cache {
			if !entry.pinned {
				evictable = true
				break
			}
		}
		if !evictable {
			r.mu.Unlock()
			return nil, errors.New("JIT program capacity reached: active routes cannot be evicted")
		}
	}
	program, err := compile()
	entry := &cacheEntry{program: program, err: err, used: r.tick}
	r.cache[key] = entry
	var retired *cacheEntry
	if len(r.cache) > r.max {
		retired = r.evictLocked(key)
	}
	if err == nil {
		entry.refs++
		r.active++
	}
	var lease *programLease
	if err == nil {
		lease = &programLease{runner: r, entry: entry, program: program}
	}
	r.mu.Unlock()
	r.closeRetired(retired)
	return lease, err
}

func (r *Runner) initCondLocked() {
	if r.cond == nil {
		r.cond = sync.NewCond(&r.mu)
	}
	if r.closeDone == nil {
		r.closeDone = make(chan struct{})
	}
}

func (r *Runner) evictLocked(keep [32]byte) *cacheEntry {
	var oldestKey [32]byte
	var oldest *cacheEntry
	for key, entry := range r.cache {
		if key == keep || entry.pinned || oldest != nil && entry.used >= oldest.used {
			continue
		}
		oldestKey, oldest = key, entry
	}
	if oldest == nil {
		return nil
	}
	delete(r.cache, oldestKey)
	oldest.retired = true
	if oldest.program != nil && oldest.refs == 0 && !oldest.closing {
		oldest.closing = true
		r.closing++
		return oldest
	}
	return nil
}

func (lease *programLease) Release() {
	if lease == nil || lease.runner == nil {
		return
	}
	lease.once.Do(func() {
		r := lease.runner
		r.mu.Lock()
		lease.entry.refs--
		var retired *cacheEntry
		if lease.entry.retired && lease.entry.refs == 0 && lease.entry.program != nil && !lease.entry.closing {
			lease.entry.closing = true
			r.closing++
			retired = lease.entry
		}
		r.mu.Unlock()

		r.closeRetired(retired)

		r.mu.Lock()
		r.active--
		r.cond.Broadcast()
		r.mu.Unlock()
	})
}

func (r *Runner) closeRetired(entry *cacheEntry) {
	if entry == nil {
		return
	}
	err := entry.program.Close()
	r.mu.Lock()
	if err != nil {
		r.closeErr = errors.Join(r.closeErr, err)
	}
	r.closing--
	r.cond.Broadcast()
	r.mu.Unlock()
}

func (r *Runner) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	r.initCondLocked()
	if r.closed {
		done := r.closeDone
		r.mu.Unlock()
		<-done
		r.mu.Lock()
		err := r.closeErr
		r.mu.Unlock()
		return err
	}
	r.closed = true
	var closeNow []*cacheEntry
	for _, entry := range r.cache {
		entry.retired = true
		if entry.program != nil && entry.refs == 0 && !entry.closing {
			entry.closing = true
			r.closing++
			closeNow = append(closeNow, entry)
		}
	}
	r.cache = nil
	r.mu.Unlock()

	for _, entry := range closeNow {
		err := entry.program.Close()
		r.mu.Lock()
		if err != nil {
			r.closeErr = errors.Join(r.closeErr, err)
		}
		r.closing--
		r.cond.Broadcast()
		r.mu.Unlock()
	}

	r.mu.Lock()
	for r.active > 0 || r.closing > 0 {
		r.cond.Wait()
	}
	runtime := r.runtime
	priorErr := r.closeErr
	r.mu.Unlock()

	var runtimeErr error
	if runtime != nil {
		runtimeErr = runtime.Close()
	}
	err := errors.Join(priorErr, runtimeErr)
	r.mu.Lock()
	r.closeErr = err
	close(r.closeDone)
	r.mu.Unlock()
	return err
}

func decodeIndexed(input []byte) (Batch, error) {
	return decodeIndexedInto(input, nil)
}

func decodeIndexedInto(input []byte, scratch *Batch) (Batch, error) {
	if len(input) < headerBytes || len(input) > headerBytes+maxResultBytes {
		return Batch{}, errors.New("invalid PPR3 result length")
	}
	if string(input[:4]) != "PPR3" || binary.LittleEndian.Uint16(input[4:6]) != 3 {
		return Batch{}, errors.New("invalid PPR3 header")
	}
	flags := binary.LittleEndian.Uint16(input[6:8])
	if flags & ^uint16(1) != 0 {
		return Batch{}, errors.New("unknown PPR3 flags")
	}
	inputCount := binary.LittleEndian.Uint32(input[8:12])
	frameCount := binary.LittleEndian.Uint32(input[12:16])
	payloadLength := binary.LittleEndian.Uint32(input[16:20])
	if inputCount > pointWireMaxItems || frameCount > maxFrames ||
		uint64(payloadLength) != uint64(len(input)-headerBytes) ||
		inputCount > frameCount || frameCount > payloadLength/frameHeaderBytes {
		return Batch{}, errors.New("invalid PPR3 counts")
	}
	return decodeIndexedFrames(input[headerBytes:], inputCount, frameCount, flags, scratch, nil, false, len(input))
}

// Both wire modes share terminal validation. Dynamic values and side outputs own
// their memory because the native buffer may be freed immediately after decode.
func decodeIndexedFrames(input []byte, inputCount, frameCount uint32, flags uint16, scratch *Batch, dictionary []string, dynamic bool, encodedBytes int) (Batch, error) {
	var batch Batch
	if scratch != nil {
		scratch.ResetForReuse()
		batch = *scratch
	}
	if cap(batch.Records) < int(inputCount) {
		batch.Records = make([]Record, inputCount)
	} else {
		batch.Records = batch.Records[:inputCount]
	}
	batch.Static = nil // scratch can come from a different program using PPS2
	batch.Protocol = "mutation-delta-v1"
	batch.EncodedBytes = encodedBytes
	if dynamic {
		batch.Protocol = "dynamic-v3"
		if cap(batch.operations) == 0 {
			batch.operations = make([]MutationOp, 0, min(int(frameCount)*2, 4096))
		}
	}
	nodes := 0
	hasErrors := false
	cursor := 0
	for range frameCount {
		if len(input)-cursor < frameHeaderBytes {
			return Batch{}, errors.New("truncated PPR3 frame")
		}
		header := input[cursor : cursor+frameHeaderBytes]
		index := binary.LittleEndian.Uint64(header[:8])
		kind, status, category, codec := header[8], TerminalStatus(header[9]), header[10], header[11]
		frameFlags := binary.LittleEndian.Uint16(header[12:14])
		if binary.LittleEndian.Uint16(header[14:16]) != 0 {
			return Batch{}, errors.New("PPR3 frame reserved bits are set")
		}
		if frameFlags != 0 &&
			(kind != 6 || status != TerminalError || frameFlags != indexedTerminalCommitPrefixErrorFlag) {
			return Batch{}, errors.New("invalid PPR3 frame flags")
		}
		length := uint64(binary.LittleEndian.Uint32(header[16:20]))
		next := uint64(cursor+frameHeaderBytes) + length
		if next > uint64(len(input)) {
			return Batch{}, errors.New("truncated PPR3 frame payload")
		}
		payload := input[cursor+frameHeaderBytes : int(next) : int(next)]
		cursor = int(next)
		if dynamic && kind != 7 {
			payload = append([]byte(nil), payload...)
		}

		if kind == 3 && index == ^uint64(0) {
			if status != TerminalNone || category != 4 || codec != 5 || frameFlags != 0 {
				return Batch{}, errors.New("invalid PPR3 global event frame")
			}
			batch.Diagnostics = append(batch.Diagnostics, payload)
			continue
		}
		if index >= uint64(inputCount) || batch.Records[index].Status != TerminalNone {
			return Batch{}, errors.New("invalid PPR3 record index")
		}
		record := &batch.Records[index]
		switch kind {
		case 1:
			if dynamic {
				return Batch{}, errors.New("dynamic output must use typed mutations")
			}
			if status != TerminalNone || category != 1 || codec != PayloadPointJSON || record.Point != nil {
				return Batch{}, errors.New("invalid PPR3 primary frame")
			}
			record.Point = payload
		case 2:
			if status != TerminalNone || category != 1 ||
				(codec != PayloadPointJSON && codec != PayloadFlatPoint) {
				return Batch{}, errors.New("invalid PPR3 emitted point frame")
			}
			record.Emitted = append(record.Emitted, payload)
		case 4:
			if status != TerminalNone || category != 3 || codec != 4 {
				return Batch{}, errors.New("invalid PPR3 stdout frame")
			}
			batch.Diagnostics = append(batch.Diagnostics, payload)
		case 5:
			if status != TerminalNone || category != 3 || codec != 5 || record.Error != nil {
				return Batch{}, errors.New("invalid PPR3 error frame")
			}
			record.Error = payload
		case 6:
			if status < TerminalOK || status > TerminalCancelled || category != 0 || codec != 0 || len(payload) != 0 {
				return Batch{}, errors.New("invalid PPR3 terminal frame")
			}
			record.Status = status
			record.CommitPrefixError = frameFlags&indexedTerminalCommitPrefixErrorFlag != 0
			hasErrors = hasErrors || status == TerminalError || status == TerminalCancelled
		case 7:
			if status != TerminalNone || category != 1 || codec != PayloadMutationDelta || record.HasMutations() {
				return Batch{}, errors.New("invalid PPR3 mutation delta frame")
			}
			if dynamic {
				mutations := pointCursor{data: payload}
				count, err := mutations.u32()
				if err != nil || count > pointWireMaxItems || uint64(count) > uint64(len(payload)) {
					return Batch{}, errors.New("invalid dynamic mutation count")
				}
				start := len(batch.operations)
				batch.operations = append(batch.operations, make([]MutationOp, count)...)
				record.Mutations = batch.operations[start:len(batch.operations):len(batch.operations)]
				record.mutationIndex = index
				for i := range record.Mutations {
					operation, err := mutations.mutationWithDictionary(&nodes, dictionary)
					if err != nil {
						return Batch{}, fmt.Errorf("dynamic record %d: %w", index, err)
					}
					record.Mutations[i] = operation
				}
				if !mutations.done() {
					return Batch{}, errors.New("trailing dynamic mutation bytes")
				}
			} else {
				record.Delta = payload
			}
		default:
			return Batch{}, fmt.Errorf("unsupported PPR3 frame kind %d", kind)
		}
	}
	if cursor != len(input) {
		return Batch{}, errors.New("trailing PPR3 bytes")
	}
	for index := range batch.Records {
		if batch.Records[index].Status == TerminalNone {
			return Batch{}, fmt.Errorf("PPR3 record %d has no terminal", index)
		}
		record := &batch.Records[index]
		switch record.Status { //nolint:exhaustive // TerminalNone was rejected immediately above.
		case TerminalOK:
			if record.CommitPrefixError || record.Error != nil || (record.Point == nil) == (!record.HasMutations()) {
				return Batch{}, fmt.Errorf("PPR3 record %d has inconsistent output", index)
			}
		case TerminalError:
			if record.Error == nil {
				return Batch{}, fmt.Errorf("PPR3 record %d error terminal has no error frame", index)
			}
			if record.CommitPrefixError {
				if record.Point != nil || !record.HasMutations() {
					return Batch{}, fmt.Errorf("PPR3 record %d has invalid committed error output", index)
				}
			} else if record.Point != nil || record.HasMutations() {
				return Batch{}, fmt.Errorf("PPR3 record %d has inconsistent output", index)
			}
		case TerminalDropped:
			if record.CommitPrefixError || record.Point != nil || record.HasMutations() || record.Error != nil {
				return Batch{}, fmt.Errorf("PPR3 record %d has inconsistent output", index)
			}
		case TerminalCancelled:
			if record.CommitPrefixError || record.Point != nil || record.HasMutations() || record.Error == nil {
				return Batch{}, fmt.Errorf("PPR3 record %d has inconsistent output", index)
			}
		}
	}
	if hasErrors != (flags&1 != 0) {
		return Batch{}, errors.New("PPR3 error flag does not match terminals")
	}
	if scratch != nil {
		*scratch = batch
	}
	return batch, nil
}
