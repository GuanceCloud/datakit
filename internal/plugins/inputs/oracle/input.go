// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oracle collect Oracle metrics by wrap a external input.
package oracle

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]
  daemon = true
  name   = "oracle"
  cmd    = "/usr/local/datakit/externals/oracle"

  ## Set true to enable election
  election = true

  ## Modify below if necessary.
  ## The password use environment variable named "ENV_INPUT_ORACLE_PASSWORD".
  args = [
    "--interval"        , "1m"                           ,
    "--host"            , "<your-oracle-host>"           ,
    "--port"            , "1521"                         ,
    "--username"        , "<oracle-user-name>"           ,
    "--service-name"    , "<oracle-service-name>"        ,
    "--slow-query-time" , "0s"                           ,
    "--log"             , "/var/log/datakit/oracle.log"  ,
  ]
  envs = [
    "ENV_INPUT_ORACLE_PASSWORD=<oracle-password>",
    "LD_LIBRARY_PATH=/opt/oracle/instantclient:$LD_LIBRARY_PATH",
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  ## Run a custom SQL query and collect corresponding metrics.
  # [[inputs.external.custom_queries]]
  #   sql = '''
  #     SELECT
  #       GROUP_ID, METRIC_NAME, VALUE
  #     FROM GV$SYSMETRIC
  #   '''
  #   metric = "oracle_custom"
  #   tags = ["GROUP_ID", "METRIC_NAME"]
  #   fields = ["VALUE"]

  #############################
  # Parameter Description (Marked with * is required field)
  #############################
  # *--interval                   : Collect interval (Default is 1m).
  # *--host                       : Oracle instance address (IP).
  # *--port                       : Oracle listen port (Default is 1521).
  # *--username                   : Oracle username.
  # *--service-name               : Oracle service name.
  # *--slow-query-time            : Oracle slow query time threshold defined. If larger than this, the executed sql will be reported.
  # *--log                        : Collector log path.
  # *ENV_INPUT_ORACLE_PASSWORD    : Oracle password.
`
)

var (
	inputName   = "oracle"
	catalogName = "db"
	l           = logger.DefaultSLogger("oracle")
)

type Input struct {
	external.Input
}

func (*Input) Run() {
	l.Info("Only for measurement documentation information, should not be here.")
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
		&systemMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelElection}
}

func defaultInput() *Input {
	return &Input{
		Input: *external.NewInput(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
