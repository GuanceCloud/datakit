package aliyunobject

import (
	"time"

	waf "github.com/aliyun/alibaba-cloud-sdk-go/services/waf-openapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	wafSampleConfig = `
#[inputs.aliyunobject.waf]
#pipeline = "aliyun_waf.p"
# ## @param - custom tags for waf object - [list of key:value element] - optional
#[inputs.aliyunobject.waf.tags]
# key1 = 'val1'
`
	wafPipelineConfig = `
	json(_,"InstanceId","name");
	json(_,"Region");
	json(_,"PayType");
	json(_,"Status");
	json(_,"InDebt");
	json(_,"SubscriptionType");
`
)

type Waf struct {
	Tags         map[string]string `toml:"tags,omitempty"`
	PipelinePath string            `toml:"pipeline,omitempty"`
}

func (e *Waf) run(ag *objectAgent) {
	var cli *waf.Client
	var err error

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
	parseObject(resp.InstanceInfo, "aliyun_waf", resp.InstanceInfo.InstanceId, e.PipelinePath, []string{}, []string{})
}
