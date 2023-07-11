// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package etcd scrape etcd prometheus exporter metrics.
package etcd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName               = "etcd"
	catalog                 = "etcd"
	defaultIntervalDuration = time.Second * 30
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Source   string        `toml:"source"`
	Interval time.Duration `toml:"interval"`

	URLs            []string `toml:"urls"`
	MeasurementName string   `toml:"measurement_name"`

	TLSOpen    bool   `toml:"tls_open"`
	CacertFile string `toml:"tls_ca"`
	CertFile   string `toml:"tls_cert"`
	KeyFile    string `toml:"tls_key"`

	TagsIgnore []string `toml:"tags_ignore"`

	Tags map[string]string `toml:"tags"`

	pm     *iprom.Prom
	Feeder io.Feeder

	Election bool `toml:"election"`
	chPause  chan bool
	pause    bool

	Tagger dkpt.GlobalTagger

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool

	urlTags map[string]urlTags

	// Input holds logger because prom have different types of instances.
	l *logger.Logger
}

type urlTags map[string]string

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
	if i.setup() {
		return
	}

	// for etcd only.
	i.Source = inputName
	i.MeasurementName = inputName

	tick := time.NewTicker(i.Interval)
	defer tick.Stop()

	i.l.Info("etcd start")

	for {
		if i.pause {
			i.l.Debug("etcd paused")
		} else {
			if err := i.collect(); err != nil {
				i.l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			i.l.Info("etcd exit")
			return

		case <-i.semStop.Wait():
			i.l.Info("etcd return")
			return

		case <-tick.C:

		case i.pause = <-i.chPause:
			// nil
		}
	}
}

func (i *Input) collect() error {
	if !i.isInitialized {
		if err := i.Init(); err != nil {
			return err
		}
	}

	ioname := inputName + "/" + i.Source

	start := time.Now()
	pts, err := i.doCollect()
	if err != nil {
		return err
	}
	if pts == nil {
		return fmt.Errorf("points got nil from doCollect")
	}

	err = i.Feeder.Feed(ioname, point.Metric, pts,
		&io.Option{CollectCost: time.Since(start)})
	if err != nil {
		i.l.Errorf("Feed: %s", err)
		i.Feeder.FeedLastError(ioname, err.Error())
	}
	return nil
}

func (i *Input) doCollect() ([]*point.Point, error) {
	i.l.Debugf("collect URLs %v", i.URLs)

	pts, err := i.Collect()
	if err != nil {
		i.l.Errorf("Collect: %s", err)
		i.Feeder.FeedLastError(i.Source, err.Error())

		// Try testing the connect
		for _, u := range i.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				i.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
			}
		}

		return nil, err
	}

	if pts == nil {
		return nil, fmt.Errorf("points got nil from Collect")
	}

	return pts, nil
}

// Collect collects metrics from all URLs.
func (i *Input) Collect() ([]*point.Point, error) {
	if i.pm == nil {
		return nil, fmt.Errorf("i.pm is nil")
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

		// append tags to points
		for k, v := range i.urlTags[u] {
			for _, pt := range pts {
				pt.AddTag([]byte(k), []byte(v))
			}
		}

		points = append(points, pts...)
	}

	return points, nil
}

func (i *Input) CollectFromHTTP(u string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromHTTPV2(u)
}

func (i *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if i.pm == nil {
		return nil, nil
	}
	return i.pm.CollectFromFileV2(filepath)
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
		time.Sleep(1 * time.Second) // sleep a while
		if err := i.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
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

func (i *Input) Init() error {
	i.l = logger.SLogger(inputName + "/" + i.Source)

	for _, u := range i.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		i.urls = append(i.urls, uu)

		if i.Election {
			i.urlTags[u] = inputs.MergeTags(i.Tagger.ElectionTags(), i.Tags, u)
		} else {
			i.urlTags[u] = inputs.MergeTags(i.Tagger.HostTags(), i.Tags, u)
		}
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(i.l), // WithLogger must in the first
		iprom.WithSource(i.Source),
		iprom.WithMeasurementName(i.MeasurementName),
		iprom.WithTLSOpen(i.TLSOpen),
		iprom.WithCacertFile(i.CacertFile),
		iprom.WithCertFile(i.CertFile),
		iprom.WithKeyFile(i.KeyFile),
		iprom.WithTagsIgnore(i.TagsIgnore),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		i.l.Warnf("iprom.NewProm: %s, ignored", err)
		return err
	}
	i.pm = pm
	i.isInitialized = true

	return nil
}

var maxPauseCh = inputs.ElectionPauseChannelLength

func defaultInput() *Input {
	return &Input{
		chPause:  make(chan bool, maxPauseCh),
		Source:   "etcd",
		Interval: defaultIntervalDuration,
		Election: true,
		Tags:     make(map[string]string),

		urlTags: map[string]urlTags{},

		semStop: cliutils.NewSem(),
		Feeder:  io.DefaultFeeder(),
		Tagger:  dkpt.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
