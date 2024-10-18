// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promscrape used to parsing promemetheuse exportor metrics.
package promscrape

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

type PromScraper struct {
	opt    *option
	client *http.Client
}

func NewPromScraper(opts ...Option) (*PromScraper, error) {
	opt := defaultOption()
	for _, fn := range opts {
		fn(opt)
	}

	client, err := buildHTTPClient(&opt.optionClientConn)
	if err != nil {
		return nil, err
	}

	return &PromScraper{
		opt:    opt,
		client: client,
	}, nil
}

func buildHTTPClient(opt *optionClientConn) (*http.Client, error) {
	clientOpts := httpcli.NewOptions()
	clientOpts.DialTimeout = opt.timeout
	clientOpts.DialKeepAlive = opt.keepAlive
	clientOpts.MaxIdleConns = 1
	clientOpts.MaxIdleConnsPerHost = 10

	if opt.tlsOpen {
		tlsconfig := dknet.TLSClientConfig{
			CaCerts:            opt.cacertFiles,
			Cert:               opt.certFile,
			CertKey:            opt.keyFile,
			InsecureSkipVerify: opt.insecureSkipVerify,
		}
		conf, err := tlsconfig.TLSConfig()
		if err != nil {
			return nil, fmt.Errorf("could not load tlsConfig %w", err)
		}
		clientOpts.TLSClientConfig = conf
	}

	return httpcli.Cli(clientOpts), nil
}

func (p *PromScraper) ScrapeURL(u string) error {
	req, err := p.newRequest(u)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code returned when scraping %q: %d", u, resp.StatusCode)
	}
	defer resp.Body.Close() //nolint

	return p.ParserStream(resp.Body)
}

func (p *PromScraper) ParserStream(in io.Reader) error {
	defaultTimestamp := time.Unix(0, 0).UnixNano() / 1e6
	isGzipped := false

	return ParseStream(in, defaultTimestamp, isGzipped, p.callbackForRow)
}

func (p *PromScraper) callbackForRow(rows []Row) error {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, row := range rows {
		measurementName, metricsName := p.splitMetricsName(row.Metric)
		var kvs point.KVs
		kvs = kvs.Add(metricsName, row.Value, false, true)

		for key, value := range p.opt.extraTags {
			kvs = kvs.AddTag(key, value)
		}
		for _, tag := range row.Tags {
			kvs = kvs.AddTag(tag.Key, tag.Value)
		}

		pts = append(pts, point.NewPointV2(measurementName, kvs, opts...))
	}

	return p.opt.callback(pts)
}

func (p *PromScraper) newRequest(u string) (*http.Request, error) {
	req, err := http.NewRequest("GET", u, nil)
	req.Header.Set("Accept", "text/plain;version=0.0.4;q=1,*/*;q=0.1")
	for k, v := range p.opt.httpHeaders {
		req.Header.Set(k, v)
	}

	s := httpcli.NewHTTPClientTraceStat(p.opt.source, p.opt.remote)
	defer s.Metrics()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), s.Trace()))

	return req, err
}

func (p *PromScraper) splitMetricsName(name string) (measurementName, metricsName string) {
	if p.opt.measurement != "" {
		return p.opt.measurement, name
	}

	startPosition := strings.IndexFunc(name, func(r rune) bool {
		return r != '_'
	})
	if startPosition == -1 || startPosition == len(name)-1 {
		return "unknown", "unknown"
	}

	name = name[startPosition:]
	// By default, measurement name and metric name are split according to the first '_' met.
	index := strings.Index(name, "_")

	switch index {
	case -1:
		return name, name
	case 0:
		return name[index:], name[index:]
	case len(name) - 1:
		return name[:index], name[:index]
	}

	// If the keepExistMetricName is true, keep the raw value for field names.
	if p.opt.keepExistMetricName {
		return name[:index], name
	}
	return name[:index], name[index+1:]
}
