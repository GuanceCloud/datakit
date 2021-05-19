package tencentobject

import (
	"context"
	"fmt"
	"net/http"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	cosSampleConfig = `
#[inputs.tencentobject.cos]

# ## @param - [list of buckets] - optional
#buckets = ['']

# ## @param - [list of excluded buckets] - optional
#exclude_buckets = ['']

# ## @param - custom tags - [list of key:value element] - optional
#[inputs.tencentobject.cos.tags]
# key1 = 'val1'
`

	cosPipelineConfig = `
json(_, Name);
json(_, Region);
json(_, CreationDate);
`
)

type Cos struct {
	Tags           map[string]string `toml:"tags,omitempty"`
	Buckets        []string          `toml:"buckets,omitempty"`
	ExcludeBuckets []string          `toml:"exclude_buckets,omitempty"`
	PipelinePath   string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (c *Cos) run(ag *objectAgent) {

	client := cos.NewClient(nil, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ag.AccessKeyID,
			SecretKey: ag.AccessKeySecret,
		},
	})

	var err error
	c.p, err = newPipeline(c.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] cos new pipeline err:%s", err.Error())
		return
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		result, _, err := client.Service.Get(context.Background())
		if err != nil {
			moduleLogger.Errorf("%s", err)
			if ag.isTest() {
				ag.testError = err
				return
			}
		} else {
			c.handleResponse(result, ag)
		}

		if ag.isTest() {
			return
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (c *Cos) handleResponse(resp *cos.ServiceGetResult, ag *objectAgent) {

	for _, bucket := range resp.Buckets {

		tags := map[string]string{
			"name": fmt.Sprintf(`OSS_%s`, bucket.Name),
		}
		ag.parseObject(bucket, "tencent_cos", bucket.Name, c.p, c.ExcludeBuckets, c.Buckets, tags)
	}
}
