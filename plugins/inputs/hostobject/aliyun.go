package hostobject

import (
	"io/ioutil"
	"net/http"
)

type aliyun struct {
	baseURL string // http://100.100.100.200/latest/meta-data
}

func metaGet(metaURL string) (res string) {

	res = Unavailable

	req, err := http.NewRequest("GET", metaURL, nil)
	if err != nil {
		l.Warn(err)
		return
	}

	resp, err := cloudCli.Do(req)
	if err != nil {
		l.Warn(err)
		return
	}

	if resp.StatusCode != 200 {
		l.Warnf("request %s: status code %d", metaURL, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	x, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Warnf("read response %s: %s", metaURL, err)
		return
	}
	res = string(x)

	return
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
		"extra_cloud_meta":      x.ExtraCloudMeta(),
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
	// FIXME: 不知何故，阿里云文档提供的 meta-URL 404 了
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

func (x *aliyun) ExtraCloudMeta() string {
	return Unavailable
}

func (x *aliyun) Region() string {
	return metaGet(x.baseURL + "/region-id")
}
