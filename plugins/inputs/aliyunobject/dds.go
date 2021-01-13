package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

const (
	ddsSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.mongodb]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	#pipeline = "aliyun_mongodb.p"
	
	# ##(optional) list of mongodb instanceid
	#db_instanceids = []
	
	# ##(optional) list of excluded mongodb instanceid
	#exclude_db_instanceids = []
`
	ddsPipelineConfig = `
json(_, DBInstanceId)
json(_, ChargeType)
json(_, RegionId)
json(_, DBInstanceType)
json(_, DBInstanceClass)
`
)

type Dds struct {
	Disable              bool     `toml:"disable"`
	DBInstancesIDs       []string `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string `toml:"exclude_db_instanceids,omitempty"`
	PipelinePath         string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (r *Dds) disabled() bool {
	return r.Disable
}

func (r *Dds) run(ag *objectAgent) {
	var cli *dds.Client
	var err error
	p, err := newPipeline(r.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] mongodb new pipeline err:%s", err.Error())
		return
	}
	r.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = dds.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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

		pageNum := 1
		pageSize := 100
		req := dds.CreateDescribeDBInstancesRequest()
		req.Scheme = "https" //nolint:goconst

		for {
			moduleLogger.Debugf("pageNume %v, pagesize %v", pageNum, pageSize)
			if len(r.DBInstancesIDs) > 0 {
				if pageNum <= len(r.DBInstancesIDs) {
					req.DBInstanceId = r.DBInstancesIDs[pageNum-1]
				} else {
					break
				}
			} else {
				req.PageNumber = requests.NewInteger(pageNum)
				req.PageSize = requests.NewInteger(pageSize)
			}
			resp, err := cli.DescribeDBInstances(req)

			select {
			case <-ag.ctx.Done():
				return
			default:
			}

			if err == nil {
				r.handleResponse(resp, ag)
			} else {
				moduleLogger.Errorf("%s", err)
				if len(r.DBInstancesIDs) > 0 {
					pageNum++
					continue
				}
				break
			}

			if len(r.DBInstancesIDs) == 0 && resp.TotalCount < resp.PageNumber*pageSize {
				break
			}

			pageNum++
			if len(r.DBInstancesIDs) == 0 {
				req.PageNumber = requests.NewInteger(pageNum)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (r *Dds) handleResponse(resp *dds.DescribeDBInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	for _, db := range resp.DBInstances.DBInstance {
		tags := map[string]string{
			"name": fmt.Sprintf("%s_%s", db.DBInstanceDescription, db.DBInstanceId),
		}
		ag.parseObject(db, "aliyun_mongodb", db.DBInstanceId, r.p, r.ExcludeDBInstanceIDs, r.DBInstancesIDs, tags)
	}

}
