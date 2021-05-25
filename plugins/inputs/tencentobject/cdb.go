package tencentobject

import (
	"fmt"
	"time"

	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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

	cdbPipelineConfig = `
json(_, InstanceId);
json(_, InstanceName);
json(_, Region);
json(_, InstanceType);
json(_, DeviceType);
json(_, Status);
json(_, Volume);
json(_, Memory);
`
)

type Cdb struct {
	Tags                 map[string]string `toml:"tags,omitempty"`
	DBInstancesIDs       []string          `toml:"db_instanceids,omitempty"`
	ExcludeDBInstanceIDs []string          `toml:"exclude_db_instanceids,omitempty"`
	PipelinePath         string            `toml:"pipeline,omitempty"`

	p *pipeline.Pipeline
}

func (c *Cdb) run(ag *objectAgent) {

	credential := ag.getCredential()
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "cdb.tencentcloudapi.com"
	var client *cdb.Client
	var err error

	c.p, err = newPipeline(c.PipelinePath)
	if err != nil {
		moduleLogger.Errorf("[error] cdb new pipeline err:%s", err.Error())
		return
	}

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

	for _, db := range resp.Response.Items {

		tags := map[string]string{
			"name": fmt.Sprintf(`%s_%s`, *db.InstanceName, *db.InstanceId),
		}
		ag.parseObject(db, "tencent_cdb", *db.InstanceId, c.p, c.ExcludeDBInstanceIDs, c.DBInstancesIDs, tags)
	}

}
