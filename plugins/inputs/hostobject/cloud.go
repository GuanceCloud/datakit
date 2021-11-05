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
	case "aliyun":
		p := &aliyun{baseURL: "http://100.100.100.200/latest/meta-data"}
		return p.Sync()

	case "aws":
		p := &aws{baseURL: "http://169.254.169.254/latest/meta-data"}
		return p.Sync()

	case "tencent":
		p := &tencent{baseURL: "http://metadata.tencentyun.com/latest/meta-data"}
		return p.Sync()
	case "azure":
		p := &azure{baseURL: "http://169.254.169.254/metadata/instance"}
		return p.Sync()
	case "hwcloud":
		p := &hwcloud{baseURL: "http://169.254.169.254/latest/meta-data"}
		return p.Sync()
	default:
		return nil, fmt.Errorf("unknown cloud_provider: %s", provider)
	}
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
		l.Warn(err)
		return nil
	}

	if resp.StatusCode != 200 {
		l.Warnf("request %s: status code %d", metaURL, resp.StatusCode)
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
		l.Warn(err)
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
