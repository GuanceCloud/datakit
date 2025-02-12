// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type promScraper struct {
	role, key      string
	urlstr, remote string
	pm             *promscrape.PromScraper

	checkPaused func() bool
	retryCount  int
	terminated  atomic.Bool
}

func newPromScraper(
	role Role,
	key string,
	urlstr string,
	checkPaused func() bool,
	opts []promscrape.Option,
) (*promScraper, error) {
	var err error
	p := promScraper{
		role:        string(role),
		key:         key,
		urlstr:      urlstr,
		checkPaused: checkPaused,
	}

	u, err := url.Parse(urlstr)
	if err == nil {
		p.remote = fmt.Sprintf(":%s%s", u.Port(), u.Path)
	}

	p.pm, err = promscrape.NewPromScraper(opts...)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *promScraper) targetURL() string  { return p.urlstr }
func (p *promScraper) resetRetryCount()   { p.retryCount = 0 }
func (p *promScraper) isTerminated() bool { return p.terminated.Load() }
func (p *promScraper) markAsTerminated()  { p.terminated.Store(true) }

func (p *promScraper) shouldScrape() bool {
	if p.checkPaused != nil {
		paused := p.checkPaused()
		return !paused
	}
	return true
}

func (p *promScraper) scrape(defaultTimestamp int64) error {
	start := time.Now()
	p.pm.SetTimestamp(defaultTimestamp)
	err := p.pm.ScrapeURL(p.urlstr)
	collectCostVec.WithLabelValues(p.role, p.key, p.remote).Observe(float64(time.Since(start)) / float64(time.Second))
	return err
}

func (p *promScraper) shouldRetry(maxScrapeRetry int) (bool, int) {
	p.retryCount++
	if p.retryCount >= maxScrapeRetry {
		return false, p.retryCount
	}
	return true, p.retryCount
}

func buildPromOptions(role Role, key string, auth *Auth, feeder dkio.Feeder, opts ...promscrape.Option) []promscrape.Option {
	source := fmt.Sprintf("kubernetesprometheus/%s::%s", role, key)
	remote := key

	callbackFn := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		if err := feeder.FeedV2(
			point.Metric,
			pts,
			dkio.WithInputName(source),
		); err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}

		collectPtsVec.WithLabelValues(string(role), key).Add(float64(len(pts)))
		return nil
	}

	res := []promscrape.Option{
		promscrape.WithSource(source),
		promscrape.WithRemote(remote),
		promscrape.WithCallback(callbackFn),
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
		token, err := os.ReadFile(auth.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, promscrape.WithBearerToken(string(bytes.TrimSpace(token)), false))
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
