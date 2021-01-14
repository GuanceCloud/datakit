package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ons"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	onsSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.rocketmq]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_rocketmq.p"
	
	# ##(optional) list of rocketmq instanceid
	#instanceids = []
	
	# ##(optional) list of excluded rocketmq instanceid
	#exclude_instanceids = []
`
	onsPipelineConfig = `
json(_, InstanceId)
json(_, InstanceStatus)
json(_, IndependentNaming)
json(_, InstanceName)
json(_, InstanceType)
`
)

type Ons struct {
	Disable            bool     `toml:"disable"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (r *Ons) disabled() bool {
	return r.Disable
}

func (r *Ons) run(ag *objectAgent) {
	if r.disabled() {
		return
	}

	var cli *ons.Client
	var err error
	p, err := newPipeline(r.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] rocketmq new pipeline err:%s", err.Error())
		return
	}
	r.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = ons.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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

		req := ons.CreateOnsInstanceInServiceListRequest()

		resp, err := cli.OnsInstanceInServiceList(req)

		if err == nil {
			r.handleResponse(resp, ag)
		} else {
			moduleLogger.Errorf("ons object: %s", err)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (r *Ons) handleResponse(resp *ons.OnsInstanceInServiceListResponse, ag *objectAgent) {

	for _, o := range resp.Data.InstanceVO {
		tags := map[string]string{
			"name": fmt.Sprintf(`%s_%s`, o.InstanceName, o.InstanceId),
		}
		ag.parseObject(o, "aliyun_rocketmq", o.InstanceId, r.p, r.ExcludeInstanceIDs, r.InstancesIDs, tags)
	}
}
