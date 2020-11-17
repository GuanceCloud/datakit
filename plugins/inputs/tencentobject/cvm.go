package tencentobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
)

type Cvm struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Cvm) run(ag *objectAgent) {
	var client *cvm.Client
	var err error

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"

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

	var objs []map[string]interface{}

	for _, inst := range resp.Response.InstanceSet {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, *inst.InstanceName, *inst.InstanceId), `tencent_cvm`, inst, *inst.InstanceId, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}

	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		if ag.isTest() {
			ag.testResult.Result = append(ag.testResult.Result, data...)
		} else {
			io.NamedFeed(data, io.Object, inputName)
		}
	} else {
		moduleLogger.Errorf("%s", err)
		return
	}
}
