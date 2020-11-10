package huaweiyunobject

import (
	"encoding/json"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	mysqlSampleConfig = `
#[inputs.huaweiyunobject.mysql]
endpoint=""

# ## @param - [list of mysql instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded mysql instanceid] - optional
#exclude_instanceids = ['']

# ## @param - custom tags for mysql object - [list of key:value element] - optional
#[inputs.huaweiyunobject.mysql.tags]
# key1 = 'val1'
`
)

type Mysql struct {
	EndPoint string            `toml:"endpoint"`
	Tags     map[string]string `toml:"tags,omitempty"`
	//ProjectID          string            `toml:"project_id"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
}

func (e *Mysql) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`rds.%s.myhuaweicloud.com`, ag.RegionID)
	}
	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		limit := 100
		offset := 0

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

			rdss, err := cli.RdsList(opts)
			if err != nil {
				moduleLogger.Errorf("%v", err)
				return
			}

			moduleLogger.Debugf("%+#v", rdss)
			e.handleResponse(rdss, ag)

			if rdss.TotalCount < offset+limit {
				break
			}

			offset = offset + limit
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Mysql) handleResponse(resp *huaweicloud.ListRdsResponse, ag *objectAgent) {

	moduleLogger.Debugf("mysql TotalCount=%d", resp.TotalCount)

	var objs []map[string]interface{}

	for _, inst := range resp.Instances {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s(%s)`, inst.Name, inst.Id), `huaweiyun_mysql`, inst, inst.Id, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
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
