// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dql implements the DataKit DQL command.
package dql

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runDQLFn = cmds.RunDQL

var (
	jsonFlag     bool
	autoJSONFlag bool
	verboseFlag  bool
	runFlag      string
	tokenFlag    string
	csvFlag      string
	forceFlag    bool
	hostFlag     string
	logFlag      string
)

// NewDQLCmd creates the DQL command.
func NewDQLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dql",
		Short: "Query DQL for various usage",
		Long:  `DQL used to query data. If no option specified, query interactively.`,
		RunE:  runDQL,
	}

	cmd.Flags().BoolVarP(&jsonFlag, "json", "J", false, "output in JSON format")
	cmd.Flags().BoolVar(&autoJSONFlag, "auto-json", false, "pretty output string if field/tag value is JSON")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "V", false, "verbosity mode")
	cmd.Flags().StringVarP(&runFlag, "run", "R", "", "run single DQL")
	cmd.Flags().StringVarP(&tokenFlag, "token", "T", "", "run query for specific token(workspace)")
	cmd.Flags().StringVar(&csvFlag, "csv", "", "Specify the directory")
	cmd.Flags().BoolVarP(&forceFlag, "force", "F", false, "overwrite csv if file exists")
	cmd.Flags().StringVarP(&hostFlag, "host", "H", "", "specify datakit host to query")
	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")

	return cmd
}

func runDQL(cmd *cobra.Command, args []string) error {
	return runDQLFn(cmds.DQLOptions{
		JSON:     jsonFlag,
		AutoJSON: autoJSONFlag,
		Verbose:  verboseFlag,
		Run:      runFlag,
		Token:    tokenFlag,
		CSV:      csvFlag,
		Force:    forceFlag,
		Host:     hostFlag,
		LogPath:  logFlag,
	})
}
