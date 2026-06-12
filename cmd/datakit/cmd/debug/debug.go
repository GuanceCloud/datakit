// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package debug implements the DataKit debug command.
package debug

import (
	"strings"

	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runDebugFn = cmds.RunDebug

const bugreportDatawayNoOptDefVal = "__datakit_default_dataway__"

var (
	logFlag                     string
	uploadLogFlag               bool
	globConfFlag                string
	regexConfFlag               string
	promConfFlag                string
	bugReportFlag               bool
	bugreportOSSFlag            string
	bugreportDatawayFlag        string
	bugreportDisableProfileFlag bool
	bugreportNMetricsFlag       int
	bugreportTagFlag            string
	inputConfFlag               string
	httpListenFlag              string
	filterFlag                  string
	dataFlag                    string
	kvFileFlag                  string
)

// NewDebugCmd creates the debug command.
func NewDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Various debug tools for DataKit",
		Long:  `Various debug options for DataKit`,
		RunE:  runDebug,
	}

	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().BoolVar(&uploadLogFlag, "upload-log", false, "upload log")
	cmd.Flags().StringVar(&globConfFlag, "glob-conf", "",
		"find the glob path and print it, provide a configuration file that contains glob statements written on separate lines.")
	cmd.Flags().StringVar(&regexConfFlag, "regex-conf", "",
		"export regex match results, provide a configuration file where the first line is a regular expression and the rest of the file is text.")
	cmd.Flags().StringVar(&promConfFlag, "prom-conf", "", "specify the prom input conf to debug")
	cmd.Flags().BoolVar(&bugReportFlag, "bug-report", false, "export DataKit running information for troubleshooting")
	cmd.Flags().StringVar(&bugreportOSSFlag, "oss", "", "upload bug report file to specified object storage(format host:bucket:ak:sk)")
	cmd.Flags().StringVar(&bugreportDatawayFlag, "bug-report-dataway", "", "upload bug report file via specified dataway URLs, comma-separated")
	cmd.Flags().Lookup("bug-report-dataway").NoOptDefVal = bugreportDatawayNoOptDefVal
	cmd.Flags().BoolVar(&bugreportDisableProfileFlag, "disable-profile", false, "disable profile collection when running bug-report")
	cmd.Flags().IntVar(&bugreportNMetricsFlag, "nmetrics", 3, "collect N batch of datakit metrics")
	cmd.Flags().StringVar(&bugreportTagFlag, "tag", "", "ping a tag to current bug report")
	cmd.Flags().StringVar(&inputConfFlag, "input-conf", "", "input TOML conf path")
	cmd.Flags().StringVar(&httpListenFlag, "http-listen", "", "setup HTTP server on debugging some inputs")
	cmd.Flags().StringVar(&filterFlag, "filter", "", "filter configure file(JSON)")
	cmd.Flags().StringVar(&dataFlag, "data", "", "data used during debugging")
	cmd.Flags().StringVar(&kvFileFlag, "kv-file", "", "kv file path")

	return cmd
}

func runDebug(cmd *cobra.Command, args []string) error {
	bugreportDatawayEnabled := cmd.Flags().Changed("bug-report-dataway")
	bugreportDataway := bugreportDatawayFlag
	if bugreportDataway == bugreportDatawayNoOptDefVal {
		bugreportDataway = ""
	}
	if bugreportDatawayEnabled && bugreportDataway == "" && len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		bugreportDataway = args[0]
	}

	return runDebugFn(cmds.DebugOptions{
		LogPath:                 logFlag,
		UploadLog:               uploadLogFlag,
		GlobConf:                globConfFlag,
		RegexConf:               regexConfFlag,
		PromConf:                promConfFlag,
		BugReport:               bugReportFlag,
		BugreportOSS:            bugreportOSSFlag,
		BugreportDataway:        bugreportDataway,
		BugreportDatawayEnabled: bugreportDatawayEnabled,
		BugreportDisableProfile: bugreportDisableProfileFlag,
		BugreportNMetrics:       bugreportNMetricsFlag,
		BugreportTag:            bugreportTagFlag,
		InputConf:               inputConfFlag,
		HTTPListen:              httpListenFlag,
		Filter:                  filterFlag,
		Data:                    dataFlag,
		KVFile:                  kvFileFlag,
	})
}
