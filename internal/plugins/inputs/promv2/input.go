// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promv2 scrape prometheus exporter metrics.
package promv2

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName     = "promv2"
	defaultSource = "not-set"
)

type Input struct {
	Source          string        `toml:"source"`
	MeasurementName string        `toml:"measurement_name"`
	URL             string        `toml:"url"`
	Interval        time.Duration `toml:"interval"`
	Election        bool          `toml:"election"`

	KeepExistMetricName bool `toml:"keep_exist_metric_name"`
	HonorTimestamps     bool `toml:"honor_timestamps"`
	DisableInstanceTag  bool `toml:"disable_instance_tag"`

	BearerTokenFile string            `toml:"bearer_token_file"`
	HTTPHeaders     map[string]string `toml:"http_headers"`
	Tags            map[string]string `toml:"tags"`
	dknet.TLSClientConfig

	endpoint *url.URL

	chPause chan bool
	pause   bool

	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
	semStop *cliutils.Sem

	lastStart  time.Time
	scraper    *promscrape.PromScraper
	pointCount int
	logger     *logger.Logger
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
	ipt.logger.Info("promv2 input started")

	if err := ipt.setup(); err != nil {
		ipt.logger.Errorf("setup failed: %s", err)
		ipt.logger.Info("promv2 input stopped")
		return
	}

	start := ntp.Now()
	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		if !ipt.pause {
			ipt.scrape(start.UnixNano())
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.logger.Info("promv2 input exiting")
			return

		case <-ipt.semStop.Wait():
			ipt.logger.Info("promv2 input stopped")
			return

		case ipt.pause = <-ipt.chPause:

		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)
		}
	}
}

func (ipt *Input) setup() error {
	if err := ipt.parseURL(); err != nil {
		return err
	}

	tags := ipt.buildTags()
	if err := ipt.setupAuth(); err != nil {
		return err
	}

	opts := ipt.buildScraperOptions(tags)
	scraper, err := promscrape.NewPromScraper(opts...)
	if err != nil {
		return err
	}

	ipt.scraper = scraper
	ipt.Interval = config.ProtectedInterval(time.Second, time.Minute*5, ipt.Interval)
	return nil
}

func (ipt *Input) parseURL() error {
	endpoint, err := url.Parse(ipt.URL)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", ipt.URL, err)
	}
	ipt.endpoint = endpoint
	return nil
}

func (ipt *Input) buildTags() map[string]string {
	tags := make(map[string]string)

	var globalTags map[string]string
	if ipt.Election {
		globalTags = ipt.tagger.ElectionTags()
		ipt.logger.Debugf("using election tags: %v", globalTags)
	} else {
		globalTags = ipt.tagger.HostTags()
		ipt.logger.Debugf("using host tags: %v", globalTags)
	}

	mergedTags := inputs.MergeTags(globalTags, ipt.Tags, ipt.URL)
	for k, v := range mergedTags {
		tags[k] = v
	}

	if !ipt.DisableInstanceTag {
		if _, ok := mergedTags["instance"]; !ok {
			tags["instance"] = ipt.endpoint.Host
		}
	}

	return tags
}

func (ipt *Input) setupAuth() error {
	if ipt.BearerTokenFile == "" {
		return nil
	}

	token, err := os.ReadFile(ipt.BearerTokenFile)
	if err != nil {
		return fmt.Errorf("read bearer token file failed: %w", err)
	}

	if _, exist := ipt.HTTPHeaders["Authorization"]; !exist {
		ipt.HTTPHeaders["Authorization"] = fmt.Sprintf("Bearer %s", strings.TrimSpace(string(token)))
	}

	return nil
}

func (ipt *Input) buildScraperOptions(tags map[string]string) []promscrape.Option {
	opts := []promscrape.Option{
		promscrape.WithSource("promv2/" + ipt.Source),
		promscrape.WithMeasurement(ipt.MeasurementName),
		promscrape.WithHTTPHeader(ipt.HTTPHeaders),
		promscrape.WithExtraTags(tags),
		promscrape.KeepExistMetricName(ipt.KeepExistMetricName),
		promscrape.HonorTimestamps(ipt.HonorTimestamps),
		promscrape.WithCallback(ipt.callback),
	}

	if ipt.hasTLSConfig() {
		opts = append(opts,
			promscrape.WithTLSOpen(true),
			promscrape.WithCacertFiles(ipt.CaCerts),
			promscrape.WithCertFile(ipt.Cert),
			promscrape.WithKeyFile(ipt.CertKey),
			promscrape.WithInsecureSkipVerify(ipt.InsecureSkipVerify),
		)
	}

	return opts
}

func (ipt *Input) hasTLSConfig() bool {
	return len(ipt.CaCerts) != 0 ||
		ipt.CertKey != "" ||
		ipt.Cert != "" ||
		ipt.InsecureSkipVerify
}

func (ipt *Input) scrape(timestamp int64) {
	ipt.lastStart = time.Now()
	ipt.pointCount = 0

	ipt.scraper.SetTimestamp(timestamp)
	if err := ipt.scraper.ScrapeURL(ipt.URL); err != nil {
		ipt.logger.Warnf("scrape failed: %s", err)
	}

	scrapeTotal.WithLabelValues(ipt.Source,
		fmt.Sprintf(":%s%s", ipt.endpoint.Port(), ipt.endpoint.Path)).Observe(float64(ipt.pointCount))
}

func (ipt *Input) callback(pts []*point.Point) error {
	if len(pts) == 0 {
		return nil
	}

	cost := time.Since(ipt.lastStart)

	if err := ipt.feeder.Feed(
		point.Metric,
		pts,
		dkio.WithCollectCost(cost),
		dkio.WithSource(dkio.FeedSource(inputName, ipt.Source)),
		dkio.WithElection(ipt.Election),
	); err != nil {
		ipt.logger.Warnf("feed metrics failed: %s", err)
	}
	ipt.pointCount += len(pts)

	return nil
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Pause() error {
	ticker := time.NewTicker(inputs.ElectionPauseTimeout)
	defer ticker.Stop()

	select {
	case ipt.chPause <- true:
		return nil
	case <-ticker.C:
		return fmt.Errorf("pause %s timeout", inputName)
	}
}

func (ipt *Input) Resume() error {
	ticker := time.NewTicker(inputs.ElectionResumeTimeout)
	defer ticker.Stop()

	select {
	case ipt.chPause <- false:
		return nil
	case <-ticker.C:
		return fmt.Errorf("resume %s timeout", inputName)
	}
}

func newProm() *Input {
	return &Input{
		Source:              defaultSource,
		KeepExistMetricName: true,
		HonorTimestamps:     true,

		pause:       false,
		chPause:     make(chan bool, inputs.ElectionPauseChannelLength),
		HTTPHeaders: make(map[string]string),
		Tags:        make(map[string]string),
		feeder:      dkio.DefaultFeeder(),
		tagger:      datakit.DefaultGlobalTagger(),
		semStop:     cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return newProm()
	})
}
