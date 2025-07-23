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
	"github.com/GuanceCloud/pipeline-go/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/operator"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/recorder"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

type pointPool struct {
	Enable           bool  `toml:"enable"`
	ReservedCapacity int64 `toml:"reserved_capacity,omitempty"`
}

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

	IntervalDeprecated   time.Duration `toml:"interval,omitempty"`
	OutputFileDeprecated string        `toml:"output_file,omitempty"`
	UUIDDeprecated       string        `toml:"uuid,omitempty"`

	PointPool *pointPool `toml:"point_pool"`

	// debug
	EnableDebugFields bool `toml:"enable_debug_fields,omitempty"`
	// pprof
	EnablePProf bool   `toml:"enable_pprof"`
	PProfListen string `toml:"pprof_listen"`

	// confd config
	Confds []*ConfdCfg `toml:"confds"`

	// DCA config
	DCAConfig *DCAConfig `toml:"dca"`

	// dk_upgrader
	DKUpgrader *DKUpgraderCfg `toml:"dk_upgrader"`

	// pipeline
	Pipeline *plval.PipelineCfg `toml:"pipeline"`

	// logging config
	LogDeprecated       string     `toml:"log,omitempty"`
	LogLevelDeprecated  string     `toml:"log_level,omitempty"`
	GinLogDeprecated    string     `toml:"gin_log,omitempty"`
	LogRotateDeprecated int        `toml:"log_rotate,omitzero"`
	Logging             *LoggerCfg `toml:"logging"`

	InstallVer string `toml:"install_version,omitempty"`

	HTTPAPI *APIConfig `toml:"http_api"`

	APMInject *APMInject `toml:"apm_inject"`

	Recorder               *recorder.Recorder `toml:"recorder"`
	IO                     *io.IOConf         `toml:"io"`
	IOCacheCountDeprecated int                `toml:"io_cache_count,omitzero"`

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

	Election *election.ElectionCfg `toml:"election"`

	// 是否已开启自动更新，通过 dk-install --ota 来开启
	AutoUpdate bool `toml:"auto_update,omitempty"`

	Tracer *tracer.Tracer `toml:"tracer,omitempty"`

	GitRepos *GitRepost `toml:"git_repos"`

	Ulimit      uint64 `toml:"ulimit"`
	DatakitUser string `toml:"datakit_user"`

	// crypto
	Crypto *configCrpto `toml:"crypto,omitempty"`

	RemoteJob *io.RemoteJob `toml:"remote_job,omitempty"`

	cmdlineMode bool
}

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
		DefaultEnabledInputs: []string{},
		PointPool: &pointPool{
			Enable:           false,
			ReservedCapacity: 4096,
		},

		GlobalHostTags:       map[string]string{},
		GlobalTagsDeprecated: map[string]string{},

		EnableDebugFields: false,
		EnablePProf:       true,
		PProfListen:       "localhost:6060",
		DatakitUser:       "root",

		Election: &election.ElectionCfg{
			Enable:             false,
			NodeWhitelist:      []string{},
			EnableNamespaceTag: false,
			Namespace:          "default",
			Tags:               map[string]string{},
		},

		Environments: map[string]string{
			"ENV_HOSTNAME": "", // not set
		}, // default nothing

		IO: &io.IOConf{
			FeedChanSize:            1,
			MaxCacheCount:           1000,
			CompactInterval:         time.Second * 10,
			AutoTimestampCorrection: true,

			Filters: nil,
		},

		Recorder: &recorder.Recorder{
			Enabled:    false,
			Path:       "",
			Encoding:   "v2",
			Duration:   time.Minute * 30,
			Inputs:     []string{},
			Categories: []string{},
		},

		Dataway: dataway.NewDefaultDataway(),

		Operator: &operator.Operator{},

		ProtectMode: true,

		HTTPAPI: &APIConfig{
			RUMOriginIPHeader:   "X-Forwarded-For",
			Listen:              "localhost:9529",
			RUMAppIDWhiteList:   []string{},
			PublicAPIs:          []string{},
			RequestRateLimit:    20,
			Timeout:             "30s",
			CloseIdleConnection: false,
			TLSConf:             &TLSConfig{},
			AllowedCORSOrigins:  []string{},
		},

		DCAConfig: &DCAConfig{
			Enable:          false,
			WebsocketServer: "ws://localhost:8000/ws",
		},

		APMInject: &APMInject{},

		DKUpgrader: &DKUpgraderCfg{
			Host: "0.0.0.0",
			Port: 9542,
		},

		Pipeline: &plval.PipelineCfg{
			IPdbType:               "iploc",
			RemotePullInterval:     "1m",
			ReferTableURL:          "",
			ReferTablePullInterval: "5m",
			DefaultPipeline:        map[string]string{},
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
			Path:     "/datakit",
			Enable:   true,
			CPUCores: 2.0,
			MemMax:   4096, // MB
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
		Crypto: &configCrpto{},
		RemoteJob: &io.RemoteJob{
			Enable: false,
			ENVs: []string{
				"OSS_BUCKET_HOST=host",
				"OSS_ACCESS_KEY_ID=key",
				"OSS_ACCESS_KEY_SECRET=secret",
				"OSS_BUCKET_NAME=bucket",
			},
			Interval: "30s",
			JavaHome: "",
		},
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == datakit.OSWindows {
		c.Logging.Log = filepath.Join(datakit.InstallDir, "log")
		c.Logging.GinLog = filepath.Join(datakit.InstallDir, "gin.log")
	}

	return c
}
