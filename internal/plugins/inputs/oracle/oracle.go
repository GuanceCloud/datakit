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
  name   = 'oracle'
  cmd    = "/usr/local/datakit/externals/oracle"

  ## Set true to enable election
  election = true

  ## The "--inputs" line below should not be modified.
  args = [
    '--interval'       , '1m'                        ,
    '--host'           , '<your-oracle-host>'        ,
    '--port'           , '1521'                      ,
    '--username'       , '<oracle-user-name>'        ,
    '--password'       , '<oracle-password>'         ,
    '--service-name'   , '<oracle-service-name>'     ,
  ]
  envs = [
    'LD_LIBRARY_PATH=/opt/oracle/instantclient:$LD_LIBRARY_PATH',
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  #############################
  # Parameter Description (Marked with * is mandatory field)
  #############################
  # *--interval       : Collect interval (Default is 1m)
  # *--host           : Oracle instance address (IP)
  # *--port           : Oracle listen port (Default is 1521)
  # *--username       : Oracle username
  # *--password       : Oracle password
  # *--service-name   : Oracle service name
`
)

var (
	inputName   = "oracle"
	catalogName = "db"
	l           = logger.DefaultSLogger("oracle")
)

type Input struct {
	external.ExternalInput
}

func (i *Input) Run() {
	l.Info("Only for measurement documentation information, should not be here.")
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
		&systemMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelElection}
}

func defaultInput() *Input {
	return &Input{
		ExternalInput: *external.NewExternalInput(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
