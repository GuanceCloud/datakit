package config

import (
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

func DefaultConfig() *Config {
	c := &Config{ //nolint:dupl
		GlobalHostTags:       map[string]string{},
		GlobalTagsDeprecated: map[string]string{},

		Election: &election.Config{
			Enable:             false,
			EnableNamespaceTag: false,
			Namespace:          "default",
			Tags:               map[string]string{},
		},

		Environments: map[string]string{
			"ENV_HOSTNAME": "", // not set
		}, // default nothing

		IOConf: &dkio.IOConfig{
			FeedChanSize:         128,
			MaxCacheCount:        64,
			MaxDynamicCacheCount: 0,
			FlushInterval:        "10s",
			OutputFileInputs:     []string{},

			EnableCache: false,
			CacheSizeGB: 1,

			Filters: map[string][]string{},
		},

		DataWayCfg: &dataway.DataWayCfg{
			URLs: []string{},
		},

		ProtectMode: true,

		HTTPAPI: &dkhttp.APIConfig{
			RUMOriginIPHeader:   "X-Forwarded-For",
			Listen:              "localhost:9529",
			RUMAppIDWhiteList:   []string{},
			PublicAPIs:          []string{},
			Timeout:             "30s",
			CloseIdleConnection: false,
		},

		DCAConfig: &dkhttp.DCAConfig{
			Enable:    false,
			Listen:    "0.0.0.0:9531",
			WhiteList: []string{},
		},
		Pipeline: &pipeline.PipelineCfg{
			IPdbType:               "-",
			RemotePullInterval:     "1m",
			ReferTableURL:          "",
			ReferTablePullInterval: "5m",
		},
		Logging: &LoggerCfg{
			Level:  "info",
			Rotate: 32,
			Log:    filepath.Join("/var/log/datakit", "log"),
			GinLog: filepath.Join("/var/log/datakit", "gin.log"),
		},

		Cgroup: &cgroup.CgroupOptions{
			Path:   "/datakit",
			Enable: true,
			CPUMax: 20.0,
			CPUMin: 5.0,
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

		Sinks: &Sinker{
			Sink: []map[string]interface{}{{}},
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
