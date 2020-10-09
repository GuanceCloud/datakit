package datakit

import (
	"testing"

	t2 "github.com/BurntSushi/toml"

	"github.com/influxdata/toml"
	"github.com/kardianos/service"
)

func TestParseDataWay(t *testing.T) {

	type tcase struct {
		url        string
		assertTrue bool
	}

	for _, url := range []*tcase{

		&tcase{url: "http://preprod-openway.cloudcare.cn/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "https://preprod-openway.cloudcare.cn/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "http://preprod-openway.cloudcare.cn:80/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "https://preprod-openway.cloudcare.cn:443/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "https://preprod-openway.cloudcare.cn:443/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "wss://preprod-openway.cloudcare.cn/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "ws://preprod-openway.cloudcare.cn/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: true},
		&tcase{url: "ws://1.2.3/v1/write/metrics?token=123&a=b&d=e&c=123_456", assertTrue: false}, // dial timeout
		&tcase{url: "", assertTrue: false}, // empty dataway url
	} {

		dw, err := ParseDataway(url.url)
		if err != nil {
			if url.assertTrue {
				t.Fatal(err)
			}
			t.Log(err)
		}

		if err := dw.Test(); err != nil {
			if url.assertTrue {
				t.Fatal(err)
			}
			t.Log(err)
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
}
