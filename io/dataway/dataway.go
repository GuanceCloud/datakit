// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"net/http"
	"time"
)

type DataWay interface {
	Init(*DataWayCfg) error
	IsLogFilter() bool
	HeartBeat() (int, error)
	DatawayList() ([]string, int, error)
	GetTokens() []string
	GetLogFilter() ([]byte, error)
	GetPipelinePull(int64) (*PullPipelineReturn, error)
	ElectionHeartbeat(string, string) ([]byte, error)
	Election(string, string) ([]byte, error)
	UpsertObjectLabels(string, []byte) (*http.Response, error)
	DeleteObjectLabels(string, []byte) (*http.Response, error)
	DQLQuery([]byte) (*http.Response, error)
	WorkspaceQuery([]byte) (*http.Response, error)
	DatakitPull(string) ([]byte, error)
}

type DataWayCfg struct {
	URLs []string `toml:"urls"`

	DeprecatedURL string `toml:"url,omitempty"`
	HTTPTimeout   string `toml:"timeout"`
	HTTPProxy     string `toml:"http_proxy"`
	Hostname      string `toml:"-"`

	DeprecatedHost   string `toml:"host,omitempty"`
	DeprecatedScheme string `toml:"scheme,omitempty"`
	DeprecatedToken  string `toml:"token,omitempty"`

	MaxIdleConnsPerHost int `toml:"max_idle_conns_per_host,omitempty"`

	TimeoutDuration time.Duration `toml:"-"`

	MaxFails int `toml:"max_fail"`

	Proxy bool `toml:"proxy,omitempty"`

	EnableHTTPTrace bool `toml:"enable_httptrace,omitempty"`
}
