package datakit

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	bstoml "github.com/BurntSushi/toml"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestInitCfg(t *testing.T) {
	c := DefaultConfig()

	tomlfile := ".main.toml"
	defer os.Remove(tomlfile)
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
	cases := []struct {
		tags   map[string]string
		expect map[string]string
		fail   bool
	}{
		{
			tags: map[string]string{
				"host": "__datakit_hostname",
				"ip":   "__datakit_ip",
				"id":   "__datakit_id",
				"uuid": "__datakit_uuid",
			},
		},

		{
			tags: map[string]string{
				"host": "$datakit_hostname",
				"ip":   "$datakit_ip",
				"id":   "$datakit_id",
				"uuid": "$datakit_uuid",
			},
		},

		{
			tags: map[string]string{
				"uuid": "some-uuid",
				"host": "some-host",
			},
			expect: map[string]string{},
		},
	}

	for _, tc := range cases {
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

		for k, v := range tc.tags {
			if tc.expect == nil {
				tu.Assert(t, v != c.GlobalTags[k], "`%s' != `%s'", v, c.GlobalTags[k])
			} else {
				tu.Assert(t, v == c.GlobalTags[k], "`%s' != `%s'", v, c.GlobalTags[k])
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

	t.Logf("%s", string(buf.Bytes()))
}

func TestLoadEnv(t *testing.T) {

	cases := []struct {
		envs   map[string]string
		expect *Config
	}{
		{
			envs: map[string]string{
				"ENV_GLOBAL_TAGS":            "a=b,c=d",
				"ENV_LOG_LEVEL":              "debug",
				"ENV_DATAWAY":                "http://host1.org,http://host2.com",
				"ENV_HOSTNAME":               "1024.coding",
				"ENV_NAME":                   "testing-datakit",
				"ENV_HTTP_LISTEN":            "localhost:9559",
				"ENV_RUM_ORIGIN_IP_HEADER":   "not-set",
				"ENV_ENABLE_PPROF":           "true",
				"ENV_DISABLE_PROTECT_MODE":   "true",
				"ENV_DEFAULT_ENABLED_INPUTS": "cpu,mem,disk",
				"ENV_ENABLE_ELECTION":        "1",
			},
			expect: &Config{
				Name:                 "testing-datakit",
				DataWay:              &DataWayCfg{URLs: []string{"http://host1.org", "http://host2.com"}},
				HTTPListen:           "localhost:9559",
				HTTPAPI:              &apiConfig{RUMOriginIPHeader: "not-set"},
				LogLevel:             "debug",
				EnablePProf:          true,
				Hostname:             "1024.coding",
				ProtectMode:          false,
				DefaultEnabledInputs: []string{"cpu", "mem", "disk"},
				EnableElection:       true,
				GlobalTags: map[string]string{
					"a": "b", "c": "d",
				},
			},
		},
		{
			envs: map[string]string{
				"ENV_ENABLE_INPUTS": "cpu,mem,disk",
			},
			expect: &Config{
				DefaultEnabledInputs: []string{"cpu", "mem", "disk"},
			},
		},

		{
			envs: map[string]string{
				"ENV_GLOBAL_TAGS": "cpu,mem,disk=sda",
			},
			expect: &Config{
				GlobalTags: map[string]string{"disk": "sda"},
			},
		},
	}

	for _, tc := range cases {
		c := &Config{}
		os.Clearenv()
		for k, v := range tc.envs {
			os.Setenv(k, v)
		}
		c.LoadEnvs()
		tu.Equals(t, tc.expect.String(), c.String())
	}
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
		os.Remove(tomlfile)
	}()

	for _, tc := range cases {

		c := DefaultConfig()

		if err := ioutil.WriteFile(tomlfile, []byte(tc.raw), os.ModePerm); err != nil {
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

		os.Remove(tomlfile)
	}
}
