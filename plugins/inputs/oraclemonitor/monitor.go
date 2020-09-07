package oraclemonitor

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[[oraclemonitor]]
  ## 采集的频度，最小粒度5m
  interval = '1m'
  ## 指标集名称，默认值oracle_monitor
  metricName = 'oracle_monitor'
  ## 实例ID(非必要属性)
  instanceId = 'oracle01'
  ## # 实例描述(非必要属性)
  instanceDesc = 'DBA团队自建Oracle单实例-booboo'
  ## oracle实例地址(ip)
  host = 'xxx.xxx.xx.x'
  ## oracle监听端口
  port = '1521'
  ## 帐号
  username = 'xxxxxx'
  ## 密码
  password = 'xxxxxx'
  ## oracle的服务名
  server = 'testdb.zhuyun'
  ## 实例类型 例如 single、dg、rac(require)
  cluster= 'single'
  ## 采集的oracle版本，支持10g, 11g, 12c
  version = '11g'
`
)

var (
	inputName = "oraclemonitor"
)

type OracleMonitor plugins.OracleMonitor

func (_ *OracleMonitor) Catalog() string {
	return "oracle"
}

func (_ *OracleMonitor) SampleConfig() string {
	return configSample
}

func (o *OracleMonitor) Run() {
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &OracleMonitor{}
	})
}
