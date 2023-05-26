// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

// APIConfig used to unmarshal HTTP API server configurations.
type APIConfig struct {
	RUMOriginIPHeader string   `toml:"rum_origin_ip_header"`
	Listen            string   `toml:"listen"`
	Disable404Page    bool     `toml:"disable_404page"`
	RUMAppIDWhiteList []string `toml:"rum_app_id_white_list"`
	PublicAPIs        []string `toml:"public_apis"`
	RequestRateLimit  float64  `toml:"request_rate_limit,omitzero"`

	Timeout             string `toml:"timeout"`
	CloseIdleConnection bool   `toml:"close_idle_connection"`
}

// DCAConfig used to unmarshal DCA HTTP API server configurations.
type DCAConfig struct {
	Enable    bool     `toml:"enable" json:"enable"`
	Listen    string   `toml:"listen" json:"listen"`
	WhiteList []string `toml:"white_list" json:"white_list"`
}
