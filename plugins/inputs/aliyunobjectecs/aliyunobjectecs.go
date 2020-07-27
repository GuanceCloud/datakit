package aliyunobjectecs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName    = `aliyunecsobject`
	moduleLogger *logger.Logger
)

func (_ *objectAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *objectAgent) Catalog() string {
	return `aliyun`
}

func (ag *objectAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	ag.ctx, ag.cancelFun = context.WithCancel(context.Background())

	go func() {
		<-datakit.Exit.Wait()
		ag.cancelFun()
	}()

	for _, inst := range ag.ECSObject {

		if inst.Interval.Duration == 0 {
			inst.Interval.Duration = time.Minute * 5
		}

		ag.wg.Add(1)
		go func(inst *AliyunCfg) {
			defer ag.wg.Done()
			ag.runInstance(inst)
		}(inst)
	}

	ag.wg.Wait()

	moduleLogger.Debugf("done")
}

func (ag *objectAgent) runInstance(cfg *AliyunCfg) {

	var cli *ecs.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = ecs.NewClientWithAccessKey(cfg.RegionID, cfg.AccessKeyID, cfg.AccessKeySecret)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		internal.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		pageNum := 1
		req := ecs.CreateDescribeInstancesRequest()
		req.PageNumber = requests.NewInteger(pageNum)
		req.PageSize = requests.NewInteger(100)

		for {
			resp, err := cli.DescribeInstances(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				ag.handleResponse(resp, cfg)
			} else {
				moduleLogger.Errorf("%s", err)
				break
			}

			if resp.TotalCount < resp.PageNumber*resp.PageSize {
				break
			}
			pageNum++
			req.PageNumber = requests.NewInteger(pageNum)
		}

		internal.SleepContext(ag.ctx, cfg.Interval.Duration)
	}

}

type objectSt struct {
	Name        string                 `json:"__name"`
	Description string                 `json:"__description"`
	Tags        map[string]interface{} `json:"__tags"`
}

func (ag *objectAgent) handleResponse(resp *ecs.DescribeInstancesResponse, cfg *AliyunCfg) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	var objs []*objectSt

	for _, inst := range resp.Instances.Instance {

		obj := &objectSt{
			Name:        fmt.Sprintf(`ECS_%s`, inst.InstanceId),
			Description: ``,
		}

		tags := map[string]interface{}{
			"__class": "ECS",
		}

		data, err := json.Marshal(&inst)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			continue
		}
		err = json.Unmarshal(data, &tags)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			continue
		}

		for k, v := range cfg.Tags {
			tags[k] = v
		}

		obj.Tags = tags

		objs = append(objs, obj)
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		io.NamedFeed(data, io.Object, inputName)
	} else {
		moduleLogger.Errorf("%s", err)
	}
}

func newAgent() *objectAgent {
	ag := &objectAgent{}
	return ag
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return newAgent()
	})
}
