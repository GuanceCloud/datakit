package aliyunobject

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	rdsSampleConfig = `
#[inputs.aliyunobject.rds]

# ## @param - [list of rds instanceid] - optional
#db_instanceids = []

# ## @param - [list of excluded rds instanceid] - optional
#exclude_db_instanceids = []

# ## @param - custom tags for rds object - [list of key:value element] - optional
#[inputs.aliyunobject.rds.tags]
# key1 = 'val1'
`
)

type Rds struct {
	Tags                 map[string]string `toml:"tags,omitempty"`
	DBInstancesIDs       []string          `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string          `toml:"exclude_db_instanceids,omitempty"`
}

func (r *Rds) run(ag *objectAgent) {
	var cli *rds.Client
	var err error

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
			moduleLogger.Infof("pageNume %v, pagesize %v", pageNum, pageSize)
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

	var objs []map[string]interface{}

	for _, db := range resp.Items.DBInstance {

		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s_%s`, db.DBInstanceDescription, db.DBInstanceId), `aliyun_rds`, db, db.DBInstanceId, r.ExcludeDBInstanceIDs, r.DBInstancesIDs); obj != nil {
			objs = append(objs, obj)
		} else {
			if err != nil {
				moduleLogger.Errorf("%s", err)
			}
		}
	}

	if len(objs) <= 0 {
		return
	}

	data, err := json.Marshal(&objs)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return
	}
	io.NamedFeed(data, io.Object, inputName)
}
