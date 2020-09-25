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

# ## @param - custom tags - [list of ecs instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded ecs instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.aliyunobject.ecs.tags]
# key1 = 'val1'
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

func (e *Ecs) handleResponse(resp *ecs.DescribeInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("ECS TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []map[string]interface{}

	for _, inst := range resp.Instances.Instance {

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

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, inst.InstanceName, inst.InstanceId),
		}

		obj[`Cpu`] = inst.Cpu
		obj[`Memory`] = inst.Memory
		obj[`CreationTime`] = inst.CreationTime
		obj[`DeletionProtection`] = inst.DeletionProtection
		obj[`ExpiredTime`] = inst.ExpiredTime
		obj[`GPUAmount`] = inst.GPUAmount
		obj[`GPUSpec`] = inst.GPUSpec
		obj[`ImageId`] = inst.ImageId
		obj[`IoOptimized`] = inst.IoOptimized
		obj[`SaleCycle`] = inst.SaleCycle
		obj[`StoppedMode`] = inst.StoppedMode
		obj[`VlanId`] = inst.VlanId

		tags := map[string]interface{}{
			`__class`:                 `aliyun_ecs`,
			`provider`:                `aliyun`,
			`ClusterId`:               inst.ClusterId,
			`DeploymentSetId`:         inst.DeploymentSetId,
			`EipAddress.AllocationId`: inst.EipAddress.AllocationId,
			`host`:                    inst.HostName,
			`HpcClusterId`:            inst.HpcClusterId,
			`InstanceChargeType`:      inst.InstanceChargeType,
			`InstanceId`:              inst.InstanceId,
			`InstanceName`:            inst.InstanceName,
			`InstanceType`:            inst.InstanceType,
			`OSName`:                  inst.OSName,
			`OSNameEn`:                inst.OSNameEn,
			`OSType`:                  inst.OSType,
			`RegionId`:                inst.RegionId,
			`ResourceGroupId`:         inst.ResourceGroupId,
			`Status`:                  inst.Status,
			`ZoneId`:                  inst.ZoneId,
			`NatIpAddress`:            inst.VpcAttributes.NatIpAddress,
			`VSwitchId`:               inst.VpcAttributes.VSwitchId,
			`VpcId`:                   inst.VpcAttributes.VpcId,
		}

		for i, ipaddr := range inst.InnerIpAddress.IpAddress {
			tags[fmt.Sprintf(`InnerIpAddress[%d]`, i)] = ipaddr
		}

		for i, ipaddr := range inst.PublicIpAddress.IpAddress {
			tags[fmt.Sprintf(`PublicIpAddress[%d]`, i)] = ipaddr
		}

		for i, ipaddr := range inst.VpcAttributes.PrivateIpAddress.IpAddress {
			tags[fmt.Sprintf(`PrivateIpAddress[%d]`, i)] = ipaddr
		}

		//tags on ecs instance
		for _, t := range inst.Tags.Tag {
			if _, have := tags[t.TagKey]; !have {
				tags[t.TagKey] = t.TagValue
			} else {
				tags[`custom_`+t.TagKey] = t.TagValue
			}
		}

		//add ecs object custom tags
		for k, v := range e.Tags {
			tags[k] = v
		}

		//add global tags
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
