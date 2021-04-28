package datakit

import (
	"fmt"
	"testing"
	"time"

	t2 "github.com/BurntSushi/toml"

	"github.com/influxdata/toml"
	//"github.com/kardianos/service"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParseDataWay(t *testing.T) {

	cases := []struct {
		url  string
		fail bool
	}{

		{
			url: "http://preprod-openway.cloudcare.cn?token=123&a=b&d=e&c=123_456;https://preprod-openway.cloudcare2.cn?token=456&a=b&d=e&c=123_456",
		},
		// {
		// 	url: "https://preprod-openway.cloudcare.cn?token=123&a=b&d=e&c=123_456",
		// },

		// {
		// 	url: "http://preprod-openway.cloudcare.cn:80?token=123&a=b&d=e&c=123_456",
		// },
		// {
		// 	url: "https://preprod-openway.cloudcare.cn:443?token=123&a=b&d=e&c=123_456",
		// },
		//
		//		{ // dial timeout
		//			url:  "http://1.2.3?token=123&a=b&d=e&c=123_456",
		//			fail: true,
		//		},

		// {
		// 	url:  "",
		// 	fail: true,
		// }, // empty dataway url
	}

	for idx, tc := range cases {

		dw, err := ParseDataway(tc.url)

		strUrls := dw.MetricURL()

		fmt.Println("=====>", strUrls)

		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Ok(t, err)
			t.Logf("[%d] %+#v", idx, dw)
		}
	}
}

func TestUnmarshalMainCfg(t *testing.T) {
	x := `
uuid = "dkid_bt6um9bksvvesmk79370"
name = "nat-datakit"
http_server_addr = "0.0.0.0:9529"
log = "/usr/local/cloudcare/dataflux/datakit/datakit.log"
log_level = "debug"
log_rotate = 1
log_upload = true
gin_log = "/usr/local/cloudcare/dataflux/datakit/gin.log"
max_post_interval = "15s"
round_interval = false
interval = "10s"
strict_mode = true
output_file = ""
default_enabled_inputs = ["cpu", "mem", "disk", "diskio", "processes", "timezone"]
install_date = 2020-09-01T06:33:09.075190972Z

[dataway]
host = "10.100.64.117:49527"
scheme = "http"
token = "__internal__"
timeout = "30s"
default_path = "/v1/write/metrics"

[global_tags]
from = "$datakit_hostname"
id = "$datakit_id"

#[dataway]
#        host = "openway.dataflux.cn:443"
#        scheme = "https"
#        token = "tkn_d24c479141bc4a6da4596f5ea5391097"
#        default_path = "/v1/write/metrics"

[agent]
interval = "10s"
round_interval = true
precision = "ns"
collection_jitter = "0s"
flush_interval = "10s"
flush_jitter = "0s"
metric_batch_size = 1000
metric_buffer_limit = 100000
utc = false
debug = false
quiet = false
logtarget = "file"
logfile = "/usr/local/cloudcare/dataflux/datakit/embed/agent.log"
logfile_rotation_interval = ""
	`

	var mc MainConfig
	if md, err := t2.Decode((x), &mc); err != nil {
		_ = md
		t.Fatal(err)
	} else {
		t.Logf("md.Undecoded: %+#v, mc: %+#v", md.Undecoded(), mc)
		t.Logf("%+#v", mc.DataWay)
	}

	data, err := TomlMarshal(&mc)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}

func TestMarshalMainCfg(t *testing.T) {

	if Cfg.MainCfg.Hostname == "" {
		Cfg.setHostname()
	}

	data, err := toml.Marshal(Cfg.MainCfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(data))

	data, err = TomlMarshal(Cfg.MainCfg)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(data)
	}
}

func TestLocalIP(t *testing.T) {
	ip, err := LocalIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("IP: %s", ip)
}

func TestGetFirstGlobalUnicastIP(t *testing.T) {
	ip, err := GetFirstGlobalUnicastIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("IP: %s", ip)
}

func TestProtectedInterval(t *testing.T) {
	cases := []struct {
		min, max, cur, expect time.Duration
		fail, unprotected     bool
	}{
		{
			min: time.Minute,
			max: time.Minute * 5,
			cur: time.Second,

			expect: time.Minute,
		},

		{
			min: time.Minute,
			max: time.Minute * 5,
			cur: time.Hour,

			expect: time.Minute * 5,
		},

		{
			min: time.Minute,
			max: time.Minute * 5,
			cur: time.Minute,

			expect: time.Minute * 5,
			fail:   true,
		},

		{
			min: time.Minute,
			max: time.Minute * 5,
			cur: time.Second,

			expect:      time.Second,
			unprotected: true,
		},
	}

	for _, c := range cases {
		Cfg.MainCfg.ProtectMode = !c.unprotected

		du := ProtectedInterval(c.min, c.max, c.cur)
		if !c.fail {
			tu.Assert(t, c.expect == du, "exp: %v, get %v", c.expect, du)
		} else {
			tu.Assert(t, c.expect != du, "exp: %v, get %v", c.expect, du)
		}
	}
}

/*
func TestServiceInstall(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatal(err)
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		t.Log(err)
	}

	if err := service.Control(svc, "install"); err != nil {
		t.Fatal(err)
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		t.Fatal(err)
	}
} */
