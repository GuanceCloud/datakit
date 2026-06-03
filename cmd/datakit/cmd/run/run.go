// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package run implements the DataKit run command.
package run

import (
	"github.com/spf13/cobra"
)

var runInContainer bool

// NewRunCmd creates the run command.
func NewRunCmd(startFn func(bool) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Select DataKit running mode",
		Long:  `Run used to select different datakit running mode(default running as service).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(startFn)
		},
	}

	cmd.Flags().BoolVarP(&runInContainer, "container", "C", false, "running in container mode")

	return cmd
}

func runRun(startFn func(bool) error) error {
	return startFn(runInContainer)
}

// GetRunInContainer returns the container flag value.
func GetRunInContainer() bool {
	return runInContainer
}
