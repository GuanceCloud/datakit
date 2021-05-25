package hostobject

type aliyun struct {
	baseURL string // http://100.100.100.200/latest/meta-data
}

func (x *aliyun) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "aliyun",
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

func (x *aliyun) Description() string {
	return Unavailable
}

func (x *aliyun) InstanceID() string {
	return metaGet(x.baseURL + "/instance-id")
}

func (x *aliyun) InstanceName() string {
	return Unavailable
}

func (x *aliyun) InstanceType() string {
	return metaGet(x.baseURL + "/instance/instance-type")
}

func (x *aliyun) InstanceChargeType() string {
	// XXX: 阿里云确实提供了 /image/market-place/charge-type 这个接口，
	// 但这并非通常意义上的 charge-type 信息。故此处不用该接口。
	return Unavailable
}

func (x *aliyun) InstanceNetworkType() string {
	return metaGet(x.baseURL + "/network-type")
}

func (x *aliyun) InstanceStatus() string {
	return Unavailable
}

func (x *aliyun) SecurityGroupID() string {
	return Unavailable
}

func (x *aliyun) PrivateIP() string {
	return metaGet(x.baseURL + "/private-ipv4")
}

func (x *aliyun) ZoneID() string {
	return metaGet(x.baseURL + "/zone-id")
}

func (x *aliyun) Region() string {
	return metaGet(x.baseURL + "/region-id")
}
