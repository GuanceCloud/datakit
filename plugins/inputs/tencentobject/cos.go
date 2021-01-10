package tencentobject

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cos "github.com/tencentyun/cos-go-sdk-v5"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
)

type Cos struct {
	Tags           map[string]string `toml:"tags,omitempty"`
	Buckets        []string          `toml:"buckets,omitempty"`
	ExcludeBuckets []string          `toml:"exclude_buckets,omitempty"`
}

func (c *Cos) run(ag *objectAgent) {

	client := cos.NewClient(nil, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ag.AccessKeyID,
			SecretKey: ag.AccessKeySecret,
		},
	})

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

	var objs []map[string]interface{}
	for _, bucket := range resp.Buckets {

		if len(c.ExcludeBuckets) > 0 {
			exclude := false
			for _, v := range c.ExcludeBuckets {
				if v == bucket.Name {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(c.Buckets) > 0 {
			include := false
			for _, v := range c.Buckets {
				if v == bucket.Name {
					include = true
					break
				}
			}

			if !include {
				continue
			}
		}

		content := map[string]interface{}{
			"Location":         bucket.Region,
			"OwnerID":          resp.Owner.ID,
			"OwnerDisplayName": resp.Owner.DisplayName,
			"CreationDate":     bucket.CreationDate,
		}

		jd, err := json.Marshal(content)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			continue
		}

		name := bucket.Name
		obj := map[string]interface{}{
			"name":    fmt.Sprintf(`OSS_%s`, name), // 目前displayName与ID一样
			"class":   "tencent_cos",
			"content": string(jd),
		}

		objs = append(objs, obj)
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err == nil {
		if ag.isTest() {
			ag.testResult.Result = append(ag.testResult.Result, data...)
		} else {
			io.NamedFeed(data, io.Object, inputName)
		}
	} else {
			moduleLogger.Errorf("%s", err)
			return
		}
	}
