package aliyunrdsslowlog

import "github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

const (
	configSample = `
#[[inputs.aliyunrdsslowlog]]
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

type AliyunRDS struct {
	RegionID        string      `toml:"region"`
	AccessKeyID     string      `toml:"accessKeyId"`
	AccessKeySecret string      `toml:"accessKeySecret"`
	Product         []string    `toml:"product"`
	Interval        string      `toml:"interval"`
	MetricName      string      `toml:"metricName"`
	client          *rds.Client `toml:"-"`
	test            bool        `toml:"-"`
	resData         []byte      `toml:"-"`
}
