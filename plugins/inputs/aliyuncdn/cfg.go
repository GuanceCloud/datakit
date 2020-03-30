package aliyuncdn

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"

const (
	aliyunCDNConfigSample = `
#[[cdn]]
#  access_key_id = ''
#  access_key_secret = ''
#  region_id = "cn-hangzhou"
#  ## 采集的频度，分钟值
#  interval = "5m"
	
#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91048.html
#     ## action
#     actionName = "DescribeDomainSrcHttpCodeData"
#     ## metric name
#     metricName = "domainSrcHttpCodeData"  
#     domainName = ""
#	  interval = 300

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/131851.html
#     actionName = "DescribeDomainQpsData"
#     metricName = "domainQpsData"
#     domainName = ""
#	  interval = 300
#     ispNameEn = ""
#     locationNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/137959.html
#     actionName = "DescribeDomainQpsDataByLayer"
#     metricName = "domainQpsDataByLayer"
#     domainName = ""
#	  interval = 300
#     ispNameEn = ""
#     locationNameEn = ""
#     layer = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/151539.html
#     actionName = "DescribeDomainSrcQpsData"
#     metricName = "domainSrcQpsData"
#     domainName = ""
#	  interval = 300


#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/141977.html
#     actionName = "DescribeDomainSrcTopUrlVisit"
#     metricName = "domainSrcTopUrlVisit"
#     domainName = ""
#     sortBy = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/141674.html
#     actionName = "DescribeDomainTopClientIpVisit"
#     metricName = "domainTopClientIpVisit"
#     domainName = ""
#     locationNameEn = ""
#     sortBy = ""
#     limit = 20

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91148.html
#     actionName = "DescribeDomainTopReferVisit"
#     metricName = "domainTopReferVisit"
#     domainName = ""
#     sortBy = ""
#     percent = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91149.html
#     actionName = "DescribeDomainTopUrlVisit"
#     metricName = "domainTopUrlVisit"
#     domainName = ""
#     sortBy = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91070.html
#     actionName = "DescribeDomainAverageResponseTime"
#     metricName = "domainQpsData"
#     domainName = ""
#     domainType = ""
#	  interval = 300
#     ispNameEn = ""
#     locationNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91088.html
#     actionName = "DescribeDomainFileSizeProportionData"
#     metricName = "domainFileSizeProportionData"
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91086.html
#     actionName = "DescribeDomainBpsDataByTimeStamp"
#     metricName = "domainBpsDataByTimeStamp"
#     domainName = ""
#     ispNames = ""
#     locationNames = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91100.html
#     actionName = "DescribeDomainISPData"
#     metricName = "domainISPData"
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91110.html
#     actionName = "DescribeDomainRealTimeBpsData"
#     metricName = "domainRealTimeBpsData"
#     domainName = ""
#     ispNameEn = ""
#     locationNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91136.html
#     actionName = "DescribeDomainRealTimeBpsData"
#     metricName = "domainRealTimeBpsData"
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91143.html
#     actionName = "DescribeDomainRealTimeSrcHttpCodeData"
#     metricName = "domainRealTimeSrcHttpCodeData"
#     domainName = ""
#     ispNameEn = ""
#     locationNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91144.html
#     actionName = "DescribeDomainRealTimeSrcTrafficData"
#     metricName = "domainRealTimeSrcTrafficData"
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91114.html
#     actionName = "DescribeDomainRealTimeByteHitRateData"
#     metricName = "domainRealTimeByteHitRateData"
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91128.html
#     actionName = "DescribeDomainBpsData"
#     metricName = "domainBpsData"
#	  interval = 300
#     domainName = ""
#     ispNameEn = ""
#     locationNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91046.html
#     actionName = "DescribeDomainBpsData"
#     metricName = "domainBpsData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91047.html
#     actionName = "DescribeDomainSrcTrafficData"
#     metricName = "domainSrcTrafficData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91050.html
#     actionName = "DescribeDomainHitRateData"
#     metricName = "domainHitRateData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91051.html
#     actionName = "DescribeDomainReqHitRateData"
#     metricName = "domainReqHitRateData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91056.html
#     actionName = "DescribeDomainHttpCodeData"
#     metricName = "domainHttpCodeData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91045.html
#     actionName = "DescribeDomainTrafficData"
#     metricName = "domainTrafficData"
#	  interval = 300
#     domainName = ""
#	  locationNameEn = ""
#     ispNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91045.html
#     actionName = "DescribeDomainTrafficData"
#     metricName = "domainTrafficData"
#	  interval = 300
#     domainName = ""
#	  locationNameEn = ""
#     ispNameEn = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91109.html
#     actionName = "DescribeDomainUvData"
#     metricName = "domainUvData"
#	  interval = 300
#     domainName = ""

#  [[cdn.actions]]
#     ## see: https://help.aliyun.com/document_detail/91105.html
#     actionName = "DescribeDomainPvData"
#     metricName = "domainPvData"
#	  interval = 300
#     domainName = ""
`
)

type Action struct {
	ActionName     string `toml:"actionName`
	MetricName     string `toml:"metricName"`
	DomainName     string `toml:"domainName"`
	Interval       int    `toml:"interval"`
	IspNameEn      string `toml:"ispNameEn"`
	LocationNameEn string `toml:"locationNameEn"`
	Layer          string `toml:"layer"`
	Merge          string `toml:"merge"`
	Field          string `toml:"field"`
	Limit          int    `toml:"limit"`
	DomainType     string `toml:"domainType"`
	IspNames       string `toml:"ispNames"`
	LocationNames  string `toml:"locationNames"`
	SortBy         string `toml:"sortBy"`
	Percent        string `toml:"percent"`
}

type CDN struct {
	RegionID        string            `toml:"region_id"`
	AccessKeyId     string            `toml:"access_key_id"`
	AccessKeySecret string            `toml:"access_key_secret"`
	Actions         []*Action         `toml:"actions"`
	Interval        internal.Duration `toml:"interval"`
}
