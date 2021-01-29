package huaweiyunobject

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	vpc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	vpcregion "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/region"
)

const (
	vpcSampleConfig = `
#[inputs.huaweiyunobject.vpc]

# ##(optional) list of instanceid
#instanceids = ['']

# ##(optional) list of excluded instanceid
#exclude_instanceids = ['']

# ##(optional)
# pipeline = ''
`
	vpcPipelineConifg = `

json(_,cidr)
json(_,status)

`
)

type Vpc struct {
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`

	PipelinePath string `toml:"pipeline,omitempty"`

	vpcCli *vpc.VpcClient

	p *pipeline.Pipeline
}

func (e *Vpc) genClient(ag *objectAgent) {

	auth := basic.NewCredentialsBuilder().
		WithAk(ag.AccessKeyID).
		WithSk(ag.AccessKeySecret).
		WithProjectId(ag.ProjectID).
		Build()

	cli := vpc.VpcClientBuilder().WithRegion(vpcregion.ValueOf(ag.RegionID)).WithCredential(auth).Build()
	e.vpcCli = vpc.NewVpcClient(cli)
}

func (e *Vpc) getInstances() (*vpcmodel.ListVpcsResponse, error) {

	var err error

	for i := 0; i < 3; i++ {
		req := &vpcmodel.ListVpcsRequest{}
		resp, err := e.vpcCli.ListVpcs(req)
		if err != nil {
			moduleLogger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}
		return resp, nil
	}

	return nil, err
}

func (e *Vpc) run(ag *objectAgent) {

	e.genClient(ag)

	pipename := e.PipelinePath
	if pipename == "" {
		pipename = inputName + "_vpc.p"
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

func (e *Vpc) handleResponse(resp *vpcmodel.ListVpcsResponse, ag *objectAgent) {

	if resp.Vpcs == nil {
		return
	}

	moduleLogger.Debugf("VPC TotalCount=%d", len(*resp.Vpcs))

	for _, s := range *resp.Vpcs {

		name := fmt.Sprintf(`%s(%s)`, s.Name, s.Id)
		class := `huaweiyun_vpc`
		err := ag.parseObject(s, name, class, s.Id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}
	}
}
