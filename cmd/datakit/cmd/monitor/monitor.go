// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package monitor implements the DataKit monitor command.
package monitor

import (
	"time"

	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runMonitorFn = cmds.RunMonitor

var (
	toFlag            string
	maxTableWidthFlag int
	logFlag           string
	refreshFlag       time.Duration
	verboseFlag       bool
	moduleFlag        string
	inputFlag         string
	pathFlag          string
	timestampFlag     int64
	dumpMetricsFlag   bool
	quantileFlag      string
)

// NewMonitorCmd creates the monitor command.
func NewMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Show datakit running statistics",
		Long:  `Monitor used to show datakit running statistics`,
		RunE:  runMonitor,
	}

	cmd.Flags().StringVar(&toFlag, "to", "", "specify the DataKit(IP:Port) to show its metrics")
	cmd.Flags().IntVarP(&maxTableWidthFlag, "max-table-width", "W", 128, "set max table cell width")
	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().DurationVarP(&refreshFlag, "refresh", "R", 5*time.Second, "refresh interval")
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "V", false, "show all statistics info, default not show goroutine and inputs config info")
	cmd.Flags().StringVarP(&moduleFlag, "module", "M", "", "show only specified module stats, seprated by ',', i.e., -M filter,inputs")
	cmd.Flags().StringVarP(&inputFlag, "input", "I", "", "show only specified inputs stats, seprated by ',', i.e., -I cpu,mem")
	cmd.Flags().StringVarP(&pathFlag, "path", "P", "", "specify the metric file path")
	cmd.Flags().StringVarP(&quantileFlag, "quantile", "Q", "", "select quantiles(50/90/99) of summaries, default are avg values")
	cmd.Flags().Int64VarP(&timestampFlag, "timestamp", "T", 0, "specify the timestamp(ms) of these metrics")
	cmd.Flags().BoolVar(&dumpMetricsFlag, "dump-metrics", false, "dump monitor metrics to local file .monitor-metrics")

	return cmd
}

func runMonitor(cmd *cobra.Command, args []string) error {
	return runMonitorFn(cmds.MonitorOptions{
		To:            toFlag,
		MaxTableWidth: maxTableWidthFlag,
		LogPath:       logFlag,
		Refresh:       refreshFlag,
		Verbose:       verboseFlag,
		Module:        moduleFlag,
		OnlyInputs:    inputFlag,
		FilePath:      pathFlag,
		TimestampMS:   timestampFlag,
		DumpMetrics:   dumpMetricsFlag,
		Quantile:      quantileFlag,
	})
}
