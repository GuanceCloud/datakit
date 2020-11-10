package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	ecsSampleConfig = `
#[inputs.huaweiyunobject.ecs]

## 地区和终端节点 https://developer.huaweicloud.com/endpoint?ECS  required
endpoint=""

# ## @param - [list of ecs instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded ecs instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.huaweiyunobject.ecs.tags]
# key1 = 'val1'
`
)

type Ecs struct {
	Tags map[string]string `toml:"tags,omitempty"`
	//	ProjectID          string            `toml:"project_id"`
	EndPoint           string   `toml:"endpoint"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
}

func (e *Ecs) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`ecs.%s.myhuaweicloud.com`, ag.RegionID)
	}

	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		limit := 100
		offset := 1

		for {

			select {
			case <-ag.ctx.Done():
				return
			default:
			}
			opts := map[string]string{
				"limit":  fmt.Sprintf("%d", limit),
				"offset": fmt.Sprintf("%d", offset),
			}

			ecss, err := cli.EcsList(opts)
			if err != nil {
				moduleLogger.Errorf("%v", err)
				return
			}
			e.handleResponse(ecss, ag)

			if ecss.Count < offset*limit {
				break
			}

			offset++
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Ecs) handleResponse(resp *huaweicloud.ListEcsResponse, ag *objectAgent) {

	moduleLogger.Debugf("ECS TotalCount=%d", resp.Count)

	var objs []map[string]interface{}

	for _, s := range resp.Servers {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, s.InstanceName, s.ID), `huaweiyun_ecs`, s, s.ID, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
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
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}
