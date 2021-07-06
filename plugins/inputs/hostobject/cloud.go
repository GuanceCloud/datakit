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

var (
	cloudCli = &http.Client{Timeout: 100 * time.Millisecond}
)

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

func (x *Input) SyncCloudInfo(provider string) (map[string]interface{}, error) {

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

	default:
		return nil, fmt.Errorf("unknown cloud_provider: %s", provider)
	}
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

	// 避免 meta 接口返回多行数据
	res = string(bytes.Replace(x, []byte{'\n'}, []byte{' '}, -1))
	return
}
