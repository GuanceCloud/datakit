package aliyunobject

import (
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	cdnSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.cdn]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_cdn.p"
	# ##(optional) list of cdn DomainName
	#domainNames = []
	# ##(optional) list of excluded cdn exclude_domainNames
	#exclude_domainNames = []
`
	cdnPipelineConifg = `
json(_, Cname)
json(_, CdnType)
json(_, DomainStatus)
json(_, SslProtocol)
json(_, ResourceGroupId)
`
)

type Cdn struct {
	Disable            bool     `toml:"disable"`
	DomainNames        []string `toml:"domainNames,omitempty"`
	ExcludeDomainNames []string `toml:"exclude_domainNames,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Cdn) disabled() bool {
	return e.Disable
}

func (e *Cdn) run(ag *objectAgent) {
	var cli *cdn.Client
	var err error
	p, err := newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] cdn new pipeline err:%s", err.Error())
		return
	}
	e.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = cdn.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		pageSize := 100
		req := cdn.CreateDescribeUserDomainsRequest()

		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", pageNum, pageSize)
			if len(e.DomainNames) > 0 {
				if pageNum <= len(e.DomainNames) {
					req.DomainName = e.DomainNames[pageNum-1]
				} else {
					break
				}
			} else {
				req.PageNumber = requests.NewInteger(pageNum)
				req.PageSize = requests.NewInteger(pageSize)
			}
			resp, err := cli.DescribeUserDomains(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				e.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if len(e.DomainNames) > 0 {
					pageNum++
					continue
				}
				break
			}
			if len(e.DomainNames) <= 0 && resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}

			pageNum++
			if len(e.DomainNames) <= 0 {
				req.PageNumber = requests.NewInteger(pageNum)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Cdn) handleResponse(resp *cdn.DescribeUserDomainsResponse, ag *objectAgent) {
	moduleLogger.Debugf("cdn TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)
	for _, inst := range resp.Domains.PageData {
		tags := map[string]string{
			"name": inst.DomainName,
		}
		ag.parseObject(inst, "aliyun_cdn", inst.DomainName, e.p, e.ExcludeDomainNames, e.DomainNames, tags)
	}
}
