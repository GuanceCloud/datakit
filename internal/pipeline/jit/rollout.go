// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/GuanceCloud/cliutils/point"
)

const MaxRolloutEntries = 4096

type ScriptIdentity struct {
	Category  string `toml:"category" json:"category"`
	Namespace string `toml:"namespace" json:"namespace"`
	Script    string `toml:"script" json:"script"`
}

type AllowRule struct {
	Category     string `toml:"category" json:"category"`
	Namespace    string `toml:"namespace" json:"namespace"`
	Script       string `toml:"script" json:"script"`
	BundleSHA256 string `toml:"bundle_sha256" json:"bundle_sha256"`
}

func (r AllowRule) Identity() ScriptIdentity {
	return ScriptIdentity{r.Category, r.Namespace, r.Script}
}

// RolloutPolicy is immutable after publication. Checks run during script load,
// while Admission.Decision is only a pair of atomic reads on the batch path.
type RolloutPolicy struct {
	mode  string
	allow map[ScriptIdentity][32]byte
	deny  map[ScriptIdentity]struct{}
}

func NewRolloutPolicy(mode string, allow []AllowRule, deny []ScriptIdentity) (*RolloutPolicy, error) {
	if mode == "" {
		mode = "auto"
	}
	if mode != "auto" && mode != "allowlist" {
		return nil, fmt.Errorf("pipeline.jit.mode must be auto or allowlist")
	}
	if len(allow)+len(deny) > MaxRolloutEntries {
		return nil, fmt.Errorf("pipeline JIT policy exceeds %d entries", MaxRolloutEntries)
	}
	p := &RolloutPolicy{mode: mode, allow: make(map[ScriptIdentity][32]byte), deny: make(map[ScriptIdentity]struct{})}
	for _, rule := range allow {
		id := rule.Identity()
		if err := validateScriptIdentity(id); err != nil {
			return nil, err
		}
		if _, exists := p.allow[id]; exists {
			return nil, fmt.Errorf("duplicate JIT allow entry: %+v", id)
		}
		digest, err := ParseSHA256(rule.BundleSHA256)
		if err != nil {
			return nil, fmt.Errorf("JIT bundle_sha256: %w", err)
		}
		p.allow[id] = digest
	}
	for _, id := range deny {
		if err := validateScriptIdentity(id); err != nil {
			return nil, err
		}
		if _, exists := p.deny[id]; exists {
			return nil, fmt.Errorf("duplicate JIT force_go entry: %+v", id)
		}
		p.deny[id] = struct{}{}
	}
	return p, nil
}

func validateScriptIdentity(id ScriptIdentity) error {
	if point.CatString(id.Category) == point.UnknownCategory {
		return fmt.Errorf("invalid JIT script category %q", id.Category)
	}
	if id.Namespace != "default" && id.Namespace != "gitrepo" && id.Namespace != "confd" && id.Namespace != "remote" {
		return fmt.Errorf("invalid JIT script namespace %q", id.Namespace)
	}
	if id.Script == "" || len(id.Script) > 1024 || strings.ContainsAny(id.Script, "\x00\r\n*?[]") {
		return errors.New("JIT script must be a nonempty exact name without wildcards")
	}
	return nil
}

func ParseSHA256(value string) ([32]byte, error) {
	var digest [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(digest) {
		return digest, errors.New("expected 64 hexadecimal SHA-256 characters")
	}
	copy(digest[:], decoded)
	return digest, nil
}

func (p *RolloutPolicy) RequiresStateless() bool { return p != nil && p.mode == "allowlist" }

func (p *RolloutPolicy) Mode() string { return p.mode }

func (p *RolloutPolicy) Check(id ScriptIdentity, bundle [32]byte) string {
	if p == nil {
		return "policy_disabled"
	}
	if _, found := p.deny[id]; found {
		return "forced_go"
	}
	if p.mode == "allowlist" {
		want, found := p.allow[id]
		if !found {
			return "not_allowed"
		}
		if want != bundle {
			return "bundle_changed"
		}
	}
	return ""
}

// HealthRegistry retains bounded tombstones across policy reloads and cache
// eviction. Old completions can affect only the exact library and bundle they
// executed. A full registry denies new admission instead of forgetting faults.
type HealthRegistry struct {
	mu       sync.Mutex
	programs map[healthKey]*programHealth
	runtimes map[[32]byte]*atomic.Bool
}

type healthKey struct{ runtime, bundle [32]byte }
type programHealth struct {
	quarantined atomic.Bool
	runtime     *atomic.Bool
}

type Admission struct {
	Bundle    [32]byte
	Stateless bool
	Reason    string
	health    *programHealth
}

func (r *HealthRegistry) Admit(runtime, bundle [32]byte, stateless bool) *Admission {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.programs == nil {
		r.programs = make(map[healthKey]*programHealth)
		r.runtimes = make(map[[32]byte]*atomic.Bool)
	}
	key := healthKey{runtime, bundle}
	h := r.programs[key]
	if h == nil {
		if len(r.programs) >= MaxRolloutEntries {
			return &Admission{Bundle: bundle, Reason: "health_capacity"}
		}
		runtimeHealth := r.runtimes[runtime]
		if runtimeHealth == nil {
			runtimeHealth = new(atomic.Bool)
			r.runtimes[runtime] = runtimeHealth
		}
		h = &programHealth{runtime: runtimeHealth}
		r.programs[key] = h
	}
	return &Admission{Bundle: bundle, Stateless: stateless, health: h}
}

// Decision returns an empty reason for native execution. blocked means that
// moving shared state to Go would be unsafe; return a failed record instead.
func (a *Admission) Decision() (reason string, blocked bool) {
	if a == nil {
		return "", false
	}
	if a.Reason != "" {
		return a.Reason, false
	}
	if a.health != nil && (a.health.quarantined.Load() || a.health.runtime.Load()) {
		return "quarantined", !a.Stateless
	}
	return "", false
}

// Quarantine returns true only for the first transition, allowing callers to
// emit one bounded metric/log rather than log each rejected Point.
func (a *Admission) Quarantine(runtimeWide bool) bool {
	if a == nil || a.health == nil {
		return false
	}
	if runtimeWide {
		return a.health.runtime.CompareAndSwap(false, true)
	}
	return a.health.quarantined.CompareAndSwap(false, true)
}
