// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"strings"
)

type gcp struct {
	baseURL string // http://169.254.169.254/computeMetadata/v1
}

func (x *gcp) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "gcp",
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

func (x *gcp) InstanceID() string {
	return metadataGetByHeaderGCP(x.baseURL + "/instance/id")
}

func (x *gcp) InstanceName() string {
	return metadataGetByHeaderGCP(x.baseURL + "/instance/name")
}

func (x *gcp) InstanceType() string {
	// projects/123456789/machineTypes/n1-standard-1
	machineType := metadataGetByHeaderGCP(x.baseURL + "/instance/machine-type")
	if machineType != Unavailable {
		parts := strings.Split(machineType, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return Unavailable
}

func (x *gcp) ZoneID() string {
	// projects/123456789/zones/us-central1-a
	zone := metadataGetByHeaderGCP(x.baseURL + "/instance/zone")
	if zone != Unavailable {
		parts := strings.Split(zone, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return Unavailable
}

func (x *gcp) Region() string {
	// us-central1-a -> us-central1
	if zone := x.ZoneID(); zone != Unavailable {
		parts := strings.Split(zone, "-")
		if len(parts) >= 2 {
			return strings.Join(parts[:len(parts)-1], "-")
		}
	}
	return Unavailable
}

func (x *gcp) PrivateIP() string {
	return metadataGetByHeaderGCP(x.baseURL + "/instance/network-interfaces/0/ip")
}

func (x *gcp) ProjectID() string {
	return metadataGetByHeaderGCP(x.baseURL + "/project/project-id")
}

func (x *gcp) Description() string {
	return Unavailable
}

func (x *gcp) InstanceChargeType() string {
	return Unavailable
}

func (x *gcp) InstanceNetworkType() string {
	return Unavailable
}

func (x *gcp) InstanceStatus() string {
	return Unavailable
}

func (x *gcp) SecurityGroupID() string {
	return Unavailable
}
