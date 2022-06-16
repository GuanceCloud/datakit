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
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

	cases := []struct {
		tags   map[string]string
		expect map[string]string
		fail   bool
	}{
		{
			tags: map[string]string{
				"host": "__datakit_hostname", // ENV host dropped during setup
				"ip":   "__datakit_ip",
				"id":   "__datakit_id",
				"uuid": "__datakit_uuid",
			},
			expect: map[string]string{
				"ip": localIP,
			},
		},

		{
			tags: map[string]string{
				"host": "$datakit_hostname", // ENV host dropped during setup
				"ip":   "$datakit_ip",
				"id":   "$datakit_id",
				"uuid": "$datakit_uuid",
			},
			expect: map[string]string{
				"ip": localIP,
			},
		},

		{
			tags: map[string]string{
				"uuid": "some-uuid",
				"host": "some-host", // ENV host dropped during setup
			},

			expect: map[string]string{
				"uuid": "some-uuid",
			},
		},
	}

	for idx, tc := range cases {
		c := DefaultConfig()
		for k, v := range tc.tags {
			c.GlobalTags[k] = v
		}

		err := c.setupGlobalTags()
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
		}

		for k, v := range c.GlobalTags {
			if tc.expect == nil {
				tu.Assert(t, v != tc.expect[k], "[case %d] `%s' != `%s', global tags: %+#v", idx, v, tc.expect[k], c.GlobalTags)
			} else {
				tu.Assert(t, v == tc.expect[k], "[case %d] `%s' != `%s', global tags: %+#v", idx, v, tc.expect[k], c.GlobalTags)
			}
		}
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
		raw  string
		fail bool
	}{
		{
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
			raw:  `abc = def`, // invalid toml
			fail: true,
		},

		{
			raw: `
name = "not-set"
http_listen=123  # invalid type
log = "log"`,
			fail: true,
		},

		{
			raw: `
name = "not-set"
log = "log"`,
			fail: false,
		},

		{
			raw: `
hostname = "should-not-set"`,
		},
	}

	tomlfile := ".main.toml"

	defer func() {
		os.Remove(tomlfile) //nolint:errcheck
	}()

	for _, tc := range cases {
		c := DefaultConfig()

		if err := ioutil.WriteFile(tomlfile, []byte(tc.raw), 0o600); err != nil {
			t.Fatal(err)
		}

		err := c.LoadMainTOML(tomlfile)
		if tc.fail {
			tu.NotOk(t, err, "")
			continue
		} else {
			tu.Ok(t, err)
		}

		t.Logf("hostname: %s", c.Hostname)

		if err := os.Remove(tomlfile); err != nil {
			t.Error(err)
		}
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

func TestGetElectionNamespace(t *testing.T) {
	Cfg = DefaultConfig()
	tu.Equals(t, GetElectionNamespace(), "")
	Cfg.Namespace = "test"
	tu.Equals(t, GetElectionNamespace(), "")

	Cfg.EnableElection = true
	tu.Equals(t, GetElectionNamespace(), "test")

	Cfg.Namespace = ""
	tu.Equals(t, GetElectionNamespace(), "default")
}
