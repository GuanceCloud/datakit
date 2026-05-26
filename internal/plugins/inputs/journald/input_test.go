// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package journald

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
		assert.Equal(t, "/rootfs", ipt.MountDir)
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

	mount_dir = "/host-root"

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
		assert.Equal(t, "/host-root", ipt.MountDir)

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

func TestCopyNodeLibsConfig(t *testing.T) {
	t.Run("default-config", func(t *testing.T) {
		ipt := defaultInput()
		assert.False(t, ipt.CopyNodeLibs)
		assert.Empty(t, ipt.CopyNodeLibsFiles)
	})

	t.Run("custom-config", func(t *testing.T) {
		ipt := defaultInput()
		config := `
copy_node_libs = true
copy_node_libs_files = ["libsystemd.so*", "libzstd.so*"]
`

		_, err := bstoml.Decode(config, ipt)
		require.NoError(t, err)
		assert.True(t, ipt.CopyNodeLibs)
		assert.Equal(t, []string{"libsystemd.so*", "libzstd.so*"}, ipt.CopyNodeLibsFiles)
	})
}

func TestPrepareHostLibraries(t *testing.T) {
	t.Run("disabled-does-not-copy", func(t *testing.T) {
		ipt := defaultInput()
		prepareNodeLibsFn = func(_, _ string, _ []string) error {
			t.Fatal("prepareNodeLibsFn should not be called")
			return nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		require.NoError(t, ipt.prepareLibraries())
		assert.Empty(t, ipt.Input.Envs)
	})

	t.Run("enabled-empty-files-outside-container-returns-error", func(t *testing.T) {
		ipt := defaultInput()
		ipt.CopyNodeLibs = true

		setDockerMode(t, false)
		prepareNodeLibsFn = func(_, _ string, _ []string) error {
			t.Fatal("prepareNodeLibsFn should not be called")
			return nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		err := ipt.prepareLibraries()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "copy_node_libs_files is required")
		assert.Empty(t, envValue(ipt.Input.Envs, "LD_LIBRARY_PATH"))
	})

	t.Run("enabled-uses-configured-file-list", func(t *testing.T) {
		ipt := defaultInput()
		ipt.CopyNodeLibs = true
		ipt.CopyNodeLibsFiles = []string{"libsystemd.so*", "libmount.so*"}

		var gotFiles []string
		var gotRootfs string
		prepareNodeLibsFn = func(rootfs, _ string, files []string) error {
			gotRootfs = rootfs
			gotFiles = append([]string(nil), files...)
			return nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		require.NoError(t, ipt.prepareLibraries())
		assert.Equal(t, "/rootfs", gotRootfs)
		assert.Equal(t, []string{"libsystemd.so*", "libmount.so*"}, gotFiles)
		assert.Contains(t, envValue(ipt.Input.Envs, "LD_LIBRARY_PATH"), defaultExternalLibDir)
	})

	t.Run("copy-node-libs-preserve-existing-environment", func(t *testing.T) {
		ipt := defaultInput()
		ipt.CopyNodeLibs = true
		ipt.CopyNodeLibsFiles = []string{"libsystemd.so*"}
		ipt.Input.Envs = []string{"CUSTOM_ENV=1"}

		prepareNodeLibsFn = func(_, _ string, _ []string) error {
			return nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		require.NoError(t, ipt.prepareLibraries())
		assert.Equal(t, "1", envValue(ipt.Input.Envs, "CUSTOM_ENV"))
		assert.Contains(t, envValue(ipt.Input.Envs, "PATH"), string(os.PathListSeparator))
		assert.Equal(t, defaultExternalLibDir, envValue(ipt.Input.Envs, "LD_LIBRARY_PATH"))
	})
}

func restoreJournaldInputSeams() {
	prepareNodeLibsFn = prepareNodeLibs
	missingSharedObjectsFn = detectMissingSharedObjects
}

func setDockerMode(t *testing.T, enabled bool) {
	t.Helper()
	orig := datakit.Docker
	datakit.Docker = enabled
	t.Cleanup(func() {
		datakit.Docker = orig
	})
}

func envValue(envs []string, key string) string {
	for _, env := range envs {
		if k, v, found := strings.Cut(env, "="); found && k == key {
			return v
		}
	}
	return ""
}

func TestApplyKubernetesMode(t *testing.T) {
	t.Run("default-mount-dir", func(t *testing.T) {
		ipt := defaultInput()
		ipt.Paths = []string{"/var/log/journal", "/run/log/journal", "relative/path", "/rootfs/custom"}

		setDockerMode(t, true)

		ipt.applyKubernetesMode()

		assert.True(t, ipt.CopyNodeLibs)
		assert.Equal(t,
			[]string{"/rootfs/var/log/journal", "/rootfs/run/log/journal", "relative/path", "/rootfs/custom"},
			ipt.Paths)
	})

	t.Run("custom-mount-dir", func(t *testing.T) {
		ipt := defaultInput()
		ipt.MountDir = "/host-root"
		ipt.Paths = []string{"/var/log/journal", "/run/log/journal", "relative/path", "/host-root/custom"}

		setDockerMode(t, true)

		ipt.applyKubernetesMode()

		assert.True(t, ipt.CopyNodeLibs)
		assert.Equal(t,
			[]string{"/host-root/var/log/journal", "/host-root/run/log/journal", "relative/path", "/host-root/custom"},
			ipt.Paths)
	})
}

func TestPrepareHostLibrariesKubernetesAuto(t *testing.T) {
	t.Run("auto-copy-missing-libs-when-files-empty", func(t *testing.T) {
		ipt := defaultInput()
		ipt.CopyNodeLibs = true
		ipt.CopyNodeLibsFiles = nil
		ipt.MountDir = "/host-root"

		setDockerMode(t, true)
		var copied [][]string
		var rootfsArgs []string
		prepareNodeLibsFn = func(rootfs, _ string, files []string) error {
			rootfsArgs = append(rootfsArgs, rootfs)
			copied = append(copied, append([]string(nil), files...))
			return nil
		}

		call := 0
		missingSharedObjectsFn = func(_, _ string) ([]string, error) {
			call++
			if call == 1 {
				return []string{"libpcre.so.1"}, nil
			}
			return nil, nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		require.NoError(t, ipt.prepareLibraries())
		require.Len(t, copied, 2)
		assert.Equal(t, []string{"/host-root", "/host-root"}, rootfsArgs)
		assert.Equal(t, []string{"libsystemd.so*"}, copied[0])
		assert.Equal(t, []string{"libpcre.so.1"}, copied[1])
		assert.Contains(t, envValue(ipt.Input.Envs, "LD_LIBRARY_PATH"), defaultExternalLibDir)
	})

	t.Run("configured-files-override-auto-logic", func(t *testing.T) {
		ipt := defaultInput()
		ipt.CopyNodeLibs = true
		ipt.CopyNodeLibsFiles = []string{"libcustom.so*"}
		ipt.MountDir = "/host-root"

		setDockerMode(t, true)
		prepareNodeLibsFn = func(rootfs, _ string, files []string) error {
			assert.Equal(t, "/host-root", rootfs)
			assert.Equal(t, []string{"libcustom.so*"}, files)
			return nil
		}
		missingSharedObjectsFn = func(_, _ string) ([]string, error) {
			t.Fatal("missingSharedObjectsFn should not be called when copy_node_libs_files is configured")
			return nil, nil
		}
		t.Cleanup(restoreJournaldInputSeams)

		require.NoError(t, ipt.prepareLibraries())
	})
}

func TestParseMissingSharedObjects(t *testing.T) {
	out := `
linux-vdso.so.1 (0x00007ffd21bea000)
libcap.so.2 => /usr/local/datakit/externals/systemd-libs/libcap.so.2 (0x00007f1dbc200000)
libpcre.so.1 => not found
libzstd.so.1 => not found
libpcre.so.1 => not found
`

	assert.Equal(t, []string{"libpcre.so.1", "libzstd.so.1"}, parseMissingSharedObjects(out))
}

func TestDetectMissingSharedObjects(t *testing.T) {
	dir := t.TempDir()
	lddPath := filepath.Join(dir, "ldd")
	script := "#!/bin/sh\n" +
		"echo \"\\tlibpcre.so.1 => not found\" \n" +
		"echo \"\\tlibcap.so.2 => /lib/libcap.so.2 (0x000)\" \n"
	require.NoError(t, os.WriteFile(lddPath, []byte(script), 0o755))

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	missing, err := detectMissingSharedObjects(dir, "libsystemd.so.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"libpcre.so.1"}, missing)
}

func TestCopyNodeLibFromMountRootCopiesAbsoluteSymlinkTarget(t *testing.T) {
	rootfs := t.TempDir()
	srcDir := filepath.Join(rootfs, "usr/lib/x86_64-linux-gnu")
	dstDir := t.TempDir()
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(rootfs, "lib/x86_64-linux-gnu"), 0o755))

	targetName := "libgcrypt.so.11.8.2"
	linkName := "libgcrypt.so.11"
	targetContent := []byte("fake-so-content-abs")

	srcTarget := filepath.Join(rootfs, "lib/x86_64-linux-gnu", targetName)
	srcLink := filepath.Join(srcDir, linkName)
	require.NoError(t, os.WriteFile(srcTarget, targetContent, 0o755))
	require.NoError(t, os.Symlink("/lib/x86_64-linux-gnu/"+targetName, srcLink))

	dstLink := filepath.Join(dstDir, linkName)
	ok, err := copyNodeLibFromMountRoot(rootfs, srcLink, dstLink)
	require.NoError(t, err)
	require.True(t, ok)

	link, err := os.Readlink(dstLink)
	require.NoError(t, err)
	assert.Equal(t, targetName, link)

	dstTarget := filepath.Join(dstDir, targetName)
	got, err := os.ReadFile(dstTarget)
	require.NoError(t, err)
	assert.Equal(t, targetContent, got)
}
