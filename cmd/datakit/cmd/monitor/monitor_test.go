// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitorCommand(t *testing.T) {
	cmd := NewMonitorCmd()
	assert.Equal(t, "monitor", cmd.Use)
	assert.Equal(t, "Show datakit running statistics", cmd.Short)
}

func TestMonitorFlags(t *testing.T) {
	cmd := NewMonitorCmd()

	assert.NotNil(t, cmd.Flag("to"))
	assert.NotNil(t, cmd.Flag("max-table-width"))
	assert.NotNil(t, cmd.Flag("log"))
	assert.NotNil(t, cmd.Flag("refresh"))
	assert.NotNil(t, cmd.Flag("verbose"))
	assert.NotNil(t, cmd.Flag("module"))
	assert.NotNil(t, cmd.Flag("input"))
	assert.NotNil(t, cmd.Flag("path"))
	assert.NotNil(t, cmd.Flag("timestamp"))
	assert.NotNil(t, cmd.Flag("dump-metrics"))

	// Test short flags
	assert.Equal(t, "W", cmd.Flag("max-table-width").Shorthand)
	assert.Equal(t, "R", cmd.Flag("refresh").Shorthand)
	assert.Equal(t, "V", cmd.Flag("verbose").Shorthand)
	assert.Equal(t, "M", cmd.Flag("module").Shorthand)
	assert.Equal(t, "I", cmd.Flag("input").Shorthand)
	assert.Equal(t, "P", cmd.Flag("path").Shorthand)
	assert.Equal(t, "T", cmd.Flag("timestamp").Shorthand)
}

func TestMonitorFlagDefaults(t *testing.T) {
	cmd := NewMonitorCmd()
	assert.Equal(t, "128", cmd.Flag("max-table-width").DefValue)
	assert.Equal(t, "5s", cmd.Flag("refresh").DefValue)
	assert.Equal(t, "/dev/null", cmd.Flag("log").DefValue)
}

func TestMonitorExecutePassesOptions(t *testing.T) {
	orig := runMonitorFn
	defer func() { runMonitorFn = orig }()

	called := false
	runMonitorFn = func(opts cmds.MonitorOptions) error {
		called = true
		assert.Equal(t, "127.0.0.1:9529", opts.To)
		assert.Equal(t, 80, opts.MaxTableWidth)
		assert.Equal(t, "/tmp/monitor.log", opts.LogPath)
		assert.Equal(t, 2*time.Second, opts.Refresh)
		assert.True(t, opts.Verbose)
		assert.Equal(t, "filter,inputs", opts.Module)
		assert.Equal(t, "cpu,mem", opts.OnlyInputs)
		assert.Equal(t, "/tmp/metrics", opts.FilePath)
		assert.Equal(t, int64(123), opts.TimestampMS)
		assert.True(t, opts.DumpMetrics)
		assert.Equal(t, "99", opts.Quantile)
		return nil
	}

	cmd := NewMonitorCmd()
	cmd.SetArgs([]string{
		"--to", "127.0.0.1:9529",
		"--max-table-width", "80",
		"--log", "/tmp/monitor.log",
		"--refresh", "2s",
		"--verbose",
		"--module", "filter,inputs",
		"--input", "cpu,mem",
		"--path", "/tmp/metrics",
		"--timestamp", "123",
		"--dump-metrics",
		"--quantile", "99",
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, called)
}
