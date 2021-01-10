package tencentobject

import (
	"encoding/json"
	"fmt"
	"time"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

const (
	cdbSampleConfig = `
#[inputs.tencentobject.cdb]

# ## @param - [list of instanceid] - optional
#db_instanceids = ['']

# ## @param - [list of excluded instanceid] - optional
#exclude_db_instanceids = ['']

# ## @param - custom tags for this object - [list of key:value element] - optional
#[inputs.tencentobject.cdb.tags]
# key1 = 'val1'
`
)

type Cdb struct {
	Tags                 map[string]string `toml:"tags,omitempty"`
	DBInstancesIDs       []string          `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string          `toml:"exclude_db_instanceids,omitempty"`
}

func (c *Cdb) run(ag *objectAgent) {

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdb.tencentcloudapi.com"
	var client *cdb.Client
	var err error

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		client, err = cdb.NewClient(credential, ag.RegionID, cpf)
		if err == nil {
			break
		}
		moduleLogger.Errorf("%s", err)
		if ag.isTest() {
			ag.testError = err
			return
		}
		datakit.SleepContext(ag.ctx, time.Second*3)
	}

	for {

		select {
		case <-ag.ctx.Done():
			return
		default:
		}

		request := cdb.NewDescribeDBInstancesRequest()

		params := "{}"
		err := request.FromJsonString(params)
		if err != nil {
			moduleLogger.Errorf("%s", err)
		} else {
			response, err := client.DescribeDBInstances(request)
			if err != nil {
				if _, ok := err.(*errors.TencentCloudSDKError); ok {
					moduleLogger.Errorf("api error, %s", err)
				} else {
					moduleLogger.Errorf("%s", err)
				}
				if ag.isTest() {
					ag.testError = err
					return
				}
			} else {
				c.handleResponse(response, ag)
			}
		}

		if ag.isTest() {
			return
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (c *Cdb) handleResponse(resp *cdb.DescribeDBInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("CDB TotalCount=%v", *resp.Response.TotalCount)

	var objs []map[string]interface{}

	for _, db := range resp.Response.Items {
		if obj, err := datakit.CloudObject2Json(fmt.Sprintf(`%s_%s`, *db.InstanceName, *db.InstanceId), `tencent_cdb`, db, *db.InstanceId, c.ExcludeDBInstanceIDs, c.DBInstancesIDs); obj != nil {
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
