package mysqlmonitor

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[mysql]]
#  ## 采集的频度，最小粒度5m
#  interval = '10s'
#  metricName = ''
#  instanceId = ''
#  instanceDesc = ''
#  host = '10.200.6.53'
#  port = '40022'
#  username = 'root'
#  password = 'root'
#  database = ''
#  
`
)

type Mysql struct {
	Interval     internal.Duration `toml:"interval"`
	MetricName   string            `toml:"metricName"`
	InstanceId   string            `toml:"instanceId"`
	Username     string            `toml:"username"`
	Password     string            `toml:"password"`
	InstanceDesc string            `toml:"instanceDesc"`
	Host         string            `toml:"host"`
	Port         string            `toml:"port"`
	Database     string            `toml:"database"`
}
