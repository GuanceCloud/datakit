// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

type ConfdCfg struct {
	Enable         bool     `toml:"enable"`          // is this backend enable
	AuthToken      string   `toml:"auth_token"`      // space
	AuthType       string   `toml:"auth_type"`       // space
	Backend        string   `toml:"backend"`         // Kind of backend，example：etcdv3 zookeeper redis consul nacos file
	BasicAuth      bool     `toml:"basic_auth"`      // basic auth, applicable: etcdv3 consul
	ClientCaKeys   string   `toml:"client_ca_keys"`  // client ca keys, applicable: etcdv3 consul
	ClientCert     string   `toml:"client_cert"`     // client cert, applicable: etcdv3 consul
	ClientKey      string   `toml:"client_key"`      // client key, applicable: etcdv3 consul redis
	ClientInsecure bool     `toml:"client_insecure"` // space
	BackendNodes   []string `toml:"nodes"`           // backend servers, applicable: etcdv3 zookeeper redis consul nacos
	Password       string   `toml:"password"`        // applicable: etcdv3 consul nacos
	Scheme         string   `toml:"scheme"`          // applicable: etcdv3 consul
	Table          string   `toml:"table"`           // space
	Separator      string   `toml:"separator"`       // redis DB number, default 0
	Username       string   `toml:"username"`        // applicable: etcdv3 consul nacos
	AppID          string   `toml:"app_id"`          // space
	UserID         string   `toml:"user_id"`         // space
	RoleID         string   `toml:"role_id"`         // space
	SecretID       string   `toml:"secret_id"`       // space
	YAMLFile       []string `toml:"file"`            // backend files
	Filter         string   `toml:"filter"`          // space
	Path           string   `toml:"path"`            // space
	Role           string   // space

	AccessKey         string `toml:"access_key"`         // for nacos
	SecretKey         string `toml:"secret_key"`         // for nacos
	CircleInterval    int    `toml:"circle_interval"`    // cycle time interval in second
	ConfdNamespace    string `toml:"confd_namespace"`    // nacos confd namespace id
	PipelineNamespace string `toml:"pipeline_namespace"` // nacos pipeline namespace id
	Region            string `toml:"region"`
}

func ConfdEnabled() bool {
	for _, confd := range Cfg.Confds {
		if confd.Enable {
			return true
		}
	}

	return false
}
