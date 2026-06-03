// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package service implements the DataKit service command.
package service

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var (
	logFlag      string
	runServiceFn = runService
)

// NewServiceCmd creates the service command.
func NewServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage datakit service",
		Long:  `Service used to manage datakit service operations like start, stop, restart, etc.`,
	}

	cmd.PersistentFlags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.AddCommand(newServiceActionCmd("start", "start datakit service", func() {
		serviceAction = cmds.ServiceActionStart
	}))
	cmd.AddCommand(newServiceActionCmd("stop", "stop datakit service", func() {
		serviceAction = cmds.ServiceActionStop
	}))
	cmd.AddCommand(newServiceActionCmd("restart", "restart datakit service", func() {
		serviceAction = cmds.ServiceActionRestart
	}))
	cmd.AddCommand(newServiceActionCmd("uninstall", "uninstall datakit service", func() {
		serviceAction = cmds.ServiceActionUninstall
	}))
	cmd.AddCommand(newServiceActionCmd("reinstall", "reinstall datakit service", func() {
		serviceAction = cmds.ServiceActionReinstall
	}))

	return cmd
}

func newServiceActionCmd(use, short string, activate func()) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			activate()
			return runServiceFn()
		},
	}
}

var serviceAction cmds.ServiceAction

func runService() error {
	return cmds.RunService(cmds.ServiceOptions{
		LogPath: logFlag,
		Action:  serviceAction,
	})
}
