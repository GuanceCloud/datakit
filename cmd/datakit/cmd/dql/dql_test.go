// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dql

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDQLCommand(t *testing.T) {
	cmd := NewDQLCmd()
	assert.Equal(t, "dql", cmd.Use)
	assert.Equal(t, "Query DQL for various usage", cmd.Short)
}

func TestDQLFlags(t *testing.T) {
	cmd := NewDQLCmd()

	// Test all long flags exist
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("auto-json"))
	assert.NotNil(t, cmd.Flag("verbose"))
	assert.NotNil(t, cmd.Flag("run"))
	assert.NotNil(t, cmd.Flag("token"))
	assert.NotNil(t, cmd.Flag("csv"))
	assert.NotNil(t, cmd.Flag("force"))
	assert.NotNil(t, cmd.Flag("host"))
	assert.NotNil(t, cmd.Flag("log"))

	// Test short flags are set correctly (Shorthand field)
	assert.Equal(t, "J", cmd.Flag("json").Shorthand)
	assert.Equal(t, "V", cmd.Flag("verbose").Shorthand)
	assert.Equal(t, "R", cmd.Flag("run").Shorthand)
	assert.Equal(t, "T", cmd.Flag("token").Shorthand)
	assert.Equal(t, "F", cmd.Flag("force").Shorthand)
	assert.Equal(t, "H", cmd.Flag("host").Shorthand)
}

func TestDQLFlagDefaults(t *testing.T) {
	cmd := NewDQLCmd()

	// Test default values
	assert.Equal(t, "false", cmd.Flag("json").DefValue)
	assert.Equal(t, "false", cmd.Flag("auto-json").DefValue)
	assert.Equal(t, "false", cmd.Flag("verbose").DefValue)
	assert.Equal(t, "", cmd.Flag("run").DefValue)
	assert.Equal(t, "", cmd.Flag("token").DefValue)
	assert.Equal(t, "", cmd.Flag("csv").DefValue)
	assert.Equal(t, "false", cmd.Flag("force").DefValue)
	assert.Equal(t, "", cmd.Flag("host").DefValue)
	assert.Equal(t, "/dev/null", cmd.Flag("log").DefValue)
}

func TestDQLExecutePassesOptions(t *testing.T) {
	orig := runDQLFn
	defer func() { runDQLFn = orig }()

	called := false
	runDQLFn = func(opts cmds.DQLOptions) error {
		called = true
		assert.True(t, opts.JSON)
		assert.True(t, opts.AutoJSON)
		assert.True(t, opts.Verbose)
		assert.Equal(t, "M::cpu:(last(`usage`))", opts.Run)
		assert.Equal(t, "token", opts.Token)
		assert.Equal(t, "/tmp/csv", opts.CSV)
		assert.True(t, opts.Force)
		assert.Equal(t, "127.0.0.1:9529", opts.Host)
		assert.Equal(t, "/tmp/dql.log", opts.LogPath)
		return nil
	}

	cmd := NewDQLCmd()
	cmd.SetArgs([]string{
		"--json",
		"--auto-json",
		"--verbose",
		"--run", "M::cpu:(last(`usage`))",
		"--token", "token",
		"--csv", "/tmp/csv",
		"--force",
		"--host", "127.0.0.1:9529",
		"--log", "/tmp/dql.log",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
