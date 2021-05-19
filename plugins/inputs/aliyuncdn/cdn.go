package aliyuncdn

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var (
	l *logger.Logger

	inputName = "aliyuncdn"
)

func (_ *CDN) Catalog() string {
	return "aliyun"
}

func (_ *CDN) SampleConfig() string {
	return aliyunCDNConfigSample
}

func (_ *CDN) Description() string {
	return ""
}

func (_ *CDN) Gather() error {
	return nil
}

func (c *CDN) Run() {
	l = logger.SLogger(inputName)

	l.Info("aliyunCDN input started...")

	interval, err := time.ParseDuration(c.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			c.run()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (r *CDN) run() {
	// check 配置
	if err := CheckCfg(r); err != nil {
		l.Errorf("config error %v", err)
	}

	// 域名采集
	cli, err := cdn.NewClientWithAccessKey(r.RegionID, r.AccessKeyID, r.AccessKeySecret)
	if err != nil {
		l.Errorf("create client failed, %s", err)
	}

	r.client = cli

	if len(r.DomainName) == 0 {
		r.DomainName = r.getDomain(r.Summary.MetricName, "")
	} else {
		for _, domain := range r.DomainName {
			r.getDomain(r.Summary.MetricName, domain)
		}
	}

	for _, action := range r.Metric.Actions {
		go r.exec(action)
	}
}

func (r *CDN) getDomain(metricName string, domain string) []string {
	var pageNumber = 1
	var pageSize = 50
	var domains = []string{}

	if metricName == "" {
		metricName = "aliyun_cdn_summary"
	}

	for {
		request := cdn.CreateDescribeUserDomainsRequest()
		request.RegionId = r.RegionID
		request.Scheme = "https"
		request.PageSize = requests.NewInteger(pageSize)
		request.PageNumber = requests.NewInteger(pageNumber)
		if domain != "" {
			request.DomainName = domain
		}

		response, err := r.client.DescribeUserDomains(request)
		if err != nil {
			l.Errorf("[DescribeUserDomainsRequest] failed, %v", err.Error())
		}

		for _, item := range response.Domains.PageData {
			if item.DomainStatus == "online" {
				domains = append(domains, item.DomainName)
			}

			for _, point := range item.Sources.Source {
				tags := map[string]string{}
				fields := map[string]interface{}{}

				tags["cdnType"] = item.CdnType
				tags["cname"] = item.Cname
				tags["domainName"] = item.DomainName
				tags["domainStatus"] = item.DomainStatus
				tags["sslProtocol"] = item.SslProtocol

				fields["gmtCreated"] = RFC3339(item.GmtCreated)
				fields["gmtModified"] = RFC3339(item.GmtModified)
				fields["resourceGroupId"] = item.ResourceGroupId
				fields["description"] = item.Description
				fields["content"] = point.Content
				fields["port"] = point.Port
				fields["priority"] = ConvertToNum(point.Priority)
				fields["type"] = point.Type
				fields["weight"] = ConvertToNum(point.Weight)

				pt, err := influxdb.NewPoint(r.Summary.MetricName, tags, fields, time.Now())
				if err != nil {
					l.Errorf("[influxdb convert point] failed, %v", err.Error())
				}

				r.resData = []byte(pt.String())

				err = io.NamedFeed([]byte(pt.String()), datakit.Metric, inputName)
			}
		}

		total := response.TotalCount
		if int64(pageNumber*pageSize) >= total {
			break
		}

		pageNumber = pageNumber + 1
	}

	return domains
}

func (r *CDN) exec(action string) error {
	et := time.Now()
	st := et.Add(-time.Minute * 10)

	p := &RunningProject{
		cfg:       r.Metric,
		client:    r.client,
		domain:    r.DomainName,
		startTime: unixTimeStrISO8601(st),
		endTime:   unixTimeStrISO8601(et),
	}

	// metricname
	p.metricName = r.Metric.MetricName
	if p.metricName == "" {
		p.metricName = "aliyun_cdn_metrics"
	}

	go p.commond(action)

	return nil
}

func (run *RunningProject) commond(action string) {
	switch action {
	case "describeDomainBpsData":
		run.DescribeDomainBpsData()
	case "describeDomainTrafficData":
		run.DescribeDomainTrafficData()
	case "describeDomainHitRateData":
		run.DescribeDomainHitRateData()
	case "describeDomainReqHitRateData":
		run.DescribeDomainReqHitRateData()
	case "describeDomainSrcBpsData":
		run.DescribeDomainSrcBpsData()
	case "describeDomainSrcTrafficData":
		run.DescribeDomainSrcTrafficData()
	case "describeDomainUvData":
		run.DescribeDomainUvData()
	case "describeDomainPvData":
		run.DescribeDomainPvData()
	case "describeDomainTopClientIpVisit":
		run.DescribeDomainTopClientIpVisit()
	case "describeDomainISPData":
		run.DescribeDomainISPData()
	case "describeDomainTopUrlVisit":
		run.DescribeDomainTopUrlVisit()
	case "describeDomainSrcTopUrlVisit":
		run.DescribeDomainSrcTopUrlVisit()
	case "describeTopDomainsByFlow":
		run.DescribeTopDomainsByFlow()
	case "describeDomainTopReferVisit":
		run.DescribeDomainTopReferVisit()
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &CDN{}
	})
}
