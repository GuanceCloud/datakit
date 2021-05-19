package aliyunddos

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"
)

const (
	configSample = `
#[[inputs.aliyunddos]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，默认值(24h)
#  interval = "24h"
`
)

type DDoS struct {
	RegionID         string        `toml:"region"`
	AccessKeyID      string        `toml:"accessKeyId"`
	AccessKeySecret  string        `toml:"accessKeySecret"`
	Interval         string        `toml:"interval"`
	IntervalDuration time.Duration `json:"-" toml:"-"`
	MetricName       string        `toml:"metricName"`
	client           *sdk.Client   `toml:"-"`
	aclient          *aegis.Client `toml:"-"`
	test             bool          `toml:"-"`
	resData          []byte        `toml:"-"`
}
