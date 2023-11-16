// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	Unavailable = "-"
	Aliyun      = "aliyun"
	AWS         = "aws"
	Tencent     = "tencent"
	Azure       = "azure"
	Hwcloud     = "hwcloud"
)

var cloudCli = &http.Client{Timeout: 100 * time.Millisecond}

//nolint:deadcode,unused
type synchronizer interface {
	Sync() (map[string]interface{}, error)
	Description() string
	InstanceID() string
	InstanceName() string
	InstanceType() string
	InstanceChargeType() string
	InstanceNetworkType() string
	InstanceStatus() string
	SecurityGroupID() string
	PrivateIP() string
	Region() string
	ZoneID() string
}

func (*Input) SyncCloudInfo(provider string) (map[string]interface{}, error) {
	defer cloudCli.CloseIdleConnections()

	switch provider {
	case Aliyun:
		p := &aliyun{baseURL: "http://100.100.100.200/latest/meta-data"}
		return p.Sync()

	case AWS:
		p := &aws{baseURL: "http://169.254.169.254/latest/meta-data"}
		return p.Sync()

	case Tencent:
		p := &tencent{baseURL: "http://metadata.tencentyun.com/latest/meta-data"}
		return p.Sync()
	case Azure:
		p := &azure{baseURL: "http://169.254.169.254/metadata/instance"}
		return p.Sync()
	case Hwcloud:
		p := &hwcloud{baseURL: "http://169.254.169.254/latest/meta-data"}
		return p.Sync()
	default:
		return nil, fmt.Errorf("unknown cloud_provider: %s", provider)
	}
}

func (ipt *Input) matchCloudProvider(cloudProvider string) bool {
	fields, err := ipt.SyncCloudInfo(cloudProvider)
	if err != nil {
		return false
	}
	instanceID, has := fields["instance_id"]
	if !has || instanceID == Unavailable {
		return false
	}
	if cloudProvider == Hwcloud || cloudProvider == AWS {
		// Both of hwcloud and aws use the same URL. They can be distinguished by
		// field 'availability-zone-id', which is present in aws but not hwcloud.
		if metaGet("http://169.254.169.254/latest/meta-data/placement/availability-zone-id") == Unavailable {
			return cloudProvider == Hwcloud
		}
		return cloudProvider == AWS
	}
	return true
}

func (ipt *Input) SetCloudProvider() error {
	cloudProviders := []string{Aliyun, AWS, Tencent, Azure, Hwcloud}
	for _, cp := range cloudProviders {
		if ipt.matchCloudProvider(cp) {
			ipt.Tags["cloud_provider"] = cp
			return nil
		}
	}
	return fmt.Errorf("did not match any cloud provider")
}

func metaGet(metaURL string) (res string) {
	if x := metadataGet(metaURL); x != nil {
		// 避免 meta 接口返回多行数据
		res = string(bytes.ReplaceAll(x, []byte{'\n'}, []byte{' '}))
		return
	}

	return Unavailable
}

func clientDo(req *http.Request, metaURL string) []byte {
	resp, err := cloudCli.Do(req)
	if err != nil {
		l.Warnf("this maybe not a cloud node: %q, ignored.", err.Error())
		return nil
	}

	if resp.StatusCode != 200 {
		l.Warnf("request %q got status code %d", metaURL, resp.StatusCode)
		return nil
	}
	defer resp.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Warnf("read response %s: %s", metaURL, err)
		return nil
	}
	return body
}

func metadataGet(metaURL string) []byte {
	req, err := http.NewRequest("GET", metaURL, nil)
	if err != nil {
		l.Errorf("http.NewRequest: %s", err)
		return nil
	}

	return clientDo(req, metaURL)
}

func metadataGetByHeader(metaURL string) []byte {
	req, err := http.NewRequest("GET", metaURL, nil)
	if err != nil {
		l.Warn(err)
		return nil
	}
	req.Header.Set("Metadata", "true")

	return clientDo(req, metaURL)
}
