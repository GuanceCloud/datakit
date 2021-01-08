package aliyunobject

import (
	"encoding/json"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	redis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	redisSampleConfig = `
#[inputs.aliyunobject.redis]

#pipeline = "aliyun_redis.p"

# ## @param - [list of redis instanceid] - optional
#instanceids = []

# ## @param - [list of excluded redis instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for redis object - [list of key:value element] - optional
#[inputs.aliyunobject.redis.tags]
# key1 = 'val1'
`
	redisPipelineConifg = `

	json(_,"InstanceName","name");
	json(_,"RegionId");
	json(_,"InstanceStatus");
	json(_,"InstanceId");
	json(_,"NetworkType");
	json(_,"ChargeType");

`
)

type Redis struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string            `toml:"pipeline,omitempty"`
}

func (e *Redis) run(ag *objectAgent) {
	var cli *redis.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = redis.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := redis.CreateDescribeInstancesRequest()
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
				break
			}

			if resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}
			pageNum++
			req.PageNumber = requests.NewInteger(pageNum)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Redis) handleResponse(resp *redis.DescribeInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("redis TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)


	for _, inst := range resp.Instances.KVStoreInstance {
		parseObject(inst, "aliyun_waf", inst.InstanceId, e.PipelinePath, e.ExcludeInstanceIDs, e.InstancesIDs)
	}
}
