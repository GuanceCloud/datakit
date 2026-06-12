// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package debug

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugCommand(t *testing.T) {
	cmd := NewDebugCmd()
	assert.Equal(t, "debug", cmd.Use)
	assert.Equal(t, "Various debug tools for DataKit", cmd.Short)
}

func TestDebugFlags(t *testing.T) {
	cmd := NewDebugCmd()

	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("upload-log"))
	assert.NotNil(t, cmd.Flag("glob-conf"))
	assert.NotNil(t, cmd.Flag("regex-conf"))
	assert.NotNil(t, cmd.Flag("prom-conf"))
	assert.NotNil(t, cmd.Flag("bug-report"))
	assert.NotNil(t, cmd.Flag("oss"))
	assert.NotNil(t, cmd.Flag("disable-profile"))
	assert.NotNil(t, cmd.Flag("nmetrics"))
	assert.NotNil(t, cmd.Flag("tag"))
	assert.NotNil(t, cmd.Flag("input-conf"))
	assert.NotNil(t, cmd.Flag("http-listen"))
	assert.NotNil(t, cmd.Flag("filter"))
	assert.NotNil(t, cmd.Flag("data"))
	assert.NotNil(t, cmd.Flag("kv-file"))
}

func TestDebugExecutePassesOptions(t *testing.T) {
	orig := runDebugFn
	defer func() { runDebugFn = orig }()

	called := false
	runDebugFn = func(opts cmds.DebugOptions) error {
		called = true
		assert.Equal(t, "/tmp/debug.log", opts.LogPath)
		assert.True(t, opts.UploadLog)
		assert.Equal(t, "/tmp/glob.conf", opts.GlobConf)
		assert.Equal(t, "/tmp/regex.conf", opts.RegexConf)
		assert.Equal(t, "/tmp/prom.conf", opts.PromConf)
		assert.True(t, opts.BugReport)
		assert.Equal(t, "host:bucket:ak:sk", opts.BugreportOSS)
		assert.Equal(t, "https://dw.example", opts.BugreportDataway)
		assert.True(t, opts.BugreportDatawayEnabled)
		assert.True(t, opts.BugreportDisableProfile)
		assert.Equal(t, 7, opts.BugreportNMetrics)
		assert.Equal(t, "case-1", opts.BugreportTag)
		assert.Equal(t, "/tmp/input.conf", opts.InputConf)
		assert.Equal(t, "127.0.0.1:9529", opts.HTTPListen)
		assert.Equal(t, "/tmp/filter.json", opts.Filter)
		assert.Equal(t, "payload", opts.Data)
		assert.Equal(t, "/tmp/kv", opts.KVFile)
		return nil
	}

	cmd := NewDebugCmd()
	cmd.SetArgs([]string{
		"--log", "/tmp/debug.log",
		"--upload-log",
		"--glob-conf", "/tmp/glob.conf",
		"--regex-conf", "/tmp/regex.conf",
		"--prom-conf", "/tmp/prom.conf",
		"--bug-report",
		"--oss", "host:bucket:ak:sk",
		"--bug-report-dataway", "https://dw.example",
		"--disable-profile",
		"--nmetrics", "7",
		"--tag", "case-1",
		"--input-conf", "/tmp/input.conf",
		"--http-listen", "127.0.0.1:9529",
		"--filter", "/tmp/filter.json",
		"--data", "payload",
		"--kv-file", "/tmp/kv",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}

func TestDebugExecutePassesBareDatawayFlag(t *testing.T) {
	orig := runDebugFn
	defer func() { runDebugFn = orig }()

	called := false
	runDebugFn = func(opts cmds.DebugOptions) error {
		called = true
		assert.True(t, opts.BugreportDatawayEnabled)
		assert.Empty(t, opts.BugreportDataway)
		assert.True(t, opts.BugreportDisableProfile)
		return nil
	}

	cmd := NewDebugCmd()
	cmd.SetArgs([]string{
		"--bug-report",
		"--bug-report-dataway",
		"--disable-profile",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
