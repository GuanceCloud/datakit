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

# ## @param - custom tags - [list of cvm instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded cvm instanceid] - optional
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
		} else {
			moduleLogger.Errorf("invalid params, %s", err)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}

}

func (e *Cvm) handleResponse(resp *cvm.DescribeInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("CVM TotalCount=%d", *resp.Response.TotalCount)

	var objs []map[string]interface{}

	for _, inst := range resp.Response.InstanceSet {

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == *inst.InstanceId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(e.InstancesIDs) > 0 {
			exclude := true
			for _, v := range e.InstancesIDs {
				if v == *inst.InstanceId {
					exclude = false
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, *inst.InstanceName, *inst.InstanceId),
		}

		obj[`Cpu`] = *inst.CPU
		obj[`Memory`] = *inst.Memory
		obj[`CreationTime`] = *inst.CreatedTime
		if inst.ExpiredTime != nil {
			obj[`ExpiredTime`] = *inst.ExpiredTime
		}
		if inst.ImageId != nil {
			obj[`ImageId`] = *inst.ImageId
		}
		if inst.RestrictState != nil {
			obj[`RestrictState`] = *inst.RestrictState
		}
		if inst.RenewFlag != nil {
			obj[`RenewFlag`] = *inst.RenewFlag
		}
		if inst.StopChargingMode != nil {
			obj[`StopChargingMode`] = *inst.StopChargingMode
		}

		tags := map[string]interface{}{
			`__class`:            `CVM`,
			`provider`:           `tencent`,
			`InstanceChargeType`: *inst.InstanceChargeType,
			`InstanceId`:         *inst.InstanceId,
			`InstanceType`:       *inst.InstanceType,
			`OSName`:             *inst.OsName,
			`RegionId`:           ag.RegionID,
			`Zone`:               *inst.Placement.Zone,
		}

		if inst.InstanceName != nil {
			tags[`InstanceName`] = *inst.InstanceName
		}

		if inst.InstanceState != nil {
			tags[`InstanceState`] = *inst.InstanceState
		}

		if inst.VirtualPrivateCloud != nil {
			if inst.VirtualPrivateCloud.VpcId != nil {
				tags[`VpcId`] = *inst.VirtualPrivateCloud.VpcId
			}
		}

		if inst.SystemDisk != nil {
			if inst.SystemDisk.DiskId != nil {
				tags["SystemDiskId"] = *inst.SystemDisk.DiskId
			}

			if inst.SystemDisk.DiskSize != nil {
				tags["SystemDiskSize"] = *inst.SystemDisk.DiskSize
			}

			if inst.SystemDisk.DiskType != nil {
				tags["SystemDiskType"] = *inst.SystemDisk.DiskType
			}
		}

		for i, disk := range inst.DataDisks {
			tags[fmt.Sprintf(`DataDisk[%d]ID`, i)] = *disk.DiskId
		}

		for i, ipaddr := range inst.PublicIpAddresses {
			tags[fmt.Sprintf(`PublicIpAddress[%d]`, i)] = *ipaddr
		}

		for i, ipaddr := range inst.PrivateIpAddresses {
			tags[fmt.Sprintf(`PrivateIpAddress[%d]`, i)] = *ipaddr
		}

		if inst.InternetAccessible != nil {
			if inst.InternetAccessible.PublicIpAssigned != nil {
				tags["PublicIpAssigned"] = *inst.InternetAccessible.PublicIpAssigned
			}
			if inst.InternetAccessible.InternetChargeType != nil {
				tags["InternetChargeType"] = *inst.InternetAccessible.InternetChargeType
			}
			if inst.InternetAccessible.InternetMaxBandwidthOut != nil {
				tags["InternetMaxBandwidthOut"] = *inst.InternetAccessible.InternetMaxBandwidthOut
			}
			if inst.InternetAccessible.BandwidthPackageId != nil {
				tags["BandwidthPackageId"] = *inst.InternetAccessible.BandwidthPackageId
			}
		}

		//tags on ecs instance
		for _, t := range inst.Tags {
			if _, have := tags[*t.Key]; !have {
				tags[*t.Key] = *t.Value
			} else {
				tags[`custom_`+*t.Key] = *t.Value
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
