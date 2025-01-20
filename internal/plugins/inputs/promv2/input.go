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

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	inputName = "promv2"
)

type Input struct {
	Source          string `toml:"source"`
	MeasurementName string `toml:"measurement_name"`

	URL      string   `toml:"url"`
	endpoint *url.URL // parsed URL

	KeepExistMetricName bool `toml:"keep_exist_metric_name"`
	DisableInstanceTag  bool `toml:"disable_instance_tag"`

	BearerTokenFile string `toml:"bearer_token_file"`
	dknet.TLSClientConfig

	HTTPHeaders map[string]string `toml:"http_headers"`
	Tags        map[string]string `toml:"tags"`

	Interval time.Duration `toml:"interval"`
	Election bool          `toml:"election"`

	chPause chan bool
	pause   bool

	Feeder dkio.Feeder
	Tagger datakit.GlobalTagger

	lastStart time.Time

	scraper *promscrape.PromScraper
	count   int
	log     *logger.Logger
}

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement { return nil }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) Catalog() string { return "prom" }

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Run() {
	ipt.log = logger.SLogger(inputName + "/" + ipt.Source)
	ipt.log.Info("start")

	if err := ipt.setup(); err != nil {
		ipt.log.Infof("init failure: %s", err)
		ipt.log.Info("exit")
		return
	}

	start := time.Now()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.log.Info("prom exit")
			return

		case ipt.pause = <-ipt.chPause:
			// nil

		case tt := <-tick.C:
			nextts := inputs.AlignTimeMillSec(tt, start.UnixMilli(), ipt.Interval.Milliseconds())
			start = time.UnixMilli(nextts)
			if !ipt.pause {
				ipt.scrape(start.UnixNano())
			}
		}
	}
}

func (ipt *Input) setup() error {
	// parse url
	u, err := url.Parse(ipt.URL)
	if err != nil {
		return err
	}
	ipt.endpoint = u

	tags := make(map[string]string)

	// on production, ipt.Tagger is empty, but we may add testing tagger here.
	var globalTags map[string]string
	if ipt.Election {
		globalTags = ipt.Tagger.ElectionTags()
		ipt.log.Infof("add global election tags %q", globalTags)
	} else {
		globalTags = ipt.Tagger.HostTags()
		ipt.log.Infof("add global host tags %q", globalTags)
	}

	mergedTags := inputs.MergeTags(globalTags, ipt.Tags, ipt.URL)
	for k, v := range mergedTags {
		tags[k] = v
	}

	// set instance tag.
	// The `instance' tag should not override.
	if !ipt.DisableInstanceTag {
		if _, ok := mergedTags["instance"]; !ok {
			tags["instance"] = u.Host
		}
	}

	if ipt.BearerTokenFile != "" {
		token, err := os.ReadFile(ipt.BearerTokenFile)
		if err != nil {
			return err
		}
		if _, exist := ipt.HTTPHeaders["Authorization"]; !exist {
			ipt.HTTPHeaders["Authorization"] = fmt.Sprintf("Bearer %s", strings.TrimSpace(string(token)))
		}
	}

	opts := []promscrape.Option{
		promscrape.WithSource("promv2/" + ipt.Source),
		promscrape.WithMeasurement(ipt.MeasurementName),
		promscrape.WithHTTPHeader(ipt.HTTPHeaders),
		promscrape.WithExtraTags(tags),
		promscrape.KeepExistMetricName(ipt.KeepExistMetricName),
		promscrape.WithCallback(ipt.callback),
	}

	if len(ipt.CaCerts) != 0 ||
		ipt.CertKey != "" ||
		ipt.Cert != "" ||
		ipt.InsecureSkipVerify {
		opts = append(opts,
			promscrape.WithTLSOpen(true),
			promscrape.WithCacertFiles(ipt.CaCerts),
			promscrape.WithCertFile(ipt.Cert),
			promscrape.WithKeyFile(ipt.CertKey),
			promscrape.WithInsecureSkipVerify(ipt.InsecureSkipVerify),
		)
	}

	ps, err := promscrape.NewPromScraper(opts...)
	if err != nil {
		return err
	}

	ipt.scraper = ps
	return nil
}

func (ipt *Input) scrape(timestamp int64) {
	ipt.lastStart = time.Now()
	// reset count
	ipt.count = 0

	ipt.scraper.SetTimestamp(timestamp)
	if err := ipt.scraper.ScrapeURL(ipt.URL); err != nil {
		ipt.log.Warn(err)
	}

	scrapeTotal.WithLabelValues(ipt.Source,
		fmt.Sprintf(":%s%s", ipt.endpoint.Port(), ipt.endpoint.Path)).Observe(float64(ipt.count))
}

func (ipt *Input) callback(pts []*point.Point) error {
	if len(pts) == 0 {
		return nil
	}

	cost := time.Since(ipt.lastStart)

	if err := ipt.Feeder.FeedV2(
		point.Metric,
		pts,
		dkio.WithCollectCost(cost),
		dkio.WithInputName(inputName+"/"+ipt.Source),
		dkio.WithElection(ipt.Election),
	); err != nil {
		ipt.log.Warnf("failed to feed prom metrics: %s", err)
	}
	ipt.count += len(pts)

	return nil
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

func newProm() *Input {
	return &Input{
		Source:      "not-set",
		pause:       false,
		chPause:     make(chan bool, inputs.ElectionPauseChannelLength),
		HTTPHeaders: make(map[string]string),
		Tags:        make(map[string]string),
		Feeder:      dkio.DefaultFeeder(),
		Tagger:      datakit.DefaultGlobalTagger(),
		Interval:    time.Second * 30,
	}
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return newProm()
	})
}
