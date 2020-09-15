package tidb

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tidb"

	sampleCfg = `
[[inputs.tidb]]
    # tidb PD Server metrics from http://HOST:PORT/pd/api/v1/stores
    # usually modify host and port
    # required
    pd_url = ["http://127.0.0.1:2379/pd/api/v1/stores"]
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.tidb.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

// 仅通过 API 采集 PD Server 指标，不对整个 TiDB 集群做采集
// 指标数量少，功能过于简单，需要后续开发 v2

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TiDB{}
	})
}

type TiDB struct {
	PDServerURL []string          `toml:"pd_url"`
	Interval    string            `toml:"interval"`
	Tags        map[string]string `toml:"tags"`

	duration time.Duration
}

func (*TiDB) SampleConfig() string {
	return sampleCfg
}

func (*TiDB) Catalog() string {
	return "db"
}

func (t *TiDB) Run() {
	l = logger.SLogger(inputName)

	if t.loadcfg() {
		return
	}
	ticker := time.NewTicker(t.duration)
	defer ticker.Stop()

	l.Infof("tidb input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := t.getMetrics()
			if err != nil {
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (t *TiDB) loadcfg() bool {
	var err error
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		t.duration, err = time.ParseDuration(t.Interval)
		if err != nil || t.duration <= 0 {
			l.Errorf("invalid interval, err %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return false
}

func (t *TiDB) getMetrics() ([]byte, error) {
	var buffer = bytes.Buffer{}
	for _, url := range t.PDServerURL {
		resp, err := http.Get(url)
		if err != nil {
			l.Errorf("http get metrics, %s", err)
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Error("read body err, %s", err)
			return nil, err
		}

		var pd PDStores
		if err := json.Unmarshal(body, &pd); err != nil {
			l.Error(err)
			return nil, err
		}
		buffer.Write(pd.Metrics(inputName, t.Tags, time.Now()))
	}
	return buffer.Bytes(), nil
}
