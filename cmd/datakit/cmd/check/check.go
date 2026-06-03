// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package check implements the DataKit check command.
package check

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runCheckFn = cmds.RunCheck

var (
	logFlag            string
	checkConfigFlag    bool
	checkConfigDirFlag string
	checkSampleFlag    bool
)

// NewCheckCmd creates the check command.
func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check inputs configurations and samples",
		Long:  `Various check tools for DataKit`,
		RunE:  runCheck,
	}

	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().BoolVar(&checkConfigFlag, "config", false, "check inputs configures and datait.conf")
	cmd.Flags().StringVar(&checkConfigDirFlag, "config-dir", "", "check configures under specified path")
	cmd.Flags().BoolVar(&checkSampleFlag, "sample", false,
		"check all inputs config sample, to ensure all sample are valid TOML")

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	return runCheckFn(cmds.CheckOptions{
		LogPath:   logFlag,
		Config:    checkConfigFlag,
		ConfigDir: checkConfigDirFlag,
		Sample:    checkSampleFlag,
	})
}
