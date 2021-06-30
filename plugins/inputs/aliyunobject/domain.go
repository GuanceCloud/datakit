package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/domain"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	domainSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.domain]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_domain.p"
	# ##(optional) list of Domain instanceid
	#instanceids = []
	
	# ##(optional) list of excluded Domain instanceid
	#exclude_instanceids = []
`
	domainPipelineConfig = `
json(_, InstanceId)
json(_, DomainStatus)
json(_, DomainName)
json(_, DomainType)
json(_, ExpirationDateStatus)
`
)

type Domain struct {
	Disable            bool     `toml:"disable"`
	InstanceIDs        []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (dm *Domain) disabled() bool {
	return dm.Disable
}

func (dm *Domain) run(ag *objectAgent) {
	var cli *domain.Client
	var err error
	p, err := newPipeline(dm.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] domain new pipeline err:%s", err.Error())
		return
	}
	dm.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = domain.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := domain.CreateQueryDomainListRequest()

		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", pageNum, pageSize)

			req.PageNum = requests.NewInteger(pageNum)
			req.PageSize = requests.NewInteger(pageSize)

			resp, err := cli.QueryDomainList(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				dm.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)

				break
			}

			if !resp.NextPage {
				break
			}

			pageNum++

			req.PageNum = requests.NewInteger(pageNum)

		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (dm *Domain) handleResponse(resp *domain.QueryDomainListResponse, ag *objectAgent) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalItemNum, resp.PageSize, resp.CurrentPageNum)

	for _, d := range resp.Data.Domain {
		tags := map[string]string{
			"name": fmt.Sprintf("%s_%s", d.InstanceId, d.DomainName),
		}
		ag.parseObject(d, "aliyun_domain", d.DomainName, dm.p, dm.ExcludeInstanceIDs, dm.InstanceIDs, tags)
	}
}
