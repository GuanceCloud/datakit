package aliyunddos

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/aegis"
)

const (
	configSample = `
#[[inputs.aliyunddos]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，最小粒度5分钟，5m
#  interval = "5m"
#  ## 指标名称，默认值(aliyun_ddos)
#  metricName = ""
`
)

type DDoS struct {
	RegionID        string `toml:"region"`
	AccessKeyID     string `toml:"accessKeyId"`
	AccessKeySecret string `toml:"accessKeySecret"`
	Interval        string `toml:"interval"`
	MetricName      string `toml:"metricName"`
	client          *sdk.Client
	aclient         *aegis.Client
}
