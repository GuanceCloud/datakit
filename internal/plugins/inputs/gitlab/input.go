// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package gitlab collect GitLab metrics
package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

var (
	_ inputs.ElectionInput = (*Input)(nil)
	_ inputs.InputV2       = (*Input)(nil)
	l                      = logger.DefaultSLogger(inputName)
	g                      = goroutine.G("inputs_gitlab")
)

const (
	inputName = "gitlab"
	catalog   = "gitlab"

	gitlabEventHeader = "X-Gitlab-Event"
	pipelineHook      = "Pipeline Hook"
	jobHook           = "Job Hook"

	sampleCfg = `
[[inputs.gitlab]]
    ## set true if you need to collect metric from url below
    enable_collect = true

    ## param type: string - default: http://127.0.0.1:80/-/metrics
    prometheus_url = "http://127.0.0.1:80/-/metrics"

    ## param type: string - optional: time units are "ms", "s", "m", "h" - default: 10s
    interval = "10s"

    ## datakit can listen to gitlab ci data at /v1/gitlab when enabled
    enable_ci_visibility = true

    ## Set true to enable election
    election = true

    ## Bearer token file for authentication
    bearer_token_file = ""

    ## HTTP headers
    [inputs.gitlab.http_headers]
    # X-Custom-Header = "custom-value"

    ## TLS configuration
    # tls_ca = "/path/to/ca.pem"
    # tls_cert = "/path/to/cert.pem"
    # tls_key = "/path/to/key.pem"
    # insecure_skip_verify = false

    ## extra tags for gitlab-ci data.
    ## these tags will not overwrite existing tags.
    [inputs.gitlab.ci_extra_tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

    ## extra tags for gitlab metrics
    [inputs.gitlab.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

type Input struct {
	EnableCollect bool              `toml:"enable_collect"`
	URL           string            `toml:"prometheus_url"`
	Interval      string            `toml:"interval"`
	Tags          map[string]string `toml:"tags"`

	EnableCIVisibility bool              `toml:"enable_ci_visibility"`
	CIExtraTags        map[string]string `toml:"ci_extra_tags"`

	BearerTokenFile string            `toml:"bearer_token_file"`
	HTTPHeaders     map[string]string `toml:"http_headers"`

	// TLS configuration
	TLSCA              string   `toml:"tls_ca"`
	TLSCert            string   `toml:"tls_cert"`
	TLSKey             string   `toml:"tls_key"`
	InsecureSkipVerify bool     `toml:"insecure_skip_verify"`
	CaCerts            []string `toml:"tls_ca_files"`

	Election bool `toml:"election"`
	pause    atomic.Bool

	semStop *cliutils.Sem // start stop signal
	reqMemo requestMemo
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger

	scraper   *promscrape.PromScraper
	interval  time.Duration
	endpoint  *url.URL
	logger    *logger.Logger
	lastStart time.Time
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) RegHTTPHandler() {
	if ipt.EnableCIVisibility {
		l.Infof("start listening to gitlab pipeline/job webhooks")
		g.Go(func(ctx context.Context) error {
			ipt.reqMemo.memoMaintainer(time.Second * 30)
			return nil
		})
		httpapi.RegHTTPHandler("POST", "/v1/gitlab", httpapi.ProtectedHandlerFunc(ipt.ServeHTTP, l))
	}
}

func (ipt *Input) Run() {
	ipt.logger = logger.SLogger(inputName)
	ipt.logger.Info("gitlab input started")

	if !ipt.EnableCollect {
		ipt.logger.Infof("metric collecting is disabled, gitlab exited")
		return
	}

	if err := ipt.setup(); err != nil {
		ipt.logger.Errorf("setup failed: %s", err)
		ipt.logger.Info("gitlab input stopped")
		return
	}

	start := ntp.Now()
	tick := time.NewTicker(ipt.interval)
	defer tick.Stop()

	ipt.logger.Infof("start collecting metrics from %s with interval %v", ipt.URL, ipt.interval)

	for {
		if !ipt.pause.Load() {
			ipt.scrape(start.UnixNano())
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.logger.Info("gitlab input exiting")
			return

		case <-ipt.semStop.Wait():
			ipt.logger.Info("gitlab input stopped")
			return

		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.interval)
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}

	if ipt.EnableCIVisibility {
		httpapi.RemoveHTTPRoute("POST", "/v1/gitlab")
	}
}

func (ipt *Input) Pause() error {
	ipt.pause.Store(true)
	return nil
}

func (ipt *Input) Resume() error {
	ipt.pause.Store(false)
	return nil
}

func (ipt *Input) setup() error {
	ipt.logger.Debug("setting up gitlab input")

	if err := ipt.parseURL(); err != nil {
		return fmt.Errorf("parse URL failed: %w", err)
	}

	tags := ipt.buildTags()
	if err := ipt.setupAuth(); err != nil {
		return fmt.Errorf("setup auth failed: %w", err)
	}

	opts := ipt.buildScraperOptions(tags)
	scraper, err := promscrape.NewPromScraper(opts...)
	if err != nil {
		return fmt.Errorf("create scraper failed: %w", err)
	}

	ipt.scraper = scraper

	// Parse interval
	dur, err := time.ParseDuration(ipt.Interval)
	if err != nil {
		return fmt.Errorf("parse interval error: %w", err)
	}
	ipt.interval = config.ProtectedInterval(time.Second, time.Minute*5, dur)

	ipt.logger.Debugf("setup completed with interval %v", ipt.interval)
	return nil
}

func (ipt *Input) parseURL() error {
	endpoint, err := url.Parse(ipt.URL)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", ipt.URL, err)
	}
	ipt.endpoint = endpoint
	ipt.logger.Debugf("parsed URL: %s", ipt.URL)
	return nil
}

func (ipt *Input) buildTags() map[string]string {
	tags := make(map[string]string)

	var globalTags map[string]string
	if ipt.Election {
		globalTags = ipt.Tagger.ElectionTags()
		ipt.logger.Debugf("using election tags: %v", globalTags)
	} else {
		globalTags = ipt.Tagger.HostTags()
		ipt.logger.Debugf("using host tags: %v", globalTags)
	}

	mergedTags := inputs.MergeTags(globalTags, ipt.Tags, ipt.URL)
	for k, v := range mergedTags {
		tags[k] = v
	}

	// Always add instance tag for GitLab
	if _, ok := mergedTags["instance"]; !ok {
		tags["instance"] = ipt.endpoint.Host
		ipt.logger.Debugf("added instance tag: %s", ipt.endpoint.Host)
	}

	ipt.logger.Debugf("built tags: %v", tags)
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
		ipt.logger.Debug("added bearer token authorization header")
	}

	return nil
}

func (ipt *Input) buildScraperOptions(tags map[string]string) []promscrape.Option {
	opts := []promscrape.Option{
		promscrape.WithSource("gitlab"),
		promscrape.WithMeasurement("gitlab"), // placeholder measurement name, final name is set in callback
		promscrape.WithHTTPHeader(ipt.HTTPHeaders),
		promscrape.WithExtraTags(tags),
		promscrape.KeepExistMetricName(true),
		promscrape.HonorTimestamps(false),
		promscrape.WithCallback(ipt.callback),
	}

	// Add TLS config if provided
	if ipt.hasTLSConfig() {
		caCerts := []string{}
		if ipt.TLSCA != "" {
			caCerts = append(caCerts, ipt.TLSCA)
		}
		if len(ipt.CaCerts) > 0 {
			caCerts = append(caCerts, ipt.CaCerts...)
		}

		opts = append(opts,
			promscrape.WithTLSOpen(true),
			promscrape.WithCacertFiles(caCerts),
			promscrape.WithCertFile(ipt.TLSCert),
			promscrape.WithKeyFile(ipt.TLSKey),
			promscrape.WithInsecureSkipVerify(ipt.InsecureSkipVerify),
		)
		ipt.logger.Debug("TLS configuration enabled")
	}

	return opts
}

func (ipt *Input) hasTLSConfig() bool {
	return ipt.TLSCA != "" || ipt.TLSCert != "" || ipt.TLSKey != "" || len(ipt.CaCerts) > 0 || ipt.InsecureSkipVerify
}

func (ipt *Input) scrape(ptTS int64) {
	ipt.lastStart = time.Now()

	if ipt.scraper == nil {
		ipt.logger.Error("scraper not initialized")
		return
	}

	ipt.logger.Debugf("start scraping metrics from %s", ipt.URL)

	// Set timestamp for scraper
	ipt.scraper.SetTimestamp(ptTS)
	if err := ipt.scraper.ScrapeURL(ipt.URL); err != nil {
		ipt.logger.Warnf("scrape failed: %s", err)
	}
}

func (ipt *Input) callback(pts []*point.Point) error {
	if len(pts) == 0 {
		ipt.logger.Debug("no points collected")
		return nil
	}

	ipt.logger.Debugf("processing %d collected points", len(pts))

	// Process each point to set measurement name and field name based on metric name.
	var processedPts []*point.Point
	measurementStats := make(map[string]int)

	for _, pt := range pts {
		tags := pt.MapTags()

		for _, field := range pt.Fields() {
			if field == nil || field.IsTag {
				continue
			}

			measurementName, fieldName := getMeasurementAndFieldNameFromMetric(field.Key)
			measurementStats[measurementName]++

			allFields := append(
				point.NewTags(tags),
				point.NewKVs(map[string]interface{}{fieldName: field.Raw()})...,
			)

			newPt := point.NewPoint(
				measurementName,
				allFields,
				point.WithTime(pt.Time()),
			)

			processedPts = append(processedPts, newPt)
		}
	}

	cost := time.Since(ipt.lastStart)

	// Log measurement distribution
	for measurement, count := range measurementStats {
		ipt.logger.Debugf("%s: %d points", measurement, count)
	}

	if err := ipt.feeder.Feed(
		point.Metric,
		processedPts,
		dkio.WithCollectCost(cost),
		dkio.WithSource("gitlab"),
		dkio.WithElection(ipt.Election),
	); err != nil {
		ipt.logger.Warnf("feed metrics failed: %s", err)
		return err
	}

	ipt.logger.Infof("collected %d points in %v", len(processedPts), cost)

	return nil
}

func getMeasurementAndFieldNameFromMetric(metricName string) (measurementName, fieldName string) {
	switch {
	case strings.HasPrefix(metricName, "gitlab_"):
		return "gitlab", strings.TrimPrefix(metricName, "gitlab_")
	case strings.HasPrefix(metricName, "http_"):
		return "gitlab_http", metricName
	default:
		return "gitlab_base", metricName
	}
}

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) Catalog() string { return catalog }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&gitlabMeasurement{},
		&gitlabBaseMeasurement{},
		&gitlabHTTPMeasurement{},
		&gitlabPipelineMeasurement{},
		&gitlabJobMeasurement{},
	}
}

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }
