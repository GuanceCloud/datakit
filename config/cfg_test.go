// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestInitCfg(t *testing.T) {
	c := DefaultConfig()

	tomlfile := ".main.toml"
	defer os.Remove(tomlfile) //nolint:errcheck
	tu.Equals(t, nil, c.InitCfg(tomlfile))
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
		tu.Equals(t, len(c.DefaultEnabledInputs), len(tc.expect))
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
				c.GlobalEnvTags[k] = v
			}

			for k, v := range tc.deprecatedGlobalTags {
				c.GlobalTagsDeprecated[k] = v
			}

			c.EnableElection = tc.election
			c.EnableElectionTag = tc.electionTag
			c.setupGlobalTags()

			// 这些预期的 tags 在 config 中必须存在
			for k, v := range tc.expectEnvTags {
				tu.Assert(t, v == c.GlobalEnvTags[k], "[%s]`%s' != `%s'", k, v, c.GlobalEnvTags[k])
			}

			for k, v := range tc.expectHostTags {
				tu.Assert(t, v == c.GlobalHostTags[k], "[%s]`%s' != `%s'", k, v, c.GlobalHostTags[k])
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
		tu.Equals(t, x, tc.expect)
	}
}

func TestDefaultToml(t *testing.T) {
	c := DefaultConfig()

	buf := new(bytes.Buffer)
	if err := bstoml.NewEncoder(buf).Encode(c); err != nil {
		l.Fatalf("encode main configure failed: %s", err.Error())
	}

	t.Logf("%s", buf.String())
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
			raw:  DatakitConfSample,
		},
	}

	tomlfile := ".main.toml"

	defer func() {
		os.Remove(tomlfile) //nolint:errcheck
	}()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()

			if err := ioutil.WriteFile(tomlfile, []byte(tc.raw), 0o600); err != nil {
				t.Fatal(err)
			}

			err := c.LoadMainTOML(tomlfile)
			if tc.fail {
				tu.NotOk(t, err, "")
				return
			} else {
				tu.Ok(t, err)
			}

			t.Logf("hostname: %s", c.Hostname)

			if err := os.Remove(tomlfile); err != nil {
				t.Error(err)
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestWriteConfigFile$ gitlab.jiagouyun.com/cloudcare-tools/datakit/config
/*
[sinks]

  [[sinks.sink]]
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S"]
    database = "db0"
    host = "1.1.1.1:8086"
    precision = "ns"
    protocol = "http"
    target = "influxdb"
    timeout = "6s"

  [[sinks.sink]]
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S"]
    database = "db1"
    host = "1.1.1.1:8087"
    precision = "ns"
    protocol = "http"
    target = "influxdb"
    timeout = "6s"

[sinks]

  [[sinks.sink]]
*/
func TestWriteConfigFile(t *testing.T) {
	c := DefaultConfig()

	cases := []struct {
		name string
		in   []map[string]interface{}
	}{
		{
			name: "has_data",
			in: []map[string]interface{}{
				{
					"target":     "influxdb",
					"categories": []string{"M", "N", "K", "O", "CO", "L", "T", "R", "S"},
					"host":       "1.1.1.1:8086",
					"protocol":   "http",
					"precision":  "ns",
					"database":   "db0",
					"timeout":    "5s",
				},
				{
					"target":       "logstash",
					"categories":   []string{"L"},
					"host":         "1.1.1.1:8080",
					"protocol":     "http",
					"request_path": "/twitter/tweet/1",
					"timeout":      "5s",
				},
			},
		},
		{
			name: "no_data",
			in: []map[string]interface{}{
				{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c.Sinks.Sink = tc.in
			mcdata, err := datakit.TomlMarshal(c)
			if err != nil {
				panic(err)
			}
			fmt.Println("=====================================================")
			fmt.Println(string(mcdata))
		})
	}
}

func TestSetupDataway(t *testing.T) {
	cases := []struct {
		name   string
		dwcfg  *dataway.DataWayCfg
		expect error
	}{
		{
			name: "check_dev_null",
			dwcfg: &dataway.DataWayCfg{
				URLs: []string{datakit.DatawayDisableURL},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := DefaultConfig()
			c.DataWayCfg = tc.dwcfg

			err := c.setupDataway()
			assert.Equal(t, tc.expect, err)
		})
	}
}
