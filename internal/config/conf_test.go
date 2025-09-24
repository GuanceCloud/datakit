// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	T "testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

// TestDefaultMainConf used to keep default config and default config sample equal.
func TestDefaultMainConf(t *testing.T) {
	c := DefaultConfig()
	c.Ulimit = 0 // ulimit diff among OS platforms

	x := DefaultConfig()
	_, err := bstoml.Decode(datakit.MainConfSample(datakit.BrandDomain), &x)
	require.NoError(t, err)

	x.DefaultEnabledInputs = x.DefaultEnabledInputs[:0] // clear
	x.GlobalHostTags = map[string]string{}              // clear:  host tags setted on default conf sample
	x.Ulimit = 0

	assert.Equal(t, c.String(), x.String())
}

func TestEnableDefaultsInputs(t *testing.T) {
	cases := []struct {
		list   string
		expect []string
	}{
		{
			list:   "a,a,b,c,d",
			expect: []string{"a", "b", "c", "d"},
		},

		{
			list:   "a,b,c,d",
			expect: []string{"a", "b", "c", "d"},
		},
	}

	c := DefaultConfig()
	for _, tc := range cases {
		c.EnableDefaultsInputs(tc.list)
		assert.Equal(t, len(c.DefaultEnabledInputs), len(tc.expect))
	}
}

func TestSetupGlobalTags(t *testing.T) {
	localIP, err := datakit.LocalIP()
	if err != nil {
		t.Fatal(err)
	}

	hn, err := os.Hostname()
	if err != nil {
		t.Fatalf("get hostname failed: %s", err.Error())
	}

	cases := []struct {
		name string

		hosttags             map[string]string
		envtags              map[string]string
		deprecatedGlobalTags map[string]string

		election, electionTag bool

		expectHostTags,
		expectEnvTags map[string]string
	}{
		{
			name: "mixed-host-and-evn-tags",
			hosttags: map[string]string{
				"ip":   "__datakit_ip",
				"host": "__datakit_hostname",

				// 此处 `__datakit_id` and `__datakit_uuid` 都被设置为 `host`
				// 即不允许出现 xxx = "__datakit_id" 这种 tag
				"some_id":   "__datakit_id",
				"some_uuid": "__datakit_uuid",

				// 但可以额外直接给一个 xxx = __datakit_hostname 这样的 tag
				"xxx": "__datakit_hostname",
			},

			envtags:        map[string]string{"cluster": "my-cluster"},
			expectHostTags: map[string]string{"ip": localIP, "host": hn, "xxx": hn},
			expectEnvTags:  map[string]string{"cluster": "my-cluster"},
		},

		{
			name:           "only-host-tags",
			hosttags:       map[string]string{"uuid": "some-uuid", "host": "some-host"},
			expectHostTags: map[string]string{"uuid": "some-uuid", "host": "some-host"},
		},

		{
			name:                 "host-tags-deprecated",
			hosttags:             map[string]string{"uuid": "some-uuid", "host": "some-host"},
			deprecatedGlobalTags: map[string]string{"tag1": "val1", "tag2": "val2"},

			expectHostTags: map[string]string{
				"uuid": "some-uuid",
				"host": "some-host",
				"tag1": "val1",
				"tag2": "val2",
			},
		},

		{
			name:          "only-env-tags",
			envtags:       map[string]string{"cluster": "my-cluster"},
			expectEnvTags: map[string]string{"cluster": "my-cluster"},
		},

		{
			name:        "enable-only-election",
			election:    true,
			electionTag: false,

			hosttags:       map[string]string{"uuid": "some-uuid", "host": "some-host"},
			envtags:        map[string]string{"cluster": "my-cluster"},
			expectEnvTags:  map[string]string{"cluster": "my-cluster"},
			expectHostTags: map[string]string{"uuid": "some-uuid", "host": "some-host"},
		},

		{
			name:        "enable-election-and-tags",
			election:    true,
			electionTag: true,

			hosttags:       map[string]string{"uuid": "some-uuid", "host": "some-host"},
			envtags:        map[string]string{"cluster": "my-cluster"},
			expectEnvTags:  map[string]string{"election_namespace": "default", "cluster": "my-cluster"},
			expectHostTags: map[string]string{"uuid": "some-uuid", "host": "some-host"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()
			for k, v := range tc.hosttags {
				c.GlobalHostTags[k] = v
			}

			for k, v := range tc.envtags {
				c.Election.Tags[k] = v
			}

			for k, v := range tc.deprecatedGlobalTags {
				c.GlobalTagsDeprecated[k] = v
			}

			c.Election.Enable = tc.election
			c.Election.EnableNamespaceTag = tc.electionTag
			c.setupGlobalTags()

			// 这些预期的 tags 在 config 中必须存在
			for k, v := range tc.expectEnvTags {
				assert.Truef(t, v == c.Election.Tags[k], "[%s]`%s' != `%s'", k, v, c.Election.Tags[k])
			}

			for k, v := range tc.expectHostTags {
				assert.Truef(t, v == c.GlobalHostTags[k], "[%s]`%s' != `%s'", k, v, c.GlobalHostTags[k])
			}
		})
	}
}

func TestProtectedInterval(t *testing.T) {
	cases := []struct {
		enabled              bool
		min, max, in, expect time.Duration
	}{
		{
			enabled: true,
			min:     time.Minute,
			max:     5 * time.Minute,
			in:      time.Second,
			expect:  time.Minute,
		},

		{
			enabled: true,
			min:     time.Minute,
			max:     5 * time.Minute,
			in:      10 * time.Minute,
			expect:  5 * time.Minute,
		},

		{
			enabled: false,
			min:     time.Minute,
			max:     5 * time.Minute,
			in:      time.Second,
			expect:  time.Second,
		},

		{
			enabled: false,
			min:     time.Minute,
			max:     5 * time.Minute,
			in:      10 * time.Minute,
			expect:  10 * time.Minute,
		},
	}

	for _, tc := range cases {
		Cfg.ProtectMode = tc.enabled
		x := ProtectedInterval(tc.min, tc.max, tc.in)
		assert.Equal(t, x, tc.expect)
	}
}

func TestUnmarshalCfg(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		fail bool
	}{
		{
			name: "real-conf",
			raw: `
	name = "not-set"
http_listen="0.0.0.0:9529"
log = "log"
log_level = "debug"
gin_log = "gin.log"
interval = "10s"
output_file = "out.data"
hostname = "iZb.1024"
default_enabled_inputs = ["cpu", "disk", "diskio", "mem", "swap", "system", "net", "hostobject"]
install_date = 2021-03-25T11:00:19Z

[dataway]
  urls = ["http://testing-openway.cloudcare.cn?token=tkn_2dc4xxxxxxxxxxxxxxxxxxxxxxxxxxxx"]
  timeout = "30s"

[global_tags]
  cluster = ""
  global_test_tag = "global_test_tag_value"
  host = "__datakit_hostname"
  project = ""
  site = ""
  lg= "tl"

[[black_lists]]
  hosts = []
  inputs = []

[[white_lists]]
  hosts = []
  inputs = []
	`,
		},

		{
			name: "invalid-toml",
			raw:  `abc = def`, // bad toml
			fail: true,
		},

		{
			name: "invalid-type",
			raw: `
name = "not-set"
http_listen=123  # invalid type
log = "log"`,
			fail: true,
		},

		{
			name: "partial-ok",
			raw: `
name = "not-set"
log = "log"`,
			fail: false,
		},

		{
			name: "partial-ok-2",
			raw: `
hostname = "should-not-set"`,
		},
		{
			name: "dk-conf-sample",
			raw:  datakit.MainConfSample(""),
		},
	}

	tomlfile := ".main.toml"

	defer func() {
		os.Remove(tomlfile) //nolint:errcheck
	}()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()

			if err := os.WriteFile(tomlfile, []byte(tc.raw), 0o600); err != nil {
				t.Fatal(err)
			}

			err := c.LoadMainTOML(tomlfile)
			if tc.fail {
				require.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			t.Logf("hostname: %s", c.hostname)

			if err := os.Remove(tomlfile); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLoadDWRetry(t *testing.T) {
	type Case struct {
		name       string
		confText   string
		retryCount int
		retryDelay time.Duration
	}

	testCases := []Case{
		{
			name: "defaultSetting",
			confText: `
[dataway]
  urls = ["http://testing-openway.cloudcare.cn?token=tkn_2dc4xxxxxxxxxxxxxxxxxxxxxxxxxxxx"]
  timeout = "30s"
`,
			retryCount: dataway.DefaultRetryCount,
			retryDelay: dataway.DefaultRetryDelay,
		},
		{
			name: "manualSetting",
			confText: `
[dataway]
  urls = ["http://testing-openway.cloudcare.cn?token=tkn_2dc4xxxxxxxxxxxxxxxxxxxxxxxxxxxx"]
  timeout = "30s"
  max_retry_count = 7
  retry_delay = "2s"
`,
			retryCount: 7,
			retryDelay: time.Second * 2,
		},
		{
			name: "zeroDelay",
			confText: `
[dataway]
  urls = ["http://testing-openway.cloudcare.cn?token=tkn_2dc4xxxxxxxxxxxxxxxxxxxxxxxxxxxx"]
  timeout = "30s"
  retry_delay = "0s"
`,
			retryCount: dataway.DefaultRetryCount,
			retryDelay: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("./", "test.main.*.conf")
			assert.NoError(t, err)

			defer os.Remove(f.Name())

			_, err = io.WriteString(f, tc.confText)
			assert.NoError(t, err)

			err = f.Close()
			assert.NoError(t, err)

			c := DefaultConfig()
			err = c.LoadMainTOML(f.Name())
			assert.NoError(t, err)

			assert.Equal(t, tc.retryCount, c.Dataway.MaxRetryCount)
			assert.Equal(t, tc.retryDelay, c.Dataway.RetryDelay)
		})
	}
}

func TestLoadResourceLimite(t *T.T) {
	t.Run(`default`, func(t *T.T) {
		conf := `
[resource_limit]
  path = "/datakit"
  cpu_max = 10.0
	cpu_cores = 2.3
  mem_max_mb = 4096
  enable = true
		`

		c := DefaultConfig()

		_, err := bstoml.Decode(conf, c)
		assert.NoError(t, err)
		assert.Equal(t, 10.0, c.ResourceLimitOptions.CPUMax)
	})
}

func Test_setupDataway(t *testing.T) {
	cases := []struct {
		name   string
		dw     *dataway.Dataway
		expect error
	}{
		{
			name: "check_dev_null",
			dw: func() *dataway.Dataway {
				x := dataway.NewDefaultDataway()
				x.URLs = []string{datakit.DatawayDisableURL}
				return x
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()
			c.Dataway = tc.dw

			err := c.setupDataway()
			assert.Equal(t, tc.expect, err)
		})
	}
}

func TestTryUpgradeCfg(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		pwd := t.TempDir()
		oldConfFile := filepath.Join(pwd, "datakit.conf")
		oldConf := DefaultConfig()
		// change some config items
		oldConf.HTTPAPI.Listen = "localhost:1234"

		assert.NoError(t, os.WriteFile(oldConfFile, []byte(oldConf.String()), datakit.ConfPerm))

		newConf := DefaultConfig()
		backup := oldConfFile + ".old"
		newConf.TryUpgradeCfg(oldConfFile, backup)

		_, err := os.Stat(backup)
		assert.NoError(t, err)

		oldConf = &Config{}
		assert.NoError(t, oldConf.LoadMainTOML(backup))
		assert.Equal(t, "localhost:1234", oldConf.HTTPAPI.Listen)
	})

	t.Run(`backup-conf-with-comments`, func(t *T.T) {
		pwd := t.TempDir()
		oldConfStr := `
################################################
# Global configures
################################################
# Default enabled input list.
default_enabled_inputs = [
  "cpu",
  "disk",
  "diskio",
  "host_processes",
  "hostobject",
  "mem",
  "net",
  "swap",
  "system",
	"fake-input",
]`

		// prepare old datakit.conf
		confPath := filepath.Join(pwd, "datakit.conf")
		assert.NoError(t, os.WriteFile(confPath, []byte(oldConfStr), datakit.ConfPerm))

		backup := confPath + ".old"

		// new version of config
		newConf := DefaultConfig()
		// upgrade old datakit.conf and backup it if required
		assert.NoError(t, newConf.TryUpgradeCfg(confPath, backup))
		assert.Empty(t, newConf.DefaultEnabledInputs) // new version datakit.conf should have no default inputs

		// reload backuped old datakit.conf
		oldConf := &Config{}
		assert.NoError(t, oldConf.LoadMainTOML(backup))

		// and we got default input list
		assert.NotEmpty(t, oldConf.DefaultEnabledInputs)
		assert.Contains(t, oldConf.DefaultEnabledInputs, "fake-input")

		oldConfBytes, err := os.ReadFile(backup)
		assert.NoError(t, err)
		assert.Contains(t, string(oldConfBytes), "# Default enabled input list.") // contains comments
	})

	t.Run(`with-comments-no-conf-items-changed`, func(t *T.T) {
		pwd := t.TempDir()
		oldConf := DefaultConfig()
		confPath := filepath.Join(pwd, "datakit.conf")

		// add comments in datakit.conf
		oldConfBytes := "# some header comments\n" + oldConf.String() + "\n# some tail comments"
		assert.NoError(t, os.WriteFile(confPath, []byte(oldConfBytes), datakit.ConfPerm))

		newConf := DefaultConfig()
		backup := filepath.Join(pwd, "datakit.conf.old")
		assert.NoError(t, newConf.TryUpgradeCfg(confPath, backup))

		assert.NoFileExists(t, backup) // no back

		// make sure file contains comments
		newConfBytes, err := os.ReadFile(confPath)
		assert.NoError(t, err)
		assert.Contains(t, string(newConfBytes), "# some header comments")
		assert.Contains(t, string(newConfBytes), "# some tail comments")

		t.Logf("%s", newConfBytes)
	})
}
