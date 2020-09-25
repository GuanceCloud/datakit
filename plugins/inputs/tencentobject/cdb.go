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

# ## @param - custom tags - [list of instanceid] - optional
#db_instanceids = ['']

# ## @param - custom tags - [list of excluded instanceid] - optional
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
			} else {
				c.handleResponse(response, ag)
			}
		}

		datakit.SleepContext(ag.ctx, ag.Interval.Duration)
	}
}

func (c *Cdb) handleResponse(resp *cdb.DescribeDBInstancesResponse, ag *objectAgent) {

	moduleLogger.Debugf("CDB TotalCount=%v", *resp.Response.TotalCount)

	var objs []map[string]interface{}

	for _, db := range resp.Response.Items {
		//moduleLogger.Debugf("dbinstanceInfo %+#v", db)

		if len(c.ExcludeDBInstanceIDs) > 0 {
			exclude := false
			for _, v := range c.ExcludeDBInstanceIDs {
				if v == *db.InstanceId {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}

		if len(c.DBInstancesIDs) > 0 {
			exclude := true
			for _, v := range c.DBInstancesIDs {
				if v == *db.InstanceId {
					exclude = false
					break
				}
			}
			if exclude {
				continue
			}
		}

		tags := map[string]interface{}{
			"__class":      "CDB",
			"provider":     "tencent",
			"InstanceId":   *db.InstanceId,
			"InstanceType": *db.InstanceType,
			"Region":       *db.Region,
			"Status":       *db.Status,
		}

		if db.ProjectId != nil {
			tags["ProjectId"] = *db.ProjectId
		}
		if db.EngineVersion != nil {
			tags["EngineVersion"] = *db.EngineVersion
		}
		if db.DeviceType != nil {
			tags["DeviceType"] = *db.DeviceType
		}
		if db.ZoneName != nil {
			tags["ZoneName"] = *db.ZoneName
		}
		if db.DeviceClass != nil {
			tags["DeviceClass"] = *db.DeviceClass
		}
		if db.ProtectMode != nil {
			tags["ProtectMode"] = *db.ProtectMode
		}
		if db.AutoRenew != nil {
			tags["AutoRenew"] = *db.AutoRenew
		}
		if db.DeployMode != nil {
			tags["DeployMode"] = *db.DeployMode
		}
		if db.CdbError != nil {
			tags["CdbError"] = *db.CdbError
		}
		if db.PayType != nil {
			tags["PayType"] = *db.PayType
		}
		if db.WanDomain != nil {
			tags["WanDomain"] = *db.WanDomain
		}
		if db.WanPort != nil {
			tags["WanPort"] = *db.WanPort
		}
		if db.WanStatus != nil {
			tags["WanStatus"] = *db.WanStatus
		}

		obj := map[string]interface{}{
			"__name":     fmt.Sprintf(`%s_%s`, *db.InstanceName, *db.InstanceId),
			"__tags":     tags,
			"Zone":       db.Zone,
			"Memory":     *db.Memory,
			"Volume":     *db.Volume,
			"CreateTime": *db.CreateTime,
			"Vport":      *db.Vport,
		}
		if db.DeadlineTime != nil {
			obj["DeadlineTime"] = *db.DeadlineTime
		}
		if db.Cpu != nil {
			obj["Cpu"] = *db.Cpu
		}
		if db.Vip != nil {
			obj["Vip"] = *db.Vip
		}
		if db.Qps != nil {
			obj["Qps"] = *db.Qps
		}
		if db.TaskStatus != nil {
			obj["TaskStatus"] = *db.TaskStatus
		}

		//add rds object custom tags
		for k, v := range c.Tags {
			tags[k] = v
		}

		//add global tags
		for k, v := range ag.Tags {
			if _, have := tags[k]; !have {
				tags[k] = v
			}
		}

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
