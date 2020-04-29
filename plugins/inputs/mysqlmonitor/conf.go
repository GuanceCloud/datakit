package mysqlmonitor

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[mysql]]
#  ## 采集的频度，最小粒度5m
#  interval = '5m'
#  ## 指标集名称，默认值(mysql_monitor)
#  metricName = ''
#  instanceId = ''
#  instanceDesc = ''
#  host = '10.200.6.53'
#  port = '3306'
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
