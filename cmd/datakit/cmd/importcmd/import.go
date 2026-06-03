// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package importcmd implements the DataKit import command.
package importcmd

import (
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var runImportNowFn = cmds.RunImportNow

var (
	importPathFlag    string
	importLogFlag     string
	importDatawayFlag []string
)

// NewImportCmd creates the import command.
func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import recorded data to Guance Cloud",
		Long:  `Import used to play recorded history data to Guance Cloud.`,
		RunE:  runImport,
	}

	cmd.Flags().StringVarP(&importPathFlag, "path", "P", filepath.Join(datakit.InstallDir, "recorder"), "point data path")
	cmd.Flags().StringVar(&importLogFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().StringSliceVarP(&importDatawayFlag, "dataway", "D", nil, "dataway list")

	return cmd
}

func runImport(cmd *cobra.Command, args []string) error {
	return runImportNowFn(cmds.ImportOptions{
		Path:        importPathFlag,
		LogPath:     importLogFlag,
		DatawayURLs: importDatawayFlag,
	}, time.Now().UnixNano())
}
