package aliyunobject

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	influxDBSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.influxdb]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_influxdb.p"
	
	# ##(optional) list of influxdb instanceid
	#instanceids = []
	
	# ##(optional) list of excluded influxdb instanceid
	#exclude_instanceids = []
`
	influxDBPipelineConfig = `
json(_, InstanceId)
json(_, RegionId)
json(_, NetworkType)
json(_, InstanceClass)
json(_, ChargeType)
    
`
)

type InfluxDB struct {
	Disable            bool     `toml:"disable"`
	InstancesIDs       []string `toml:"instanceids,omitempty"`
	ExcludeInstanceIDs []string `toml:"exclude_instanceids,omitempty"`
	PipelinePath       string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (e *InfluxDB) disabled() bool {
	return e.Disable
}

func (e *InfluxDB) run(ag *objectAgent) {
	var cli *sdk.Client
	var err error
	p, err := newPipeline(e.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] influxdb new pipeline err:%s", err.Error())
		return
	}
	e.p = p
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
	for _, inst := range gjson.Get(resp, "InstanceList").Array() {
		name := inst.Get("InstanceAlias").String()
		id := inst.Get("InstanceId").String()
		tags := map[string]string{
			"name": fmt.Sprintf("%s_%s", name, id),
		}
		ag.parseObject(inst.String(), "aliyun_influxdb", id, e.p, e.ExcludeInstanceIDs, e.InstancesIDs, tags)
	}
}
