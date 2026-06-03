// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package version implements the DataKit version command.
package version

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runVersionFn = cmds.RunVersion

var (
	logFlag            string
	disableUpgradeInfo bool
)

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version and upgrade info",
		Long:  `Version used to handle version related functions.`,
		RunE:  runVersion,
	}

	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().BoolVar(&disableUpgradeInfo, "upgrade-info-off", false, "do not show upgrade info")

	return cmd
}

func runVersion(cmd *cobra.Command, args []string) error {
	return runVersionFn(cmds.VersionOptions{
		LogPath:            logFlag,
		DisableUpgradeInfo: disableUpgradeInfo,
	})
}
