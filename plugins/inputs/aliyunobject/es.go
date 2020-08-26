package aliyunobject

import (
	"encoding/json"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/elasticsearch"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	elasticsearchSampleConfig = `
#[inputs.aliyunobject.elasticsearch]

# ## @param - custom tags - [list of elasticsearch instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded elasticsearch instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.aliyunobject.elasticsearch.tags]
# key1 = 'val1'
`
)

type Elasticsearch struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
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
			moduleLogger.Infof("pageNume %v, pagesize %v", page, size)
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

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == inst.InstanceId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`:                       inst.Description,
			`clientNodeConfiguration`:      inst.ClientNodeConfiguration,
			`createdAt`:                    inst.CreatedAt,
			`elasticDataNodeConfiguration`: inst.ElasticDataNodeConfiguration,
			`esVersion`:                    inst.EsVersion,
			`kibanaConfiguration`:          inst.KibanaConfiguration,
			`masterConfiguration`:          inst.MasterConfiguration,
			`networkConfig`:                inst.NetworkConfig,
			`nodeAmount`:                   inst.NodeAmount,
			`nodeSpec`:                     inst.NodeSpec,
		}

		tags := map[string]interface{}{
			`__class`:                `aliyun_elasticsearch`,
			`provider`:               `aliyun`,
			`InstanceId`:             inst.InstanceId,
			`advancedDedicateMaster`: inst.AdvancedDedicateMaster,
			`dedicateMaster`:         inst.DedicateMaster,
			`paymentType`:            inst.PaymentType,
			`ResourceGroupId`:        inst.ResourceGroupId,
			`Status`:                 inst.Status,
		}

		//tags on es instance
		for _, t := range inst.Tags {
			if _, have := tags[t.TagKey]; !have {
				tags[t.TagKey] = t.TagValue
			} else {
				tags[`custom_`+t.TagKey] = t.TagValue
			}
		}

		//add es object custom tags
		for k, v := range e.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj["__tags"] = tags

		objs = append(objs, obj)
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
