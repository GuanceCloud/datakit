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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/check"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/debug"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/dql"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/importcmd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/install"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/monitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/run"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/tool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmd/version"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/core"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	releaseVersion    string
	inputsReleaseType string
	lite              string
	eLinker           string
	startDataKitFn    = startDataKit
)

var rootCmd = &cobra.Command{
	Use:   "datakit",
	Short: "DataKit data collection agent",
	Long: `DataKit is an open source, integrated data collection agent, which provides full
platform (Linux/Windows/macOS) support and has comprehensive data collection capability,
covering various scenarios such as host, container, middleware, tracing, logging and
	security inspection.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return startDataKitFn(false)
	},
}

const rootUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func SetBuildInfo(releaseVer, inputsType, liteValue, eLinkerValue string) {
	releaseVersion = releaseVer
	inputsReleaseType = inputsType
	lite = liteValue
	eLinker = eLinkerValue
}

// Execute runs the root command.
func Execute() error {
	applyBuildInfo()
	if isCompletionRequest(os.Args[1:]) {
		if err := executeFilteredCompletion(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return nil
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func applyBuildInfo() {
	if releaseVersion != "" {
		datakit.Version = releaseVersion
	}

	if v, err := strconv.ParseBool(lite); err == nil {
		datakit.Lite = v
	}
	if v, err := strconv.ParseBool(eLinker); err == nil {
		datakit.ELinker = v
	}

	cmds.ReleaseVersion = releaseVersion
	cmds.InputsReleaseType = inputsReleaseType
	cmds.Lite = datakit.Lite
	cmds.ELinker = datakit.ELinker
}

func startDataKit(runInContainer bool) error {
	return core.Start(releaseVersion, inputsReleaseType, lite, eLinker, runInContainer)
}

func isCompletionRequest(args []string) bool {
	return len(args) > 0 && (args[0] == cobra.ShellCompRequestCmd || args[0] == cobra.ShellCompNoDescRequestCmd)
}

func executeFilteredCompletion(args []string) error {
	var buf bytes.Buffer

	origOut := rootCmd.OutOrStdout()
	origErr := rootCmd.ErrOrStderr()
	defer func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(origOut)
		rootCmd.SetErr(origErr)
	}()

	rootCmd.SetArgs(args)
	rootCmd.SetOut(&buf)
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	_, err := os.Stdout.WriteString(filterHelpCompletion(buf.String()))
	return err
}

func filterHelpCompletion(output string) string {
	lines := strings.SplitAfter(output, "\n")
	var b strings.Builder

	for _, line := range lines {
		name := strings.TrimSpace(strings.SplitN(line, "\t", 2)[0])
		if name == "help" {
			continue
		}
		b.WriteString(line)
	}

	return b.String()
}

//nolint:gochecknoinits
func init() {
	rootCmd.SetUsageTemplate(rootUsageTemplate)
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				target := strings.Join(args, " ")
				return fmt.Errorf("datakit help %s is not supported; use datakit %s --help instead", target, target)
			}
			return fmt.Errorf("datakit help is not supported; use datakit --help instead")
		},
	})

	// Register all subcommands
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(dql.NewDQLCmd())
	rootCmd.AddCommand(pipeline.NewPipelineCmd())
	rootCmd.AddCommand(service.NewServiceCmd())
	rootCmd.AddCommand(monitor.NewMonitorCmd())
	rootCmd.AddCommand(install.NewInstallCmd())
	rootCmd.AddCommand(debug.NewDebugCmd())
	rootCmd.AddCommand(tool.NewToolCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(importcmd.NewImportCmd())
	rootCmd.AddCommand(run.NewRunCmd(startDataKitFn))
}
