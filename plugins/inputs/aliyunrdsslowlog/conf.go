package aliyunrdsslowlog

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[rdsslowlog]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，最小粒度24小时
#  interval = "24h"
#  ## mysql/sqlserver
#  product = ["mysql"]
#  metricName = ""
#  
`
)

type RDSslowlog struct {
	RegionID        string            `toml:"region"`
	AccessKeyID     string            `toml:"accessKeyId"`
	AccessKeySecret string            `toml:"accessKeySecret"`
	Product         []string          `toml:"product"`
	Interval        internal.Duration `toml:"interval"`
	MetricName      string            `toml:"metricName"`
}
