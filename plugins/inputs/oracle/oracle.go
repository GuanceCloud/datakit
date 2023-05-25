// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oracle collect Oracle metrics by wrap a external input.
package oracle

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]
  daemon = true
  name = 'oracle'
  cmd  = "/usr/local/datakit/externals/oracle"

  ## Set true to enable election
  election = true

  args = [
    '--interval'       , '1m'                        ,
    '--host'           , '<your-oracle-host>'        ,
    '--port'           , '1521'                      ,
    '--username'       , '<oracle-user-name>'        ,
    '--password'       , '<oracle-password>'         ,
    '--service-name'   , '<oracle-service-name>'     ,
  ]
  envs = [
    'LD_LIBRARY_PATH=/opt/oracle/instantclient_19_8:$LD_LIBRARY_PATH',
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  #############################
  # 参数说明(标 * 为必选项)
  #############################
  # *--interval       : 采集的频度，最小粒度 5m
  # *--host           : Oracle 实例地址(ip)
  #  --port           : Oracle 监听端口
  # *--username       : Oracle 用户名
  # *--password       : Oracle 密码
  # *--service-name   : Oracle 的服务名
  # *--query          : 自定义查询语句，格式为 <sql:metricName:tags>，sql 为自定义采集的语句，tags 填入使用 tag 字段
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
	l.Info("oracle started...")
	i.ExternalInput.Run()
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
