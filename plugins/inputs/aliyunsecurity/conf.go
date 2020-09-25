package aliyunsecurity

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/sas"
)

const (
	configSample = `
#[[inputs.aliyunsecurity]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度
#  interval = "10m"
`
)

type Security struct {
	RegionID         string `toml:"region"`
	AccessKeyID      string `toml:"accessKeyId"`
	AccessKeySecret  string `toml:"accessKeySecret"`
	Interval         string `toml:"interval"`
	MetricName       string `toml:"metricName"`
	client           *sas.Client
	aclient          *aegis.Client
	IntervalDuration time.Duration
}
