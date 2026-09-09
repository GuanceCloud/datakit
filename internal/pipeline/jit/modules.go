// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// ModuleSnapshot owns a copy of sources and a stable identity. Build once when
// loading a generation, not per batch. Do not mutate the caller map concurrently
// with NewModuleSnapshot.
type ModuleSnapshot struct {
	entry         string
	sources       map[string]string
	key           [32]byte
	identity      ScriptIdentity
	bundle        [32]byte
	externalRules bool
}

// Identity identifies the stateful owner and its immutable source closure.
func (snapshot *ModuleSnapshot) Identity() [32]byte { return snapshot.key }

func (snapshot *ModuleSnapshot) HasDependencies() bool { return len(snapshot.sources) > 1 }

func (snapshot *ModuleSnapshot) ScriptIdentity() ScriptIdentity { return snapshot.identity }
func (snapshot *ModuleSnapshot) EntrySource() string            { return snapshot.sources[snapshot.entry] }
func (snapshot *ModuleSnapshot) BundleIdentity() [32]byte       { return snapshot.bundle }
func (snapshot *ModuleSnapshot) RequiresExternalRules() bool    { return snapshot.externalRules }

// WithExternalRules records rules that the native compile ABI cannot receive.
// Allowlist admission rejects these snapshots before compilation.
func (snapshot *ModuleSnapshot) WithExternalRules(patterns map[string]string) *ModuleSnapshot {
	next := snapshot.WithRules(patterns)
	next.externalRules = true
	return next
}

// WithRules binds rollout approval to the exact grammar configuration too.
// The program cache identity is unchanged; only immutable admission metadata
// changes. Keys and length-prefixed values make the digest reproducible.
func (snapshot *ModuleSnapshot) WithRules(patterns map[string]string) *ModuleSnapshot {
	next := *snapshot
	h := sha256.New()
	h.Write([]byte("datakit-jit-rollout-v1\x00pipeline-go-1.4.3-datakit\x00"))
	h.Write(snapshot.key[:])
	keys := make([]string, 0, len(patterns))
	for key := range patterns {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range []string{key, patterns[key]} {
			var size [8]byte
			binary.LittleEndian.PutUint64(size[:], uint64(len(value)))
			h.Write(size[:])
			h.Write([]byte(value))
		}
	}
	copy(next.bundle[:], h.Sum(nil))
	return &next
}

// BoundModules holds one compiled version throughout encoding and processing,
// even if its cache entry is invalidated meanwhile. Close releases that lease,
// not the shared Runner, and waits for ongoing Process calls.
type BoundModules struct {
	mu     sync.RWMutex
	lease  *programLease
	source string
}

func (r *Runner) BindModules(snapshot *ModuleSnapshot) (*BoundModules, error) {
	lease, err := r.acquireModules(snapshot)
	if err != nil {
		return nil, err
	}
	return &BoundModules{lease: lease, source: snapshot.sources[snapshot.entry]}, nil
}

// BindSource pins one ordinary program for negotiation, encoding and execution.
func (r *Runner) BindSource(source string) (*BoundModules, error) {
	lease, err := r.acquireProgram(source)
	if err != nil {
		return nil, err
	}
	return &BoundModules{lease: lease, source: source}, nil
}

func (bound *BoundModules) Projection(source string) (InputProjection, error) {
	bound.mu.RLock()
	defer bound.mu.RUnlock()
	if bound.lease == nil || source != bound.source {
		return InputProjection{}, errors.New("closed or mismatched module binding")
	}
	return programProjection(bound.lease.program)
}

func (bound *BoundModules) Process(source string, input []byte) (Batch, error) {
	return bound.ProcessContext(context.Background(), source, input)
}

func (bound *BoundModules) ProcessContext(ctx context.Context, source string, input []byte) (Batch, error) {
	if ctx == nil {
		return Batch{}, errors.New("nil JIT process context")
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	bound.mu.RLock()
	defer bound.mu.RUnlock()
	if bound.lease == nil || source != bound.source {
		return Batch{}, errors.New("closed or mismatched module binding")
	}
	if len(input) > maxInputBytes {
		return Batch{}, fmt.Errorf("JIT input exceeds %d bytes", maxInputBytes)
	}
	return processProgramContext(ctx, bound.lease.program, input)
}

// ProcessInto keeps the same generation lease while reusing caller-owned
// result containers. Callers must finish consuming scratch before reuse.
func (bound *BoundModules) ProcessInto(source string, input []byte, scratch *Batch) (Batch, error) {
	bound.mu.RLock()
	defer bound.mu.RUnlock()
	if bound.lease == nil || source != bound.source {
		return Batch{}, errors.New("closed or mismatched module binding")
	}
	if len(input) > maxInputBytes {
		return Batch{}, fmt.Errorf("JIT input exceeds %d bytes", maxInputBytes)
	}
	if program, ok := bound.lease.program.(interface {
		ProcessIndexedInto([]byte, *Batch) (Batch, error)
	}); ok {
		return program.ProcessIndexedInto(input, scratch)
	}
	return bound.lease.program.ProcessIndexed(input)
}

// ProcessContextInto keeps cancellation and the generation lease while reusing
// the same synchronous batch scratch as ProcessInto.
func (bound *BoundModules) ProcessContextInto(ctx context.Context, source string, input []byte, scratch *Batch) (Batch, error) {
	if ctx == nil {
		return Batch{}, errors.New("nil JIT process context")
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	if !HasExecutionControl(ctx) {
		return bound.ProcessInto(source, input, scratch)
	}
	bound.mu.RLock()
	defer bound.mu.RUnlock()
	if bound.lease == nil || source != bound.source {
		return Batch{}, errors.New("closed or mismatched module binding")
	}
	if len(input) > maxInputBytes {
		return Batch{}, fmt.Errorf("JIT input exceeds %d bytes", maxInputBytes)
	}
	if program, ok := bound.lease.program.(interface {
		ProcessIndexedContextInto(context.Context, []byte, *Batch) (Batch, error)
	}); ok {
		return program.ProcessIndexedContextInto(ctx, input, scratch)
	}
	return processProgramContext(ctx, bound.lease.program, input)
}

func (bound *BoundModules) Close() error {
	bound.mu.Lock()
	defer bound.mu.Unlock()
	if bound.lease != nil {
		bound.lease.Release()
		bound.lease = nil
	}
	return nil
}

func NewModuleSnapshot(entry string, sources map[string]string) (*ModuleSnapshot, error) {
	return NewScopedModuleSnapshot("", "", entry, sources)
}

// NewScopedModuleSnapshot separates stateful instances across namespace and
// category. SourceHash from the runtime still identifies code, not its owner.
func NewScopedModuleSnapshot(namespace, category, entry string, sources map[string]string) (*ModuleSnapshot, error) {
	if entry == "" || len(sources) == 0 || len(sources) > 128 {
		return nil, errors.New("invalid JIT module snapshot")
	}
	if _, ok := sources[entry]; !ok {
		return nil, errors.New("JIT entry missing from snapshot")
	}
	snapshot := &ModuleSnapshot{entry: entry, sources: make(map[string]string, len(sources))}
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
		snapshot.sources[name] = source
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	hash.Write([]byte("\x00datakit-jit-modules-v1\x00"))
	write := func(value string) {
		var length [8]byte
		binary.LittleEndian.PutUint64(length[:], uint64(len(value)))
		hash.Write(length[:])
		hash.Write([]byte(value))
	}
	write(entry)
	write(namespace)
	write(category)
	for _, name := range names {
		write(name)
		write(snapshot.sources[name])
	}
	copy(snapshot.key[:], hash.Sum(nil))
	snapshot.identity = ScriptIdentity{Category: category, Namespace: namespace, Script: entry}
	snapshot.bundle = snapshot.key
	return snapshot, nil
}

func (r *Runner) acquireModules(snapshot *ModuleSnapshot) (*programLease, error) {
	if r == nil || snapshot == nil || snapshot.sources == nil {
		return nil, errors.New("JIT runner or module snapshot is nil")
	}
	if r.profile != "pipeline-go-1.4.3-datakit" {
		return nil, errors.New("module snapshots require pipeline-go-1.4.3-datakit")
	}
	return r.acquireCompiled(snapshot.key, func() (Program, error) {
		runtime, ok := r.runtime.(ModuleRuntime)
		if !ok {
			return nil, errors.New("JIT runtime lacks module snapshot compilation")
		}
		return runtime.CompileModules(snapshot.entry, snapshot.sources)
	})
}

func (r *Runner) PrepareModules(snapshot *ModuleSnapshot) error {
	lease, err := r.acquireModules(snapshot)
	if lease != nil {
		lease.Release()
	}
	return err
}

// CheckModules pins a successfully admitted snapshot just like Check, so LRU
// eviction cannot reset state owned by an active generation.
func (r *Runner) CheckModules(snapshot *ModuleSnapshot) CheckResult {
	if r == nil {
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonRunnerUnavailable, Detail: "JIT runner is nil"}
	}
	lease, err := r.acquireModules(snapshot)
	if err != nil {
		return CheckResult{Route: RoutePipelineGo, Reason: CheckReasonCompileError, Detail: err.Error()}
	}
	return r.checkLease(lease, snapshot.key)
}

func (r *Runner) ProcessModules(snapshot *ModuleSnapshot, input []byte) (Batch, error) {
	return r.ProcessModulesContext(context.Background(), snapshot, input)
}

func (r *Runner) ProcessModulesContext(ctx context.Context, snapshot *ModuleSnapshot, input []byte) (Batch, error) {
	if ctx == nil {
		return Batch{}, errors.New("nil JIT process context")
	}
	if err := ctx.Err(); err != nil {
		return Batch{}, err
	}
	if len(input) > maxInputBytes {
		return Batch{}, fmt.Errorf("JIT input exceeds %d bytes", maxInputBytes)
	}
	lease, err := r.acquireModules(snapshot)
	if err != nil {
		return Batch{}, err
	}
	defer lease.Release()
	return processProgramContext(ctx, lease.program, input)
}

func (r *Runner) InvalidateModules(snapshot *ModuleSnapshot) {
	if r != nil && snapshot != nil {
		r.invalidateKey(snapshot.key)
	}
}
