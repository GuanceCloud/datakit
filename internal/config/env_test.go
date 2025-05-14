// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/offload"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
)

func TestLoadEnv(t *testing.T) {
	cases := []struct {
		name   string
		envs   map[string]string
		expect *Config
	}{
		{
			name: "test-dataway-ntp",
			envs: map[string]string{
				"ENV_DATAWAY_NTP_INTERVAL": "5m",
				"ENV_DATAWAY_NTP_DIFF":     "30s",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Dataway.NTP.Interval = 5 * time.Minute
				cfg.Dataway.NTP.SyncOnDiff = 30 * time.Second

				return cfg
			}(),
		},
		{
			name: "test-recorder-envs",
			envs: map[string]string{
				"ENV_ENABLE_RECORDER":     "on",
				"ENV_RECORDER_PATH":       "/path/to/recorder",
				"ENV_RECORDER_ENCODING":   "v2",
				"ENV_RECORDER_DURATION":   "30s",
				"ENV_RECORDER_INPUTS":     "cpu,mem",
				"ENV_RECORDER_CATEGORIES": "metric,logging",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Recorder.Enabled = true
				cfg.Recorder.Path = "/path/to/recorder"
				cfg.Recorder.Encoding = "v2"
				cfg.Recorder.Duration = 30 * time.Second
				cfg.Recorder.Inputs = []string{"cpu", "mem"}
				cfg.Recorder.Categories = []string{"metric", "logging"}

				return cfg
			}(),
		},

		{
			name: "test-point-pool-envs",
			envs: map[string]string{
				"ENV_POINT_POOL_RESERVED_CAPACITY": "1234",
				"ENV_ENABLE_POINT_POOL":            "on",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.PointPool.Enable = false
				cfg.PointPool.ReservedCapacity = 1234

				return cfg
			}(),
		},

		{
			name: `bad-sinkers`,
			envs: map[string]string{
				"ENV_SINKER": `[ some bad json `,
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				return cfg
			}(),
		},

		{
			name: "normal",
			envs: map[string]string{
				"ENV_GLOBAL_HOST_TAGS": "a=b,c=d",
				"ENV_GLOBAL_TAGS":      "x=y,m=n", // deprecated, not used

				"ENV_LOG_LEVEL":                       "debug",
				"ENV_LOG_ROTATE_BACKUP":               "10",
				"ENV_LOG_ROTATE_SIZE_MB":              "128",
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_HOSTNAME":                        "1024.coding",
				"ENV_NAME":                            "testing-datakit",
				"ENV_HTTP_LISTEN":                     "localhost:9559",
				"ENV_HTTP_LISTEN_SOCKET":              "/var/run/datakit/datakit.sock",
				"ENV_RUM_ORIGIN_IP_HEADER":            "not-set",
				"ENV_ENABLE_PPROF":                    "true",
				"ENV_DISABLE_PROTECT_MODE":            "true",
				"ENV_DEFAULT_ENABLED_INPUTS":          "cpu,mem,disk",
				"ENV_ENABLE_ELECTION":                 "1",
				"ENV_NAMESPACE":                       "some-default",
				"ENV_DISABLE_404PAGE":                 "on",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "123",
				"ENV_DATAWAY_TLS_INSECURE":            "on",
				"ENV_REQUEST_RATE_LIMIT":              "1234",
				"ENV_HTTP_ALLOWED_CORS_ORIGINS":       "https://foo,https://bar",
				"ENV_DATAWAY_ENABLE_HTTPTRACE":        "any",
				"ENV_DATAWAY_HTTP_PROXY":              "http://1.2.3.4:1234",
				"ENV_HTTP_CLOSE_IDLE_CONNECTION":      "on",
				"ENV_HTTP_TIMEOUT":                    "10s",
				"ENV_HTTP_ENABLE_TLS":                 "yes",
				"ENV_HTTP_TLS_CRT":                    "/path/to/datakit/tls.crt",
				"ENV_HTTP_TLS_KEY":                    "/path/to/datakit/tls.key",

				"ENV_ENABLE_ELECTION_NAMESPACE_TAG":              "ok",
				"ENV_PIPELINE_OFFLOAD_RECEIVER":                  offload.DKRcv,
				"ENV_PIPELINE_OFFLOAD_ADDRESSES":                 "http://aaa:123,http://1.2.3.4:1234",
				"ENV_PIPELINE_DEFAULT_PIPELINE":                  `{"xxx":"a.p"}`,
				"ENV_PIPELINE_DISABLE_HTTP_REQUEST_FUNC":         "true",
				"ENV_PIPELINE_HTTP_REQUEST_HOST_WHITELIST":       `["guance.com", "10.0.0.1"]`,
				"ENV_PIPELINE_HTTP_REQUEST_CIDR_WHITELIST":       `["10.0.0.0/8"]`,
				"ENV_PIPELINE_HTTP_REQUEST_DISABLE_INTERNAL_NET": "true",
			},
			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.Name = "testing-datakit"

				cfg.Dataway.URLs = []string{"http://host1.org", "http://host2.com"}

				cfg.Dataway.MaxIdleConnsPerHost = 123
				cfg.Dataway.HTTPProxy = "http://1.2.3.4:1234"
				cfg.Dataway.EnableHTTPTrace = true
				cfg.Dataway.IdleTimeout = 90 * time.Second
				cfg.Dataway.HTTPTimeout = 30 * time.Second
				cfg.Dataway.ContentEncoding = "v2"
				cfg.Dataway.MaxRetryCount = dataway.DefaultRetryCount
				cfg.Dataway.InsecureSkipVerify = true
				cfg.Dataway.RetryDelay = dataway.DefaultRetryDelay
				cfg.Dataway.MaxRawBodySize = dataway.DefaultMaxRawBodySize
				cfg.Dataway.GlobalCustomerKeys = []string{}
				cfg.Dataway.GZip = true

				cfg.HTTPAPI.AllowedCORSOrigins = []string{"https://foo", "https://bar"}
				cfg.HTTPAPI.RUMOriginIPHeader = "not-set"
				cfg.HTTPAPI.Listen = "localhost:9559"
				cfg.HTTPAPI.ListenSocket = "/var/run/datakit/datakit.sock"
				cfg.HTTPAPI.Disable404Page = true
				cfg.HTTPAPI.RequestRateLimit = 1234.0
				cfg.HTTPAPI.Timeout = "10s"
				cfg.HTTPAPI.CloseIdleConnection = true
				cfg.HTTPAPI.TLSConf.Cert = "/path/to/datakit/tls.crt"
				cfg.HTTPAPI.TLSConf.PrivKey = "/path/to/datakit/tls.key"

				cfg.Logging.Level = "debug"
				cfg.Logging.RotateBackups = 10
				cfg.Logging.Rotate = 128

				cfg.Pipeline.Offload = &offload.OffloadConfig{}
				cfg.Pipeline.Offload.Receiver = offload.DKRcv
				cfg.Pipeline.Offload.Addresses = []string{"http://aaa:123", "http://1.2.3.4:1234"}

				cfg.Pipeline.DefaultPipeline = map[string]string{"xxx": "a.p"}
				cfg.Pipeline.DisableHTTPRequestFunc = true
				cfg.Pipeline.HTTPRequestHostWhitelist = []string{"guance.com", "10.0.0.1"}
				cfg.Pipeline.HTTPRequestCIDRWhitelist = []string{"10.0.0.0/8"}
				cfg.Pipeline.HTTPRequestDisableInternalNet = true

				cfg.EnablePProf = true
				cfg.Hostname = "1024.coding"
				cfg.ProtectMode = false
				cfg.DefaultEnabledInputs = []string{"cpu", "mem", "disk"}

				cfg.Election.Enable = true
				cfg.Election.EnableNamespaceTag = true
				cfg.Election.Namespace = "some-default"

				cfg.GlobalHostTags = map[string]string{
					"a": "b",
					"c": "d",
				}

				cfg.Election.Tags = map[string]string{
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
				cfg.IO.Filters = map[string]filter.FilterConditions{
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
				cfg.IO.Filters = map[string]filter.FilterConditions{
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
				cfg.HTTPAPI.RequestRateLimit = 20.0
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
				cfg.Pipeline.IPdbType = "iploc"
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
			name: "test-disable-env_protected",
			envs: map[string]string{
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "-1",
				"ENV_DATAWAY_TIMEOUT":                 "1m",
				"ENV_DATAWAY_ENABLE_HTTPTRACE":        "on",
				"ENV_DATAWAY_MAX_IDLE_CONNS":          "100",
				"ENV_DATAWAY_IDLE_TIMEOUT":            "100s",
				"ENV_DATAWAY_CONTENT_ENCODING":        "v2",
				"ENV_SINKER_GLOBAL_CUSTOMER_KEYS":     " , key1,key2 ,",
				"ENV_DATAWAY_MAX_RETRY_COUNT":         "8",
				"ENV_DATAWAY_RETRY_DELAY":             "5s",
				"ENV_DATAWAY_MAX_RAW_BODY_SIZE":       strconv.Itoa(1024 * 32),
				"ENV_DATAWAY_ENABLE_SINKER":           "set",

				"ENV_DISABLE_PROTECT_MODE": "set",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.ProtectMode = false

				cfg.Dataway.URLs = []string{"http://host1.org", "http://host2.com"}
				cfg.Dataway.MaxIdleConnsPerHost = 0
				cfg.Dataway.MaxIdleConns = 100
				cfg.Dataway.EnableHTTPTrace = true
				cfg.Dataway.IdleTimeout = 100 * time.Second
				cfg.Dataway.HTTPTimeout = time.Minute
				cfg.Dataway.GlobalCustomerKeys = []string{"key1", "key2"}
				cfg.Dataway.MaxRetryCount = 8
				cfg.Dataway.RetryDelay = time.Second * 5
				cfg.Dataway.MaxRawBodySize = 1024 * 32
				cfg.Dataway.ContentEncoding = "v2"
				cfg.Dataway.EnableSinker = true
				cfg.Dataway.GZip = true

				return cfg
			}(),
		},

		{
			name: "test-env_protected",
			envs: map[string]string{
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "-1",
				"ENV_DATAWAY_TIMEOUT":                 "1m",
				"ENV_DATAWAY_ENABLE_HTTPTRACE":        "on",
				"ENV_DATAWAY_MAX_IDLE_CONNS":          "100",
				"ENV_DATAWAY_IDLE_TIMEOUT":            "100s",
				"ENV_DATAWAY_CONTENT_ENCODING":        "v2",
				"ENV_SINKER_GLOBAL_CUSTOMER_KEYS":     " , key1,key2 ,",
				"ENV_DATAWAY_MAX_RETRY_COUNT":         "8",
				"ENV_DATAWAY_RETRY_DELAY":             "5s",
				"ENV_DATAWAY_MAX_RAW_BODY_SIZE":       strconv.Itoa(1024 * 32),
				"ENV_DATAWAY_ENABLE_SINKER":           "set",

				"ENV_DISABLE_PROTECT_MODE": "", // not-set
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.Dataway.URLs = []string{"http://host1.org", "http://host2.com"}
				cfg.Dataway.MaxIdleConnsPerHost = 0
				cfg.Dataway.MaxIdleConns = 100
				cfg.Dataway.EnableHTTPTrace = true
				cfg.Dataway.IdleTimeout = 100 * time.Second
				cfg.Dataway.HTTPTimeout = time.Minute
				cfg.Dataway.GlobalCustomerKeys = []string{"key1", "key2"}
				cfg.Dataway.MaxRetryCount = 8
				cfg.Dataway.RetryDelay = time.Second * 5
				cfg.Dataway.MaxRawBodySize = dataway.MinimalRawBodySize
				cfg.Dataway.ContentEncoding = "v2"
				cfg.Dataway.EnableSinker = true
				cfg.Dataway.GZip = true

				return cfg
			}(),
		},

		{
			name: "test-ENV_DATAWAY*",
			envs: map[string]string{
				"ENV_DATAWAY":                         "http://host1.org,http://host2.com",
				"ENV_DATAWAY_MAX_IDLE_CONNS_PER_HOST": "-1",
				"ENV_DATAWAY_TIMEOUT":                 "1m",
				"ENV_DATAWAY_ENABLE_HTTPTRACE":        "on",
				"ENV_DATAWAY_MAX_IDLE_CONNS":          "100",
				"ENV_DATAWAY_IDLE_TIMEOUT":            "100s",
				"ENV_DATAWAY_CONTENT_ENCODING":        "v2",
				"ENV_SINKER_GLOBAL_CUSTOMER_KEYS":     " , key1,key2 ,",
				"ENV_DATAWAY_MAX_RETRY_COUNT":         "8",
				"ENV_DATAWAY_RETRY_DELAY":             "5s",
				"ENV_DATAWAY_MAX_RAW_BODY_SIZE":       strconv.Itoa(1024 * 1024 * 32),
				"ENV_DATAWAY_ENABLE_SINKER":           "set",
				"ENV_DATAWAY_TLS_INSECURE":            "on",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.Dataway.URLs = []string{"http://host1.org", "http://host2.com"}
				cfg.Dataway.MaxIdleConnsPerHost = 0
				cfg.Dataway.MaxIdleConns = 100
				cfg.Dataway.EnableHTTPTrace = true
				cfg.Dataway.IdleTimeout = 100 * time.Second
				cfg.Dataway.HTTPTimeout = time.Minute
				cfg.Dataway.GlobalCustomerKeys = []string{"key1", "key2"}
				cfg.Dataway.MaxRetryCount = 8
				cfg.Dataway.RetryDelay = time.Second * 5
				cfg.Dataway.MaxRawBodySize = 1024 * 1024 * 32
				cfg.Dataway.ContentEncoding = "v2"
				cfg.Dataway.EnableSinker = true
				cfg.Dataway.GZip = true
				cfg.Dataway.InsecureSkipVerify = true

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
				"ENV_IO_FLUSH_WORKERS":  "1",

				"ENV_IO_FEED_CHAN_SIZE":       "123",
				"ENV_IO_FEED_GLOBAL_BLOCKING": "1",

				"ENV_IO_CACHE_CLEAN_INTERVAL": "100s",
				"ENV_IO_CACHE_ALL":            "on",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.IO.FeedChanSize = 1 // force reset to 1
				cfg.IO.MaxCacheCount = 8192

				cfg.IO.FeedChanSize = 123
				cfg.IO.CompactInterval = 2 * time.Second
				cfg.IO.CompactWorkers = 1

				return cfg
			}(),
		},

		{
			name: "disable-dw-gzip",
			envs: map[string]string{
				"ENV_DATAWAY_DISABLE_GZIP": "on",
			},

			expect: func() *Config {
				cfg := DefaultConfig()

				cfg.Dataway.GZip = false

				return cfg
			}(),
		},

		{
			name: "test-k8s-node-name",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME": "node1",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Hostname = "node1"

				return cfg
			}(),
		},

		{
			name: "test-k8s-cluster-node-name",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME":         "node1",
				"ENV_K8S_CLUSTER_NODE_NAME": "testing-env-node1",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.Hostname = "testing-env-node1"

				return cfg
			}(),
		},

		{
			name: "test-point-pool",
			envs: map[string]string{
				"ENV_POINT_POOL_RESERVED_CAPACITY": "12345",
				"ENV_DISABLE_POINT_POOL":           "yes",
			},

			expect: func() *Config {
				cfg := DefaultConfig()
				cfg.PointPool.Enable = false
				cfg.PointPool.ReservedCapacity = 12345

				return cfg
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()

			// setup envs
			for k, v := range tc.envs {
				assert.NoError(t, os.Setenv(k, v))
			}

			// load them
			assert.NoError(t, c.LoadEnvs())
			assert.Equal(t, tc.expect.String(), c.String(), "expect\n%s, get\n%s", tc.expect.String(), c.String())

			t.Cleanup(func() {
				for k := range tc.envs {
					assert.NoError(t, os.Unsetenv(k))
				}
			})
		})
	}
}

func TestSetNodenameAsHostname(t *testing.T) {
	cases := []struct {
		name                                       string
		envs                                       map[string]string
		expectHostname                             string
		expectNodeNamePrefix, expectNodeNameSuffix string
	}{
		{
			name: "test-prefix-nodeName",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME":         "host-abc",
				"ENV_K8S_CLUSTER_NODE_NAME": "cluster_host-abc",
			},
			expectHostname:       "cluster_host-abc",
			expectNodeNamePrefix: "cluster_",
			expectNodeNameSuffix: "",
		},
		{
			name: "test-prefix-nodeName-2",
			envs: map[string]string{
				"NODE_NAME":                 "host-abc",
				"ENV_K8S_CLUSTER_NODE_NAME": "cluster_host-abc",
			},
			expectHostname:       "cluster_host-abc",
			expectNodeNamePrefix: "cluster_",
			expectNodeNameSuffix: "",
		},
		{
			name: "test-suffix-nodeName",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME":         "host-abc",
				"ENV_K8S_CLUSTER_NODE_NAME": "host-abc_k8s",
			},
			expectHostname:       "host-abc_k8s",
			expectNodeNamePrefix: "",
			expectNodeNameSuffix: "_k8s",
		},
		{
			name: "test-prefix-and-suffix-nodeName",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME":         "host-abc",
				"ENV_K8S_CLUSTER_NODE_NAME": "cluster_host-abc_k8s",
			},
			expectHostname:       "cluster_host-abc_k8s",
			expectNodeNamePrefix: "cluster_",
			expectNodeNameSuffix: "_k8s",
		},
		{
			name: "test-only-nodeName",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME": "host-abc",
			},
			expectHostname:       "host-abc",
			expectNodeNamePrefix: "",
			expectNodeNameSuffix: "",
		},
		{
			name: "test-only-cluster-nodeName",
			envs: map[string]string{
				"ENV_K8S_CLUSTER_NODE_NAME": "cluster_host-abc_k8s",
			},
			expectHostname:       "cluster_host-abc_k8s",
			expectNodeNamePrefix: "",
			expectNodeNameSuffix: "",
		},
		{
			name: "test-no-match-nodeName",
			envs: map[string]string{
				"ENV_K8S_NODE_NAME":         "host-abc",
				"ENV_K8S_CLUSTER_NODE_NAME": "cluster_host-def_k8s",
			},
			expectHostname:       "cluster_host-def_k8s",
			expectNodeNamePrefix: "",
			expectNodeNameSuffix: "",
		},
		{
			name:                 "test-empty",
			envs:                 map[string]string{}, // empty
			expectHostname:       "",
			expectNodeNamePrefix: "",
			expectNodeNameSuffix: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()

			// setup envs
			for k, v := range tc.envs {
				assert.NoError(t, os.Setenv(k, v))
			}

			c.setNodenameAsHostname()

			assert.Equal(t, tc.expectHostname, c.Hostname)
			assert.Equal(t, tc.expectNodeNamePrefix, nodeNamePrefix)
			assert.Equal(t, tc.expectNodeNameSuffix, nodeNameSuffix)

			t.Cleanup(func() {
				for k := range tc.envs {
					assert.NoError(t, os.Unsetenv(k))
				}

				// reset global variables
				nodeNamePrefix = ""
				nodeNameSuffix = ""
			})
		})
	}
}
