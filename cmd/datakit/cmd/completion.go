// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cmd implements the DataKit Cobra command tree.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	completionPrint bool
	completionPath  string
	completionForce bool
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|powershell|fish|zsh]",
	Short: "Install shell completion scripts",
	Long: `Install shell completion scripts for DataKit.

Specify the target shell explicitly, for example: datakit completion zsh --force.
If no shell is specified, DataKit will try to detect it from the SHELL environment
variable. This may fail when running through sudo or other restricted environments.
Use --print to output the generated script instead of installing it.
`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"bash", "powershell", "fish", "zsh"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var shell string
		if len(args) == 1 {
			shell = args[0]
		}

		res, err := installCompletion(rootCmd, completionInstallOptions{
			Shell: shell,
			Print: completionPrint,
			Path:  completionPath,
			Force: completionForce,
		})
		if err != nil {
			return err
		}

		if completionPrint {
			_, err = os.Stdout.WriteString(res.Script)
			return err
		}

		msg := fmt.Sprintf("completion for %s installed to %s", res.Shell, res.Path)
		if res.Detected {
			msg = fmt.Sprintf("detected shell %s, %s", res.Shell, msg)
		}
		if res.InDocker {
			msg += " (inside container filesystem)"
		}
		if _, err := fmt.Fprintln(os.Stdout, msg); err != nil {
			return err
		}
		_, err = fmt.Fprintln(os.Stdout, res.Activation)
		return err
	},
}

//nolint:gochecknoinits
func init() {
	completionCmd.Flags().BoolVar(&completionPrint, "print", false, "print generated script instead of installing it")
	completionCmd.Flags().StringVar(&completionPath, "path", "", "install completion to the specified path")
	completionCmd.Flags().BoolVar(&completionForce, "force", false, "overwrite the target path if it already exists")
}
