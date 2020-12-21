package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	ecsSampleConfig = `
#[inputs.aliyunobject.ecs]

# ## @param - [list of ecs instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded ecs instanceid] - optional
#exclude_instanceids = ['']
`
)

type Ecs struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Ecs) run(ag *objectAgent) {
	var cli *ecs.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = ecs.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
		if err == nil {
			break
		} else {
			moduleLogger.Errorf("%s", err)
			if ag.isTest() {
				ag.testError = err
				return
			}
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
		req := ecs.CreateDescribeInstancesRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		if len(e.InstancesIDs) > 0 {
			data, err := json.Marshal(e.InstancesIDs)
			if err == nil {
				req.InstanceIds = string(data)
			}
		}

		for {
			resp, err := cli.DescribeInstances(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				e.handleResponse(resp, ag)
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

func (e *Ecs) handleResponse(resp *ecs.DescribeInstancesResponse, ag *objectAgent) {

	var objs []map[string]interface{}

	for _, inst := range resp.Instances.Instance {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, inst.InstanceName, inst.InstanceId), `aliyun_ecs`, inst, inst.InstanceId, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
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
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return
		}

		if ag.mode == "debug" {
			fmt.Printf("%s", string(data))
		} else {
			io.NamedFeed(data, io.Object, inputName)
		}
	}

}
