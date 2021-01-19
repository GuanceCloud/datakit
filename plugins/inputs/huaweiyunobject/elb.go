package huaweiyunobject

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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
#type=""

## 地区和终端节点 https://developer.huaweicloud.com/endpoint?ELB
#endpoint=""

# ## @param - [list of Elb instanceid] - optional
#instanceids = []

# ## @param - [list of excluded Elb instanceid] - optional
#exclude_instanceids = []

# 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名的脚本，作为其默认 pipeline 配置
# pipeline = "huaweiyun_elb_object.p"
`
	elbPipelineConfig = `

json(_,type)
json(_,admin_state_up)
	`
)

type Elb struct {
	Type         string `toml:"type"`
	EndPoint     string `toml:"endpoint"`
	PipelinePath string `toml:"pipeline,omitempty"`
	//	ProjectID          string            `toml:"project_id"`

	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`

	p *pipeline.Pipeline
}

func (e *Elb) run(ag *objectAgent) {

	if e.EndPoint == `` {
		e.EndPoint = fmt.Sprintf(`elb.%s.myhuaweicloud.com`, ag.RegionID)
	}

	if e.PipelinePath != `` {
		p, err := pipeline.NewPipelineByScriptPath(e.PipelinePath)
		if err != nil {
			moduleLogger.Errorf("[error]  elb new pipeline err:%s", err.Error())
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

	for _, lb := range resp.Loadbalancers {

		err := ag.parseObject(lb, fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID), `huaweiyun_elb`, lb.ID, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}

	}
}

func (e *Elb) handResponseV2(lbs []elb.LoadbalancerV2, ag *objectAgent) {
	moduleLogger.Debugf("Elb TotalCount=%d", len(lbs))

	for _, lb := range lbs {

		name := fmt.Sprintf(`%s(%s)`, lb.Name, lb.ID)
		class := `huaweiyun_elb`
		err := ag.parseObject(lb, name, class, lb.ID, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}

}
