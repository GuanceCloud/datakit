// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tool

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolCommand(t *testing.T) {
	cmd := NewToolCmd()
	assert.Equal(t, "tool", cmd.Use)
	assert.Equal(t, "Various tools for DataKit", cmd.Short)
}

func TestToolFlags(t *testing.T) {
	cmd := NewToolCmd()

	// Test key flags exist
	assert.NotNil(t, cmd.Flag("grokq"))
	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("show-cloud-info"))
	assert.NotNil(t, cmd.Flag("ipinfo"))
	assert.NotNil(t, cmd.Flag("workspace-info"))
	assert.NotNil(t, cmd.Flag("dump-samples"))
	assert.NotNil(t, cmd.Flag("default-main-conf"))
	assert.Nil(t, cmd.Flag("setup-completer-script"))
	assert.Nil(t, cmd.Flag("completer-script"))
	assert.NotNil(t, cmd.Flag("parse-lp"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("update-ipdb"))
}

func TestToolExecutePassesOptions(t *testing.T) {
	orig := runToolFn
	defer func() { runToolFn = orig }()

	called := false
	runToolFn = func(opts cmds.ToolOptions) error {
		called = true
		assert.True(t, opts.GrokQ)
		assert.Equal(t, "/tmp/tool.log", opts.LogPath)
		assert.True(t, opts.ShowCloudInfo)
		assert.Equal(t, "8.8.8.8", opts.IPInfo)
		assert.True(t, opts.WorkspaceInfo)
		assert.Equal(t, "/tmp/samples", opts.DumpSamples)
		assert.True(t, opts.DefaultMainConf)
		assert.Equal(t, "/tmp/lp", opts.ParseLineProtocol)
		assert.True(t, opts.JSON)
		assert.True(t, opts.UpdateIPDB)
		assert.Equal(t, "/tmp/conf", opts.ParseKVFile)
		assert.Equal(t, "/tmp/kv", opts.KVFile)
		assert.True(t, opts.RemoveApmAutoInject)
		assert.Equal(t, "dk-runc", opts.ChangeDockerContainersRuntime)
		assert.True(t, opts.IngestionCanary)
		assert.Equal(t, "logging-index", opts.IngestionCanaryIndex)
		return nil
	}

	cmd := NewToolCmd()
	cmd.SetArgs([]string{
		"--grokq",
		"--log", "/tmp/tool.log",
		"--show-cloud-info",
		"--ipinfo", "8.8.8.8",
		"--workspace-info",
		"--dump-samples", "/tmp/samples",
		"--default-main-conf",
		"--parse-lp", "/tmp/lp",
		"--json",
		"--update-ipdb",
		"--parse-kv-file", "/tmp/conf",
		"--kv-file", "/tmp/kv",
		"--remove-apm-auto-inject",
		"--change-docker-containers-runtime", "dk-runc",
		"--ingestion-canary",
		"--ingestion-canary-index", "logging-index",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
