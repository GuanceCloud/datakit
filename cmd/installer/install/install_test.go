package install

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestUpdateLagacyConfig(_ *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	updateLagacyConfig("/usr/local/cloudcare/forethought/datakit")
}

func TestUpgradeMainConfigure(t *testing.T) {
	x := `
uuid = "dkid_bthjoijksvvcdnrlot0g"
name = "nat-datakit"
http_server_addr = "0.0.0.0:9529"
interval_duration = 0
log = "/usr/local/cloudcare/dataflux/datakit/datakit.log"
log_level = "debug"
log_rotate = 32
log_upload = false
gin_log = "/usr/local/cloudcare/dataflux/datakit/gin.log"
max_post_interval = "15s"
RoundInterval = false
interval = "10s"
output_file = ""
hostname = "ubt-server"
default_enabled_inputs = ["cpu", "mem", "disk", "diskio", "processes", "timezone"]
install_date = 2020-09-17T10:36:58Z

[dataway]
host = "10.100.64.117:49527"
token = "tkn_7cf6db7eb3224a88ab94870cbxxxxxxx"
scheme = "http"
timeout = "30s"
default_path = "/v1/write/metrics"

[global_tags]

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
  debug = true
  quiet = false
  logtarget = "file"
  logfile = "/usr/local/cloudcare/dataflux/datakit/embed/agent.log"
  logfile_rotation_interval = ""
  logfile_rotation_max_size = "32MB"
  logfile_rotation_max_archives = 5
  omit_hostname = true`

	cfg := datakit.DefaultConfig()
	mcp := "./dk.conf"
	if err := ioutil.WriteFile(mcp, []byte(x), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	if err := cfg.LoadMainConfig(mcp); err != nil {
		l.Fatalf("LoadMainConfig(): %s", err.Error())
	}
	l.Debugf("%+#v", cfg.MainCfg.DataWay)

	upgradeMainConfigure(cfg, mcp)

	if err := cfg.LoadMainConfig(mcp); err != nil {
		l.Fatalf("LoadMainConfig(): %s", err.Error())
	}

	l.Debugf("%+#v", cfg.MainCfg)
	l.Debugf("%+#v", cfg.MainCfg.DataWay)

	os.RemoveAll(mcp)
}
