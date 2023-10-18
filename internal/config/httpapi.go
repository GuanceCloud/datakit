// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
)

type TLSConfig struct {
	Cert    string `toml:"cert"`
	PrivKey string `toml:"privkey"`
}

// APIConfig used to unmarshal HTTP API server configurations.
type APIConfig struct {
	RUMOriginIPHeader   string     `toml:"rum_origin_ip_header"`
	Listen              string     `toml:"listen"`
	Disable404Page      bool       `toml:"disable_404page"`
	RUMAppIDWhiteList   []string   `toml:"rum_app_id_white_list"`
	PublicAPIs          []string   `toml:"public_apis"`
	RequestRateLimit    float64    `toml:"request_rate_limit,omitzero"`
	Timeout             string     `toml:"timeout"`
	CloseIdleConnection bool       `toml:"close_idle_connection"`
	TLSConf             *TLSConfig `toml:"tls"`
}

func (conf *APIConfig) HTTPSEnabled() bool {
	if conf.TLSConf != nil {
		if finfo, err := os.Stat(conf.TLSConf.Cert); err != nil {
			return false
		} else if finfo.IsDir() {
			return false
		}
		if finfo, err := os.Stat(conf.TLSConf.PrivKey); err != nil {
			return false
		} else if finfo.IsDir() {
			return false
		}

		return true
	}

	return false
}

// DCAConfig used to unmarshal DCA HTTP API server configurations.
type DCAConfig struct {
	Enable    bool     `toml:"enable" json:"enable"`
	Listen    string   `toml:"listen" json:"listen"`
	WhiteList []string `toml:"white_list" json:"white_list"`
}
