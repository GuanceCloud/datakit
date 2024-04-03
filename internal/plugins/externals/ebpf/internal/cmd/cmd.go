//go:build linux
// +build linux

// Package cmd implements datakit-ebpf command
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/cmd/run"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "datakit-ebpf",
		Short: "datakit-ebpf is a system and network observer based on BPF and eBPF.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	rootCmd.AddCommand(newCompletionCmd(rootCmd))
	rootCmd.AddCommand(run.NewRunCmd())
	return rootCmd
}

func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
