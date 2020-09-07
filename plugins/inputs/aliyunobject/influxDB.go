package aliyunobject

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	influxDBSampleConfig = `
#[inputs.aliyunobject.influxdb]

# ## @param - custom tags - [list of influxdb instanceid] - optional
#instanceids = []

# ## @param - custom tags - [list of excluded influxdb instanceid] - optional
#exclude_instanceids = []

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.aliyunobject.influxdb.tags]
# key1 = 'val1'
`
)

type InfluxDB struct {
	Tags               map[string]string `toml:"tags,omitempty"`
	InstancesIDs       []string          `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string          `toml:"exclude_instanceids,omitempty"`
}

func (e *InfluxDB) run(ag *objectAgent) {
	var cli *sdk.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = sdk.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		pageNum := 1
		pageSize := 100
		for {
			resp, err := DescribeHiTSDBInstanceList(*cli, pageSize, pageNum)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}
			result := resp.GetHttpContentString()
			if err == nil {
				e.handleResponse(result, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				break
			}

			if gjson.Get(result, "Total").Int() < gjson.Get(result, "PageSize").Int()*gjson.Get(result, "PageNumber").Int() {
				break
			}
			pageNum++
		}
		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func DescribeHiTSDBInstanceList(client sdk.Client, pageSize int, pageNumber int) (response *responses.CommonResponse, err error) {
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https"
	request.Domain = "hitsdb.aliyuncs.com"
	request.Version = "2017-06-01"
	request.ApiName = "DescribeHiTSDBInstanceList"
	request.QueryParams["PageNumber"] = strconv.Itoa(pageNumber)
	request.QueryParams["PageSize"] = strconv.Itoa(pageSize)
	return client.ProcessCommonRequest(request)
}

func (e *InfluxDB) handleResponse(resp string, ag *objectAgent) {
	var objs []map[string]interface{}
	for _, inst := range gjson.Get(resp, "InstanceList").Array() {

		if len(e.ExcludeInstanceIDs) > 0 {
			exclude := false
			for _, v := range e.ExcludeInstanceIDs {
				if v == inst.Get("InstanceId").String() {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}
		if len(e.InstancesIDs) > 0 {
			contain := false
			for _, v := range e.InstancesIDs {
				if v == inst.Get("InstanceId").String() {
					contain = true
					break
				}
			}
			if !contain {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`:          inst.Get("InstanceAlias").String(),
			`GmtCreated`:      inst.Get("GmtCreated").String(),
			`GmtExpire`:       inst.Get("GmtExpire").String(),
			`InstanceStorage`: inst.Get("InstanceStorage").String(),
			`UserId`:          inst.Get("UserId").String(),
		}

		tags := map[string]interface{}{
			`__class`:        `aliyun_influxdb`,
			`provider`:       `aliyun`,
			`InstanceId`:     inst.Get("InstanceId").String(),
			`ZoneId`:         inst.Get("ZoneId").String(),
			`ChargeType`:     inst.Get("ChargeType").String(),
			`InstanceStatus`: inst.Get("InstanceStatus").String(),
			`NetworkType`:    inst.Get("NetworkType").String(),
			`RegionId`:       inst.Get("RegionId").String(),
			`EngineType`:     inst.Get("EngineType").String(),
			`InstanceClass`:  inst.Get("InstanceClass").String(),
		}

		for k, v := range e.Tags {
			tags[k] = v
		}
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
