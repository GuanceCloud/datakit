package hostobject

type tencent struct {
	baseURL string // http://metadata.tencentyun.com/latest
}

func (x *tencent) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "tencent",
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

func (x *tencent) Description() string {
	return Unavailable
}

func (x *tencent) InstanceID() string {
	return metaGet(x.baseURL + "/instance-id")
}

func (x *tencent) InstanceName() string {
	return metaGet(x.baseURL + "/instance-name")
}

func (x *tencent) InstanceType() string {
	return metaGet(x.baseURL + "/instance/instance-type")
}

func (x *tencent) InstanceChargeType() string {
	return metaGet(x.baseURL + "/payment/charge-type")
}

func (x *tencent) InstanceNetworkType() string {
	return Unavailable
}

func (x *tencent) InstanceStatus() string {
	return Unavailable
}

func (x *tencent) SecurityGroupID() string {
	return metaGet(x.baseURL + "/instance/security-group")
}

func (x *tencent) PrivateIP() string {
	return metaGet(x.baseURL + "/local-ipv4")
}

func (x *tencent) ZoneID() string {
	return metaGet(x.baseURL + "/placement/zone")
}

func (x *tencent) Region() string {
	return metaGet(x.baseURL + "/placement/region")
}
