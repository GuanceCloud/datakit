package aliyunsecurity

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[security]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度
#  interval = "10m"
#  ## 指标名称，默认值(aliyun_security)
#  metricName = ""
`
)

type Security struct {
	RegionID        string            `toml:"region"`
	AccessKeyID     string            `toml:"accessKeyId"`
	AccessKeySecret string            `toml:"accessKeySecret"`
	Interval        internal.Duration `toml:"interval"`
	MetricName      string            `toml:"metricName"`
}
