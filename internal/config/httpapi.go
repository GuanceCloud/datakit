// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"time"
)

type TLSConfig struct {
	Cert    string `toml:"cert"`
	PrivKey string `toml:"privkey"`
}

type APIConfig struct {
	RUMOriginIPHeader string   `toml:"rum_origin_ip_header"`
	ListenSocket      string   `toml:"listen_socket"`
	Listen            string   `toml:"listen"`
	Disable404Page    bool     `toml:"disable_404page"`
	RUMAppIDWhiteList []string `toml:"rum_app_id_white_list"`
	PublicAPIs        []string `toml:"public_apis"`

	RequestRateLimit      float64       `toml:"request_rate_limit"`
	RequestRateLimitTTL   time.Duration `toml:"request_rate_limit_ttl"`
	RequestRateLimitBurst int           `toml:"request_rate_limit_burst"`

	IdleTimeout       time.Duration `toml:"http_idle_timeout"`
	ReadTimeout       time.Duration `toml:"http_read_timeout"`
	ReadHeaderTimeout time.Duration `toml:"http_read_head_timeout"`
	WriteTimeout      time.Duration `toml:"http_write_timeout"`

	TLSConf            *TLSConfig `toml:"tls"`
	DisableWhitelist   bool       `toml:"disable_whitelist"`
	AllowedCORSOrigins []string   `toml:"allowed_cors_origins"`
}

func DefaultAPIConfig() *APIConfig {
	return &APIConfig{
		RUMOriginIPHeader: "X-Forwarded-For",
		Listen:            "localhost:9529",
		RUMAppIDWhiteList: []string{},
		PublicAPIs:        []string{},

		DisableWhitelist:  false, // 添加这一行
		IdleTimeout:       time.Second * 60,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 30,
		WriteTimeout:      time.Second * 30,

		RequestRateLimit:      100,
		RequestRateLimitTTL:   time.Second * 60,
		RequestRateLimitBurst: 500, // 5 X ratelimit

		TLSConf:            &TLSConfig{},
		AllowedCORSOrigins: []string{},
	}
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
	Enable          bool   `toml:"enable" json:"enable"`
	WebsocketServer string `toml:"websocket_server" json:"websocket_server"`
}
