package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	slbSampleConfig = `
#[inputs.aliyunobject.slb]

# ## @param - custom tags - [list of slb instanceid] - optional
#instanceids = ['']

# ## @param - custom tags - [list of excluded slb instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for slb object - [list of key:value element] - optional
#[inputs.aliyunobject.slb.tags]
# key1 = 'val1'
`
)

type Slb struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (s *Slb) run(ag *objectAgent) {
	var cli *slb.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = slb.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := slb.CreateDescribeLoadBalancersRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		if len(s.InstancesIDs) > 0 {
			data, err := json.Marshal(s.InstancesIDs)
			if err == nil {
				req.LoadBalancerId = string(data)
			}
		}

		for {
			resp, err := cli.DescribeLoadBalancers(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				s.handleResponse(resp, ag)
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

func (s *Slb) handleResponse(resp *slb.DescribeLoadBalancersResponse, ag *objectAgent) {

	moduleLogger.Debugf("SLB TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []map[string]interface{}

	for _, inst := range resp.LoadBalancers.LoadBalancer {

		if len(s.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range s.ExcludeInstanceIDs {
				if v == inst.LoadBalancerId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s(%s)`, inst.LoadBalancerName, inst.LoadBalancerId),
		}

		obj[`NetworkType`] = inst.NetworkType
		obj[`CreationTime`] = inst.CreateTime
		obj[`AddressIPVersion`] = inst.AddressIPVersion

		tags := map[string]interface{}{
			`__class`:            `aliyun_slb`,
			`provider`:           `aliyun`,
			`InternetChargeType`: inst.InternetChargeType,
			`ResourceGroupId`:    inst.ResourceGroupId,
			`LoadBalancerId`:     inst.LoadBalancerId,
			`LoadBalancerName`:   inst.LoadBalancerName,
			`LoadBalancerStatus`: inst.LoadBalancerStatus,
			`PayType`:            inst.PayType,
			`Address`:            inst.Address,
			`AddressType`:        inst.AddressType,
			`RegionId`:           inst.RegionId,
			`MasterZoneId`:       inst.MasterZoneId,
			`SlaveZoneId`:        inst.SlaveZoneId,
			`VSwitchId`:          inst.VSwitchId,
			`VpcId`:              inst.VpcId,
		}

		//add ecs object custom tags
		for k, v := range s.Tags {
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

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
