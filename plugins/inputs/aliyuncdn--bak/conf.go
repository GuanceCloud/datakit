package aliyuncdn

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	aliyunCDNConfigSample = `
#[[cdn]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  domain = ["xxxx"]

// domain
#  [cdn.summary]
#     ## metric name
#     metricName = "aliyun_cdn_summary"
#     interval = "5m"

#   ## 流量带宽
#  [cdn.metrics]
#     ## metric name
#     metricName = "aliyun_cdn_metrics"  
#     interval = "5m"
`
)

type Config struct {
	MetricName string            `toml:"metricName"`
	Interval   internal.Duration `toml:"interval"`
}

type CDN struct {
	RegionID        string   `toml:"region"`
	AccessKeyID     string   `toml:"accessKeyId"`
	AccessKeySecret string   `toml:"accessKeySecret"`
	DomainName      []string `toml:"domain"`
	// Config          map[string]*Config
	Metrics Config `toml:"metrics"`
	Summary Config `toml:"summary"`
}
