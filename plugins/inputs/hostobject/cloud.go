package hostobject

import (
	"fmt"
	"net/http"
	"time"
)

const (
	Unavailable = "-"
)

var (
	cloudCli = &http.Client{Timeout: 3 * time.Second}
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
	ExtraCloudMeta() string
}

func (x *Input) syncCloudInfo(provider string) (map[string]interface{}, error) {

	defer cloudCli.CloseIdleConnections()

	switch provider {
	case "aliyun":
		p := &aliyun{baseURL: "http://100.100.100.200/latest/meta-data"}
		return p.Sync()

	case "aws":
		return nil, fmt.Errorf("TODO")
	case "tencent":
		return nil, fmt.Errorf("TODO")
	default:
		return nil, fmt.Errorf("unknown cloud_provider: %s", provider)
	}
}
