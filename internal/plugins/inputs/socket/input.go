// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package socket collect socket metrics
package socket

import (
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	KillGrace = 5 * time.Second
	TCP       = "tcp"
	UDP       = "udp"
)

var (
	inputName   = "socket"
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second * 10
	maxInterval = time.Second * 60

	// interface assert.
	_ inputs.ElectionInput = (*input)(nil)
)

type input struct {
	DestURL    []string         `toml:"dest_url"`
	Interval   datakit.Duration `toml:"interval"` // 单位为秒
	UDPTimeOut datakit.Duration `toml:"udp_timeout"`
	TCPTimeOut datakit.Duration `toml:"tcp_timeout"`

	collectCache []*point.Point
	Tags         map[string]string `toml:"tags"`
	semStop      *cliutils.Sem     // start stop signal
	platform     string

	Election bool `toml:"election"`
	pause    bool
	pauseCh  chan bool

	feeder dkio.Feeder
	tagger datakit.GlobalTagger

	urlTags []map[string]string
}

func (i *input) Catalog() string {
	return "socket"
}

func (i *input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("socket input started")

	// setup tags of multiple dest URLs
	for _, urlStr := range i.DestURL {
		u, err := url.Parse(urlStr)
		if err != nil {
			l.Warnf("url.Parse %q failed: %s, ignored", u, err)
			continue
		}

		var globalTags map[string]string
		if i.Election {
			globalTags = i.tagger.ElectionTags()
			l.Infof("add global election tags %q", globalTags)
		} else {
			globalTags = i.tagger.HostTags()
			l.Infof("add global host tags %q", globalTags)
		}

		i.urlTags = append(i.urlTags, inputs.MergeTags(globalTags, i.Tags, urlStr))
	}
}

func (i *input) Run() {
	i.setup()

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	for {
		if i.pause {
			l.Debugf("election failed, skipped")
		} else {
			i.collectCache = i.collectCache[:0]
			start := time.Now()

			i.Collect()

			if len(i.collectCache) > 0 {
				if err := i.feeder.FeedV2(point.Metric, i.collectCache,
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(i.Election),
					dkio.WithInputName(inputName),
				); err != nil {
					l.Errorf("Feed: %s, ignored", err)
				}
			}
		}

		select {
		case <-tick.C:

		case i.pause = <-i.pauseCh:
			l.Infof("set input %q paused?(%v)", inputName, i.pause)

		case <-datakit.Exit.Wait():
			l.Infof("socket input exit")
			return

		case <-i.semStop.Wait():
			l.Infof("socket input return")
			return
		}
	}
}

func (i *input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.OSLabelMac}
}

func (i *input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TCPMeasurement{},
		&UDPMeasurement{},
	}
}

func (i *input) Collect() {
	if len(i.DestURL) == 0 {
		l.Warnf("input socket have no URLs")
		return
	}

	for idx, u := range i.DestURL {
		resURL, err := url.Parse(u)
		if err != nil {
			l.Errorf("url.Parse: %s, ignored", err.Error())
			continue
		}

		switch resURL.Scheme {
		case TCP:
			pt := i.collectTCP(resURL.Hostname(), resURL.Port())
			if pt != nil {
				for k, v := range i.urlTags[idx] {
					pt.AddTag(k, v)
				}

				i.collectCache = append(i.collectCache, pt)
			}

		case UDP:
			pt := i.collectUDP(resURL.Hostname(), resURL.Port())
			if pt != nil {
				for k, v := range i.urlTags[idx] {
					pt.AddTag(k, v)
				}
				i.collectCache = append(i.collectCache, pt)
			}

		default:
			l.Warnf("unknown scheme %q", resURL.Scheme)

			i.feeder.FeedLastError(fmt.Sprintf("unknown scheme %q", resURL.Scheme),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Metric))
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &input{
			tagger: datakit.DefaultGlobalTagger(),
			feeder: dkio.DefaultFeeder(),

			Election: true,
			pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),

			Interval: datakit.Duration{Duration: time.Second * 30},
			semStop:  cliutils.NewSem(),
			platform: runtime.GOOS,

			UDPTimeOut: datakit.Duration{Duration: time.Second * 10},
			TCPTimeOut: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
