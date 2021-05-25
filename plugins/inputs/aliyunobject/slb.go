package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	slbSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.slb]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path

	#pipeline = "aliyun_slb.p"
	# ##(optional) list of slb instanceid
	#instanceids = ['']
	
	# ##(optional) list of excluded slb instanceid
	#exclude_instanceids = ['']
`
	slbPipelineConfig = `
json(_, LoadBalancerId)
json(_, LoadBalancerStatus)
json(_, PayType)
json(_, Address)
json(_, RegionId)

`
)

type Slb struct {
	Disable            bool     `toml:"disable"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (s *Slb) disabled() bool {
	return s.Disable
}

func (s *Slb) run(ag *objectAgent) {
	var cli *slb.Client
	var err error
	p, err := newPipeline(s.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] slb new pipeline err:%s", err.Error())
		return
	}
	s.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = slb.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		if ag.isTest() {
			ag.testError = err
			return
		}
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		req := slb.CreateDescribeLoadBalancersRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		if len(s.InstancesIDs) > 0 {
			data, err := json.Marshal(s.InstancesIDs)
			if err == nil {
				req.LoadBalancerId = string(data)
			}
		}

		for {
			resp, err := cli.DescribeLoadBalancers(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				s.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if ag.isTest() {
					ag.testError = err
					return
				}
				break
			}

			if resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}
			pageNum++
			req.PageNumber = requests.NewInteger(pageNum)
		}

		if ag.isTest() {
			break
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (s *Slb) handleResponse(resp *slb.DescribeLoadBalancersResponse, ag *objectAgent) {

	moduleLogger.Debugf("SLB TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	for _, inst := range resp.LoadBalancers.LoadBalancer {
		tags := map[string]string{
			"name": fmt.Sprintf(`%s_%s`, inst.LoadBalancerName, inst.LoadBalancerId),
		}
		ag.parseObject(inst, "aliyun_slb", inst.LoadBalancerId, s.p, s.ExcludeInstanceIDs, s.InstancesIDs, tags)

	}
}
