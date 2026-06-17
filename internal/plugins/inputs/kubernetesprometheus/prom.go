// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type promScraper struct {
	role   string // e.g. pod/node
	urlstr string
	pm     *promscrape.PromScraper

	key      string // e.g. namespace/name
	job      string
	host     string // e.g. 172.16.10.10
	instance string // e.g. 172.16.10.10:8080
	remote   string // e.g. :8080/metrics
	feeder   dkio.Feeder

	checkPaused func() bool
	retryCount  int
	nextRetry   time.Time
	terminated  atomic.Bool
}

const aggregateMetricLabel = "all"

func newPromScraper(
	role Role,
	key string,
	urlstr string,
	measurement string,
	feeder dkio.Feeder,
	checkPaused func() bool,
	opts []promscrape.Option,
) (*promScraper, error) {
	var err error
	p := promScraper{
		role:        string(role),
		urlstr:      urlstr,
		key:         key,
		job:         key, // default use key
		feeder:      feeder,
		checkPaused: checkPaused,
	}
	if measurement != "" {
		p.job = measurement
	}

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	p.host = splitHost(u.Host)
	p.instance = u.Host
	p.remote = fmt.Sprintf(":%s%s", u.Port(), u.Path)

	p.pm, err = promscrape.NewPromScraper(opts...)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *promScraper) targetURL() string { return p.urlstr }
func (p *promScraper) resetRetryCount() {
	p.retryCount = 0
	p.nextRetry = time.Time{}
}
func (p *promScraper) isTerminated() bool { return p.terminated.Load() }

func (p *promScraper) markAsTerminated() {
	if p.terminated.Swap(true) {
		return
	}
	p.recordUp(0, 0)
}

func (p *promScraper) shouldScrape() bool {
	if p.checkPaused != nil {
		paused := p.checkPaused()
		return !paused
	}
	return true
}

func (p *promScraper) canScrape(now time.Time) bool {
	return p.nextRetry.IsZero() || !now.Before(p.nextRetry)
}

func (p *promScraper) recordFailure(interval time.Duration) (int, time.Time) {
	p.retryCount++
	delay := scrapeBackoff(interval, p.retryCount)
	p.nextRetry = ntp.Now().Add(delay + scrapeJitter(delay))
	return p.retryCount, p.nextRetry
}

func (p *promScraper) scrape(ctx context.Context, defaultTimestamp int64) error {
	start := time.Now()

	p.pm.SetTimestamp(defaultTimestamp)
	err := p.pm.ScrapeURLWithContext(ctx, p.urlstr)
	if err != nil {
		p.recordUp(0, defaultTimestamp)
	} else {
		p.recordUp(1, defaultTimestamp)
	}

	collectCostVec.WithLabelValues(p.role, aggregateMetricLabel, aggregateMetricLabel).
		Observe(float64(time.Since(start)) / float64(time.Second))
	return err
}

func (p *promScraper) recordUp(up int, timestamp int64) {
	var kvs point.KVs
	kvs = kvs.AddTag("job", p.job)
	kvs = kvs.AddTag("instance", p.instance)
	kvs = kvs.AddTag("host", p.host)
	kvs = kvs.Add("up", up)

	if timestamp == 0 {
		timestamp = ntp.Now().UnixNano()
	}

	pt := point.NewPoint("collector", kvs, append(point.DefaultMetricOptions(), point.WithTimestamp(timestamp))...)

	if err := p.feeder.Feed(
		point.Metric,
		[]*point.Point{pt},
		dkio.WithSource("kubernetesprometheus-collector"),
		dkio.WithElection(true),
	); err != nil {
		klog.Warnf("failed to feed collector metrics: %s, ignored", err)
	}
}

func buildPromOptions(role Role, key string, auth *Auth, feeder dkio.Feeder, opts ...promscrape.Option) []promscrape.Option {
	const source = "kubernetesprometheus"
	remote := string(role)

	callbackFn := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		if err := feeder.Feed(point.Metric, pts, dkio.WithSource(source)); err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}
		collectPtsVec.WithLabelValues(string(role), aggregateMetricLabel).Add(float64(len(pts)))
		return nil
	}

	res := []promscrape.Option{
		promscrape.WithSource(source),
		promscrape.WithRemote(remote),
		promscrape.WithCallback(callbackFn),
		promscrape.WithRequestTimeout(defaultScrapeTimeout),
		promscrape.WithMaxBodySize(defaultMaxScrapeSize),
		promscrape.WithMaxSamples(defaultMaxSamplesPerScrape),
		promscrape.WithMaxLabels(defaultMaxLabelsPerSample),
	}
	res = append(res, opts...)

	if tlsOpts, err := buildPromOptionsWithAuth(auth); err != nil {
		klog.Warnf("%s %s has unexpected tls config %s", role, key, err)
	} else {
		res = append(res, tlsOpts...)
	}

	return res
}

func buildPromOptionsWithAuth(auth *Auth) ([]promscrape.Option, error) {
	var opts []promscrape.Option

	if auth.BearerTokenFile != "" {
		opts = append(opts, promscrape.WithBearerTokenFile(auth.BearerTokenFile))
	}

	if auth.TLSConfig != nil {
		opts = append(
			opts,
			promscrape.WithTLSOpen(true),
			promscrape.WithCacertFiles(auth.TLSConfig.CaCerts),
			promscrape.WithCertFile(auth.TLSConfig.Cert),
			promscrape.WithKeyFile(auth.TLSConfig.CertKey),
			promscrape.WithInsecureSkipVerify(auth.TLSConfig.InsecureSkipVerify),
		)
	}

	return opts, nil
}
