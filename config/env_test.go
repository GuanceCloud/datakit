// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestLoadEnv(t *testing.T) {
	cases := []struct {
		name   string
		envs   map[string]string
		expect *Config
	}{
		{
			name: "normal",
			envs: map[string]string{
				"ENV_GLOBAL_HOST_TAGS":                "a=b,c=d",
				"ENV_GLOBAL_TAGS":                     "x=y,m=n",
				"ENV_LOG_LEVEL":                       "debug",
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_HOSTNAME":                        "1024.coding",
				"ENV_NAME":                            "testing-datakit",
				"ENV_HTTP_LISTEN":                     "localhost:9559",
				"ENV_RUM_ORIGIN_IP_HEADER":            "not-set",
				"ENV_ENABLE_PPROF":                    "true",
				"ENV_DISABLE_PROTECT_MODE":            "true",
				"ENV_DEFAULT_ENABLED_INPUTS":          "cpu,mem,disk",
				"ENV_ENABLE_ELECTION":                 "1",
				"ENV_NAMESPACE":                       "some-default",
				"ENV_DISABLE_404PAGE":                 "on",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "123",
				"ENV_REQUEST_RATE_LIMIT":              "1234",
				"ENV_DATAWAY_ENABLE_HTTPTRACE":        "any",
				"ENV_DATAWAY_HTTP_PROXY":              "http://1.2.3.4:1234",
				"ENV_HTTP_CLOSE_IDLE_CONNECTION":      "on",
				"ENV_HTTP_TIMEOUT":                    "10s",
				"ENV_ENABLE_ELECTION_NAMESPACE_TAG":   "ok",
			},
			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.Name = "testing-datakit"
				cfg.DataWayCfg = &dataway.DataWayCfg{
					URLs:                []string{"http://host1.org", "http://host2.com"},
					MaxIdleConnsPerHost: 123,
					HTTPProxy:           "http://1.2.3.4:1234",
					Proxy:               true,
					EnableHTTPTrace:     true,
				}

				cfg.HTTPAPI.RUMOriginIPHeader = "not-set"
				cfg.HTTPAPI.Listen = "localhost:9559"
				cfg.HTTPAPI.Disable404Page = true
				cfg.HTTPAPI.RequestRateLimit = 1234.0
				cfg.HTTPAPI.Timeout = "10s"
				cfg.HTTPAPI.CloseIdleConnection = true

				cfg.Logging.Level = "debug"

				cfg.EnablePProf = true
				cfg.Hostname = "1024.coding"
				cfg.ProtectMode = false
				cfg.DefaultEnabledInputs = []string{"cpu", "mem", "disk"}

				cfg.EnableElection = true
				cfg.EnableElectionTag = true
				cfg.ElectionNamespace = "some-default"

				cfg.GlobalHostTags = map[string]string{
					"a": "b",
					"c": "d",
					"x": "y",
					"m": "n",
				}

				cfg.GlobalEnvTags = map[string]string{
					"election_namespace": "some-default",
				}

				return cfg
			}(),
		},

		{
			name: "test-ENV_IO_FILTERS",
			envs: map[string]string{
				"ENV_IO_FILTERS": `
					{
					  "logging":[
							"{ source = 'datakit' and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}",
							"{ source = 'abc' and ( host in ['ubt-dev-02', 'tanb-ubt-dev-test-1'] )}"
						]
					}`,
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.IOConf.Filters = map[string][]string{
					"logging": {
						`{ source = 'datakit' and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}`,
						"{ source = 'abc' and ( host in ['ubt-dev-02', 'tanb-ubt-dev-test-1'] )}",
					},
				}
				return cfg
			}(),
		},

		{
			name: "test-ENV_IO_FILTERS-with-bad-json",
			envs: map[string]string{
				"ENV_IO_FILTERS": `
					{
					  "logging":[
							"{ source = 'datakit' and ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )}",
							"{ source = 'abc' and ( host in ['ubt-dev-02', 'tanb-ubt-dev-test-1'] )}"
						], # bad json
					}`,
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				return cfg
			}(),
		},

		{
			name: "test-ENV_IO_FILTERS-with-bad-condition",
			envs: map[string]string{
				"ENV_IO_FILTERS": `
					{
					  "logging":[
							"{ source = 'datakit' and-xx ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )} # and-xx is invalid"
						]
					}`,
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.IOConf.Filters = map[string][]string{
					"logging": {
						"{ source = 'datakit' and-xx ( host in ['ubt-dev-01', 'tanb-ubt-dev-test'] )} # and-xx is invalid",
					},
				}
				return cfg
			}(),
		},

		{
			name: "test-ENV_RUM_APP_ID_WHITE_LIST",
			envs: map[string]string{
				"ENV_RUM_APP_ID_WHITE_LIST": "appid-1,appid-2",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.HTTPAPI.RUMAppIDWhiteList = []string{"appid-1", "appid-2"}
				return cfg
			}(),
		},

		{
			name: "test-ENV_HTTP_PUBLIC_APIS",
			envs: map[string]string{
				"ENV_HTTP_PUBLIC_APIS": "/v1/write/rum",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.HTTPAPI.PublicAPIs = []string{"/v1/write/rum"}
				return cfg
			}(),
		},

		{
			name: "test-ENV_REQUEST_RATE_LIMIT",
			envs: map[string]string{
				"ENV_REQUEST_RATE_LIMIT": "1234.0",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.HTTPAPI.RequestRateLimit = 1234.0
				return cfg
			}(),
		},

		{
			name: "bad-ENV_REQUEST_RATE_LIMIT",
			envs: map[string]string{
				"ENV_REQUEST_RATE_LIMIT": "0.1234.0",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.HTTPAPI.RequestRateLimit = 0
				return cfg
			}(),
		},

		{
			name: "test-ENV_IPDB",
			envs: map[string]string{
				"ENV_IPDB": "iploc",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Pipeline.IPdbType = "iploc"
				return cfg
			}(),
		},

		{
			name: "test-unknown-ENV_IPDB",
			envs: map[string]string{
				"ENV_IPDB": "unknown-ipdb",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Pipeline.IPdbType = "-"
				return cfg
			}(),
		},

		{
			name: "test-ENV_ENABLE_INPUTS",
			envs: map[string]string{
				"ENV_ENABLE_INPUTS": "cpu,mem,disk",
			},
			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.DefaultEnabledInputs = []string{"cpu", "mem", "disk"}
				return cfg
			}(),
		},

		{
			name: "test-ENV_GLOBAL_TAGS",
			envs: map[string]string{
				"ENV_GLOBAL_TAGS": "cpu,mem,disk=sda",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.GlobalHostTags = map[string]string{"disk": "sda"}
				return cfg
			}(),
		},

		{
			name: "test-ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST",
			envs: map[string]string{
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "-1",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.DataWayCfg = &dataway.DataWayCfg{
					URLs:                []string{"http://host1.org", "http://host2.com"},
					MaxIdleConnsPerHost: 0,
				}

				return cfg
			}(),
		},

		{
			name: "test-io-envs",
			envs: map[string]string{
				"ENV_IO_MAX_CACHE_COUNT": "8192",

				"ENV_IO_ENABLE_CACHE":      "hahahah",
				"ENV_IO_CACHE_MAX_SIZE_GB": "8",

				"ENV_IO_FLUSH_INTERVAL": "2s",
				"ENV_IO_BLOCKING_MODE":  "on",
				"ENV_IO_QUEUE_SIZE":     "123",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.IOConf.FeedChanSize = 123
				cfg.IOConf.MaxCacheCount = 8192
				cfg.IOConf.MaxDynamicCacheCount = 8192

				cfg.IOConf.EnableCache = true
				cfg.IOConf.CacheSizeGB = 8
				cfg.IOConf.FlushInterval = "2s"
				cfg.IOConf.BlockingMode = true

				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()
			os.Clearenv()
			for k, v := range tc.envs {
				if err := os.Setenv(k, v); err != nil {
					t.Fatal(err)
				}
			}
			if err := c.LoadEnvs(); err != nil {
				t.Error(err)
				return
			}

			a := tc.expect.String()
			b := c.String()
			tu.Equals(t, a, b)
		})
	}
}
