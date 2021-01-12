package tencentobject

import (
	"fmt"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	cvmSampleConfig = `
#[inputs.tencentobject.cvm]

# ## @param - [list of cvm instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded cvm instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags - [list of key:value element] - optional
#[inputs.tencentobject.cvm.tags]
# key1 = 'val1'
`

	cvmPipelineConfig = `
json(_, Name);
json(_, Region);
json(_, CreationDate);
`
)

type Cvm struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Cvm) run(ag *objectAgent) {
	var client *cvm.Client
	var err error

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"

	e.p, err = newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] cvm new pipeline err:%s", err.Error())
		return
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		client, err = cvm.NewClient(credential, ag.RegionID, cpf)
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

		request := cvm.NewDescribeInstancesRequest()

		params := "{}"
		err = request.FromJsonString(params)
		if err == nil {
			response, err := client.DescribeInstances(request)
			if _, ok := err.(*errors.TencentCloudSDKError); ok {
				moduleLogger.Errorf("An API error has returned: %s", err)
			} else {
				if err != nil {
					moduleLogger.Errorf("%s", err)
				} else {
					e.handleResponse(response, ag)
				}
			}
			if err != nil && ag.isTest() {
				ag.testError = err
			}
		} else {
			moduleLogger.Errorf("invalid params, %s", err)
			if ag.isTest() {
				ag.testError = err
			}
		}

		if ag.isTest() {
			return
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}

}

func (e *Cvm) handleResponse(resp *cvm.DescribeInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("CVM TotalCount=%d", *resp.Response.TotalCount)

	for _, inst := range resp.Response.InstanceSet {

		tags := map[string]string{
			"name": fmt.Sprintf(`%s(%s)`, *inst.InstanceName, *inst.InstanceId),
		}
		ag.parseObject(inst, "tencent_cvm", *inst.InstanceId, e.p, e.ExcludeInstanceIDs, e.InstancesIDs, tags)
	}
}
