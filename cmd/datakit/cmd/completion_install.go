// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cmd implements the DataKit Cobra command tree.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
)

var (
	bashCompletionCandidates = []string{
		"/usr/share/bash-completion/completions/datakit",
		"/etc/bash_completion.d/datakit",
	}
	zshCompletionCandidates = []string{
		"/usr/share/zsh/site-functions/_datakit",
		"/usr/local/share/zsh/site-functions/_datakit",
	}
)

type completionInstallOptions struct {
	Shell string
	Print bool
	Path  string
	Force bool
}

type completionInstallResult struct {
	Shell      string
	Path       string
	Script     string
	Detected   bool
	InDocker   bool
	Activation string
}

func installCompletion(cmd *cobra.Command, opts completionInstallOptions) (*completionInstallResult, error) {
	shell := opts.Shell
	detected := false
	if shell == "" {
		var err error
		shell, err = detectCurrentShell()
		if err != nil {
			return nil, err
		}
		detected = true
	}

	script, err := generateCompletionScript(cmd, shell)
	if err != nil {
		return nil, err
	}

	res := &completionInstallResult{
		Shell:      shell,
		Script:     script,
		Detected:   detected,
		InDocker:   config.IsDockerRuntime(),
		Activation: activationHint(shell, opts.Path),
	}

	if opts.Print {
		return res, nil
	}

	targetPath, err := resolveInstallPath(shell, opts.Path)
	if err != nil {
		return nil, err
	}

	if err := ensureWritableTarget(targetPath, opts.Force); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil { //nolint:gosec
		return nil, err
	}

	if err := os.WriteFile(targetPath, []byte(script), 0o644); err != nil { //nolint:gosec
		return nil, err
	}

	res.Path = targetPath
	res.Activation = activationHint(shell, targetPath)

	return res, nil
}

func resolveInstallPath(shell, path string) (string, error) {
	if path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch shell {
	case shellBash:
		for _, candidate := range bashCompletionCandidates {
			if canUseTarget(candidate, home) {
				return candidate, nil
			}
		}
		return filepath.Join(home, ".local/share/bash-completion/completions/datakit"), nil
	case shellFish:
		return filepath.Join(home, ".config/fish/completions/datakit.fish"), nil
	case shellZsh:
		for _, candidate := range zshCompletionCandidates {
			if canUseTarget(candidate, home) {
				return candidate, nil
			}
		}
		return filepath.Join(home, ".zfunc/_datakit"), nil
	case shellPowerShell:
		if completionPath := powershellCompletionPath(home); completionPath != "" {
			return completionPath, nil
		}
	}

	return "", fmt.Errorf("no writable install target found for shell %s", shell)
}

func powershellCompletionPath(home string) string {
	if home == "" {
		return ""
	}

	if os.PathSeparator == '\\' {
		return filepath.Join(home, "Documents", "PowerShell", "Completions", "datakit.ps1")
	}

	return filepath.Join(home, ".config", "powershell", "completions", "datakit.ps1")
}

func canUseTarget(path, home string) bool {
	dir := filepath.Dir(path)
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		f, err := os.CreateTemp(dir, ".dk-completion-*")
		if err != nil {
			return false
		}
		_ = f.Close()
		_ = os.Remove(f.Name())
		return true
	}

	return home != "" && strings.HasPrefix(dir, home)
}

func ensureWritableTarget(path string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("completion target already exists: %s", path)
	}

	return nil
}

func activationHint(shell, path string) string {
	switch shell {
	case shellBash:
		if path != "" {
			return fmt.Sprintf("reload your shell or run: source %s", path)
		}
	case shellFish:
		return "restart fish or open a new shell session"
	case shellZsh:
		if path != "" && filepath.Base(path) == "_datakit" {
			zfuncDir := filepath.Dir(path)
			zshrcBlock := "\\n# DataKit completion\\n" +
				"fpath=(%s $fpath)\\n" +
				"autoload -Uz compinit\\n" +
				"compinit\\n" +
				"autoload -Uz _datakit\\n" +
				"compdef _datakit datakit\\n"
			return fmt.Sprintf(
				"run this command to enable completion: "+
					"grep -q 'fpath=(%s $fpath)' ~/.zshrc || "+
					"printf '"+zshrcBlock+"' >> ~/.zshrc; "+
					"rm -rf ~/.zcompdump*; source ~/.zshrc; autoload -Uz _datakit; compdef _datakit datakit",
				zfuncDir,
				zfuncDir,
			)
		}
		return "run compinit again or open a new zsh session"
	case shellPowerShell:
		if path != "" {
			return fmt.Sprintf("reload your shell or run: . %s", path)
		}
	}

	return "open a new shell session to load the completion"
}
