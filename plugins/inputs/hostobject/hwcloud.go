package hostobject

import (
	"encoding/json"
	"fmt"
)

type hwcloud struct {
	baseURL string
}

type hwMetaData struct {
	InstanceID   string `json:"uuid"`
	InstanceName string `json:"name"`
	InstanceType string `json:"instance_type"`
	Region       string `json:"region_id"`
	ZoneID       string `json:"availability_zone"`
}

type netWorkData struct {
	NetWork []*netWorks `json:"networks"`
	Links   []*links    `json:"links"`
}

type netWorks struct {
	InstanceNetworkType string `json:"type"`
}

type links struct {
	PrivateIP string `json:"local_ipv4"`
}

const (
	hwMetadataURL  = "http://169.254.169.254/openstack/latest/meta_data.json"
	netWorkDataURL = "http://169.254.169.254/openstack/latest/network_data.json"
)

func (x *hwcloud) Sync() (map[string]interface{}, error) {
	return map[string]interface{}{
		"cloud_provider":        "hwcloud",
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

func (x *hwcloud) getHwMetaData() *hwMetaData {
	resp := metadataGet(hwMetadataURL)
	model := &hwMetaData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func getNetWorkData() *netWorkData {
	resp := metadataGet(netWorkDataURL)
	model := &netWorkData{}
	err := json.Unmarshal(resp, &model)
	if err != nil {
		l.Errorf("marshal json failed: %s", err)
		return nil
	}
	return model
}

func (x *hwcloud) Description() string {
	return Unavailable
}

func (x *hwcloud) InstanceID() string {
	if hwmetadata := x.getHwMetaData(); hwmetadata != nil {
		return hwmetadata.InstanceID
	}
	return Unavailable
}

func (x *hwcloud) InstanceName() string {
	if hwmetadata := x.getHwMetaData(); hwmetadata != nil {
		return hwmetadata.InstanceName
	}
	return Unavailable
}

func (x *hwcloud) InstanceType() string {
	if hwmetadata := x.getHwMetaData(); hwmetadata != nil {
		return hwmetadata.InstanceType
	}
	return Unavailable
}

func (x *hwcloud) InstanceChargeType() string {
	return Unavailable
}

func (x *hwcloud) InstanceNetworkType() string {
	if networkdata := getNetWorkData(); networkdata != nil {
		res := []string{}
		for _, v := range networkdata.NetWork {
			res = append(res, v.InstanceNetworkType)
		}
		InstanceNetworkTypes := ""
		for _, v := range res {
			InstanceNetworkTypes = fmt.Sprintf(InstanceNetworkTypes + v + " ")
		}
		return InstanceNetworkTypes
	}
	return Unavailable
}

func (x *hwcloud) InstanceStatus() string {
	return Unavailable
}

func (x *hwcloud) SecurityGroupID() string {
	return metaGet(x.baseURL + "/security-groups")
}

func (x *hwcloud) PrivateIP() string {
	if networkdata := getNetWorkData(); networkdata != nil {
		res := []string{}
		for _, v := range networkdata.Links {
			res = append(res, v.PrivateIP)
		}
		privateIP := ""
		for _, v := range res {
			privateIP = fmt.Sprintf(privateIP + v + " ")
		}
		return privateIP
	}
	return Unavailable
}

func (x *hwcloud) ZoneID() string {
	if hwmetadata := x.getHwMetaData(); hwmetadata != nil {
		return hwmetadata.ZoneID
	}
	return Unavailable
}

func (x *hwcloud) Region() string {
	if hwmetadata := x.getHwMetaData(); hwmetadata != nil {
		return hwmetadata.Region
	}
	return Unavailable
}
