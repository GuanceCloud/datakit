package aliyunobject

import (
	"time"

	waf "github.com/aliyun/alibaba-cloud-sdk-go/services/waf-openapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	wafSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.waf]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_waf.p"
	# ## @param - custom tags for waf object - [list of key:value element] - optional
	
`
	wafPipelineConfig = `
json(_, Region)
json(_, PayType)
json(_, Status)
json(_, InDebt)
json(_, SubscriptionType)
`
)

type Waf struct {
	Disable      bool              `toml:"disable"`
	Tags         map[string]string `toml:"tags,omitempty"`
	PipelinePath string            `toml:"pipeline,omitempty"`
	p            *pipeline.Pipeline
}

func (e *Waf) disabled() bool {
	return e.Disable
}

func (e *Waf) run(ag *objectAgent) {
	var cli *waf.Client
	var err error
	p, err := newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] waf new pipeline err:%s", err.Error())
		return
	}
	e.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = waf.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := waf.CreateDescribeInstanceInfoRequest()
		resp, err := cli.DescribeInstanceInfo(req)
		if err != nil {
			moduleLogger.Errorf("waf object: %s", err)
			break
		}
		e.handleResponse(resp, ag)
		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Waf) handleResponse(resp *waf.DescribeInstanceInfoResponse, ag *objectAgent) {
	if resp.InstanceInfo.PayType == 0 {
		moduleLogger.Warnf("%s", "waf payType 0")
		return
	}
	tags := map[string]string{
		"name": resp.InstanceInfo.InstanceId,
	}
	ag.parseObject(resp.InstanceInfo, "aliyun_waf", resp.InstanceInfo.InstanceId, e.p, []string{}, []string{}, tags)
}
