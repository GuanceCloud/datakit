// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package install implements the DataKit install command.
package install

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runInstallFn = cmds.RunInstall

var (
	logFlag        string
	telegrafFlag   bool
	scheckFlag     bool
	ipdbFlag       string
	symbolToolFlag bool
)

// NewInstallCmd creates the install command.
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install DataKit related packages and plugins",
		Long:  `Install used to install DataKit related packages and plugins`,
		RunE:  runInstall,
	}

	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().BoolVar(&telegrafFlag, "telegraf", false, "install Telegraf")
	cmd.Flags().BoolVar(&scheckFlag, "scheck", false, "install SCheck")
	cmd.Flags().StringVar(&ipdbFlag, "ipdb", "", "install IP database")
	cmd.Flags().BoolVar(&symbolToolFlag, "symbol-tools", false,
		"install tools for symbolizing crash backtrace address, including Android command line tools, ProGuard, Android-NDK, atosl, etc ...")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	return runInstallFn(cmds.InstallOptions{
		LogPath:     logFlag,
		Telegraf:    telegrafFlag,
		Scheck:      scheckFlag,
		IPDB:        ipdbFlag,
		SymbolTools: symbolToolFlag,
	})
}
