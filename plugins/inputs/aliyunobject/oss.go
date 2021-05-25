package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	ossSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.oss]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_oss.p"
	# ##(optional) list of oss buckets
	#buckets = []
	
	# ##(optional) list of excluded oss instanceid
	#exclude_buckets = []
`
	ossPipelineConfig = `
json(_, Location)
json(_, StorageClass)
json(_, CreationDate)

`
)

type Oss struct {
	Disable        bool     `toml:"disable"`
	Buckets        []string `toml:"buckets,omitempty"`
	ExcludeBuckets []string `toml:"exclude_buckets,omitempty"`
	PipelinePath   string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (o *Oss) disabled() bool {
	return o.Disable
}

func (o *Oss) run(ag *objectAgent) {
	var cli *oss.Client
	var err error
	p, err := newPipeline(o.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] oss new pipeline err:%s", err.Error())
		return
	}
	o.p = p
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

	for _, bucket := range lsRes.Buckets {
		tags := map[string]string{
			"name": fmt.Sprintf(`OSS_%s`, bucket.Name),
		}
		ag.parseObject(bucket, "aliyun_oss", bucket.Name, o.p, o.Buckets, o.ExcludeBuckets, tags)
	}
}
