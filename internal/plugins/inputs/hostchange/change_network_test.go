// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkConfigChecker_Init(t *testing.T) {
	t.Parallel()

	input := &Input{}
	checker := &NetworkConfigChecker{}

	assert.NoError(t, checker.Init(input))
	assert.Equal(t, input, checker.input)
}

func TestNetworkConfigChecker_shouldIgnoreInterface(t *testing.T) {
	t.Parallel()

	checker := &NetworkConfigChecker{
		IgnoreInterfaces: []string{"lo", "docker*"},
	}

	// Test exact match
	assert.True(t, checker.shouldIgnoreInterface("lo"))
	// Test wildcard match
	assert.True(t, checker.shouldIgnoreInterface("docker0"))
	assert.True(t, checker.shouldIgnoreInterface("docker1"))
	// Test non-matching
	assert.False(t, checker.shouldIgnoreInterface("eth0"))
	assert.False(t, checker.shouldIgnoreInterface("ens33"))
}

func TestCompareStringSlices(t *testing.T) {
	t.Parallel()

	// Test equal slices
	assert.True(t, compareStringSlices([]string{"a", "b", "c"}, []string{"a", "b", "c"}))
	// Test empty slices
	assert.True(t, compareStringSlices([]string{}, []string{}))
	// Test different lengths
	assert.False(t, compareStringSlices([]string{"a", "b"}, []string{"a", "b", "c"}))
	// Test different contents
	assert.False(t, compareStringSlices([]string{"a", "b", "c"}, []string{"a", "b", "d"}))
	// Test different order
	assert.False(t, compareStringSlices([]string{"a", "b", "c"}, []string{"c", "b", "a"}))
}

func TestNetworkConfigChecker_CreateNetworkChangeItem(t *testing.T) {
	t.Parallel()

	checker := &NetworkConfigChecker{}
	changeItem := checker.createNetworkChangeItem(NetworkChangeParams{
		ChangeID:   ChangeIDNetworkInterface,
		IfaceName:  "eth0",
		OldIP:      "192.168.1.100",
		NewIP:      "192.168.1.200",
		OldStatus:  true,
		NewStatus:  true,
		ConfigType: "interface",
		ChangeType: "modify",
	})

	assert.Equal(t, ChangeIDNetworkInterface, changeItem.ChangeID)
	assert.Greater(t, changeItem.ChangeTimestampMicro, int64(0))
	assert.NotEmpty(t, changeItem.Title)
	assert.NotEmpty(t, changeItem.Message)
}
