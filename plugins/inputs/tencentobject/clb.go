package tencentobject

import (
	"fmt"
	"time"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	clbSampleConfig = `
#[inputs.tencentobject.clb]

# ## @param - [list of clb instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded clb instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for clb object - [list of key:value element] - optional
#[inputs.tencentobject.clb.tags]
# key1 = 'val1'
`

	clbPipelineConfig = `
json(_, LoadBalancerId);
json(_, LoadBalancerName);
json(_, Domain);
json(_, Status);
`
)

type Clb struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (c *Clb) run(ag *objectAgent) {

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "clb.tencentcloudapi.com"
	var client *clb.Client
	var err error

	c.p, err = newPipeline(c.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] clb new pipeline err:%s", err.Error())
		return
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		client, err = clb.NewClient(credential, ag.RegionID, cpf)
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

		request := clb.NewDescribeLoadBalancersRequest()

		params := "{}"
		err := request.FromJsonString(params)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		} else {
			response, err := client.DescribeLoadBalancers(request)
			if err != nil {
				if _, ok := err.(*errors.TencentCloudSDKError); ok {
					moduleLogger.Errorf("api error, %s", err)
				} else {
					moduleLogger.Errorf("%s", err)
				}
				if ag.isTest() {
					ag.testError = err
					return
				}
			} else {
				c.handleResponse(response, ag)
			}
		}

		if ag.isTest() {
			return
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (c *Clb) handleResponse(resp *clb.DescribeLoadBalancersResponse, ag *objectAgent) {

	moduleLogger.Debugf("CLB TotalCount=%d", *resp.Response.TotalCount)

	for _, inst := range resp.Response.LoadBalancerSet {

		tags := map[string]string{
			"name": fmt.Sprintf(`%s(%s)`, *inst.LoadBalancerName, *inst.LoadBalancerId),
		}
		ag.parseObject(inst, "tencent_clb", *inst.LoadBalancerId, c.p, c.ExcludeInstanceIDs, c.InstancesIDs, tags)
	}
}
