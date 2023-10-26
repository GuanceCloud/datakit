// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package db2 collect IBM Db2 metrics by wrap a external input.
package db2

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
  name   = 'db2'
  cmd    = "/usr/local/datakit/externals/db2"

  ## Set true to enable election
  election = true

  ## The "--inputs" line below should not be modified.
  args = [
    '--interval'       , '1m'                        ,
    '--host'           , '<db2-host>'                ,
    '--port'           , '50000'                     ,
    '--username'       , 'db2inst1'                  ,
    '--password'       , '<db2-password>'            ,
    '--database'       , '<db2-database-name>'       ,
    '--service-name'   , '<db2-service-name>'        ,
  ]
  envs = [
    'LD_LIBRARY_PATH=/opt/ibm/clidriver/lib:$LD_LIBRARY_PATH',
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  #############################
  # Parameter Description (Marked with * is mandatory field)
  #############################
  # *--interval       : Collect interval (Default is 1m)
  # *--host           : IBM Db2 nstance address (IP)
  # *--port           : IBM Db2 listen port (Default is 50000)
  # *--username       : IBM Db2 username (Default is db2inst1)
  # *--password       : IBM Db2 password
  # *--database       : IBM Db2 database name
  #  --service-name   : IBM Db2 service name
`
)

var (
	inputName   = "db2"
	catalogName = "db"
	l           = logger.DefaultSLogger(inputName)
)

type Input struct {
	external.Input
}

func (ipt *Input) Run() {
	l.Info("Only for measurement documentation information, should not be here.")
}

func (ipt *Input) Catalog() string { return catalogName }

func (ipt *Input) SampleConfig() string { return configSample }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&instanceMeasurement{},
		&databaseMeasurement{},
		&bufferPoolMeasurement{},
		&tableSpaceMeasurement{},
		&transactionLogMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
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
