package tidb

import (
	"bytes"
	"encoding/json"
	"fmt"
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

    [inputs.tidb.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

// 仅通过 API 采集 PD Server 指标，不对整个 TiDB 集群做采集
// 指标数量少，功能过于简单，需要后续开发 v2

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &TiDB{
			Interval: datakit.Cfg.IntervalDeprecated,
			Tags:     make(map[string]string),
		}
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

	if t.initCfg() {
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
			if err := io.NamedFeed(data, datakit.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (t *TiDB) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := t.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return false
}

func (t *TiDB) loadCfg() (err error) {
	t.duration, err = time.ParseDuration(t.Interval)
	if err != nil {
		err = fmt.Errorf("invalid interval, %s", err.Error())
		return
	} else if t.duration <= 0 {
		err = fmt.Errorf("invalid interval, cannot be less than zero")
		return
	}
	return
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
