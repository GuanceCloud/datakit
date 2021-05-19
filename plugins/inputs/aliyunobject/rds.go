package aliyunobject

import (
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

const (
	rdsSampleConfig = `
# ##(optional)
#[inputs.aliyunobject.rds]
	# ##(optional) ignore this object, default is false
	#disable = false
	# ##(optional) pipeline script path
	
	#pipeline = "aliyun_rds.p"
	# ##(optional) list of rds instanceid
	#db_instanceids = []
	
	# ##(optional) list of excluded rds instanceid
	#exclude_db_instanceids = []
`
	rdsPipelineConfig = `
json(_, DBInstanceId)
json(_, DBInstanceType)
json(_, RegionId)
json(_, Engine)
json(_, DBInstanceClass)
`
)

type Rds struct {
	Disable              bool     `toml:"disable"`
	DBInstancesIDs       []string `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string `toml:"exclude_db_instanceids,omitempty"`
	PipelinePath         string   `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (r *Rds) disabled() bool {
	return r.Disable
}

func (r *Rds) run(ag *objectAgent) {
	var cli *rds.Client
	var err error
	p, err := newPipeline(r.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] rds new pipeline err:%s", err.Error())
		return
	}
	r.p = p
	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		cli, err = rds.NewClientWithAccessKey(ag.RegionID, ag.AccessKeyID, ag.AccessKeySecret)
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
		req := rds.CreateDescribeDBInstancesRequest()

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

			if len(r.DBInstancesIDs) <= 0 && resp.TotalRecordCount < resp.PageNumber*pageSize {
				break
			}

			pageNum++
			if len(r.DBInstancesIDs) <= 0 {
				req.PageNumber = requests.NewInteger(pageNum)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (r *Rds) handleResponse(resp *rds.DescribeDBInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalRecordCount, resp.PageRecordCount, resp.PageNumber)

	for _, db := range resp.Items.DBInstance {
		tags := map[string]string{
			"name": fmt.Sprintf(`%s_%s`, db.DBInstanceDescription, db.DBInstanceId),
		}
		ag.parseObject(db, "aliyun_rds", db.DBInstanceId, r.p, r.ExcludeDBInstanceIDs, r.DBInstancesIDs, tags)
	}
}
