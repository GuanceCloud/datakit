package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/elasticsearch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	elasticsearchSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.elasticsearch]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_elasticsearch.p"
	# ##(optional) list of elasticsearch instanceid
	#instanceids = []
	# ##(optional) list of excluded elasticsearch instanceid
	#exclude_instanceids = []
`
	elasticsearchPipelineConifg = `
json(_, instanceId)
json(_, paymentType)
json(_, status)
json(_, dedicateMaster)
json(_, resourceGroupId)
`
)

type Elasticsearch struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *Elasticsearch) disabled() bool {
	return e.Disable
}

func (e *Elasticsearch) run(ag *objectAgent) {
	var cli *elasticsearch.Client
	var err error
	p, err := newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] elasticsearch new pipeline err:%s", err.Error())
		return
	}
	e.p = p
	for {
		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = elasticsearch.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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

		page := 1
		size := 100
		req := elasticsearch.CreateListInstanceRequest()
		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", page, size)
			if len(e.InstancesIDs) > 0 {
				if page <= len(e.InstancesIDs) {
					req.InstanceId = e.InstancesIDs[page-1]
				} else {
					break
				}
			} else {
				req.Page = requests.NewInteger(page)
				req.Size = requests.NewInteger(size)
			}
			resp, err := cli.ListInstance(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				e.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if len(e.InstancesIDs) > 0 {
					page++
					continue
				}
				break
			}

			if len(e.InstancesIDs) <= 0 && resp.Headers.XTotalCount < page*size {
				break
			}

			page++
			if len(e.InstancesIDs) <= 0 {
				req.Page = requests.NewInteger(page)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Elasticsearch) handleResponse(resp *elasticsearch.ListInstanceResponse, ag *objectAgent) {
	for _, inst := range resp.Result {
		tags := map[string]string{
			"name": fmt.Sprintf("%s_%s", inst.Description, inst.InstanceId),
		}
		ag.parseObject(inst, "aliyun_elasticsearch", inst.InstanceId, e.p, e.ExcludeInstanceIDs, e.InstancesIDs, tags)
	}

}
