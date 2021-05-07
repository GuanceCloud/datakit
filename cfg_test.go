package datakit

import (
	"testing"
	// "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	// "github.com/influxdata/toml"
	//"github.com/kardianos/service"
)

var cfg = `
	name = ""
#http_server_addr = "0.0.0.0:9529"
http_listen="0.0.0.0:9529"
https_port = 443
inner_grpc_port = 0
interval_duration = 0
log = "/usr/local/cloudcare/dataflux/datakit/log"
log_level = "debug"
log_rotate = 32
gin_log = "/usr/local/cloudcare/dataflux/datakit/gin.log"
interval = "10s"
#output_file = "/usr/local/cloudcare/dataflux/datakit/mmm34.log"
hostname = "iZbp152ke14timzud0du15Z"
default_enabled_inputs = ["cpu", "disk", "diskio", "mem", "swap", "system", "net", "hostobject"]
install_date = 2021-03-25T11:00:19Z

[dataway]
  url = "http://testing-openway.cloudcare.cn/v1/write/metrics?token=tkn_2dc438b6693711eb8ff97aeee04b54af"
  timeout = "30s"

[global_tags]
  cluster = ""
  global_test_tag = "global_test_tag_value"
  host = "__datakit_hostname"
  project = ""
  site = ""
  lg= "tl"

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
  logfile_rotation_max_size = "32MB"
  logfile_rotation_max_archives = 5
  omit_hostname = true

[[black_lists]]
  hosts = []
  inputs = []

[[white_lists]]
  hosts = []
  inputs = []
	`

func TestUnmarshalMainCfg(t *testing.T) {
	Cfg.DoLoadMainConfig([]byte(cfg))

	t.Log(Cfg.MainCfg.DataWay.Urls)
}
