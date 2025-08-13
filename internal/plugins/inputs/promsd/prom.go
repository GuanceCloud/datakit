// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promsd

import (
	"bytes"
	"net"
	"net/url"
	"os"
	"sync/atomic"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/promscrape"
)

type scraper interface {
	targetURL() string
	scrape(timestamp int64) error
	isTerminated() bool
	markAsTerminated()
}

type promScraper struct {
	urlstr     string
	pm         *promscrape.PromScraper
	terminated atomic.Bool
}

func newPromScraper(urlstr string, opts []promscrape.Option) (*promScraper, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	p := promScraper{urlstr: u.String()}

	tags := map[string]string{
		"host":     splitHost(u.Host),
		"instance": u.Host,
	}

	p.pm, err = promscrape.NewPromScraper(append(opts, promscrape.WithExtraTags(tags))...)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *promScraper) targetURL() string  { return p.urlstr }
func (p *promScraper) isTerminated() bool { return p.terminated.Load() }
func (p *promScraper) markAsTerminated() {
	p.terminated.Store(true)
}

func (p *promScraper) scrape(defaultTimestamp int64) error {
	p.pm.SetTimestamp(defaultTimestamp)
	err := p.pm.ScrapeURL(p.urlstr)
	return err
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

	if auth.TLSClientConfig != nil && len(auth.TLSClientConfig.CaCerts) > 0 {
		opts = append(
			opts,
			promscrape.WithTLSOpen(true),
			promscrape.WithCacertFiles(auth.TLSClientConfig.CaCerts),
			promscrape.WithCertFile(auth.TLSClientConfig.Cert),
			promscrape.WithKeyFile(auth.TLSClientConfig.CertKey),
			promscrape.WithInsecureSkipVerify(auth.TLSClientConfig.InsecureSkipVerify),
		)
	}

	return opts, nil
}

func splitHost(remote string) string {
	host := remote

	// try get 'host' tag from remote URL.
	if u, err := url.Parse(remote); err == nil && u.Host != "" { // like scheme://host:[port]/...
		host = u.Host
		if ip, _, err := net.SplitHostPort(u.Host); err == nil {
			host = ip
		}
	} else { // not URL, only IP:Port
		if ip, _, err := net.SplitHostPort(remote); err == nil {
			host = ip
		}
	}

	if host == "localhost" || net.ParseIP(host).IsLoopback() {
		return ""
	}

	return host
}
