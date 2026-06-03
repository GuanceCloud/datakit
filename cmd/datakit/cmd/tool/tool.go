// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tool implements the DataKit tool command.
package tool

import (
	"github.com/spf13/cobra"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
)

var runToolFn = cmds.RunTool

var (
	grokQFlag                         bool
	logFlag                           string
	showCloudInfoFlag                 bool
	ipinfoFlag                        string
	workspaceInfoFlag                 bool
	dumpSamplesFlag                   string
	defaultMainConfFlag               bool
	parseLineProtocolFlag             string
	jsonFlag                          bool
	updateIPDBFlag                    bool
	parseKVFileFlag                   string
	kvFileFlag                        string
	removeApmAutoInjectFlag           bool
	changeDockerContainersRuntimeFlag string
	ingestionCanaryFlag               bool
	ingestionCanaryIndexFlag          string
)

// NewToolCmd creates the tool command.
func NewToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Various tools for DataKit",
		Long:  `Various tools for DataKit`,
		RunE:  runTool,
	}

	cmd.Flags().BoolVar(&grokQFlag, "grokq", false, "query groks interactively")
	cmd.Flags().StringVar(&logFlag, "log", cmds.CommonLogFlag(), "log path")
	cmd.Flags().BoolVar(&showCloudInfoFlag, "show-cloud-info", false,
		"show current host's cloud info(currently support aliyun/tencent/aws/hwcloud/azure)")
	cmd.Flags().StringVar(&ipinfoFlag, "ipinfo", "", "show IP geo info")
	cmd.Flags().BoolVar(&workspaceInfoFlag, "workspace-info", false, "show workspace info")
	cmd.Flags().StringVar(&dumpSamplesFlag, "dump-samples", "", "dump all inputs samples")
	cmd.Flags().BoolVar(&defaultMainConfFlag, "default-main-conf", false, "print default datakit.conf")
	cmd.Flags().StringVar(&parseLineProtocolFlag, "parse-lp", "", "parse line-protocol file")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "output in JSON format(partially supported)")
	cmd.Flags().BoolVar(&updateIPDBFlag, "update-ipdb", false, "update local IPDB")
	cmd.Flags().StringVar(&parseKVFileFlag, "parse-kv-file", "", "parse input conf file with kv replaced")
	cmd.Flags().StringVar(&kvFileFlag, "kv-file", "", "specify the kv file path")
	cmd.Flags().BoolVar(&removeApmAutoInjectFlag, "remove-apm-auto-inject", false, "remove apm-auto-inject")
	cmd.Flags().StringVar(&changeDockerContainersRuntimeFlag, "change-docker-containers-runtime", "",
		"change the runtime of the created container, the value is runc or dk-runc")
	cmd.Flags().BoolVar(&ingestionCanaryFlag, "ingestion-canary", false, "test ingestion latency for metric/logging/tracing data")
	cmd.Flags().StringVar(&ingestionCanaryIndexFlag, "ingestion-canary-index", "default", "storage index for logging data (only for ingestion-canary)")

	return cmd
}

func runTool(cmd *cobra.Command, args []string) error {
	return runToolFn(cmds.ToolOptions{
		GrokQ:                         grokQFlag,
		LogPath:                       logFlag,
		ShowCloudInfo:                 showCloudInfoFlag,
		IPInfo:                        ipinfoFlag,
		WorkspaceInfo:                 workspaceInfoFlag,
		DumpSamples:                   dumpSamplesFlag,
		DefaultMainConf:               defaultMainConfFlag,
		ParseLineProtocol:             parseLineProtocolFlag,
		JSON:                          jsonFlag,
		UpdateIPDB:                    updateIPDBFlag,
		ParseKVFile:                   parseKVFileFlag,
		KVFile:                        kvFileFlag,
		RemoveApmAutoInject:           removeApmAutoInjectFlag,
		ChangeDockerContainersRuntime: changeDockerContainersRuntimeFlag,
		IngestionCanary:               ingestionCanaryFlag,
		IngestionCanaryIndex:          ingestionCanaryIndexFlag,
	})
}
