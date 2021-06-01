package io

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
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
output_file = "/usr/local/cloudcare/dataflux/datakit/mmm.log"
hostname = "iZbp152ke14timzud0du15Z"
default_enabled_inputs = ["cpu", "disk", "diskio", "mem", "swap", "system", "net", "hostobject"]
install_date = 2021-03-25T11:00:19Z

[dataway]
  url = "https://openway.dataflux.cn?token=tkn_76d2d1efd3ff43db984497bfb4f3c25a"
  http_proxy = "http://127.0.0.1:8080"
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

var x *IO

func StartCollect() error {
	// l = logger.SLogger("dialtesting_io")
	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		x.StartIO(true)
	}()

	l.Debugf("io: %+#v", x)

	return nil
}

func FeedX(name, category string, pt *Point, opt *Option) error {
	pts := []*Point{}
	pts = append(pts, pt)

	return x.DoFeed(pts, category, name, opt)
}

func TestFlush(t *testing.T) {
	urls := []string{"https://openway.dataflux.cn/v1/write/metrics?token=tkn_76d2d1efd3ff43db984497bfb4f3c25a",
		"https://openway.dataflux.cn/v1/write/metrics?token=tkn_a5cbdacf23214966aa382ae0182e972b",
	}
	defaultMaxCacheCnt = int64(1024)
	x = NewIO(defaultMaxCacheCnt)
	x.dw, _ = datakit.ParseDataway(urls)
	x.FlushInterval = 1 * time.Second

	StartCollect()

	for i := 0; i < 100; i++ {
		l.Debug("loop index", i)
		time.Sleep(1 * time.Second)
		tags := map[string]string{
			"abc": "123",
		}

		fields := map[string]interface{}{
			"value1": 123,
			"value2": 234,
		}

		data, err := MakePoint("test", tags, fields, time.Now())
		if err != nil {
			l.Warnf("make metric failed: %s", err.Error)
		}

		FeedX("test", datakit.Metric, data, &Option{})
	}
}

func TestUnmarshalMainCfg(t *testing.T) {
	config.Cfg.DoLoadMainConfig([]byte(cfg))

	t.Log(config.Cfg.DataWay.HttpProxy)
}

func TestPushData(t *testing.T) {
	config.Cfg.DoLoadMainConfig([]byte(cfg))

	defaultMaxCacheCnt = int64(1024)
	x = NewIO(defaultMaxCacheCnt)
	x.dw, _ = datakit.ParseDataway(config.Cfg.DataWay.Urls)
	x.FlushInterval = 1 * time.Second

	StartCollect()

	for i := 0; i < 100; i++ {
		l.Debug("loop index", i)
		time.Sleep(1 * time.Second)
		tags := map[string]string{
			"abc": "123",
		}

		fields := map[string]interface{}{
			"value1": 123,
			"value2": 234,
		}

		data, err := MakePoint("test_proxy", tags, fields, time.Now())
		if err != nil {
			l.Warnf("make metric failed: %s", err.Error)
		}

		FeedX("test", datakit.Metric, data, &Option{})
	}
}
