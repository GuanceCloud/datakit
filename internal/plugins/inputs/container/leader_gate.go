// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"sync"
	"sync/atomic"
)

type leaderGate struct {
	mu              sync.Mutex
	electionEnabled bool
	allowed         atomic.Bool
	touched         bool
	requested       bool
	onChange        func(bool)
}

func newLeaderGate() *leaderGate {
	g := &leaderGate{requested: true}
	g.allowed.Store(true)
	return g
}

func (g *leaderGate) ConfigureElection(enabled bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.electionEnabled = enabled
	if !enabled {
		g.allowed.Store(true)
	} else if !g.touched {
		g.allowed.Store(false)
	} else {
		g.allowed.Store(g.requested)
	}
	g.notify()
}

func (g *leaderGate) Pause() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.touched = true
	g.requested = false
	g.allowed.Store(!g.electionEnabled)
	g.notify()
}

func (g *leaderGate) Resume() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.touched = true
	g.requested = true
	g.allowed.Store(true)
	g.notify()
}

func (g *leaderGate) Allowed() bool {
	if g == nil {
		return true
	}
	return g.allowed.Load()
}

// Subscribe delivers the current state too, so election decisions made before
// the Kubernetes collector is constructed cannot be lost.
func (g *leaderGate) Subscribe(onChange func(bool)) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onChange = onChange
	g.notify()
}

func (g *leaderGate) notify() {
	if g.onChange != nil {
		g.onChange(g.allowed.Load())
	}
}
