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
)

func Test_setupDefaultInputs(t *T.T) {
	t.Run("install", func(t *T.T) {
		c := config.DefaultConfig()
		setupDefaultInputs(c,
			"", // no list specified: use all default
			[]string{"1", "2", "3"}, false)
		assert.Equal(t, []string{"1", "2", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-white-list", func(t *T.T) {
		c := config.DefaultConfig()
		setupDefaultInputs(c,
			"2,mem", // white list, with extra input 'mem'
			[]string{"1", "2", "3"}, false)

		assert.Equal(t, []string{
			"-1",
			"-3",
			"2",
			"mem",
		}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-white-list", func(t *T.T) {
		c := config.DefaultConfig()

		c.DefaultEnabledInputs = []string{"disk"}

		setupDefaultInputs(c,
			"2,mem", // white list, with extra input 'mem'
			[]string{"1", "2", "3"}, true)

		assert.Equal(t, []string{"1", "2", "3", "disk"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-blacklist", func(t *T.T) {
		c := config.DefaultConfig()
		setupDefaultInputs(c,
			"-2", // black list
			[]string{"1", "2", "3"}, false)

		assert.Equal(t, []string{"-2", "1", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("install-with-black-white-list", func(t *T.T) {
		c := config.DefaultConfig()
		setupDefaultInputs(c,
			"-2,1", // mixed w/b list: only black list applied
			[]string{"1", "2", "3"}, false)

		assert.Equal(t, []string{"-2", "1", "3"}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-empty-list", func(t *T.T) {
		c := config.DefaultConfig()
		setupDefaultInputs(c, "", []string{"1", "2", "3"}, true)

		assert.Equal(t, []string{"-1", "-2", "-3"}, c.DefaultEnabledInputs)
	})

	t.Run("upgrade-with-black-list", func(t *T.T) {
		c := config.DefaultConfig()
		c.DefaultEnabledInputs = []string{"-1"}

		setupDefaultInputs(c, "", []string{"1", "2", "3"}, true)

		// NOTE: under blacklist, new added inputs are accepted
		assert.Equal(t, []string{"-1", "2", "3"}, c.DefaultEnabledInputs)
	})
}

func TestUpgradeMainConfig(t *T.T) {
	cases := []struct {
		name string
		old,
		expect *config.Config
	}{
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
				c.IntervalDeprecated = "100s"

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.IO.MaxCacheCount = 1000 // auto reset to 10000
				c.IO.FlushInterval = "100s"

				return c
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			got := upgradeMainConfig(tc.old)
			assert.Equal(t, tc.expect.String(), got.String())

			t.Logf("%s", got.String())

			c := config.DefaultConfig()
			if _, err := bstoml.Decode(got.String(), c); err != nil {
				t.Errorf("bstoml.Decode: %s", err)
			} else {
				assert.Equal(t, tc.expect.String(), c.String())
			}
		})
	}
}
