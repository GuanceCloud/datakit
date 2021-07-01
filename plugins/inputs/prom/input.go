package prom

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName   = "prom"
	catalogName = "prom"

	sampleCfg = `
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9100/metrics"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  # metric_name_filter = ["cpu"]

  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = ""

  ## 指标集名称
  # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  # 如果配置measurement_name, 则不进行指标名称的切割
  # 最终的指标集名称会添加上measurement_prefix前缀
  # measurement_name = "prom"

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

	## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

	## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  #[[inputs.prom.measurements]]
  #  prefix = "cpu_"
  #  name = "cpu"

  # [[inputs.prom.measurements]]
  # prefix = "mem_"
  # name = "mem"

	## 自定义认证方式，目前仅支持 Bearer Token
	# [inputs.prom.auth]
	# type = "bearer_token"
	# token = "xxxxxxxx"
	# token_file = "/tmp/token"

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: inputName}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

type Rule struct {
	Pattern string `toml:"pattern"`
	Prefix  string `toml:"prefix"`
	Name    string `toml:"name"`
}

type Input struct {
	URL               string   `toml:"url"`
	MetricTypes       []string `toml:"metric_types"`
	MetricNameFilter  []string `toml:"metric_name_filter"`
	MeasurementPrefix string   `toml:"measurement_prefix"`
	MeasurementName   string   `toml:"measurement_name"`
	Measurements      []Rule   `toml:"measurements"`

	Interval string `toml:"interval"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	Tags map[string]string `toml:"tags"`
	Auth map[string]string `toml:"auth"`

	SampleCfg string

	client       *http.Client
	duration     time.Duration
	collectTime  time.Time
	collectCache []inputs.Measurement

	chPause chan bool
	pause   bool
}

func (i *Input) SampleConfig() string {
	return i.SampleCfg
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (i *Input) Catalog() string {
	return catalogName

}

func (i *Input) extendSelfTag(tags map[string]string) {
	if i.Tags != nil {
		for k, v := range i.Tags {
			tags[k] = v
		}
	}
}

func (i *Input) GetReq(url string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if len(i.Auth) > 0 {
		authType := i.Auth["type"]
		if authFunc, ok := AuthMaps[authType]; ok {
			req, err = authFunc(i.Auth, url)
		} else {
			req, err = http.NewRequest("GET", url, nil)
		}
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}
	return req, err
}

func (i *Input) Collect() error {
	req, err := i.GetReq(i.URL)
	if err != nil {
		return err
	}
	r, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	measurements, err := PromText2Metrics(r.Body, i)
	if err != nil {
		return err
	}

	i.collectCache = append(i.collectCache, measurements...)

	return nil
}

func (i *Input) GetCachedPoints() []*io.Point {
	points := []*io.Point{}
	for _, m := range i.collectCache {
		point, err := m.LineProto()
		if err != nil {
			l.Warn("invalid measurement")
		} else {
			points = append(points, point)
		}
	}

	return points
}

const (
	maxInterval = 10 * time.Minute
	minInterval = 1 * time.Second
)

func (i *Input) InitClient() error {
	client, err := i.createHTTPClient()
	if err != nil {
		return err
	}
	i.client = client
	return nil
}

func (i *Input) SetClient(client *http.Client) {
	i.client = client
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	duration, err := time.ParseDuration(i.Interval)

	if err != nil {
		l.Error(fmt.Errorf("invalid interval, %s", err.Error()))
		return
	} else if duration <= 0 {
		l.Error(fmt.Errorf("invalid interval, cannot be less than zero"))
		return
	}

	i.duration = config.ProtectedInterval(minInterval, maxInterval, duration)

	err = i.InitClient()
	if err != nil {
		l.Error(err.Error())
	}

	defer i.stop()

	tick := time.NewTicker(i.duration)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("prom exit")
			return

		case <-tick.C:
			if i.pause {
				continue
			}

			start := time.Now()
			if err := i.Collect(); err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			} else {
				if len(i.collectCache) > 0 {
					continue
				}

				err := inputs.FeedMeasurement(inputName,
					datakit.Metric,
					i.collectCache,
					&io.Option{CollectCost: time.Since(start)})
				if err != nil {
					io.FeedLastError(inputName, err.Error())
					l.Errorf(err.Error())
				}
				i.collectCache = i.collectCache[:0]
			}
		}
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) stop() {
	i.client.CloseIdleConnections()
}

func (i *Input) createHTTPClient() (*http.Client, error) {
	if i.client != nil {
		return i.client, nil
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if i.TLSOpen {
		tc, err := TLSConfig(i.CacertFile, i.CertFile, i.KeyFile)
		if err != nil {
			return nil, err
		} else {
			client.Transport = &http.Transport{
				TLSClientConfig: tc,
			}
		}
	}
	return client, nil
}

func NewProm(sampleCfg string) *Input {
	return &Input{
		SampleCfg: sampleCfg,
		chPause:   make(chan bool, 1),
	}
}

func init() {
	inputs.Add("prom", func() inputs.Input {
		return NewProm(sampleCfg)
	})
}
