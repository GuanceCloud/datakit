package oraclemonitor

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[oracle]]
#  ## 采集的频度，最小粒度5m
#  interval = '10s'
#  metricName = ''
#  instanceId = ''
#  instanceDesc = ''
#  server = '10.200.6.53'
#  port = '40022'
#  username = 'root'
#  password = 'root'
#  sid = ''
#  name = ''
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
	Server       string            `toml:"server"`
	Port         string            `toml:"port"`
	Sid          string            `toml:"sid"`
	Name         string            `toml:"name"`
	TType        string            `toml:"type"`
}
