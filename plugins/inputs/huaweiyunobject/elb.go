package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/elb"
)

const (
	classicType          = `经典型`
	sharedType           = `共享型`
	sharedTypeEnterprise = `共享型_企业项目`
	elbSampleConfig      = `
#[inputs.huaweiyunobject.elb]

## elb type: 经典型、共享型、共享型_企业项目 - requried
type=""

## 地区和终端节点 https://developer.huaweicloud.com/endpoint?ELB
endpoint=""

# ## @param - custom tags - [list of Elb instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded Elb instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for Elb object - [list of key:value element] - optional
#[inputs.huaweiyunobject.elb.tags]
# key1 = 'val1'
`
)

type Elb struct {
	Type     string `toml:"type"`
	EndPoint string `toml:"endpoint"`
	//	ProjectID          string            `toml:"project_id"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Elb) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`elb.%s.myhuaweicloud.com`, ag.RegionID)
	}

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		switch e.Type {
		case classicType:
			elistV1, err := cli.ElbV1List(nil)
			if err != nil {
				moduleLogger.Errorf(`get elblist v1, %v`, err)
				return
			}

			e.handResponseV1(elistV1, ag)

		case sharedType, sharedTypeEnterprise:
			//			moduleLogger.Debugf(`cli %+#v`, cli)
			e.doV2Action(cli, ag)

		default:
			moduleLogger.Warnf(`wrong type`)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Elb) doV2Action(cli *huaweicloud.HWClient, ag *objectAgent) {

	var marker string
	limit := 100
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		opt := map[string]string{
			"limit":        fmt.Sprintf("%d", limit),
			"page_reverse": fmt.Sprintf("%v", true),
			"marker":       marker,
		}

		switch e.Type {
		case sharedType:
			elbsV20, err := cli.ElbV20List(opt)
			if err != nil {
				moduleLogger.Errorf(`get elblist v2.0, %v`, err)
				return
			}

			length := len(elbsV20.Loadbalancers)
			e.handResponseV2(elbsV20.Loadbalancers, ag)

			if length < limit {
				return
			}

			marker = elbsV20.Loadbalancers[length-1].ID
		case sharedTypeEnterprise:
			elbsV2, err := cli.ElbV2List(opt)
			if err != nil {
				moduleLogger.Errorf(`get elblist v2, %v`, err)
				return
			}

			length := len(elbsV2.Loadbalancers)
			e.handResponseV2(elbsV2.Loadbalancers, ag)

			if length < limit {
				return
			}

			marker = elbsV2.Loadbalancers[length-1].ID

		default:
			moduleLogger.Warnf(`wrong type`)
		}

	}

}

func (e *Elb) handResponseV1(resp *elb.ListLoadbalancersV1, ag *objectAgent) {

	moduleLogger.Debugf("Elb TotalCount=%d", resp.InstanceNum)
	var objs []map[string]interface{}

	for _, lb := range resp.Loadbalancers {
		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == lb.ID {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(e.InstancesIDs) > 0 {
			include := false
			for _, v := range e.InstancesIDs {
				if v == lb.ID {
					include = true
					break
				}
			}

			if !include {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID),
		}

		obj[`admin_state_up`] = lb.AdminStateUp
		obj[`bandwidth`] = lb.Bandwidth
		obj[`creation_time`] = lb.CreateTime
		obj[`description`] = lb.Description
		obj[`id`] = lb.ID

		obj[`name`] = lb.Name
		obj[`update_time`] = lb.UpdateTime
		obj[`vip_address`] = lb.VipAddress
		obj[`vip_subnet_id`] = lb.VipSubnetID

		tags := map[string]interface{}{
			`__class`:           `huaweiyun_elb`,
			`provider`:          `huaweiyun`,
			`status`:            lb.Status,
			`type`:              lb.Type,
			`vpc_id`:            lb.VpcID,
			`security_group_id`: lb.SecurityGroupID,
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

func (e *Elb) handResponseV2(lbs []elb.LoadbalancerV2, ag *objectAgent) {
	moduleLogger.Debugf("Elb TotalCount=%d", len(lbs))
	var objs []map[string]interface{}

	for _, lb := range lbs {
		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == lb.ID {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(e.InstancesIDs) > 0 {
			include := false
			for _, v := range e.InstancesIDs {
				if v == lb.ID {
					include = true
					break
				}
			}

			if !include {
				continue
			}
		}

		listeners, err := json.Marshal(lb.Listeners)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}

		pools, err := json.Marshal(lb.Pools)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}

		obj := map[string]interface{}{
			`__name`:                fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID),
			`provisioning_status`:   lb.ProvisioningStatus,
			`tenant_id`:             lb.TenantID,
			`updated_at`:            lb.UpdatedAt,
			`vip_port_id`:           lb.VipPortID,
			`admin_state_up`:        lb.AdminStateUp,
			`create_at`:             lb.CreatedAt,
			`enterprise_project_id`: lb.EnterpriseProjectID,
			`description`:           lb.Description,
			`id`:                    lb.ID,
			`name`:                  lb.Name,
			`listeners`:             listeners,
			`operating_status`:      lb.OperatingStatus,
			`pools`:                 pools,
			`vip_address`:           lb.VipAddress,
			`vip_subnet_id`:         lb.VipSubnetID,
		}

		tags := map[string]interface{}{
			`__class`:    `huaweiyun_elb`,
			`__provider`: `huaweiyun`,
			`project_id`: lb.ProjectID,
			`provider`:   lb.Provider,
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
