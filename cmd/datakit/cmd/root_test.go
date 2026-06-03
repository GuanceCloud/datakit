// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmd

import (
	"bytes"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetRootCommandForTest() {
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	if flag := rootCmd.Flags().Lookup("help"); flag != nil {
		_ = flag.Value.Set("false")
		flag.Changed = false
	}
}

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "datakit", rootCmd.Use)
	assert.Equal(t, "DataKit data collection agent", rootCmd.Short)
	assert.NotNil(t, rootCmd.Long)
}

func TestRootCommandHasCompletionSubcommand(t *testing.T) {
	found := false
	for _, subcmd := range rootCmd.Commands() {
		if subcmd.Name() == "completion" {
			found = true
			break
		}
	}
	assert.True(t, found, "root command should have completion subcommand")
}

func TestRootHelpCommandDisabled(t *testing.T) {
	cmd := rootCmd
	defer resetRootCommandForTest()

	cmd.SetArgs([]string{"help"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "datakit help is not supported")
	assert.Contains(t, err.Error(), "datakit --help")
}

func TestRootHelpCommandSuggestsSubcommandHelp(t *testing.T) {
	cmd := rootCmd
	defer resetRootCommandForTest()

	cmd.SetArgs([]string{"help", "service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "datakit help service is not supported")
	assert.Contains(t, err.Error(), "datakit service --help")
}

func TestRootHelpOutputDoesNotShowHelpCommand(t *testing.T) {
	cmd := rootCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})
	defer resetRootCommandForTest()

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "\n  help")
}

func TestRootLogFlagDefaultsMatchPlatform(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		flag := cmd.Flag("log")
		if flag == nil {
			flag = cmd.PersistentFlags().Lookup("log")
		}
		if flag == nil {
			continue
		}
		assert.Equal(t, cmds.CommonLogFlag(), flag.DefValue, cmd.Name())
	}
}

func TestFilterHelpCompletion(t *testing.T) {
	out := filterHelpCompletion("check\tCheck inputs\nhelp\t\nversion\tShow version\n:4\n")
	assert.Contains(t, out, "check\tCheck inputs")
	assert.Contains(t, out, "version\tShow version")
	assert.Contains(t, out, ":4\n")
	assert.NotContains(t, out, "help")
}

func TestRootCommandStartsDatakitByDefault(t *testing.T) {
	orig := startDataKitFn
	defer func() {
		startDataKitFn = orig
		resetRootCommandForTest()
	}()

	called := false
	runInContainer := true
	startDataKitFn = func(v bool) error {
		called = true
		runInContainer = v
		return nil
	}

	rootCmd.SetArgs(nil)
	err := rootCmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
	assert.False(t, runInContainer)
}
