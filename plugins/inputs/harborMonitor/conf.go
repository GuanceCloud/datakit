package harborMonitor

const (
	harborConfigSample = `
#[[inputs.harborMonitor]]
#  ## 指标集名称
#  metricName = 'harbor'
#  ## 镜像仓库域名
#  domain = ''
#  ## 域名是否支持https
#  https = true
#  ## 采集频度
#  interval = '1h'
#  ## 帐号
#  username = ''
#  ## 密码
#  password = ''
`
)

type HarborMonitor struct {
	MetricName string
	Domain     string
	Https      bool
	Username   string
	Password   string
	Interval   string
	test       bool   `toml:"-"`
	resData    []byte `toml:"-"`
}
