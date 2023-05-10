// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package installer

import (
	"testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestUpgradeMainConfig(t *testing.T) {
	cases := []struct {
		name string
		old,
		expect *config.Config
	}{
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
				c.Election = &config.ElectionCfg{
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
				c.Election = &config.ElectionCfg{
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
					Log:    "/some/path",
					GinLog: "/some/gin/log",
					Level:  "debug",
					Rotate: 128,
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
				c.OutputFileDeprecated = "/some/messy/file"
				c.IntervalDeprecated = "100s"

				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.IO.MaxCacheCount = 1000 // auto reset to 10000
				c.IO.OutputFile = "/some/messy/file"
				c.IO.FlushInterval = "100s"

				return c
			}(),
		},

		{
			name: `upgrade-sinker`,
			old: func() *config.Config {
				c := config.DefaultConfig()
				c.SinkersDeprecated = &config.SinkerDeprecated{Arr: []*dataway.Sinker{
					{
						Categories: []string{"M", "L"},
						Filters:    []string{},
						URL:        "some-dw-sinker",
					},
				}}
				return c
			}(),

			expect: func() *config.Config {
				c := config.DefaultConfig()
				c.Dataway.Sinkers = []*dataway.Sinker{
					{
						Categories: []string{"M", "L"},
						Filters:    []string{},
						URL:        "some-dw-sinker",
					},
				}

				t.Logf("c.dataway: %#v", c.Dataway)
				return c
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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
