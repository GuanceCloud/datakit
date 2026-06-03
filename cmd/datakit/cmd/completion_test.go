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

func TestCompletionCommand(t *testing.T) {
	assert.Equal(t, "completion", completionCmd.Name())
	assert.Equal(t, "Install shell completion scripts", completionCmd.Short)
	assert.Equal(t, []string{"bash", "powershell", "fish", "zsh"}, completionCmd.ValidArgs)
}

func TestCompletionCommandArgs(t *testing.T) {
	assert.NotNil(t, completionCmd.Args)
}

func TestCompletionCommandRunE(t *testing.T) {
	assert.NotNil(t, completionCmd.RunE)
}

func TestDetectCurrentShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	shell, err := detectCurrentShell()
	require.NoError(t, err)
	assert.Equal(t, "zsh", shell)
}

func TestDetectCurrentShellRequiresKnownShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/unknown-shell")

	shell, err := detectCurrentShell()
	require.Error(t, err)
	assert.Empty(t, shell)
	assert.Contains(t, err.Error(), "unable to detect current shell")
}

func TestNormalizeShell(t *testing.T) {
	cases := map[string]string{
		"/bin/bash":                           "bash",
		"/usr/local/bin/zsh":                  "zsh",
		"/opt/homebrew/bin/fish":              "fish",
		"C:\\Program Files\\PowerShell\\pwsh": "powershell",
		"powershell":                          "powershell",
		"unknown":                             "",
		"":                                    "",
	}

	for value, want := range cases {
		t.Run(value, func(t *testing.T) {
			assert.Equal(t, want, normalizeShell(value))
		})
	}
}

func TestGenerateCompletionScript(t *testing.T) {
	for _, shell := range []string{"bash", "powershell", "fish", "zsh"} {
		script, err := generateCompletionScript(rootCmd, shell)
		require.NoError(t, err)
		assert.NotEmpty(t, script)
		assert.Contains(t, script, "datakit")
	}
}

func TestGenerateCompletionScriptInvalidShell(t *testing.T) {
	_, err := generateCompletionScript(rootCmd, "invalid")
	require.Error(t, err)
}

func TestInstallCompletionToCustomPath(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "datakit.bash")

	res, err := installCompletion(rootCmd, completionInstallOptions{
		Shell: "bash",
		Path:  target,
	})
	require.NoError(t, err)
	assert.Equal(t, target, res.Path)

	data, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestInstallCompletionPrintOnly(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "datakit.bash")

	res, err := installCompletion(rootCmd, completionInstallOptions{
		Shell: "bash",
		Path:  target,
		Print: true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Script)

	_, err = os.Stat(target)
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}
