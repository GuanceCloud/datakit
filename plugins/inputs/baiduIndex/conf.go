package baiduIndex

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	baiduIndexConfigSample = `
#[[baiduIndex]]
#  metricName = 'baiduIndex'
#  interval = '10s'
#  cookie = ''
#  keywords = ''
`
)

type BaiduIndexCfg struct {
	MetricName string
	Cookie     string
	Keywords   string
	Interval   internal.Duration `toml:"interval"`
}
