package scanport

import (
	"time"
)

const (
	configSample = `
#[inputs.scanport]
#  ## 采集的频度，最小粒度10m
#  interval = '10m'
#  ## 指标集名称，默认值scanport
#  metricName = ''
#  ## 网络cidr
#  cidr = ''
#  ## ip集合
#  ips = []
#  ## 协议
#  protocol = ["tcp", "udp"]
#  ## 端口范围
#  portStart = 1
#  portEnd = 65535
#  desc = ''
#`
)

type Scanport struct {
	Metric           string            `json:"metricName" toml:"metricName"`
	Interval         string            `json:"interval" toml:"interval"`
	Cidr             string            `json:"cidr" toml:"cidr"`
	Ips              []string          `json:"ips" toml:"ips"`
	Protocol         []string          `json:"protocol" toml:"protocol"`
	PortStart        int               `json:"portStart" toml:"portStart"`
	PortEnd          int               `json:"portEnd" toml:"portEnd"`
	Desc             string            `json:"desc" toml:"desc"`
	Tags             map[string]string `json:"tags" toml:"tags"`
	Timeout          time.Duration     `json:"-" toml:"-"`
	IntervalDuration time.Duration     `json:"-" toml:"-"`
}
