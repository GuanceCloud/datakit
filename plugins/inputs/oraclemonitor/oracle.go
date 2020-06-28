// +build linux,amd64

package oraclemonitor

import (
	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
#[[oracle]]
#  ## 采集的频度，最小粒度5m
#  interval = '10s'
#  ## 指标集名称，默认值oracle_monitor
#  metricName = ''
#  ## 实例ID(非必要属性)
#  instanceId = ''
#  ## # 实例描述(非必要属性)
#  instanceDesc = ''
#  ## oracle实例地址(ip)
#  host = ''
#  ## oracle监听端口
#  port = ''
#  ## 帐号
#  username = ''
#  ## 密码
#  password = ''
#  ## oracle的服务名
#  server = ''
#  ## 实例类型 例如 单实例、DG、RAC 等，非必要属性
#  type= 'singleInstance'
#`
)

type oraclemonitor struct{}

func (_ *oraclemonitor) Catalog() string {
	return "oracle"
}

func (_ *oraclemonitor) SampleConfig() string {
	return configSample
}

func (_ *oraclemonitor) Description() string {
	return ""
}

func (_ *oraclemonitor) Gather(telegraf.Accumulator) error {
	return nil
}

func (_ *oraclemonitor) Init() error {
	return nil
}

func (o *oraclemonitor) Start() error {
	// TODO: start oraclemonitor binary
	return nil
}

func init() {
	inputs.Add("oraclemonintor", func() inputs.Input {
		// TODO: check if binary oraclemonintor ready
		return &oraclemonitor{}
	})
}
