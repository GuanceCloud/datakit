// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package prom scrape prometheus exporter metrics.
package prom

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "prom"
	catalog   = "prom"
)

// defaultMaxFileSize is the default max response body size, in bytes.
// This field is used only when metrics are written to file, i.e. Output is configured.
// If the size of response body is over defaultMaxFileSize, metrics will be discarded.
// 32 MB.
const defaultMaxFileSize int64 = 32 * 1024 * 1024

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Source   string `toml:"source"`
	Interval string `toml:"interval"`
	Timeout  string `toml:"timeout"`

	URL               string       `toml:"url,omitempty"` // Deprecated
	URLs              []string     `toml:"urls"`
	IgnoreReqErr      bool         `toml:"ignore_req_err"`
	MetricTypes       []string     `toml:"metric_types"`
	MetricNameFilter  []string     `toml:"metric_name_filter"`
	MeasurementPrefix string       `toml:"measurement_prefix"`
	MeasurementName   string       `toml:"measurement_name"`
	Measurements      []iprom.Rule `json:"measurements"`
	Output            string       `toml:"output"`
	MaxFileSize       int64        `toml:"max_file_size"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	TagsIgnore  []string            `toml:"tags_ignore"`
	TagsRename  *iprom.RenameTags   `toml:"tags_rename"`
	AsLogging   *iprom.AsLogging    `toml:"as_logging"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers"`

	Tags map[string]string `toml:"tags"`

	Auth map[string]string `toml:"auth"`

	pm *iprom.Prom

	Election bool `toml:"election"`
	chPause  chan bool
	pause    bool

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return catalog }

func (i *Input) SetTags(m map[string]string) {
	if i.Tags == nil {
		i.Tags = make(map[string]string)
	}
	for k, v := range m {
		if _, ok := i.Tags[k]; !ok {
			i.Tags[k] = v
		}
	}
}

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	if i.setup() {
		return
	}

	tick := time.NewTicker(i.pm.Option().GetIntervalDuration())
	defer tick.Stop()

	l.Info("prom start")

	for {
		if i.pause {
			l.Debug("prom paused")
		} else {
			if err := i.RunningCollect(); err != nil {
				l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			l.Info("prom exit")
			return

		case <-i.semStop.Wait():
			l.Info("prom return")
			return

		case <-tick.C:

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

const defaultIntervalDuration = time.Second * 30

func (i *Input) GetIntervalDuration() time.Duration {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			l.Infof("prom init error: %s", err)
			return defaultIntervalDuration
		}
	}
	return i.pm.Option().GetIntervalDuration()
}

func (i *Input) RunningCollect() error {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			return err
		}
	}

	ioname := inputName + "/" + i.Source

	start := time.Now()
	pts := i.doCollect()
	if i.AsLogging != nil && i.AsLogging.Enable {
		// Feed measurement as logging.
		for _, pt := range pts {
			// We need to feed each point separately because
			// each point might have different measurement name.
			if err := io.Feed(pt.Name, datakit.Logging, []*point.Point{pt},
				&io.Option{CollectCost: time.Since(start)}); err != nil {
				l.Errorf("Feed: %s", err)
				io.FeedLastError(ioname, err.Error())
			}
		}
	} else {
		if err := io.Feed(ioname, datakit.Metric, pts,
			&io.Option{CollectCost: time.Since(start)}); err != nil {
			l.Errorf("Feed: %s", err)
			io.FeedLastError(ioname, err.Error())
		}
	}
	return nil
}

func (i *Input) doCollect() []*point.Point {
	l.Debugf("collect URLs %v", i.URLs)

	// If Output is configured, data is written to local file specified by Output.
	// Data will no more be written to datakit io.
	if i.Output != "" {
		err := i.WriteMetricText2File()
		if err != nil {
			l.Errorf("WriteMetricText2File: %s", err.Error())
		}
		return nil
	}

	pts, err := i.Collect()
	if err != nil {
		l.Errorf("Collect: %s", err)
		io.FeedLastError(i.Source, err.Error())

		// Try testing the connect
		for _, u := range i.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return nil
	}

	if len(pts) == 0 {
		l.Warnf("no data")
		return nil
	}

	return pts
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) setup() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(5 * time.Second) // sleep a while
		if err := i.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (i *Input) Init() error {
	l = logger.SLogger(inputName + "/" + i.Source)

	if i.URL != "" {
		i.URLs = append(i.URLs, i.URL)
	}
	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		i.urls = append(i.urls, uu)
	}

	kvIgnore := iprom.IgnoreTagKeyValMatch{}
	for k, arr := range i.IgnoreTagKV {
		for _, x := range arr {
			if re, err := regexp.Compile(x); err != nil {
				l.Warnf("regexp.Compile('%s'): %s, ignored", x, err)
			} else {
				kvIgnore[k] = append(kvIgnore[k], re)
			}
		}
	}

	// toml 不支持匿名字段的 marshal，JSON 支持
	opt := &iprom.Option{
		Source:   i.Source,
		Interval: i.Interval,
		Timeout:  i.Timeout,

		URL:  i.URL,
		URLs: i.URLs,

		MetricTypes: i.MetricTypes,

		IgnoreReqErr: i.IgnoreReqErr,

		MetricNameFilter:  i.MetricNameFilter,
		MeasurementPrefix: i.MeasurementPrefix,
		MeasurementName:   i.MeasurementName,
		Measurements:      i.Measurements,

		TLSOpen:    i.TLSOpen,
		CacertFile: i.CacertFile,
		CertFile:   i.CertFile,
		KeyFile:    i.KeyFile,

		Tags:        i.Tags,
		TagsIgnore:  i.TagsIgnore,
		IgnoreTagKV: kvIgnore,
		HTTPHeaders: i.HTTPHeaders,

		RenameTags:  i.TagsRename,
		AsLogging:   i.AsLogging,
		Output:      i.Output,
		MaxFileSize: i.MaxFileSize,
		Auth:        i.Auth,

		Election: i.Election,
	}

	pm, err := iprom.NewProm(opt)
	if err != nil {
		l.Warnf("prom.NewProm: %s, ignored", err)
		return err
	}
	i.pm = pm
	i.isInitialized = true

	return nil
}

// Collect collects metrics from all URLs.
func (i *Input) Collect() ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	var points []*point.Point
	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		var pts []*point.Point
		if uu.Scheme != "http" && uu.Scheme != "https" {
			pts, err = i.CollectFromFile(u)
		} else {
			pts, err = i.CollectFromHTTP(u)
		}
		if err != nil {
			return nil, err
		}
		points = append(points, pts...)
	}
	return points, nil
}

func (i *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromHTTP(u)
}

func (i *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromFile(filepath)
}

// WriteMetricText2File collects from all URLs and then
// directly writes them to file specified by field Output.
func (i *Input) WriteMetricText2File() error {
	// Remove if file already exists.
	if _, err := os.Stat(i.Output); err == nil {
		if err := os.Remove(i.Output); err != nil {
			return err
		}
	}
	for _, u := range i.URLs {
		if err := i.pm.WriteMetricText2File(u); err != nil {
			return err
		}
		stat, err := os.Stat(i.Output)
		if err != nil {
			return err
		}
		if stat.Size() > i.MaxFileSize {
			return fmt.Errorf("file size is too large, max: %d, got: %d", i.MaxFileSize, stat.Size())
		}
	}
	return nil
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func NewProm() *Input {
	return &Input{
		chPause:     make(chan bool, maxPauseCh),
		MaxFileSize: defaultMaxFileSize,
		Source:      "prom",
		Interval:    "30s",
		Election:    true,

		semStop: cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return NewProm()
	})
}
