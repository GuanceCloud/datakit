package huaweiyunobject

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	//"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/elb"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	elb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v2"
	elbmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v2/model"
	elbregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v2/region"
)

const (
	classicType          = `经典型`
	sharedType           = `共享型`
	sharedTypeEnterprise = `共享型_企业项目`
	elbSampleConfig      = `
#[inputs.huaweiyunobject.elb]

# ##(optional) list of Elb instanceid
#instanceids = []

# ##(optional) list of excluded Elb instanceid
#exclude_instanceids = []

# ##(optional)
# pipeline = "huaweiyun_elb_object.p"
`
	elbPipelineConfig = `

json(_,admin_state_up)
	`
)

type Elb struct {
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline

	elbCli *elb.ElbClient
}

func (e *Elb) genClient(ag *objectAgent) {

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	cli := elb.ElbClientBuilder().WithRegion(elbregion.ValueOf(ag.RegionID)).WithCredential(auth).Build()
	e.elbCli = elb.NewElbClient(cli)
}

func (e *Elb) getInstances() (*elbmodel.ListLoadbalancersResponse, error) {

	var err error

	for i := 0; i < 3; i++ {
		req := &elbmodel.ListLoadbalancersRequest{}
		resp, err := e.elbCli.ListLoadbalancers(req)
		if err != nil {
			moduleLogger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}
		return resp, nil
	}

	return nil, err
}

func (e *Elb) run(ag *objectAgent) {

	e.genClient(ag)

	pipename := e.PipelinePath
	if pipename == "" {
		pipename = inputName + "_elb.p"
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

func (e *Elb) handleResponse(resp *elbmodel.ListLoadbalancersResponse, ag *objectAgent) {

	if resp.Loadbalancers == nil {
		return
	}

	moduleLogger.Debugf("Elb TotalCount=%d", len(*resp.Loadbalancers))

	for _, lb := range *resp.Loadbalancers {

		err := ag.parseObject(lb, fmt.Sprintf(`%s(%s)`, lb.Name, lb.Id), `huaweiyun_elb`, lb.Id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}
}
