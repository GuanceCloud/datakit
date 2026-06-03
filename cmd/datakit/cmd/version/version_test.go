// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package version

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCmd()
	assert.Equal(t, "version", cmd.Use)
	assert.Equal(t, "Show version and upgrade info", cmd.Short)
}

func TestVersionFlags(t *testing.T) {
	cmd := NewVersionCmd()

	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("upgrade-info-off"))
}

func TestVersionExecutePassesOptions(t *testing.T) {
	orig := runVersionFn
	defer func() { runVersionFn = orig }()

	called := false
	runVersionFn = func(opts cmds.VersionOptions) error {
		called = true
		assert.Equal(t, "/tmp/version.log", opts.LogPath)
		assert.True(t, opts.DisableUpgradeInfo)
		return nil
	}

	cmd := NewVersionCmd()
	cmd.SetArgs([]string{"--log", "/tmp/version.log", "--upgrade-info-off"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
