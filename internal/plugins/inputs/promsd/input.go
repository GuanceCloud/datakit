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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
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
	FileSD   *FileSD       `toml:"file_sd_config"`

	Tags map[string]string `toml:"tags"`

	pauseChan      chan bool
	isPaused       bool
	isWorkerActive bool

	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
	stopSem *cliutils.Sem

	logger *logger.Logger
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
	ipt.logger = logger.SLogger(inputName + "/" + ipt.Source)
	ipt.logger.Info("promsd input started")
	ipt.setup()

	var ctx context.Context
	var cancel context.CancelFunc
	scraperChan := make(chan scraper, 10)

	for {
		if !ipt.isPaused && !ipt.isWorkerActive {
			ctx, cancel = context.WithCancel(context.Background())
			ipt.startWorker(ctx, scraperChan)
			ipt.isWorkerActive = true
		}

		select {
		case <-datakit.Exit.Wait():
			if cancel != nil {
				cancel()
			}
			ipt.logger.Info("promsd input exiting")
			return

		case <-ipt.stopSem.Wait():
			if cancel != nil {
				cancel()
			}
			ipt.logger.Info("promsd input stopped")
			return

		case ipt.isPaused = <-ipt.pauseChan:
			if ipt.isPaused && cancel != nil {
				cancel()
				ipt.isWorkerActive = false
			}
		}
	}
}

func (ipt *Input) setup() {
	if ipt.HTTPSD != nil {
		ipt.HTTPSD.RefreshInterval = config.ProtectedInterval(minRefreshInterval, maxRefreshInterval, ipt.HTTPSD.RefreshInterval)
		ipt.HTTPSD.SetLogger(ipt.logger)
	}
	if ipt.ConsulSD != nil {
		ipt.ConsulSD.RefreshInterval = config.ProtectedInterval(minRefreshInterval, maxRefreshInterval, ipt.ConsulSD.RefreshInterval)
		ipt.ConsulSD.SetLogger(ipt.logger)
	}
	if ipt.FileSD != nil {
		ipt.FileSD.RefreshInterval = config.ProtectedInterval(minRefreshInterval, maxRefreshInterval, ipt.FileSD.RefreshInterval)
		ipt.FileSD.SetLogger(ipt.logger)
	}
}

func (ipt *Input) startWorker(ctx context.Context, scraperChan chan scraper) {
	promOptions, err := ipt.buildPromOptions()
	if err != nil {
		ipt.logger.Warnf("failed to build prom options: %s", err)
		return
	}

	g := datakit.G(inputName + "/" + ipt.Source)

	for i := 0; i < workerCount; i++ {
		workerName := fmt.Sprintf("worker-%d", i)

		// 100ms 以内随机启动
		randSleep := rand.Intn(100) // nolint:gosec
		time.Sleep(time.Duration(randSleep) * time.Millisecond)

		g.Go(func(_ context.Context) error {
			startScraperConsumer(ctx, ipt.logger, workerName, ipt.Scrape.Interval, scraperChan)
			return nil
		})
	}

	ipt.startServiceDiscovery(ctx, g, promOptions, scraperChan)
}

func (ipt *Input) startServiceDiscovery(ctx context.Context, g *goroutine.Group, promOptions []promscrape.Option, scraperChan chan scraper) {
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

	if ipt.FileSD != nil {
		g.Go(func(_ context.Context) error {
			ipt.FileSD.StartScraperProducer(ctx, ipt.Scrape, promOptions, scraperChan)
			return nil
		})
	}
}

func (ipt *Input) buildPromOptions() ([]promscrape.Option, error) {
	if ipt.Scrape == nil {
		return nil, fmt.Errorf("scrape config is required")
	}

	var globalTags map[string]string
	if ipt.Election {
		globalTags = ipt.tagger.ElectionTags()
		ipt.logger.Infof("using election tags: %v", globalTags)
	} else {
		globalTags = ipt.tagger.HostTags()
		ipt.logger.Infof("using host tags: %v", globalTags)
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
	if err := ipt.feeder.Feed(
		point.Metric,
		pts,
		dkio.WithSource(dkio.FeedSource(inputName, ipt.Source)),
		dkio.WithElection(ipt.Election),
	); err != nil {
		ipt.logger.Warnf("failed to feed prom metrics: %s", err)
	}
	return nil
}

func (ipt *Input) Terminate() {
	if ipt.stopSem != nil {
		ipt.stopSem.Close()
	}
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseChan <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseChan <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

// nolint:gochecknoinits
func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Source:    "not-set",
			Election:  true,
			isPaused:  false,
			pauseChan: make(chan bool, inputs.ElectionPauseChannelLength),
			Tags:      make(map[string]string),
			feeder:    dkio.DefaultFeeder(),
			tagger:    datakit.DefaultGlobalTagger(),
			stopSem:   cliutils.NewSem(),
		}
	})
}
