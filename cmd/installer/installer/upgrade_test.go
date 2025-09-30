// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	T "testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

func Test_setupDefaultInputs(t *T.T) {
	t.Run("install", func(t *T.T) {
		opt := DefaultInstallArgs()
		c := config.DefaultConfig()
		opt.setupDefaultInputs(c, []string{"1", "2", "3"})
		assert.Equal(t, []string{"1", "2", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-white-list", func(t *T.T) {
		c := config.DefaultConfig()

		opt := DefaultInstallArgs()
		opt.EnableInputs = "2,mem"

		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		assert.Equal(t, []string{
			"-1",
			"-3",
			"2",
			"mem",
		}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-merged-white-list", func(t *T.T) {
		c := config.DefaultConfig()
		c.DefaultEnabledInputs = []string{"disk"}
		opt := DefaultInstallArgs()
		opt.EnableInputs = "2,mem" // white list, with extra input 'mem'
		opt.FlagDKUpgrade = true

		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		assert.Equal(t, []string{"-1", "-3", "2", "mem"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-blacklist", func(t *T.T) {
		c := config.DefaultConfig()
		opt := DefaultInstallArgs()
		opt.EnableInputs = "-2"
		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		assert.Equal(t, []string{"-2", "1", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-mixed-black-white-list", func(t *T.T) {
		c := config.DefaultConfig()
		opt := DefaultInstallArgs()
		opt.EnableInputs = "-2,1" // mixed w/b list: only black list applied
		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		assert.Equal(t, []string{"-2", "1", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-empty-list", func(t *T.T) {
		c := config.DefaultConfig()
		opt := DefaultInstallArgs()
		opt.FlagDKUpgrade = true
		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		assert.Equal(t, []string{"-1", "-2", "-3"}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-black-list", func(t *T.T) {
		c := config.DefaultConfig()
		c.DefaultEnabledInputs = []string{"-1"}

		opt := DefaultInstallArgs()
		opt.FlagDKUpgrade = true
		opt.setupDefaultInputs(c, []string{"1", "2", "3"})

		// NOTE: under blacklist, new added inputs are accepted
		assert.Equal(t, []string{"-1", "2", "3"}, c.DefaultEnabledInputs)
	})
}

func Test_upgradeMainConfInstance(t *T.T) {
	cases := []struct {
		name string
		old,
		expect *config.Config
	}{
		{
			name: "upgrade-http-api-limit",
			old: func() *config.Config {
				c := config.DefaultConfig()

				// set to old version's default values
				c.HTTPAPI.RequestRateLimit = 20.0

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()

				t.Logf("c.HTTPAPI: %+#v", c.HTTPAPI)
				return c
			}(),
		},

		{
			name: "upgrade-enable-pprof",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.EnablePProf = false
				c.PProfListen = ""
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				return c
			}(),
		},

		{
			name: "upgrade-http-timeout",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.DeprecatedHTTPTimeout = "10m"
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.HTTPTimeout = 10 * time.Minute
				return c
			}(),
		},

		{
			name: "upgrade-invalid-http-timeout",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.DeprecatedHTTPTimeout = "10min"
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.HTTPTimeout = 30 * time.Second // use default
				return c
			}(),
		},

		{
			name: "upgrade-election",

			old: func() *config.Config {
				c := config.DefaultConfig()
				c.ElectionNamespaceDeprecated = "ns-abc"
				c.GlobalEnvTagsDeprecated = map[string]string{
					"tag1": "val1",
				}
				c.EnableElectionDeprecated = true
				c.EnableElectionTagDeprecated = true

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Election = &election.ElectionCfg{
					Namespace:          "ns-abc",
					Enable:             true,
					NodeWhitelist:      []string{},
					EnableNamespaceTag: true,
					Tags: map[string]string{
						"tag1": "val1",
					},
				}

				return c
			}(),
		},

		{
			name: "upgrade-election-another",

			old: func() *config.Config {
				c := config.DefaultConfig()
				c.NamespaceDeprecated = "ns-abc"
				c.GlobalEnvTagsDeprecated = map[string]string{
					"tag1": "val1",
				}
				c.EnableElectionDeprecated = true
				c.EnableElectionTagDeprecated = true

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Election = &election.ElectionCfg{
					Namespace:          "ns-abc",
					Enable:             true,
					NodeWhitelist:      []string{},
					EnableNamespaceTag: true,
					Tags: map[string]string{
						"tag1": "val1",
					},
				}

				return c
			}(),
		},

		{
			name: "upgrade-logging",

			old: func() *config.Config {
				c := config.DefaultConfig()
				c.LogDeprecated = "/some/path"
				c.LogLevelDeprecated = "debug"
				c.GinLogDeprecated = "/some/gin/log"
				c.LogRotateDeprecated = 128

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Logging = &config.LoggerCfg{
					Log:           "/some/path",
					GinLog:        "/some/gin/log",
					Level:         "debug",
					Rotate:        128,
					RotateBackups: 5,
				}

				return c
			}(),
		},

		{
			name: "upgrade-http",

			old: func() *config.Config {
				c := config.DefaultConfig()
				c.HTTPListenDeprecated = ":12345"
				c.Disable404PageDeprecated = true

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.HTTPAPI.Listen = ":12345"
				c.HTTPAPI.Disable404Page = true

				return c
			}(),
		},

		{
			name: "upgrade-io",

			old: func() *config.Config {
				c := config.DefaultConfig()
				c.IOCacheCountDeprecated = 10
				c.IntervalDeprecated = 100 * time.Second

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.IO.MaxCacheCount = 1000 // auto reset to 10000
				c.IO.CompactInterval = 100 * time.Second

				return c
			}(),
		},

		{
			name: "default-encoding-v2",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.ContentEncoding = "v1"

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				return c
			}(),
		},

		{
			name: "set-default-raw-body-size",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.MaxRawBodySize = dataway.DeprecatedDefaultMaxRawBodySize

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				return c
			}(),
		},

		{
			name: "default-enable-point-pool",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.PointPool.Enable = false

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				return c
			}(),
		},

		{
			name: "apply-old-cpu-max-limit",
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.ResourceLimitOptions.CPUCores = 0            // clear
				c.ResourceLimitOptions.CPUMaxDeprecated = 20.0 // old cpu-max exist, do not apply cpu-cores
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.ResourceLimitOptions.CPUMaxDeprecated = 0.0
				c.ResourceLimitOptions.CPUCores = resourcelimit.CPUMaxToCores(20.0) // convert cpu-max to cpu-cores
				return c
			}(),
		},

		{
			name: "apply-new-cpu-cores-limit",
			old: func() *config.Config {
				c := config.DefaultConfig() // old cpu-max not set, use cpu-cores
				c.ResourceLimitOptions.CPUCores = 0.5
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.ResourceLimitOptions.CPUCores = 0.5
				return c
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			got := upgradeMainConfInstance(tc.old)
			assert.Equal(t, tc.expect.String(), got.String())

			c := config.DefaultConfig()
			_, err := bstoml.Decode(got.String(), c)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect.String(), c.String())
		})
	}
}
