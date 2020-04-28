package aliyuncdn

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	aliyunCDNConfigSample = `
#[[cdn]]
#  accessKeyId = ''
#  accessKeySecret = ''
#  region = "cn-hangzhou"
#  ## 采集的频度，分钟值
#  interval = "5m"
#  domains = [""]

#  [cdn.summary]
#     ## metric name
#     metricName = "aliyun_cdn_summary"

#  [cdn.metric]
#     ## metric name
#     actions = ["DescribeDomainBpsData"]
#     metricName = "aliyun_cdn_metrics"
#     ispNameEn = ""
#     locationNameEn = ""
#     sortBy = ""
#     layer = ""
#     merge = ""
#     field = ""
#     limit = ""
#     domainType=""
#     ispNames= ""
#     percent=""
`
)

type Metric struct {
	Actions        []string `toml:"actions`
	MetricName     string   `toml:"metricName"`
	IspNameEn      string   `toml:"ispNameEn"`
	LocationNameEn string   `toml:"locationNameEn"`
	Layer          string   `toml:"layer"`
	Merge          string   `toml:"merge"`
	Field          string   `toml:"field"`
	DomainType     string   `toml:"domainType"`
	IspNames       string   `toml:"ispNames"`
	LocationNames  string   `toml:"locationNames"`
	Percent        string   `toml:"percent"`
	SortBy         string   `toml:"sortBy"`
}

type Summary struct {
	MetricName string `toml:"metricName"`
}

type CDN struct {
	RegionID        string            `toml:"region"`
	AccessKeyID     string            `toml:"accessKeyId"`
	AccessKeySecret string            `toml:"accessKeySecret"`
	DomainName      []string          `toml:"domains"`
	Metric          *Metric           `toml:"metric"`
	Summary         *Summary          `toml:"summary"`
	Interval        internal.Duration `toml:"interval"`
}
