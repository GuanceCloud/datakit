package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/elasticsearch"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	elasticsearchSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.elasticsearch]
    # ##(optional) ignore this object, default is false
    #disable = false

    # ##(optional) list of elasticsearch instanceid
    #instanceids = []

    # ##(optional) list of excluded elasticsearch instanceid
    #exclude_instanceids = []
`
)

type Elasticsearch struct {
	Disable            bool              `toml:"disable"`
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *Elasticsearch) disabled() bool {
	return e.Disable
}

func (e *Elasticsearch) run(ag *objectAgent) {
	var cli *elasticsearch.Client
	var err error

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
	var objs []map[string]interface{}

	for _, inst := range resp.Result {

		if obj, err := datakit.CloudObject2Json(inst.Description, `aliyun_elasticsearch`, inst, inst.InstanceId, e.ExcludeInstanceIDs, e.InstancesIDs); obj != nil {
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
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	if ag.isDebug() {
		fmt.Printf("%s\n", string(data))
	} else {
		io.NamedFeed(data, io.Object, inputName)
	}
}
