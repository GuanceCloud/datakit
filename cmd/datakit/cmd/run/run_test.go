// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	cmd := NewRunCmd(func(bool) error { return nil })
	assert.Equal(t, "run", cmd.Use)
	assert.Equal(t, "Select DataKit running mode", cmd.Short)
}

func TestRunFlags(t *testing.T) {
	cmd := NewRunCmd(func(bool) error { return nil })

	assert.NotNil(t, cmd.Flag("container"))
	assert.Equal(t, "C", cmd.Flag("container").Shorthand)
}

func TestRunExecutePassesContainerFlag(t *testing.T) {
	runInContainer = false

	called := false
	passed := false
	cmd := NewRunCmd(func(v bool) error {
		called = true
		passed = v
		return nil
	})
	cmd.SetArgs([]string{"--container"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
	assert.True(t, passed)
}
