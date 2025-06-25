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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
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
	feeder dkio.Feeder

	Election bool `toml:"election"`
	chPause  chan bool
	pause    bool

	Tagger datakit.GlobalTagger

	urls []*url.URL

	semStop *cliutils.Sem // start stop signal

	isInitialized bool

	urlTags map[string]urlTags

	start time.Time

	// Input holds logger because prom have different types of instances.
	l *logger.Logger
}

type urlTags map[string]string

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&etcdMeasurement{},
	}
}

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return catalog }

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Run() {
	if ipt.setup() {
		return
	}

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	ipt.l.Info("etcd start")

	ipt.start = ntp.Now()

	for {
		if ipt.pause {
			ipt.l.Debug("etcd paused")
		} else {
			if err := ipt.collect(); err != nil {
				ipt.l.Warn(err)
			}
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.l.Info("etcd exit")
			return

		case <-ipt.semStop.Wait():
			ipt.l.Info("etcd return")
			return

		case tt := <-tick.C:
			ipt.start = inputs.AlignTime(tt, ipt.start, ipt.Interval)

		case ipt.pause = <-ipt.chPause:
			// nil
		}
	}
}

func (ipt *Input) collect() error {
	if !ipt.isInitialized {
		if err := ipt.Init(); err != nil {
			return err
		}
	}

	ioname := inputName + "/" + ipt.Source

	start := time.Now()
	pts, err := ipt.doCollect()
	if err != nil {
		return err
	}
	if pts == nil {
		return fmt.Errorf("points got nil from doCollect")
	}

	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(ioname)); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed measurement: %s", err)
	}

	return nil
}

func (ipt *Input) doCollect() ([]*point.Point, error) {
	ipt.l.Debugf("collect URLs %v", ipt.URLs)

	pts, err := ipt.Collect()
	if err != nil {
		ipt.l.Errorf("Collect: %s", err)

		ioname := inputName + "-" + ipt.Source
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorSource(ioname),
		)

		// Try testing the connect
		for _, u := range ipt.urls {
			if err := net.RawConnect(u.Hostname(), u.Port(), time.Second*3); err != nil {
				ipt.l.Errorf("failed to connect to %s:%s, %s", u.Hostname(), u.Port(), err)
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
func (ipt *Input) Collect() ([]*point.Point, error) {
	if ipt.pm == nil {
		return nil, fmt.Errorf("i.pm is nil")
	}
	var points []*point.Point
	for _, u := range ipt.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		var pts []*point.Point
		if uu.Scheme != "http" && uu.Scheme != "https" {
			pts, err = ipt.CollectFromFile(u)
		} else {
			pts, err = ipt.pm.CollectFromHTTPV2(u, iprom.WithTimestamp(ipt.start.UnixNano()))
		}
		if err != nil {
			return nil, err
		}

		// append tags to points
		for k, v := range ipt.urlTags[u] {
			for _, pt := range pts {
				pt.AddTag(k, v)
			}
		}

		points = append(points, pts...)
	}

	return points, nil
}

func (ipt *Input) CollectFromFile(filepath string) ([]*point.Point, error) {
	if ipt.pm == nil {
		return nil, nil
	}
	return ipt.pm.CollectFromFileV2(filepath)
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) setup() bool {
	// for etcd only.
	ipt.Source = inputName
	ipt.MeasurementName = inputName

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}
		time.Sleep(1 * time.Second) // sleep a while
		if err := ipt.Init(); err != nil {
			continue
		} else {
			break
		}
	}

	return false
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (ipt *Input) Init() error {
	l = logger.SLogger(inputName)
	ipt.l = logger.SLogger(inputName + "/" + ipt.Source)

	for _, u := range ipt.URLs {
		uu, err := url.Parse(u)
		if err != nil {
			return err
		}
		ipt.urls = append(ipt.urls, uu)

		if ipt.Election {
			ipt.urlTags[u] = inputs.MergeTags(ipt.Tagger.ElectionTags(), ipt.Tags, u)
		} else {
			ipt.urlTags[u] = inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, u)
		}
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(ipt.l), // WithLogger must in the first
		iprom.WithSource(ipt.Source),
		iprom.WithMeasurementName(ipt.MeasurementName),
		iprom.WithTLSOpen(ipt.TLSOpen),
		iprom.WithCacertFiles([]string{ipt.CacertFile}),
		iprom.WithCertFile(ipt.CertFile),
		iprom.WithKeyFile(ipt.KeyFile),
		iprom.WithTagsIgnore(ipt.TagsIgnore),
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		ipt.l.Warnf("iprom.NewProm: %s, ignored", err)
		return err
	}
	ipt.pm = pm
	ipt.isInitialized = true

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
		feeder:  dkio.DefaultFeeder(),
		Tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
