package aliyunobject

import (
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	redis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"time"
)

const (
	redisSampleConfig = `
#[inputs.aliyunobject.redis]

# ## @param - custom tags - [list of redis instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded redis instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for redis object - [list of key:value element] - optional
#[inputs.aliyunobject.redis.tags]
# key1 = 'val1'
`
)

type Redis struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
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
		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == inst.InstanceId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		tags := map[string]interface{}{
			"__class":          "aliyun_redis",
			"provider":       "aliyun",
			"RegionId":         inst.RegionId,
			"ArchitectureType": inst.ArchitectureType,
			"ChargeType":       inst.ChargeType,
			"EngineVersion":    inst.EngineVersion,
			"ResourceGroupId":  inst.ResourceGroupId,
			"VSwitchId":        inst.VSwitchId,
			"VpcId":            inst.VpcId,
			"ZoneId":           inst.ZoneId,
			"ConnectionMode":   inst.ConnectionMode,
			"InstanceId":       inst.InstanceId,
			"InstanceStatus":   inst.InstanceStatus,
			"InstanceType":     inst.InstanceType,
			"NetworkType":      inst.NetworkType,
			"NodeType":         inst.NodeType,
			"PackageType":      inst.PackageType,
			"ReplacateId":      inst.ReplacateId,
			"SearchKey":        inst.SearchKey,
			"InstanceClass":    inst.InstanceClass,
			"PrivateIp":        inst.PrivateIp,
		}

		obj := map[string]interface{}{
			"__name":              inst.InstanceName,
			"DestroyTime":         inst.DestroyTime,
			"CreateTime":          inst.CreateTime,
			"Bandwidth":           inst.Bandwidth,
			"Capacity":            inst.Capacity,
			"Config":              inst.Config,
			"ConnectionDomain":    inst.ConnectionDomain,
			"Connections":         inst.Connections,
			"EndTime":             inst.EndTime,
			"Port":                inst.Port,
			"QPS":                 inst.QPS,
			"HasRenewChangeOrder": inst.HasRenewChangeOrder,
			"IsRds":               inst.IsRds,
			"UserName":            inst.UserName,
		}

		for _, t := range inst.Tags.Tag {
			if _, have := tags[t.Key]; !have {
				tags[t.Key] = t.Value
			} else {
				tags[`custom_`+t.Key] = t.Value
			}
		}

		for k, v := range e.Tags {
			tags[k] = v
		}

		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj["__tags"] = tags

		objs = append(objs, obj)
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
