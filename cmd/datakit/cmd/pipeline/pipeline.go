// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pipeline implements the DataKit pipeline command.
package pipeline

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runPipelineFn = cmds.RunPipeline

var (
	categoryFlag string
	nsFlag       string
	nameFlag     string
	logFlag      string
	txtFlag      string
	fileFlag     string
	tableFlag    bool
	dateFlag     bool
)

// NewPipelineCmd creates the pipeline command.
func NewPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Debug pipeline scripts",
		Long:  `Pipeline used to debug exists pipeline script.`,
		RunE:  runPipeline,
	}

	cmd.Flags().StringVarP(&categoryFlag, "category", "C", "logging", "data category (logging, metric, ...)")
	cmd.Flags().StringVarP(&nsFlag, "namespace", "N", "default", "namespace (default, gitrepo, remote)")
	cmd.Flags().StringVarP(&nameFlag, "name", "P", "", "pipeline name")
	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().StringVarP(&txtFlag, "txt", "T", "", "text string for the pipeline or grok")
	cmd.Flags().StringVarP(&fileFlag, "file", "F", "", "text file path for the pipeline or grok")
	cmd.Flags().BoolVar(&tableFlag, "tab", false, "output result in table format")
	cmd.Flags().BoolVar(&dateFlag, "date", false, "append date display(according to local timezone) on timestamp")

	return cmd
}

func runPipeline(cmd *cobra.Command, args []string) error {
	return runPipelineFn(cmds.PipelineOptions{
		Category:  categoryFlag,
		Namespace: nsFlag,
		Name:      nameFlag,
		LogPath:   logFlag,
		Text:      txtFlag,
		FilePath:  fileFlag,
		Table:     tableFlag,
		Date:      dateFlag,
	})
}
