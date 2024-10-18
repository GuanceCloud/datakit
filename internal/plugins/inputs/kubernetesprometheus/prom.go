// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
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
	terminated  atomic.Bool

	interval time.Duration
	lastTime time.Time
}

func newPromScraper(
	role Role,
	key string,
	urlstr string,
	interval time.Duration,
	checkPaused func() bool,
	opts []promscrape.Option,
) (*promScraper, error) {
	var err error
	p := promScraper{
		role:        string(role),
		key:         key,
		urlstr:      urlstr,
		checkPaused: checkPaused,
		interval:    interval,
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
func (p *promScraper) isTerminated() bool { return p.terminated.Load() }
func (p *promScraper) markAsTerminated()  { p.terminated.Store(true) }

func (p *promScraper) shouldScrape() bool {
	if p.lastTime.IsZero() {
		p.lastTime = time.Now()
		return true
	}
	if time.Since(p.lastTime) < p.interval {
		return false
	}
	if p.checkPaused != nil {
		paused := p.checkPaused()
		return !paused
	}
	return true
}

func (p *promScraper) scrape() error {
	p.lastTime = time.Now()
	err := p.pm.ScrapeURL(p.urlstr)
	scrapeTargetCost.WithLabelValues(p.role, p.key, p.remote).Observe(float64(time.Since(p.lastTime)) / float64(time.Second))
	return err
}

func buildPromOptions(role Role, key string, feeder dkio.Feeder, opts ...promscrape.Option) []promscrape.Option {
	source := "kubernetesprometheus/" + string(role)
	remote := key

	callbackFn := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		if err := feeder.FeedV2(
			point.Metric,
			pts,
			dkio.WithInputName(source),
			dkio.DisableGlobalTags(true),
			dkio.WithElection(true),
		); err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}

		collectPtsCounter.WithLabelValues(string(role), key).Add(float64(len(pts)))
		return nil
	}

	res := []promscrape.Option{
		promscrape.WithSource(source),
		promscrape.WithRemote(remote),
		promscrape.WithCallback(callbackFn),
	}
	res = append(res, opts...)
	return res
}

func buildPromOptionsWithAuth(auth *Auth) ([]promscrape.Option, error) {
	var opts []promscrape.Option

	if auth.BearerTokenFile != "" {
		token, err := os.ReadFile(auth.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, promscrape.WithBearerToken(string(token)))
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
