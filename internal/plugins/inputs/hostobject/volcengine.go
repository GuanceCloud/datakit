// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import "fmt"

const volcMetaRootURL = "http://100.96.0.96/latest"

type volcEcs struct {
	baseURL string
}

func (x *volcEcs) Sync() (map[string]any, error) {
	return map[string]any{
		"cloud_provider":        VolcEngine,
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

func (x *volcEcs) Description() string {
	return Unavailable
}

func (x *volcEcs) InstanceID() string {
	return metaGet(x.baseURL + "/instance_id")
}

func (x *volcEcs) InstanceName() string {
	return metaGet(x.baseURL + "/hostname")
}

func (x *volcEcs) InstanceType() string {
	return metaGet(x.baseURL + "/instance_type_id")
}

func (x *volcEcs) InstanceChargeType() string {
	return metaGet(x.baseURL + "/payment/charge_type")
}

func (x *volcEcs) InstanceNetworkType() string {
	return Unavailable
}

func (x *volcEcs) InstanceStatus() string {
	return Unavailable
}

func (x *volcEcs) SecurityGroupID() string {
	mac := metaGet(x.baseURL + "/mac")
	return metaGet(x.baseURL + fmt.Sprintf(
		"/network/interfaces/macs/%s/security_group_ids", mac))
}

func (x *volcEcs) PrivateIP() string {
	return metaGet(x.baseURL + "/private_ipv4")
}

func (x *volcEcs) ZoneID() string {
	return metaGet(x.baseURL + "/availability_zone")
}

func (x *volcEcs) Region() string {
	return metaGet(x.baseURL + "/region_id")
}
