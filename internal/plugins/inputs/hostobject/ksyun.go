// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

const ksyunMetaRootURL = "http://11.255.255.100:8775/latest/meta-data"

type ksyun struct {
	baseURL string
}

func (x *ksyun) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        Ksyun,
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
		"project_id":            x.ProjectID(),
	}, nil
}

func (x *ksyun) Description() string {
	return Unavailable
}

func (x *ksyun) InstanceID() string {
	return metaGet(x.baseURL + "/instance-id")
}

func (x *ksyun) InstanceName() string {
	return metaGet(x.baseURL + "/hostname")
}

func (x *ksyun) InstanceType() string {
	return metaGet(x.baseURL + "/host-type")
}

func (x *ksyun) InstanceChargeType() string {
	return Unavailable
}

func (x *ksyun) InstanceNetworkType() string {
	return Unavailable
}

func (x *ksyun) InstanceStatus() string {
	return Unavailable
}

func (x *ksyun) SecurityGroupID() string {
	return metaGet(x.baseURL + "/securitygroup-ids")
}

func (x *ksyun) PrivateIP() string {
	return metaGet(x.baseURL + "/local-ipv4")
}

func (x *ksyun) ZoneID() string {
	return metaGet(x.baseURL + "/placement/zone")
}

func (x *ksyun) Region() string {
	return metaGet(x.baseURL + "/placement/region")
}

func (x *ksyun) ProjectID() string {
	return Unavailable
}
