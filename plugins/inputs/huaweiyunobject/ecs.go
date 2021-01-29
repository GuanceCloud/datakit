package huaweiyunobject

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	ecsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
)

const (
	ecsSampleConfig = `
#[inputs.huaweiyunobject.ecs]

# ##(optional) list of ecs instanceid
#instanceids = ['']

# ##(optional) list of excluded ecs instanceid
#exclude_instanceids = ['']

# ##(optional)
# pipeline = ''
`
	ecsPipelineConifg = `

json(_,hostId)
json(_,tenant_id)
json(_,host_status)

`
)

type Ecs struct {
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`

	PipelinePath string `toml:"pipeline,omitempty"`

	ecsCli *ecs.EcsClient

	p *pipeline.Pipeline
}

func (e *Ecs) genClient(ag *objectAgent) {

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	cli := ecs.EcsClientBuilder().WithRegion(ecsregion.ValueOf(ag.RegionID)).WithCredential(auth).Build()
	e.ecsCli = ecs.NewEcsClient(cli)
}

func (e *Ecs) getInstances() (*ecsmodel.ListServersDetailsResponse, error) {

	var err error

	for i := 0; i < 3; i++ {
		req := &ecsmodel.ListServersDetailsRequest{}
		resp, err := e.ecsCli.ListServersDetails(req)
		if err != nil {
			moduleLogger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}
		return resp, nil
	}

	return nil, err
}

func (e *Ecs) run(ag *objectAgent) {

	e.genClient(ag)

	pipename := e.PipelinePath
	if pipename == "" {
		pipename = inputName + "_ecs.p"
	}
	e.p = getPipeline(pipename)

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		resp, err := e.getInstances()
		if err == nil {
			e.handleResponse(resp, ag)
		} else {
			moduleLogger.Errorf("%v", err)
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Ecs) handleResponse(resp *ecsmodel.ListServersDetailsResponse, ag *objectAgent) {

	if resp.Servers == nil {
		return
	}

	moduleLogger.Debugf("ECS TotalCount=%d", *resp.Count)

	for _, s := range *resp.Servers {

		name := fmt.Sprintf(`%s(%s)`, s.Name, s.Id)
		class := `huaweiyun_ecs`
		err := ag.parseObject(s, name, class, s.Id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}
}
