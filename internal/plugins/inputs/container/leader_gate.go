// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import "sync/atomic"

type leaderGate struct {
	electionEnabled atomic.Bool
	allowed         atomic.Bool
	touched         atomic.Bool
}

func newLeaderGate() *leaderGate {
	g := &leaderGate{}
	g.allowed.Store(true)
	return g
}

func (g *leaderGate) ConfigureElection(enabled bool) {
	g.electionEnabled.Store(enabled)
	if !enabled {
		g.allowed.Store(true)
		return
	}

	if !g.touched.Load() {
		g.allowed.Store(false)
	}
}

func (g *leaderGate) Pause() {
	g.touched.Store(true)
	g.allowed.Store(false)
}

func (g *leaderGate) Resume() {
	g.touched.Store(true)
	g.allowed.Store(true)
}

func (g *leaderGate) Allowed() bool {
	if g == nil || !g.electionEnabled.Load() {
		return true
	}
	return g.allowed.Load()
}
