package huaweiyunobject

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	rds "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rds/v3"
	rdsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rds/v3/model"
	rdsregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rds/v3/region"
)

const (
	rdsSampleConfig = `
#[inputs.huaweiyunobject.rds]

# ##(optional) list of instanceid
#instanceids = ['']

# ##(optional) list of excluded instanceid
#exclude_instanceids = ['']

# ##(optional)
# pipeline = ''

`
	rdsPipelineConfig = `

json(_,switch_strategy)
json(_,charge_info)
json(_,region)

`
)

type Rds struct {
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline

	rdsCli *rds.RdsClient
}

func (e *Rds) genClient(ag *objectAgent) {

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	cli := rds.RdsClientBuilder().WithRegion(rdsregion.ValueOf(ag.RegionID)).WithCredential(auth).Build()
	e.rdsCli = rds.NewRdsClient(cli)
}

func (e *Rds) getInstances() (*rdsmodel.ListInstancesResponse, error) {

	var err error

	for i := 0; i < 3; i++ {
		req := &rdsmodel.ListInstancesRequest{}
		resp, err := e.rdsCli.ListInstances(req)
		if err != nil {
			moduleLogger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}
		return resp, nil
	}

	return nil, err
}

func (e *Rds) run(ag *objectAgent) {

	e.genClient(ag)

	pipename := e.PipelinePath
	if pipename == "" {
		pipename = inputName + "_rds.p"
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

func (e *Rds) handleResponse(resp *rdsmodel.ListInstancesResponse, ag *objectAgent) {

	if resp.Instances == nil {
		return
	}

	moduleLogger.Debugf("rds TotalCount=%d", *resp.TotalCount)

	for _, inst := range *resp.Instances {

		name := fmt.Sprintf(`%s(%s)`, inst.Name, inst.Id)
		class := `huaweiyun_rds`
		err := ag.parseObject(inst, name, class, inst.Id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}

}
