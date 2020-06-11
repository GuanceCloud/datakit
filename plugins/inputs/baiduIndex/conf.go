package baiduIndex

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	configSample = `
#[[baidu]]
#  ## 认证cookie
#  cookie = ''
#  keywords = ["测试"]
#  kind = 'new'
#  ## 采集的频度，最小粒度24小时
#  interval = "24h"
#  ## 指标名，默认值(baiduIndex)
#  metricName = ""
#  
`
)

type Baidu struct {
	Cookie     string            `toml:"cookie"`
	Keywords   []string          `toml:"keywords"`
	Kind       string            `toml:"kind"`
	Interval   internal.Duration `toml:"interval"`
	MetricName string            `toml:"metricName"`
}
