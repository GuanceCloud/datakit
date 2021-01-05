package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	redis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	redisSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.redis]
    # ##(optional) ignore this object, default is false
    #disable = false
	
    # ##(optional) list of redis instanceid
    #instanceids = []

    # ##(optional) list of excluded redis instanceid
    #exclude_instanceids = []
`
)

type Redis struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Redis) disabled() bool {
	return e.Disable
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

	var objs []map[string]interface{}

	for _, inst := range resp.Instances.KVStoreInstance {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, inst.InstanceName, inst.InstanceId), `aliyun_redis`, inst, inst.InstanceId, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
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
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	if ag.isDebug() {
		fmt.Printf("%s\n", string(data))
	} else {
		io.NamedFeed(data, io.Object, inputName)
	}
}
