package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	ossSampleConfig = `
#[inputs.aliyunobject.oss]

# ## @param - custom tags - [list of oss buckets] - optional
#buckets = []

# ## @param - custom tags - [list of excluded oss instanceid] - optional

#exclude_buckets = []

# ## @param - custom tags for ecs object - [list of key:value element] - optional
#[inputs.aliyunobject.oss.tags]
# key1 = 'val1'

`
)

type Oss struct {
	Tags           map[string]string `toml:"tags,omitempty"`
	Buckets        []string          `toml:"buckets,omitempty"`
	ExcludeBuckets []string          `toml:"exclude_buckets,omitempty"`
}

func (o *Oss) run(ag *objectAgent) {
	var cli *oss.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = oss.New(`oss-`+ag.RegionID+`.aliyuncs.com`, ag.AccessKeyID, ag.AccessKeySecret)
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

		marker := ``
		pageSize := 500
		pageNum := 1

		for {
			lsRes, err := cli.ListBuckets(oss.Marker(marker), oss.MaxKeys(pageSize))

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				moduleLogger.Debugf("pageNum=%v totalCount=%v, marker=%v count=%v", pageNum, (pageNum-1)*pageSize+len(lsRes.Buckets), marker, len(lsRes.Buckets))

				o.handleResponse(&lsRes, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				break
			}

			if len(lsRes.Buckets) < pageSize {
				break
			}
			pageNum++
			marker = lsRes.NextMarker
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (o *Oss) handleResponse(lsRes *oss.ListBucketsResult, ag *objectAgent) {

	var objs []map[string]interface{}
	for _, bucket := range lsRes.Buckets {

		if len(o.ExcludeBuckets) > 0 {
			exclude := false
			for _, v := range o.ExcludeBuckets {
				if v == bucket.Name {
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
			"XMLName":      bucket.XMLName,
			"Ower.XMLName": lsRes.Owner.XMLName,
		}

		tags := map[string]interface{}{
			"__class":           "aliyun_oss",
			"provider":          "aliyun",
			"Location":          bucket.Location,
			"Owner.ID":          lsRes.Owner.ID,
			"Owner.DisplayName": lsRes.Owner.DisplayName,
			"StorageClass":      bucket.StorageClass,
		}

		//add oss object custom tags
		for k, v := range o.Tags {
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
