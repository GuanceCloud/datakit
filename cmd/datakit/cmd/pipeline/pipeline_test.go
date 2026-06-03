// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineCommand(t *testing.T) {
	cmd := NewPipelineCmd()
	assert.Equal(t, "pipeline", cmd.Use)
	assert.Equal(t, "Debug pipeline scripts", cmd.Short)
}

func TestPipelineFlags(t *testing.T) {
	cmd := NewPipelineCmd()

	assert.NotNil(t, cmd.Flag("category"))
	assert.NotNil(t, cmd.Flag("namespace"))
	assert.NotNil(t, cmd.Flag("name"))
	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("txt"))
	assert.NotNil(t, cmd.Flag("file"))
	assert.NotNil(t, cmd.Flag("tab"))
	assert.NotNil(t, cmd.Flag("date"))

	// Test short flags
	assert.Equal(t, "C", cmd.Flag("category").Shorthand)
	assert.Equal(t, "N", cmd.Flag("namespace").Shorthand)
	assert.Equal(t, "P", cmd.Flag("name").Shorthand)
	assert.Equal(t, "T", cmd.Flag("txt").Shorthand)
	assert.Equal(t, "F", cmd.Flag("file").Shorthand)
}

func TestPipelineExecutePassesOptions(t *testing.T) {
	orig := runPipelineFn
	defer func() { runPipelineFn = orig }()

	called := false
	runPipelineFn = func(opts cmds.PipelineOptions) error {
		called = true
		assert.Equal(t, "logging", opts.Category)
		assert.Equal(t, "gitrepo", opts.Namespace)
		assert.Equal(t, "demo.p", opts.Name)
		assert.Equal(t, "/tmp/pipeline.log", opts.LogPath)
		assert.Equal(t, "message", opts.Text)
		assert.Equal(t, "/tmp/input.log", opts.FilePath)
		assert.True(t, opts.Table)
		assert.True(t, opts.Date)
		return nil
	}

	cmd := NewPipelineCmd()
	cmd.SetArgs([]string{
		"--category", "logging",
		"--namespace", "gitrepo",
		"--name", "demo.p",
		"--log", "/tmp/pipeline.log",
		"--txt", "message",
		"--file", "/tmp/input.log",
		"--tab",
		"--date",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
