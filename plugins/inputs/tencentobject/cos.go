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

# ## @param - custom tags - [list of buckets] - optional
#buckets = ['']

# ## @param - custom tags - [list of excluded buckets] - optional
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
		} else {
			c.handleResponse(result, ag)
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

		name := bucket.Name
		obj := map[string]interface{}{
			"__name":       fmt.Sprintf(`OSS_%s`, name), // 目前displayName与ID一样
			"CreationDate": bucket.CreationDate,
		}

		tags := map[string]interface{}{
			"__class":          "COS",
			"provider":         "tencent",
			"Location":         bucket.Region,
			"OwnerID":          resp.Owner.ID,
			"OwnerDisplayName": resp.Owner.DisplayName,
		}

		//add oss object custom tags
		for k, v := range c.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

		obj[`__tags`] = tags

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
