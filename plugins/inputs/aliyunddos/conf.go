package aliyunddos

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[ddos]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，最小粒度5分钟，5m
#  interval = "24h"
#  ## 指标名称，默认值(aliyun_ddos)
#  metricName = ""
`
)

type DDoS struct {
	RegionID        string            `toml:"region"`
	AccessKeyID     string            `toml:"accessKeyId"`
	AccessKeySecret string            `toml:"accessKeySecret"`
	Interval        internal.Duration `toml:"interval"`
	MetricName      string            `toml:"metricName"`
}
