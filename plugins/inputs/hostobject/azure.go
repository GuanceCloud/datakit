package hostobject

import "encoding/json"

type azure struct {
	baseURL string
}

type azureMetaData struct {
	Compute *compute `json:"compute"`
}

type compute struct {
	Region       string `json:"location"`
	InstanceName string `json:"name"`
	InstanceID   string `json:"vmId"`
	ZoneID       string `json:"zone"`
	InstanceType string `json:"vmSize"`
}

func (x *azure) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "azure",
		"description":           x.Description(),
		"instance_id":           x.InstanceID(),
		"instance_name":         x.InstanceName(),
		"instance_type":         x.InstanceType(),
		"instance_charge_type":  x.InstanceChargeType(),
		"instance_network_type": x.InstanceNetworkType(),
		"instance_status":       x.InstanceStatus(),
		"security_group_id":     x.SecurityGroupID(),
		"private_ip":            x.PrivateIP(),
		"zone_id":               x.ZoneID(),
		"region":                x.Region(),
	}, nil
}

func (x *azure) getAzureMetaData() *azureMetaData {
	resp := metadataGetByHeader(x.baseURL + "?api-version=2021-02-01") // API要求修改header
	if resp == nil {
		return nil
	}
	model := &azureMetaData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func (x *azure) Description() string {
	return Unavailable
}

func (x *azure) InstanceID() string {
	if azureMetaData := x.getAzureMetaData(); azureMetaData != nil {
		return azureMetaData.Compute.InstanceID
	}
	return Unavailable
}

func (x *azure) InstanceName() string {
	if azureMetaData := x.getAzureMetaData(); azureMetaData != nil {
		return azureMetaData.Compute.InstanceName
	}
	return Unavailable
}

func (x *azure) InstanceType() string {
	if azureMetaData := x.getAzureMetaData(); azureMetaData != nil {
		return azureMetaData.Compute.InstanceType
	}
	return Unavailable
}

func (x *azure) InstanceChargeType() string {
	return Unavailable
}

func (x *azure) InstanceNetworkType() string {
	return Unavailable
}

func (x *azure) InstanceStatus() string {
	return Unavailable
}

func (x *azure) SecurityGroupID() string {
	return Unavailable
}

func (x *azure) PrivateIP() string {
	const privateIPURL = "/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01&format=text"
	res := metadataGetByHeader(x.baseURL + privateIPURL)
	return string(res)
}

func (x *azure) ZoneID() string {
	if azureMetaData := x.getAzureMetaData(); azureMetaData != nil {
		return azureMetaData.Compute.ZoneID
	}
	return Unavailable
}

func (x *azure) Region() string {
	if azureMetaData := x.getAzureMetaData(); azureMetaData != nil {
		return azureMetaData.Compute.Region
	}
	return Unavailable
}
