// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderGateKeepsEarlyResume(t *testing.T) {
	gate := newLeaderGate()

	gate.ConfigureElection(true)
	gate.Pause()
	assert.False(t, gate.Allowed())

	gate.Resume()
	assert.True(t, gate.Allowed())

	gate.ConfigureElection(true)
	assert.True(t, gate.Allowed())
}

func TestLeaderGateDisabledElectionAllowsCollection(t *testing.T) {
	gate := newLeaderGate()

	gate.ConfigureElection(true)
	assert.False(t, gate.Allowed())

	gate.ConfigureElection(false)
	assert.True(t, gate.Allowed())
}

func TestLeaderGateDeliversDecisionsBeforeCollectorStarts(t *testing.T) {
	for _, leader := range []bool{false, true} {
		t.Run(map[bool]string{false: "pause", true: "resume"}[leader], func(t *testing.T) {
			gate := newLeaderGate()
			if leader {
				gate.Resume()
			} else {
				gate.Pause()
			}
			gate.ConfigureElection(true)
			var observed bool
			gate.Subscribe(func(allowed bool) { observed = allowed })
			assert.Equal(t, leader, observed)
			assert.Equal(t, leader, gate.Allowed())
			gate.Pause()
			assert.False(t, observed)
			gate.Resume()
			assert.True(t, observed)
		})
	}
}
