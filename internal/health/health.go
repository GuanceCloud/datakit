// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package health tracks whether DataKit components still make progress.
package health

import (
	"sync"
	"time"
)

// Options controls when a component is considered unable to recover.
type Options struct {
	FailureTimeout time.Duration
	SilenceTimeout time.Duration
}

// FailedComponent identifies a component that requires process recovery.
type FailedComponent struct {
	Name     string `json:"name"`
	Instance string `json:"instance,omitempty"`
}

// Snapshot is the aggregate liveness response.
type Snapshot struct {
	Live             bool              `json:"live"`
	CheckedAt        time.Time         `json:"checked_at"`
	FailedComponents []FailedComponent `json:"failed_components"`
}

// Registry owns component health reporters.
type Registry struct {
	mu         sync.RWMutex
	components map[string]*reporter
	now        func() time.Time
}

// Reporter is the liveness reporting contract used by DataKit components.
type Reporter interface {
	Alive()
	Failure()
	Close()
}

type reporter struct {
	registry     *Registry
	key          string
	name         string
	instance     string
	options      Options
	lastReport   time.Time
	failureSince time.Time
}

// Default is the process-wide component health registry.
var Default = NewRegistry()

// NewRegistry creates an empty health registry.
func NewRegistry() *Registry {
	return &Registry{components: make(map[string]*reporter), now: time.Now}
}

// Register adds or replaces a component instance in the registry.
// Registered components affect process liveness when they stop recovering.
func (r *Registry) Register(name, instance string, opts Options) Reporter {
	reporter := &reporter{
		registry:   r,
		key:        name + "\x00" + instance,
		name:       name,
		instance:   instance,
		options:    opts,
		lastReport: r.now(),
	}
	r.mu.Lock()
	r.components[reporter.key] = reporter
	r.mu.Unlock()
	return reporter
}

// Snapshot evaluates all registered components at the current time.
func (r *Registry) Snapshot() Snapshot {
	now := r.now()
	result := Snapshot{Live: true, CheckedAt: now, FailedComponents: []FailedComponent{}}

	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, reporter := range r.components {
		if !reporter.liveAt(now) {
			result.Live = false
			result.FailedComponents = append(result.FailedComponents, FailedComponent{
				Name: reporter.name, Instance: reporter.instance,
			})
		}
	}
	return result
}

// Alive records that the component is making progress and clears prior failures.
func (r *reporter) Alive() {
	r.registry.mu.Lock()
	r.lastReport = r.registry.now()
	r.failureSince = time.Time{}
	r.registry.mu.Unlock()
}

// Failure records that the component could not recover during this work cycle.
func (r *reporter) Failure() {
	now := r.registry.now()
	r.registry.mu.Lock()
	r.lastReport = now
	if r.failureSince.IsZero() {
		r.failureSince = now
	}
	r.registry.mu.Unlock()
}

// Close removes this component instance from its registry.
func (r *reporter) Close() {
	r.registry.mu.Lock()
	if r.registry.components[r.key] == r {
		delete(r.registry.components, r.key)
	}
	r.registry.mu.Unlock()
}

func (r *reporter) liveAt(now time.Time) bool {
	if r.options.FailureTimeout > 0 && !r.failureSince.IsZero() &&
		now.Sub(r.failureSince) >= r.options.FailureTimeout {
		return false
	}
	if r.options.SilenceTimeout > 0 && now.Sub(r.lastReport) >= r.options.SilenceTimeout {
		return false
	}
	return true
}
