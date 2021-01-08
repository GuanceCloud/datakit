package aliyunobject

import (
	"encoding/json"
	"time"

	waf "github.com/aliyun/alibaba-cloud-sdk-go/services/waf-openapi"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	wafSampleConfig = `
#[inputs.aliyunobject.waf]

# ## @param - custom tags for waf object - [list of key:value element] - optional
#[inputs.aliyunobject.waf.tags]
# key1 = 'val1'
`
)

type Waf struct {
	Tags map[string]string `toml:"tags,omitempty"`
}

func (e *Waf) run(ag *objectAgent) {
	var cli *waf.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = waf.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := waf.CreateDescribeInstanceInfoRequest()
		resp, err := cli.DescribeInstanceInfo(req)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			break
		}
		e.handleResponse(resp, ag)
		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (e *Waf) handleResponse(resp *waf.DescribeInstanceInfoResponse, ag *objectAgent) {
	if resp.InstanceInfo.PayType == 0 {
		moduleLogger.Warnf("%s", "waf payType 0")
		return
	}
	var objs []map[string]interface{}

	content := map[string]interface{}{
		"InDebt":           resp.InstanceInfo.InDebt,
		"InstanceId":       resp.InstanceInfo.InstanceId,
		"PayType":          resp.InstanceInfo.PayType,
		"Region":           resp.InstanceInfo.Region,
		"Status":           resp.InstanceInfo.Status,
		"SubscriptionType": resp.InstanceInfo.SubscriptionType,
		"EndDate":          resp.InstanceInfo.EndDate,
		"RemainDay":        resp.InstanceInfo.RemainDay,
		"Trial":            resp.InstanceInfo.Trial,
	}

	jd, err := json.Marshal(content)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}

	obj := map[string]interface{}{
		"name":    resp.InstanceInfo.InstanceId,
		"class":   "aliyun_waf",
		"content": string(jd),
	}

	objs = append(objs, obj)
	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	io.NamedFeed(data, io.Object, inputName)
}
