package aliyunrdsslowlog

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[rdsslowlog]]
#  ## 阿里云ak信息
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，最小粒度24小时
#  interval = "24h"
#  ## mysql/sqlserver
#  product = ["mysql", "sqlserver"]
#  ## 指标名，默认值(aliyun_rds_slow_log)
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
