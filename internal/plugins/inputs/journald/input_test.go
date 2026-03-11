// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package journald

import (
	"testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDecode(t *testing.T) {
	t.Run("default-config", func(t *testing.T) {
		ipt := defaultInput()
		assert.NotNil(t, ipt)
		assert.Equal(t, true, ipt.TailOnly)
		assert.Equal(t, 1000, ipt.MaxEntriesPerBatch)
		assert.Equal(t, true, ipt.SaveCursor)
		assert.Contains(t, ipt.CursorFile, "journald.cursor")
		assert.Len(t, ipt.Paths, 2)
		assert.Equal(t, []string{"/var/log/journal", "/run/log/journal"}, ipt.Paths)
	})

	t.Run("custom-config", func(t *testing.T) {
		ipt := defaultInput()

		config := `
paths = [
  "/custom/journal/path",
  "/another/journal/path"
]

units = ["nginx.service", "docker.service"]

priorities = ["err", "warning", "crit"]

exclude_fields = [
  "_BOOT_ID",
  "_MACHINE_ID"
]

tail_only = false
max_entries_per_batch = 500

save_cursor = false
cursor_file = "/tmp/journald_test.pos"

rootfs_prefix = "/host"

[tags]
  source = "custom_journald"
  environment = "test"
`
		_, err := bstoml.Decode(config, ipt)
		assert.NoError(t, err)

		assert.Len(t, ipt.Paths, 2)
		assert.Equal(t, "/custom/journal/path", ipt.Paths[0])
		assert.Equal(t, "/another/journal/path", ipt.Paths[1])

		assert.Len(t, ipt.Units, 2)
		assert.Equal(t, "nginx.service", ipt.Units[0])
		assert.Equal(t, "docker.service", ipt.Units[1])

		assert.Len(t, ipt.Priorities, 3)
		assert.Equal(t, "err", ipt.Priorities[0])
		assert.Equal(t, "warning", ipt.Priorities[1])
		assert.Equal(t, "crit", ipt.Priorities[2])

		assert.Len(t, ipt.ExcludeFields, 2)
		assert.Equal(t, "_BOOT_ID", ipt.ExcludeFields[0])
		assert.Equal(t, "_MACHINE_ID", ipt.ExcludeFields[1])

		assert.Equal(t, false, ipt.TailOnly)
		assert.Equal(t, 500, ipt.MaxEntriesPerBatch)

		assert.Equal(t, false, ipt.SaveCursor)
		assert.Equal(t, "/tmp/journald_test.pos", ipt.CursorFile)

		assert.NotNil(t, ipt.Input.Tags)
		assert.Equal(t, "custom_journald", ipt.Input.Tags["source"])
		assert.Equal(t, "test", ipt.Input.Tags["environment"])
	})

	t.Run("sample-config", func(t *testing.T) {
		ipt := defaultInput()
		_, err := bstoml.Decode(sampleConfig, ipt)
		assert.NoError(t, err)
		assert.Equal(t, true, ipt.TailOnly)
		assert.Equal(t, 1000, ipt.MaxEntriesPerBatch)
		assert.Equal(t, true, ipt.SaveCursor)
		assert.Contains(t, ipt.CursorFile, "journald.cursor")
	})
}

func TestInputCatalog(t *testing.T) {
	ipt := defaultInput()
	assert.Equal(t, "logging", ipt.Catalog())
}

func TestMeasurementInfo(t *testing.T) {
	ipt := defaultInput()
	measurements := ipt.SampleMeasurement()
	assert.Len(t, measurements, 1)

	info := measurements[0].Info()
	assert.NotNil(t, info)
	assert.Equal(t, "journald", info.Name)
	assert.Equal(t, "logging", info.Cat.String())
	assert.Contains(t, info.Desc, "Systemd journal logs")

	// Check tags
	require.NotNil(t, info.Tags)
	assert.Contains(t, info.Tags, "service")

	// Check fields
	require.NotNil(t, info.Fields)
	assert.Contains(t, info.Fields, "message")
	assert.Contains(t, info.Fields, "priority")
	assert.Contains(t, info.Fields, "status")
	assert.Contains(t, info.Fields, "pid")
	assert.Contains(t, info.Tags, "host")
	assert.Contains(t, info.Fields, "_UID")
	assert.Contains(t, info.Fields, "_GID")
	assert.Contains(t, info.Fields, "_SYSTEMD_UNIT")
	assert.Contains(t, info.Fields, "_BOOT_ID")
	assert.Contains(t, info.Fields, "_MACHINE_ID")
}

func TestArgumentBuilding(t *testing.T) {
	t.Run("default-args", func(t *testing.T) {
		ipt := defaultInput()
		ipt.buildArgs()

		args := ipt.Input.Args
		assert.Contains(t, args, "--paths")
		assert.Contains(t, args, "/var/log/journal,/run/log/journal")
		assert.Contains(t, args, "--tail-only")
		assert.Contains(t, args, "--max-entries")
		assert.Contains(t, args, "1000")
		assert.Contains(t, args, "--save-cursor")
		assert.Contains(t, args, "--cursor-file")
	})

	t.Run("custom-args", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Units = []string{"nginx.service", "docker.service"}
		ipt.Priorities = []string{"err", "warning"}
		ipt.ExcludeFields = []string{"_BOOT_ID"}
		ipt.TailOnly = false
		ipt.MaxEntriesPerBatch = 500
		ipt.SaveCursor = false

		ipt.buildArgs()

		args := ipt.Input.Args
		assert.Contains(t, args, "--units")
		assert.Contains(t, args, "nginx.service,docker.service")
		assert.Contains(t, args, "--priorities")
		assert.Contains(t, args, "err,warning")
		assert.Contains(t, args, "--exclude-fields")
		assert.Contains(t, args, "_BOOT_ID")
		assert.NotContains(t, args, "--tail-only")
		assert.Contains(t, args, "--max-entries")
		assert.Contains(t, args, "500")
		assert.NotContains(t, args, "--save-cursor")
	})
}

func TestBinaryPathDetection(t *testing.T) {
	ipt := defaultInput()

	// Test that it can find the binary in the local build directory
	ipt.Input.Cmd = "./journald"
	binaryPath := ipt.findJournaldBinary()
	// Should find the binary we built
	if binaryPath == "" {
		// Try alternative path
		ipt.Input.Cmd = "journald"
		binaryPath = ipt.findJournaldBinary()
	}
	// Binary should be found in one of the locations
	// Note: This test may fail if binary is not built, which is OK for CI
	if binaryPath == "" {
		t.Skip("journald binary not found, skipping test (run 'make' to build)")
	}
	assert.NotEmpty(t, binaryPath)
}
