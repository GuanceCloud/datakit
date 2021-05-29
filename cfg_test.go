package datakit

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	// "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	// "github.com/influxdata/toml"
	//"github.com/kardianos/service"

	bstoml "github.com/BurntSushi/toml"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestGenerateDatakitID(t *testing.T) {
	t.Logf("%s", GenerateDatakitID())
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
	}

	Docker = true

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

	var raw = `
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
	`

	id := "dkid_for_testing"

	if err := ioutil.WriteFile(".main.toml", []byte(raw), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(".id", []byte(id), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	defer func() {
		os.Remove(".id")
		os.Remove(".main.toml")
	}()

	c := DefaultConfig()
	if err := c.LoadMainTOML(".main.toml", ".id"); err != nil {
		t.Error(err)
	}
}

//func TestLoadEnv(t *testing.T) {
//	os.Setenv("ENV_ENABLE_INPUTS", "a,b,c,d")
//	os.Setenv("ENV_GLOBAL_TAGS", "a=b,c=d")
//	os.Setenv("ENV_LOG_LEVEL", "debug")
//	os.Setenv("ENV_LOG_LEVEL", "debug")
//	os.Setenv("ENV_UUID", "dkid_12345")
//	os.Setenv("ENV_DATAWAY", "https://openway.dataflux.cn?token=tkn_mocked")
//
//	Docker = true
//	UUIDFile = ".dk.id"
//	mcp := "mcp.conf"
//
//	os.Remove(UUIDFile)
//	os.Remove(mcp)
//
//	c := DefaultConfig()
//
//	if err := c.LoadEnvs(mcp); err != nil {
//		t.Error(err)
//	}
//}
