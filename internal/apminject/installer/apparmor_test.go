// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build (linux && amd64) || (linux && arm64)
// +build linux,amd64 linux,arm64

package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withAppArmorPaths(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	oldEnabledPath := appArmorEnabledPath
	oldProfilesPath := appArmorProfilesPath
	oldAbstractionPath := appArmorAbstractionPath

	appArmorEnabledPath = filepath.Join(tmpDir, "enabled")
	appArmorProfilesPath = filepath.Join(tmpDir, "profiles")
	appArmorAbstractionPath = filepath.Join(tmpDir, "apparmor.d", "abstractions", appArmorAbstractionName)

	t.Cleanup(func() {
		appArmorEnabledPath = oldEnabledPath
		appArmorProfilesPath = oldProfilesPath
		appArmorAbstractionPath = oldAbstractionPath
	})

	return tmpDir
}

func TestAppArmorEnabled(t *testing.T) {
	_ = withAppArmorPaths(t)

	assert.False(t, appArmorEnabled())

	require.NoError(t, os.WriteFile(appArmorEnabledPath, []byte("Y\n"), 0o644))
	assert.True(t, appArmorEnabled())

	require.NoError(t, os.WriteFile(appArmorEnabledPath, []byte("N\n"), 0o644))
	assert.False(t, appArmorEnabled())

	require.NoError(t, os.Remove(appArmorEnabledPath))
	require.NoError(t, os.MkdirAll(filepath.Dir(appArmorProfilesPath), 0o755))
	require.NoError(t, os.WriteFile(appArmorProfilesPath, []byte("profile-a\n"), 0o644))
	assert.True(t, appArmorEnabled())
}

func TestAppArmorAbstractionContent(t *testing.T) {
	content := appArmorAbstractionContent("/usr/local/datakit")

	assert.Contains(t, content, "#include <abstractions/datakit-apm-inject>")
	assert.Contains(t, content, "/etc/ld.so.preload r,")
	assert.Contains(t, content, "/usr/local/datakit/apm_inject/inject/apm_launcher.so mr,")
	assert.Contains(t, content, "/usr/local/datakit/apm_inject/inject/rewriter ix,")
	assert.Contains(t, content, "/usr/local/datakit/apm_inject/lib/** mr,")
	assert.Contains(t, content, "/tmp/dk_inject_rewrite_* rw,")
	assert.Contains(t, content, "/var/run/datakit/datakit.sock rw,")
}

func TestPrepareAppArmorAbstraction(t *testing.T) {
	_ = withAppArmorPaths(t)
	installDir := filepath.Join(t.TempDir(), "datakit")

	enabled, err := prepareAppArmorAbstraction(installDir)
	require.NoError(t, err)
	assert.False(t, enabled)

	require.NoError(t, os.WriteFile(appArmorEnabledPath, []byte("Y\n"), 0o644))

	enabled, err = prepareAppArmorAbstraction(installDir)
	require.NoError(t, err)
	assert.True(t, enabled)

	data, err := os.ReadFile(appArmorAbstractionPath)
	require.NoError(t, err)
	assert.Equal(t, appArmorAbstractionContent(installDir), string(data))

	enabled, err = prepareAppArmorAbstraction(installDir)
	require.NoError(t, err)
	assert.True(t, enabled)
}
