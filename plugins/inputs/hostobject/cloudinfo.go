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
	var info cloudInfo

	switch provider {
	case "aliyun":
		cli, err := ecs.NewClientWithAccessKey(auth.RegionID, auth.AK, auth.SK)
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return nil
		}
		_ = cli

		req := ecs.CreateDescribeInstancesRequest()
		instIDsData, err := json.Marshal([]string{instanceid})
		if err != nil {
			moduleLogger.Errorf("%s", err)
			return nil
		}
		req.InstanceIds = string(instIDsData)

		resp, err := cli.DescribeInstances(req)

		if err != nil {
			moduleLogger.Errorf("%s", err)
			return nil
		}

		_ = resp

	default:
		return nil
	}

	return &info
}
