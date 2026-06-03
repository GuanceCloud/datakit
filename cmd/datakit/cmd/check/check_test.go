// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package check

import (
	"errors"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCommand(t *testing.T) {
	cmd := NewCheckCmd()
	assert.Equal(t, "check", cmd.Use)
	assert.Equal(t, "Check inputs configurations and samples", cmd.Short)
}

func TestCheckFlags(t *testing.T) {
	cmd := NewCheckCmd()

	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("config"))
	assert.NotNil(t, cmd.Flag("config-dir"))
	assert.NotNil(t, cmd.Flag("sample"))
}

func TestCheckExecutePassesOptions(t *testing.T) {
	orig := runCheckFn
	defer func() { runCheckFn = orig }()

	called := false
	runCheckFn = func(opts cmds.CheckOptions) error {
		called = true
		assert.Equal(t, "/tmp/check.log", opts.LogPath)
		assert.True(t, opts.Config)
		assert.Equal(t, "/tmp/conf.d", opts.ConfigDir)
		assert.True(t, opts.Sample)
		return nil
	}

	cmd := NewCheckCmd()
	cmd.SetArgs([]string{"--log", "/tmp/check.log", "--config", "--config-dir", "/tmp/conf.d", "--sample"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}

func TestCheckExecuteReturnsRunError(t *testing.T) {
	orig := runCheckFn
	defer func() { runCheckFn = orig }()

	wantErr := errors.New("check failed")
	runCheckFn = func(opts cmds.CheckOptions) error {
		return wantErr
	}

	cmd := NewCheckCmd()
	err := cmd.Execute()
	require.ErrorIs(t, err, wantErr)
}
