// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package importcmd

import (
	"path/filepath"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportCommand(t *testing.T) {
	cmd := NewImportCmd()
	assert.Equal(t, "import", cmd.Use)
	assert.Equal(t, "Import recorded data to Guance Cloud", cmd.Short)
}

func TestImportFlags(t *testing.T) {
	cmd := NewImportCmd()

	assert.NotNil(t, cmd.Flag("path"))
	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("dataway"))

	// Test short flag
	assert.Equal(t, "P", cmd.Flag("path").Shorthand)
	assert.Equal(t, "D", cmd.Flag("dataway").Shorthand)
	assert.Equal(t, filepath.Join(datakit.InstallDir, "recorder"), cmd.Flag("path").DefValue)
	assert.Equal(t, cmds.CommonLogFlag(), cmd.Flag("log").DefValue)
}

func TestImportExecutePassesOptions(t *testing.T) {
	orig := runImportNowFn
	defer func() { runImportNowFn = orig }()

	called := false
	runImportNowFn = func(opts cmds.ImportOptions, when int64) error {
		called = true
		assert.Equal(t, "/tmp/points", opts.Path)
		assert.Equal(t, "/tmp/import.log", opts.LogPath)
		assert.Equal(t, []string{"https://dw1.example", "https://dw2.example"}, opts.DatawayURLs)
		assert.NotZero(t, when)
		return nil
	}

	cmd := NewImportCmd()
	cmd.SetArgs([]string{
		"--path", "/tmp/points",
		"--log", "/tmp/import.log",
		"--dataway", "https://dw1.example,https://dw2.example",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
