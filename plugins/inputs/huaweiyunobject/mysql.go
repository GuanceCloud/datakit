package huaweiyunobject

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud"
)

const (
	mysqlSampleConfig = `
#[inputs.huaweiyunobject.mysql]
#endpoint=""

# ## @param - [list of mysql instanceid] - optional
#instanceids = ['']

# ## @param - [list of excluded mysql instanceid] - optional
#exclude_instanceids = ['']

# 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名的脚本，作为其默认 pipeline 配置
# pipeline = "huaweiyun_mysql_object.p"

`
	mysqlPipelineConfig = `

json(_,switch_strategy)
json(_,charge_info)
json(_,region)

`
)

type Mysql struct {
	EndPoint string `toml:"endpoint"`

	//ProjectID          string            `toml:"project_id"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`

	PipelinePath string `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Mysql) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`rds.%s.myhuaweicloud.com`, ag.RegionID)
	}
	cli := huaweicloud.NewHWClient(ag.AccessKeyID, ag.AccessKeySecret, e.EndPoint, ag.ProjectID, moduleLogger)

	if e.PipelinePath != `` {
		p, err := pipeline.NewPipelineByScriptPath(e.PipelinePath)
		if err != nil {
			moduleLogger.Errorf("[error] mysql new pipeline err:%s", err.Error())
			return
		}
		e.p = p
	}

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

	for _, inst := range resp.Instances {

		name := fmt.Sprintf(`%s(%s)`, inst.Name, inst.Id)
		class := `huaweiyun_mysql`
		err := ag.parseObject(inst, name, class, inst.Id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}

}
