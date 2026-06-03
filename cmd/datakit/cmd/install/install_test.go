// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package install

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCommand(t *testing.T) {
	cmd := NewInstallCmd()
	assert.Equal(t, "install", cmd.Use)
	assert.Equal(t, "Install DataKit related packages and plugins", cmd.Short)
}

func TestInstallFlags(t *testing.T) {
	cmd := NewInstallCmd()

	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("telegraf"))
	assert.NotNil(t, cmd.Flag("scheck"))
	assert.NotNil(t, cmd.Flag("ipdb"))
	assert.NotNil(t, cmd.Flag("symbol-tools"))
}

func TestInstallExecutePassesOptions(t *testing.T) {
	orig := runInstallFn
	defer func() { runInstallFn = orig }()

	called := false
	runInstallFn = func(opts cmds.InstallOptions) error {
		called = true
		assert.Equal(t, "/tmp/install.log", opts.LogPath)
		assert.True(t, opts.Telegraf)
		assert.True(t, opts.Scheck)
		assert.Equal(t, "iploc", opts.IPDB)
		assert.True(t, opts.SymbolTools)
		return nil
	}

	cmd := NewInstallCmd()
	cmd.SetArgs([]string{"--log", "/tmp/install.log", "--telegraf", "--scheck", "--ipdb", "iploc", "--symbol-tools"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
