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

# ## @param - [list of rocketmq instanceid] - optional
#instanceids = []

# ## @param - [list of excluded rocketmq instanceid] - optional
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

	var objs []map[string]interface{}

	for _, o := range resp.Data.InstanceVO {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`ons_%s`, o.InstanceId), `aliyun_rocketmq`, o, o.InstanceId, r.ExcludeInstanceIDs, r.InstancesIDs); obj != nil {
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
	io.NamedFeed(data, io.Object, inputName)
}
