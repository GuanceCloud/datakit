package hostobject

type aws struct {
	baseURL string // http://100.100.100.200/latest/meta-data
}

func (x *aws) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "aws",
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

func (x *aws) Description() string {
	return Unavailable
}

func (x *aws) InstanceID() string {
	return metaGet(x.baseURL + "/instance-id")
}

func (x *aws) InstanceName() string {
	return Unavailable
}

func (x *aws) InstanceType() string {
	return metaGet(x.baseURL + "/instance-type")
}

func (x *aws) InstanceChargeType() string {
	return Unavailable
}

func (x *aws) InstanceNetworkType() string {
	return Unavailable
}

func (x *aws) InstanceStatus() string {
	return Unavailable
}

func (x *aws) SecurityGroupID() string {
	return metaGet(x.baseURL + "/security-groups")
}

func (x *aws) PrivateIP() string {
	return metaGet(x.baseURL + "/local-ipv4")
}

func (x *aws) ZoneID() string {
	return metaGet(x.baseURL + "/placement/availability-zone-id")
}

func (x *aws) Region() string {
	// 这个在 AWS 文档是没有的：
	//  https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html
	return metaGet(x.baseURL + "/placement/region")
}
