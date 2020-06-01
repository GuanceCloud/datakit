package harborMonitor

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	baiduIndexConfigSample = `
#[[harbor]]
#  metricName = 'harbor'
#  domain = ''
#  interval = '10s'
#  username = ''
#  password = ''
`
)

type HarborCfg struct {
	MetricName string
	Domain     string
	Username   string
	Password   string
	Interval   internal.Duration `toml:"interval"`
}
