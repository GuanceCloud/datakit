package tracerouter

import "time"

const (
	configSample = `
#[[inputs.tracerouter]]
#  ## 采集的频度，最小粒度10m
#  interval = '10m'
#  ## 指标集名称，默认值scanport
#  metricName = ''
#  ## trace addr ip or domain
#  addr = ''
#`
)

type TraceRouter struct {
	Metric           string        `json:"metricName" toml:"metricName"`
	Interval         string        `json:"interval" toml:"interval"`
	Addr             string        `json:"interval" toml:"addr"`
	IntervalDuration time.Duration `json:"-" toml:"-"`
}
