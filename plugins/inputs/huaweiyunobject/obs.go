package huaweiyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/sdk/huaweicloud/obs"
)

const (
	obsSampleConfig = `
#[inputs.huaweiyunobject.obs]
endpoint=""

# ## @param - custom tags - [list of obs instanceid] - optional
#buckets = []

# ## @param - custom tags - [list of excluded obs instanceid] - optional
#exclude_buckets = []

# ## @param - custom tags for obs object - [list of key:value element] - optional
#[inputs.huaweiyunobject.obs.tags]
# key1 = 'val1'
`
)

type Obs struct {
	EndPoint       string            `toml:"endpoint"`
	Tags           map[string]string `toml:"tags,omitempty"`
	Buckets        []string          `toml:"buckets,omitempty"`
	ExcludeBuckets []string          `toml:"exclude_buckets,omitempty"`
}

func (o *Obs) run(ag *objectAgent) {
	var cli *obs.ObsClient
	var err error
	if o.EndPoint == `` {
		o.EndPoint = fmt.Sprintf(`obs.%s.myhuaweicloud.com`, ag.RegionID)
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

	var objs []map[string]interface{}

	for _, bk := range resp.Buckets {

		if len(o.ExcludeBuckets) > 0 {
			exclude := false
			for _, v := range o.ExcludeBuckets {
				if v == bk.Name {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(o.Buckets) > 0 {
			include := false
			for _, v := range o.Buckets {
				if v == bk.Name {
					include = true
					break
				}
			}

			if !include {
				continue
			}
		}

		obj := map[string]interface{}{
			`__name`: fmt.Sprintf(`%s`, bk.Name),
		}

		owner, err := json.Marshal(resp.Owner)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`Owener`] = owner

		xmlName, err := json.Marshal(bk.XMLName)
		if err != nil {
			moduleLogger.Errorf(`%s, ignore`, err.Error())
			return
		}
		obj[`Bucket.XMLName`] = xmlName

		obj[`Name`] = bk.Name

		obj[`CreationDate`] = bk.CreationDate

		tags := map[string]interface{}{
			`__class`:           `huaweiyun_obs`,
			`provider`:          `huaweiyun`,
			`Location`:          bk.Location,
			`Owner.ID`:          resp.Owner.ID,
			`Owner.DisplayName`: resp.Owner.DisplayName,
		}

		//add obs object custom tags
		for k, v := range o.Tags {
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
