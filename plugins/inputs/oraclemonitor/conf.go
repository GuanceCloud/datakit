// +build linux,amd64

package oraclemonitor

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

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
#
`
)

type Oracle struct {
	Interval     internal.Duration `toml:"interval"`
	MetricName   string            `toml:"metricName"`
	InstanceId   string            `toml:"instanceId"`
	Username     string            `toml:"username"`
	Password     string            `toml:"password"`
	InstanceDesc string            `toml:"instanceDesc"`
	Host         string            `toml:"host"`
	Port         string            `toml:"port"`
	Server       string            `toml:"server"`
	TType        string            `toml:"type"`
}
