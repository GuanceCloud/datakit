package aliyuncdn

import "github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"

const (
	aliyunCDNConfigSample = `
#[[inputs.aliyuncdn]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，分钟值
#  interval = "5m"
#  domains = []

#  [inputs.aliyuncdn.summary]
#     ## metric name
#     metricName = "aliyun_cdn_summary"

#  [inputs.aliyuncdn.metric]
#     ## 该参数为阿里云api action, 支持以下action
#     ## describeDomainBpsData,describeDomainTrafficData,describeDomainHitRateData,describeDomainReqHitRateData,describeDomainSrcBpsData,describeDomainSrcTrafficData,describeDomainUvData,describeDomainPvData,
#     ## describeDomainTopClientIpVisit, describeDomainISPData, describeDomainTopUrlVisit, describeDomainSrcTopUrlVisit, describeTopDomainsByFlow, describeDomainTopReferVisit
#     actions = ["describeDomainBpsData"]
#     metricName = "aliyun_cdn_metrics"
#     ## 在describeDomainBpsData、describeDomainTrafficData action下可以配置该参数，根据业务场景选择(非必须参数)
#     ispNameEn = ""
#     ## 在describeDomainBpsData、describeDomainTrafficData、describeDomainTopClientIpVisit action下可以配置该参数，根据业务场景选择(非必须参数)
#     locationNameEn = ""
#     ## 在describeDomainTopClientIpVisit action下可以配置该参数，根据业务场景选择(非必须参数)
#     sortBy = ""`
)

type Metric struct {
	Actions        []string `toml:"actions"`
	MetricName     string   `toml:"metricName"`
	IspNameEn      string   `toml:"ispNameEn"`
	LocationNameEn string   `toml:"locationNameEn"`
	SortBy         string   `toml:"sortBy"`
}

type Summary struct {
	MetricName string `toml:"metricName"`
}

type CDN struct {
	RegionID        string   `toml:"region"`
	AccessKeyID     string   `toml:"accessKeyId"`
	AccessKeySecret string   `toml:"accessKeySecret"`
	DomainName      []string `toml:"domains"`
	Metric          *Metric  `toml:"metric"`
	Summary         *Summary `toml:"summary"`
	Interval        string   `toml:"interval"`
	client          *cdn.Client
	test            bool   `toml:"-"`
	resData         []byte `toml:"-"`
}
