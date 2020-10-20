package scanport

import (
	"time"
)

const (
	configSample = `
#[[inputs.scanport]]
#  ## 采集的频度，最小粒度10m
#  interval = '10m'
#  ## ip或域名，或cidr网络192.168.1.1/30, (default "127.0.0.1")
#  targets = ''
#  ## 端口范围, 如，80,81,88-1000 (default 80)
#  port = ''
#  ## 超时时长(毫秒) (default 200)
#  timeout = 200
#  ## 协称数 (default 100)
#  process = 100
#  desc = ''
#`
)

type Scanport struct {
	Metric           string            `json:"metricName" toml:"metricName"`
	Interval         string            `json:"interval" toml:"interval"`
	Targets          string            `json:"targets" toml:"targets"`
	Port             string            `json:"port" toml:"port"`
	Desc             string            `json:"desc" toml:"desc"`
	Tags             map[string]string `json:"tags" toml:"tags"`
	Timeout          int               `json:"timeout" toml:"timeout"`
	Process          int               `json:"process" toml:"process"`
	IntervalDuration time.Duration     `json:"-" toml:"-"`
	test             bool              `toml:"-"`
	resData          []byte            `toml:"-"`
}
