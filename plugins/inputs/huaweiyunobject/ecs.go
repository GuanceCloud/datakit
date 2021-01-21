package huaweiyunobject

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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

# 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名的脚本，作为其默认 pipeline 配置
# pipeline = "huaweiyun_ecs_object.p"
`
	ecsPipelineConifg = `

json(_,hostId)
json(_,tenant_id)
json(_,host_status)

`
)

type Ecs struct {

	//	ProjectID          string            `toml:"project_id"`
	EndPoint           string   `toml:"endpoint"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`

	PipelinePath string `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Ecs) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`ecs.%s.myhuaweicloud.com`, ag.RegionID)
	}

	if e.PipelinePath != `` {
		p, err := pipeline.NewPipelineByScriptPath(e.PipelinePath)
		if err != nil {
			moduleLogger.Errorf("[error] ecs new pipeline err:%s", err.Error())
			return
		}
		e.p = p
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

	for _, s := range resp.Servers {

		name := fmt.Sprintf(`%s(%s)`, s.InstanceName, s.ID)
		class := `huaweiyun_ecs`
		err := ag.parseObject(s, name, class, s.ID, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)

		}
	}
}
