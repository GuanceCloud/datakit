// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/operator"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

type Config struct {
	DefaultEnabledInputs []string `toml:"default_enabled_inputs"`

	BlackList []*inputHostList `toml:"black_lists,omitempty"`
	WhiteList []*inputHostList `toml:"white_lists,omitempty"`

	UUID    string `toml:"-"`
	RunMode int    `toml:"-"`

	Name     string `toml:"name,omitempty"`
	Hostname string `toml:"-"`

	// http config: TODO: merge into APIConfig
	HTTPBindDeprecated   string `toml:"http_server_addr,omitempty"`
	HTTPListenDeprecated string `toml:"http_listen,omitempty"`

	IntervalDeprecated   string `toml:"interval,omitempty"`
	OutputFileDeprecated string `toml:"output_file,omitempty"`
	UUIDDeprecated       string `toml:"uuid,omitempty"` // deprecated

	// pprof
	EnablePProf bool   `toml:"enable_pprof"`
	PProfListen string `toml:"pprof_listen"`

	// confd config
	Confds []*ConfdCfg `toml:"confds"`

	// DCA config
	DCAConfig *DCAConfig `toml:"dca"`

	// pipeline
	Pipeline *pipeline.PipelineCfg `toml:"pipeline"`

	// logging config
	LogDeprecated       string     `toml:"log,omitempty"`
	LogLevelDeprecated  string     `toml:"log_level,omitempty"`
	GinLogDeprecated    string     `toml:"gin_log,omitempty"`
	LogRotateDeprecated int        `toml:"log_rotate,omitzero"`
	Logging             *LoggerCfg `toml:"logging"`

	InstallVer string `toml:"install_version,omitempty"`

	HTTPAPI *APIConfig `toml:"http_api"`

	IO                     *IOConf `toml:"io"`
	IOCacheCountDeprecated int     `toml:"io_cache_count,omitzero"`

	Dataway  *dataway.Dataway   `toml:"dataway"`
	Operator *operator.Operator `toml:"-"`

	GlobalHostTags       map[string]string `toml:"global_host_tags"`
	GlobalTagsDeprecated map[string]string `toml:"global_tags,omitempty"`

	Environments                   map[string]string                   `toml:"environments"`
	ResourceLimitOptionsDeprecated *resourcelimit.ResourceLimitOptions `toml:"cgroup,omitempty"`
	ResourceLimitOptions           *resourcelimit.ResourceLimitOptions `toml:"resource_limit"`

	Disable404PageDeprecated bool `toml:"disable_404page,omitempty"`
	ProtectMode              bool `toml:"protect_mode"`

	EnableElectionDeprecated    bool              `toml:"enable_election,omitempty"`
	EnableElectionTagDeprecated bool              `toml:"enable_election_tag,omitempty"`
	ElectionNamespaceDeprecated string            `toml:"election_namespace,omitempty"`
	NamespaceDeprecated         string            `toml:"namespace,omitempty"` // 避免跟 k8s 的 namespace 概念混淆
	GlobalEnvTagsDeprecated     map[string]string `toml:"global_env_tags,omitempty"`

	Election *ElectionCfg `toml:"election"`

	// 是否已开启自动更新，通过 dk-install --ota 来开启
	AutoUpdate bool `toml:"auto_update,omitempty"`

	Tracer *tracer.Tracer `toml:"tracer,omitempty"`

	GitRepos *GitRepost `toml:"git_repos"`

	Ulimit uint64 `toml:"ulimit"`
}

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
		GlobalHostTags:       map[string]string{},
		GlobalTagsDeprecated: map[string]string{},

		EnablePProf: true,
		PProfListen: "localhost:6060",

		Election: &ElectionCfg{
			Enable:             false,
			EnableNamespaceTag: false,
			Namespace:          "default",
			Tags:               map[string]string{},
		},

		Environments: map[string]string{
			"ENV_HOSTNAME": "", // not set
		}, // default nothing

		IO: &IOConf{
			FeedChanSize:     1,
			MaxCacheCount:    1000,
			FlushInterval:    "10s",
			OutputFileInputs: []string{},

			// Enable disk cache on datakit send fail.
			EnableCache:        false,
			CacheSizeGB:        10,
			CacheCleanInterval: "5s",

			Filters: nil,
		},

		Dataway: &dataway.Dataway{
			URLs:        []string{"not-set"},
			HTTPTimeout: 30 * time.Second,
			IdleTimeout: 90 * time.Second,
		},
		Operator: &operator.Operator{},

		ProtectMode: true,

		HTTPAPI: &APIConfig{
			RUMOriginIPHeader:   "X-Forwarded-For",
			Listen:              "localhost:9529",
			RUMAppIDWhiteList:   []string{},
			PublicAPIs:          []string{},
			Timeout:             "30s",
			CloseIdleConnection: false,
			TLSConf:             &TLSConfig{},
		},

		DCAConfig: &DCAConfig{
			Enable:    false,
			Listen:    "0.0.0.0:9531",
			WhiteList: []string{},
		},

		Pipeline: &pipeline.PipelineCfg{
			IPdbType:               "-",
			RemotePullInterval:     "1m",
			ReferTableURL:          "",
			ReferTablePullInterval: "5m",
			Offload: &offload.OffloadConfig{
				Receiver:  offload.DKRcv,
				Addresses: []string{},
			},
		},

		Logging: &LoggerCfg{
			Level:         "info",
			Rotate:        logger.MaxSize,
			RotateBackups: logger.MaxBackups,
			Log:           filepath.Join("/var/log/datakit", "log"),
			GinLog:        filepath.Join("/var/log/datakit", "gin.log"),
		},

		ResourceLimitOptions: &resourcelimit.ResourceLimitOptions{
			Path:   "/datakit",
			Enable: true,
			CPUMax: 20.0,
			MemMax: 4096, // MB
		},

		GitRepos: &GitRepost{
			PullInterval: "1m",
			Repos: []*GitRepository{
				{
					Enable:                false,
					URL:                   "",
					SSHPrivateKeyPath:     "",
					SSHPrivateKeyPassword: "",
					Branch:                "master",
				},
			},
		},

		Ulimit: func() uint64 {
			switch runtime.GOOS {
			case "linux":
				return uint64(64000)
			case "darwin":
				return uint64(10240)
			default:
				return uint64(0)
			}
		}(),
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == datakit.OSWindows {
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	}

	return c
}
