package huaweiyunobject

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/obs"
)

const (
	obsSampleConfig = `
#[inputs.huaweiyunobject.obs]
#endpoint=""

# ## @param - [list of obs instanceid] - optional
#buckets = []

# ## @param - [list of excluded obs instanceid] - optional
#exclude_buckets = []

# 如果 pipeline 未配置，则在 pipeline 目录下寻找跟 source 同名的脚本，作为其默认 pipeline 配置
# pipeline = "huaweiyun_obs_object.p"
`
	obsPipelineConifg = `

json(_,Location)

`
)

type Obs struct {
	EndPoint       string   `toml:"endpoint"`
	Buckets        []string `toml:"buckets,omitempty"`
	ExcludeBuckets []string `toml:"exclude_buckets,omitempty"`

	PipelinePath string `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (o *Obs) run(ag *objectAgent) {
	var cli *obs.ObsClient
	var err error
	if o.EndPoint == `` {
		o.EndPoint = fmt.Sprintf(`obs.%s.myhuaweicloud.com`, ag.RegionID)
	}

	if o.PipelinePath != `` {
		p, err := pipeline.NewPipelineByScriptPath(o.PipelinePath)
		if err != nil {
			moduleLogger.Errorf("[error] obs new pipeline err:%s", err.Error())
			return
		}
		o.p = p
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = obs.New(ag.AccessKeyID, ag.AccessKeySecret, o.EndPoint)

		if err == nil {
			break
		}
		moduleLogger.Errorf("%v", err)

		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		input := &obs.ListBucketsInput{}
		input.QueryLocation = true
		buckets, err := cli.ListBuckets(input)
		if err != nil {
			moduleLogger.Errorf("%v", err)

			return
		}
		o.handleResponse(buckets, ag)

		moduleLogger.Debugf("%+#v", buckets)

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (o *Obs) handleResponse(resp *obs.ListBucketsOutput, ag *objectAgent) {

	moduleLogger.Debugf("obs Count=%d", len(resp.Buckets))

	for _, bk := range resp.Buckets {

		name := fmt.Sprintf(`%s`, bk.Name)
		class := `huaweiyun_obs`
		err := ag.parseObject(bk, name, class, bk.Name, o.p, o.ExcludeBuckets, o.Buckets)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		}

	}

}
