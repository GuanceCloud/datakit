package aliyuncdn

import (
	"context"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type AliyunCDN struct {
	CDN              []*CDN `toml:"cdn"`
	runningInstances []*RunningInstance
	ctx              context.Context
	cancelFun        context.CancelFunc
	accumulator      telegraf.Accumulator
	logger           *models.Logger
}

type RunningInstance struct {
	cfg             *CDN
	agent           *AliyunCDN
	logger          *models.Logger
	runningProjects []*RunningProject
	metricName      string
	client          *cdn.Client
	domains         []string
}

func (_ *AliyunCDN) SampleConfig() string {
	return aliyunCDNConfigSample
}

func (_ *AliyunCDN) Description() string {
	return ""
}

func (_ *AliyunCDN) Gather(telegraf.Accumulator) error {
	return nil
}

func (c *AliyunCDN) Start(acc telegraf.Accumulator) error {
	if len(c.CDN) == 0 {
		c.logger.Warnf("no configuration found")
		return nil
	}

	c.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `aliyuncdn`,
	}

	c.logger.Infof("aliyun cdn start...")

	c.accumulator = acc

	for _, instCfg := range c.CDN {
		r := &RunningInstance{
			cfg:    instCfg,
			agent:  c,
			logger: c.logger,
		}

		c.runningInstances = append(c.runningInstances, r)

		go r.run(c.ctx)
	}

	return nil
}

func (cdn *AliyunCDN) Stop() {
	cdn.cancelFun()
}

func (r *RunningInstance) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		// check 配置
		if err := CheckCfg(r.cfg); err != nil {
			r.logger.Errorf("config error %v", err)
			return nil
		}

		// 域名采集
		r.domains = r.cfg.DomainName
		cli, err := cdn.NewClientWithAccessKey(r.cfg.RegionID, r.cfg.AccessKeyID, r.cfg.AccessKeySecret)
		if err != nil {
			r.logger.Errorf("create client failed, %s", err)
			return err
		}

		r.client = cli

		if len(r.domains) == 0 {
			r.domains = r.getDomain(r.cfg.Summary.MetricName, "")
		} else {
			for _, domain := range r.domains {
				r.getDomain(r.cfg.Summary.MetricName, domain)
			}
		}

		for _, action := range r.cfg.Metric.Actions {
			go r.exec(ctx, action)
		}

		internal.SleepContext(ctx, 10*time.Second)
	}

	return nil
}

func (r *RunningInstance) getDomain(metricName string, domain string) []string {
	var pageNumber = 1
	var pageSize = 50
	var domains = []string{}

	if metricName == "" {
		metricName = "aliyun_cdn_summary"
	}

	for {
		request := cdn.CreateDescribeUserDomainsRequest()
		request.RegionId = r.cfg.RegionID
		request.Scheme = "https"
		request.PageSize = requests.NewInteger(pageSize)
		request.PageNumber = requests.NewInteger(pageNumber)
		if domain != "" {
			request.DomainName = domain
		}

		response, err := r.client.DescribeUserDomains(request)
		if err != nil {
			r.logger.Errorf("[DescribeUserDomainsRequest] failed, %v", err.Error())
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

				r.agent.accumulator.AddFields(metricName, fields, tags)
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

func (r *RunningInstance) exec(ctx context.Context, action string) error {
	et := time.Now()
	st := et.Add(-time.Minute * 10)

	p := &RunningProject{
		accumulator: r.agent.accumulator,
		cfg:         r.cfg.Metric,
		client:      r.client,
		logger:      r.logger,
		domain:      r.domains,
		startTime:   unixTimeStrISO8601(st),
		endTime:     unixTimeStrISO8601(et),
	}

	// metricname
	p.metricName = r.cfg.Metric.MetricName
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
	inputs.Add("aliyuncdn", func() telegraf.Input {
		ac := &AliyunCDN{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
