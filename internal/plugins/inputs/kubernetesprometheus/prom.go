// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

type promTarget struct {
	urlstr string
	pm     *iprom.Prom

	shouldScrape func() bool
	lastTime     time.Time
}

func newPromTarget(ctx context.Context, urlstr string, interval time.Duration, election bool, opts []iprom.PromOption) (*promTarget, error) {
	var err error
	p := promTarget{urlstr: urlstr}

	p.pm, err = iprom.NewProm(opts...)
	if err != nil {
		return nil, err
	}

	p.shouldScrape = func() bool {
		if election {
			paused, exists := pauseFrom(ctx)
			if exists && paused {
				return false
			}
		}

		if p.lastTime.IsZero() {
			p.lastTime = time.Now()
			return true
		}
		if time.Since(p.lastTime) < interval {
			return false
		}

		return true
	}

	return &p, nil
}

func (p *promTarget) url() string { return p.urlstr }
func (p *promTarget) scrape() error {
	if !p.shouldScrape() {
		return nil
	}
	p.lastTime = time.Now()
	_, err := p.pm.CollectFromHTTPV2(p.urlstr)
	return err
}

func buildPromOptions(role Role, key string, feeder dkio.Feeder, opts ...iprom.PromOption) []iprom.PromOption {
	name := string(role) + "::" + key

	callbackFn := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		if err := feeder.FeedV2(
			point.Metric,
			pts,
			dkio.WithInputName(name),
			dkio.DisableGlobalTags(true),
			dkio.WithElection(true),
		); err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}

		collectPtsCounter.WithLabelValues(string(role), key).Add(float64(len(pts)))
		return nil
	}

	res := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(name),
		iprom.WithMaxBatchCallback(1, callbackFn),
	}
	res = append(res, opts...)
	return res
}

func buildPromOptionsWithAuth(auth *Auth) ([]iprom.PromOption, error) {
	var opts []iprom.PromOption

	if auth.BearerTokenFile != "" {
		token, err := os.ReadFile(auth.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, iprom.WithBearerToken(string(token)))
	}

	if auth.TLSConfig != nil {
		opts = append(
			opts,
			iprom.WithTLSOpen(true),
			iprom.WithCacertFiles(auth.TLSConfig.CaCerts),
			iprom.WithCertFile(auth.TLSConfig.Cert),
			iprom.WithKeyFile(auth.TLSConfig.CertKey),
			iprom.WithInsecureSkipVerify(auth.TLSConfig.InsecureSkipVerify),
		)
	}

	return opts, nil
}
