package baiduIndex

const (
	configSample = `
#[[inputs.baidu]]
#  ## 认证cookie，必填
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

type BaiduIndex struct {
	Cookie     string   `toml:"cookie"`
	Keywords   []string `toml:"keywords"`
	Kind       string   `toml:"kind"`
	Interval   string   `toml:"interval"`
	MetricName string   `toml:"metricName"`
}
