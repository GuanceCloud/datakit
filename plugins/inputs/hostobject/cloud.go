package hostobject

import (
	"encoding/json"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

type cloudInfo struct {
	InstanceName string `json:"instance_name"`
	Region       string `json:"region"`
	InstanceType string `json:"instance_type"`
	Status       string `json:"status"`
}

type CloudAuth struct {
	AK       string
	SK       string
	RegionID string

	STK string
}

func getCloudInfo(auth *CloudAuth, provider string, instanceid string) *cloudInfo {

	if instanceid == "" {
		moduleLogger.Errorf("instanceid cannot be empty")
		return nil
	}

	switch provider {
	case "aliyun":
		return aliyun(auth, instanceid)
	default:
		moduleLogger.Errorf("provider '%s' not supported")
		return nil
	}
}

func aliyun(auth *CloudAuth, instanceid string) *cloudInfo {

	var info cloudInfo

	cli, err := ecs.NewClientWithAccessKey(auth.RegionID, auth.AK, auth.SK)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil
	}

	req := ecs.CreateDescribeInstancesRequest()
	instIDsData, err := json.Marshal([]string{instanceid})
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil
	}
	req.InstanceIds = string(instIDsData)

	if auth.STK != "" {
		req.QueryParams["SecurityToken"] = auth.STK
		req.FormParams["SecurityToken"] = auth.STK
	}

	resp, err := cli.DescribeInstances(req)

	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil
	}

	if len(resp.Instances.Instance) == 0 {
		moduleLogger.Errorf("instance %s not fount", instanceid)
		return nil
	}

	info.InstanceName = resp.Instances.Instance[0].InstanceName
	info.Region = resp.Instances.Instance[0].RegionId
	info.InstanceType = resp.Instances.Instance[0].InstanceType
	info.Status = resp.Instances.Instance[0].Status

	return &info
}
