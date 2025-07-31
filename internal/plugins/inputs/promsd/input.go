// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promsd implements Prometheus service discovery functions.
package promsd

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

var _ inputs.ElectionInput = (*Input)(nil)

type Input struct {
	Source   string `toml:"source"`
	Election bool   `toml:"election"`

	Scrape   *ScrapeConfig `toml:"scrape"`
	HTTPSD   *HTTPSD       `toml:"http_sd_config"`
	ConsulSD *ConsulSD     `toml:"consul_sd_config"`

	Tags map[string]string `toml:"tags"`

	chPause       chan bool
	pause         bool
	workerRunning bool

	Feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
	semStop *cliutils.Sem // start stop signal

	log *logger.Logger
}

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{inputs.DefaultEmptyMeasurement}
}

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return "prom" }

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Run() {
	ipt.log = logger.SLogger(inputName + "/" + ipt.Source)
	ipt.log.Info("start")
	ipt.setup()

	var ctx context.Context
	var cancel context.CancelFunc
	scraperChan := make(chan scraper, 10)

	for {
		if !ipt.pause && !ipt.workerRunning {
			ctx, cancel = context.WithCancel(context.Background())
			ipt.startWorker(ctx, scraperChan)
			ipt.workerRunning = true
		}

		select {
		case <-datakit.Exit.Wait():
			if cancel != nil {
				cancel()
			}
			ipt.log.Info("promsd exit")
			return

		case <-ipt.semStop.Wait():
			if cancel != nil {
				cancel()
			}
			ipt.log.Info("promsd return")
			return

		case ipt.pause = <-ipt.chPause:
			if ipt.pause && cancel != nil {
				cancel()
				ipt.workerRunning = false
			}
		}
	}
}

func (ipt *Input) setup() {
	if ipt.HTTPSD != nil {
		ipt.HTTPSD.SetLogger(ipt.log)
	}
	if ipt.ConsulSD != nil {
		ipt.ConsulSD.SetLogger(ipt.log)
	}
}

func (ipt *Input) startWorker(ctx context.Context, scraperChan chan scraper) {
	promOptions, err := ipt.buildPromOptions()
	if err != nil {
		ipt.log.Warnf("failed of build prom options, err: %s", err)
		return
	}

	g := datakit.G(inputName + "/" + ipt.Source)

	for i := 0; i < workerNumber; i++ {
		name := fmt.Sprintf("worker-%d", i)

		randSleep := rand.Intn(100 /*100 ms*/) //nolint:gosec
		time.Sleep(time.Duration(randSleep) * time.Millisecond)

		g.Go(func(_ context.Context) error {
			startScraperConsumer(ctx, ipt.log, name, ipt.Scrape.Interval, scraperChan)
			return nil
		})
	}

	if ipt.HTTPSD != nil {
		g.Go(func(_ context.Context) error {
			ipt.HTTPSD.StartScraperProducer(ctx, ipt.Scrape, promOptions, scraperChan)
			return nil
		})
	}

	if ipt.ConsulSD != nil {
		g.Go(func(_ context.Context) error {
			ipt.ConsulSD.StartScraperProducer(ctx, ipt.Scrape, promOptions, scraperChan)
			return nil
		})
	}
}

func (ipt *Input) buildPromOptions() ([]promscrape.Option, error) {
	if ipt.Scrape == nil {
		return nil, fmt.Errorf("unexpected scrape config")
	}

	var globalTags map[string]string
	if ipt.Election {
		globalTags = ipt.Tagger.ElectionTags()
		ipt.log.Infof("add global election tags %q", globalTags)
	} else {
		globalTags = ipt.Tagger.HostTags()
		ipt.log.Infof("add global host tags %q", globalTags)
	}
	mergedTags := inputs.MergeTags(globalTags, ipt.Tags, "")

	opts := []promscrape.Option{
		promscrape.WithSource(inputName + "/" + ipt.Source),
		promscrape.KeepExistMetricName(ipt.Scrape.KeepExistMetricName),
		promscrape.WithExtraTags(mergedTags),
		promscrape.WithHTTPHeader(ipt.Scrape.HTTPHeaders),
		promscrape.WithCallback(ipt.callbackFn),
	}

	if ipt.Scrape.Auth != nil {
		authOpts, err := buildPromOptionsWithAuth(ipt.Scrape.Auth)
		if err != nil {
			return nil, err
		}
		opts = append(opts, authOpts...)
	}

	return opts, nil
}

func (ipt *Input) callbackFn(pts []*point.Point) error {
	if len(pts) == 0 {
		return nil
	}
	if err := ipt.Feeder.Feed(
		point.Metric,
		pts,
		dkio.WithSource(dkio.FeedSource(inputName, ipt.Source)),
		dkio.WithElection(ipt.Election),
	); err != nil {
		ipt.log.Warnf("failed to feed prom metrics: %s", err)
	}
	return nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Source:   "not-set",
			Election: true,
			pause:    false,
			chPause:  make(chan bool, inputs.ElectionPauseChannelLength),
			Tags:     make(map[string]string),
			Feeder:   dkio.DefaultFeeder(),
			Tagger:   datakit.DefaultGlobalTagger(),
			semStop:  cliutils.NewSem(),
		}
	})
}
