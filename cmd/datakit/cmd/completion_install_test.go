// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveInstallPathFish(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := resolveInstallPath("fish", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config/fish/completions/datakit.fish"), path)
}

func TestResolveInstallPathBashFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig := bashCompletionCandidates
	bashCompletionCandidates = []string{filepath.Join(t.TempDir(), "missing", "datakit")}
	t.Cleanup(func() { bashCompletionCandidates = orig })

	path, err := resolveInstallPath("bash", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".local/share/bash-completion/completions/datakit"), path)
}

func TestResolveInstallPathZshFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	orig := zshCompletionCandidates
	zshCompletionCandidates = []string{filepath.Join(t.TempDir(), "missing", "_datakit")}
	t.Cleanup(func() { zshCompletionCandidates = orig })

	path, err := resolveInstallPath("zsh", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".zfunc/_datakit"), path)
}

func TestResolveInstallPathUsesCustomPath(t *testing.T) {
	target := filepath.Join(t.TempDir(), "datakit-completion")

	path, err := resolveInstallPath("bash", target)
	require.NoError(t, err)
	assert.Equal(t, target, path)
}

func TestResolveInstallPathPowerShell(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := resolveInstallPath("powershell", "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "powershell", "completions", "datakit.ps1"), path)
	assert.NotContains(t, path, "Microsoft.PowerShell_profile.ps1")
}

func TestResolveInstallPathRejectsUnsupportedShell(t *testing.T) {
	_, err := resolveInstallPath("unknown", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no writable install target found")
}

func TestCanUseTargetExistingWritableDir(t *testing.T) {
	dir := t.TempDir()
	assert.True(t, canUseTarget(filepath.Join(dir, "datakit"), t.TempDir()))
}

func TestCanUseTargetAllowsHomeFallback(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "missing", "datakit")
	assert.True(t, canUseTarget(target, home))
}

func TestEnsureWritableTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "datakit")
	require.NoError(t, os.WriteFile(target, []byte("x"), 0o644))

	err := ensureWritableTarget(target, false)
	require.Error(t, err)

	require.NoError(t, ensureWritableTarget(target, true))
}

func TestActivationHint(t *testing.T) {
	assert.Contains(t, activationHint("bash", "/tmp/datakit"), "source /tmp/datakit")
	assert.Contains(t, activationHint("powershell", "/tmp/profile.ps1"), ". /tmp/profile.ps1")
	assert.Contains(t, activationHint("fish", ""), "restart fish")
	assert.Contains(t, activationHint("zsh", ""), "compinit")
	assert.Equal(t,
		"run this command to enable completion: grep -q 'fpath=(/tmp/.zfunc $fpath)' ~/.zshrc || printf '\\n# DataKit completion\\nfpath=(/tmp/.zfunc $fpath)\\nautoload -Uz compinit\\ncompinit\\nautoload -Uz _datakit\\ncompdef _datakit datakit\\n' >> ~/.zshrc; rm -rf ~/.zcompdump*; source ~/.zshrc; autoload -Uz _datakit; compdef _datakit datakit",
		activationHint("zsh", filepath.Join("/tmp", ".zfunc", "_datakit")))
	assert.Contains(t, activationHint("unknown", ""), "open a new shell")
}
