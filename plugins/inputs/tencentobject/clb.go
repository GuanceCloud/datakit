package tencentobject

import (
	"encoding/json"
	"fmt"
	"time"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	clbSampleConfig = `
#[inputs.tencentobject.clb]

# ## @param - custom tags - [list of clb instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded clb instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for clb object - [list of key:value element] - optional
#[inputs.tencentobject.clb.tags]
# key1 = 'val1'
`
)

type Clb struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (c *Clb) run(ag *objectAgent) {

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "clb.tencentcloudapi.com"
	var client *clb.Client
	var err error

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
			} else {
				c.handleResponse(response, ag)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (c *Clb) handleResponse(resp *clb.DescribeLoadBalancersResponse, ag *objectAgent) {

	moduleLogger.Debugf("CLB TotalCount=%d", *resp.Response.TotalCount)

	var objs []map[string]interface{}

	for _, inst := range resp.Response.LoadBalancerSet {

		if len(c.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range c.ExcludeInstanceIDs {
				if v == *inst.LoadBalancerId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(c.InstancesIDs) > 0 {
			exclude := true
			for _, v := range c.InstancesIDs {
				if v == *inst.LoadBalancerId {
					exclude = false
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, *inst.LoadBalancerName, *inst.LoadBalancerId),
		}

		if inst.CreateTime != nil {
			obj[`CreationTime`] = *inst.CreateTime
		}
		if inst.StatusTime != nil {
			obj[`StatusTime`] = *inst.StatusTime
		}
		if inst.ExpireTime != nil {
			obj[`ExpireTime`] = *inst.ExpireTime
		}
		if inst.AddressIPVersion != nil {
			obj[`AddressIPVersion`] = *inst.AddressIPVersion
		}
		if inst.Forward != nil {
			obj[`Forward`] = *inst.Forward
		}
		if inst.Status != nil {
			obj[`Status`] = *inst.Status
		}

		tags := map[string]interface{}{
			`__class`:        `CLB`,
			`provider`:       `tencent`,
			`LoadBalancerId`: *inst.LoadBalancerId,
		}

		if inst.ChargeType != nil {
			tags["ChargeType"] = *inst.ChargeType
		}
		if inst.Domain != nil {
			tags["Domain"] = *inst.Domain
		}
		if inst.LoadBalancerType != nil {
			tags["LoadBalancerType"] = *inst.LoadBalancerType
		}
		if inst.LoadBalancerName != nil {
			tags["LoadBalancerName"] = *inst.LoadBalancerName
		}
		if inst.Status != nil {
			tags["Status"] = *inst.Status
		}
		if inst.ProjectId != nil {
			tags["ProjectId"] = *inst.ProjectId
		}
		if inst.OpenBgp != nil {
			tags["OpenBgp"] = *inst.OpenBgp
		}
		if inst.SubnetId != nil {
			tags["SubnetId"] = *inst.SubnetId
		}
		if inst.IsolatedTime != nil {
			tags["IsolatedTime"] = *inst.IsolatedTime
		}
		if inst.VpcId != nil {
			tags["VpcId"] = *inst.VpcId
		}
		if inst.NetworkAttributes != nil {
			if inst.NetworkAttributes.InternetChargeType != nil {
				tags["InternetChargeType"] = *inst.NetworkAttributes.InternetChargeType
			}
			if inst.NetworkAttributes.InternetMaxBandwidthOut != nil {
				tags["InternetMaxBandwidthOut"] = *inst.NetworkAttributes.InternetMaxBandwidthOut
			}
			if inst.NetworkAttributes.BandwidthpkgSubType != nil {
				tags["BandwidthpkgSubType"] = *inst.NetworkAttributes.BandwidthpkgSubType
			}
		}
		if inst.Isolation != nil {
			tags["Isolation"] = *inst.Isolation
		}
		if inst.TargetRegionInfo != nil {
			if inst.TargetRegionInfo.Region != nil {
				tags["TargetRegion"] = *inst.TargetRegionInfo.Region
			}
			if inst.TargetRegionInfo.VpcId != nil {
				tags["TargetRegionVpcId"] = *inst.TargetRegionInfo.VpcId
			}
		}
		if inst.AnycastZone != nil {
			tags["AnycastZone"] = *inst.AnycastZone
		}
		if inst.VipIsp != nil {
			tags["VipIsp"] = *inst.VipIsp
		}
		if inst.IsDDos != nil {
			tags["IsDDos"] = *inst.IsDDos
		}
		if inst.MasterZone != nil {
			if inst.MasterZone.ZoneName != nil {
				tags["Zone"] = *inst.MasterZone.Zone
			}
		}

		//add custom tags
		for k, v := range c.Tags {
			tags[k] = v
		}

		//add global custom tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj["__tags"] = tags

		objs = append(objs, obj)
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
