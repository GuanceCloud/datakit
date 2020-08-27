package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ons"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	onsSampleConfig = `
#[inputs.aliyunobject.rocketmq]

# ## @param - custom tags - [list of rocketmq instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded rocketmq instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for rocketmq object - [list of key:value element] - optional
#[inputs.aliyunobject.rocketmq.tags]
# key1 = 'val1'
`
)

type Ons struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (r *Ons) run(ag *objectAgent) {
	var cli *ons.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = ons.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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

		req := ons.CreateOnsInstanceInServiceListRequest()
		moduleLogger.Debugf("%+#v", req)

		resp, err := cli.OnsInstanceInServiceList(req)

		if err == nil {
			r.handleResponse(resp, ag)
		} else {
			moduleLogger.Errorf("%s", err)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (r *Ons) handleResponse(resp *ons.OnsInstanceInServiceListResponse, ag *objectAgent) {

	moduleLogger.Debugf("%+#v", resp)
	var objs []*map[string]interface{}

	for _, o := range resp.Data.InstanceVO {

		inc := false
		for _, isid := range ag.Ons.InstancesIDs {
			if isid == o.InstanceId {
				inc = true
				break
			}
		}

		if len(ag.Ons.InstancesIDs) > 0 && !inc {
			continue
		}

		exclude := false
		for _, isId := range ag.Ons.ExcludeInstanceIDs {
			if o.InstanceId == isId {
				exclude = true
				break
			}
		}

		if exclude {
			continue
		}

		tags := map[string]interface{}{
			"__class":           "aliyun_rocketmq",
			"__provider":        "aliyun",
			"InstanceId":        o.InstanceId,
			"IndependentNaming": o.IndependentNaming,
			"InstanceName":      o.InstanceName,
			"InstanceStatus":    o.InstanceStatus,
			"InstanceType":      o.InstanceType,
		}

		for _, t := range o.Tags.Tag {
			tags[t.Key] = t.Value
		}

		//add Ons object custom tags
		for k, v := range r.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj := &map[string]interface{}{
			"__name":      fmt.Sprintf(`ons_%s`, o.InstanceId),
			"__tags":      tags,
			"ReleaseTime": o.ReleaseTime,
		}
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
