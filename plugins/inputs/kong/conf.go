package kong

import "time"

const (
	configSample = `
#[[inputs.kong]]
#  addr = 'http://127.0.0.1:8001/metrics'
#  ## 采集的频度，最小粒度5分钟
#  interval = "5m"
#  ## 指标名，默认值(kong)
#  #[inputs.kong.tags]
#  #tags1 = "value1"
#  
`
)

type Kong struct {
	Addr             string            `toml:"addr"`
	Interval         string            `toml:"interval"`
	Metric           string            `toml:"metricName"`
	Tags             map[string]string `toml:"tags"`
	IntervalDuration time.Duration     `json:"-" toml:"-"`
}
