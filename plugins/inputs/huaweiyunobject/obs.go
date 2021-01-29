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

# ##(optional) ist of obs instanceid
#buckets = []

# ##(optional) list of excluded obs instanceid
#exclude_buckets = []

# ##(optional)
# pipeline = ''
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

	pipename := o.PipelinePath
	if pipename == "" {
		pipename = inputName + "_obs.p"
	}
	o.p = getPipeline(pipename)

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
