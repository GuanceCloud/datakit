// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cmd implements the DataKit Cobra command tree.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	shellBash       = "bash"
	shellFish       = "fish"
	shellZsh        = "zsh"
	shellPowerShell = "powershell"
)

func detectCurrentShell() (string, error) {
	if shell := normalizeShell(os.Getenv("SHELL")); shell != "" {
		return shell, nil
	}

	if runtime.GOOS == "windows" {
		return shellPowerShell, nil
	}

	return "", fmt.Errorf("unable to detect current shell, please specify one of: bash, powershell, fish, zsh")
}

func normalizeShell(value string) string {
	if value == "" {
		return ""
	}

	base := strings.ToLower(filepath.Base(strings.ReplaceAll(value, "\\", "/")))

	switch base {
	case shellBash:
		return shellBash
	case "pwsh", shellPowerShell:
		return shellPowerShell
	case shellFish:
		return shellFish
	case shellZsh:
		return shellZsh
	default:
		return ""
	}
}

func generateCompletionScript(cmd *cobra.Command, shell string) (string, error) {
	var buf bytes.Buffer

	switch shell {
	case shellBash:
		if err := cmd.GenBashCompletionV2(&buf, true); err != nil {
			return "", err
		}
	case shellPowerShell:
		if err := cmd.GenPowerShellCompletionWithDesc(&buf); err != nil {
			return "", err
		}
	case shellFish:
		if err := cmd.GenFishCompletion(&buf, true); err != nil {
			return "", err
		}
	case shellZsh:
		if err := cmd.GenZshCompletion(&buf); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	return buf.String(), nil
}
