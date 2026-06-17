// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promscrape used to parsing promemetheuse exportor metrics.
package promscrape

import (
	"context"
	"errors"
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
	opt       *option
	client    *http.Client
	timestamp int64 // unit nanoseconds
}

func NewPromScraper(opts ...Option) (*PromScraper, error) {
	opt := defaultOption()
	for _, fn := range opts {
		fn(opt)
	}
	if opt.bearerTokenFile != "" {
		opt.bearerTokenSource = newCachedFileBearerTokenSource(opt.bearerTokenFile, opt.bearerTokenRefreshInterval)
	}

	client, err := buildHTTPClient(&opt.optionClientConn)
	if err != nil {
		return nil, err
	}

	return &PromScraper{
		opt:       opt,
		client:    client,
		timestamp: -1, // not set
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

	client := httpcli.Cli(clientOpts)
	client.Timeout = opt.requestTimeout
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.ResponseHeaderTimeout = opt.requestTimeout
	}
	return client, nil
}

func (p *PromScraper) SetTimestamp(timestamp int64) {
	p.timestamp = timestamp
}

func (p *PromScraper) ScrapeURL(u string) error {
	return p.ScrapeURLWithContext(context.Background(), u)
}

func (p *PromScraper) ScrapeURLWithContext(ctx context.Context, u string) error {
	req, err := p.newRequest(ctx, u)
	if err != nil {
		return &ScrapeError{URL: u, Err: err}
	}

	s := httpcli.GetTracer(p.opt.source, p.opt.remote, "")
	defer s.Metrics()
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), s.Trace()))

	resp, err := p.client.Do(req)
	if err != nil {
		return &ScrapeError{URL: u, Err: err}
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			p.resetBearerToken()
		}
		return &ScrapeError{URL: u, StatusCode: resp.StatusCode}
	}

	reader := io.Reader(resp.Body)
	if p.opt.maxBodySize > 0 {
		reader = &limitReader{reader: reader, remaining: p.opt.maxBodySize}
	}
	return p.ParserStream(reader)
}

func (p *PromScraper) ParserStream(in io.Reader) error {
	defaultTimestamp := time.Unix(0, 0).UnixNano() / 1e6
	isGzipped := false
	samples := 0

	return ParseStream(in, defaultTimestamp, isGzipped, func(rows []Row) error {
		if err := p.validateRows(rows, &samples); err != nil {
			return err
		}
		return p.callbackForRow(rows)
	})
}

func (p *PromScraper) validateRows(rows []Row, samples *int) error {
	if p.opt.maxSamples > 0 && *samples+len(rows) > p.opt.maxSamples {
		return fmt.Errorf("prometheus sample limit exceeded: %d", p.opt.maxSamples)
	}
	*samples += len(rows)

	for i := range rows {
		row := &rows[i]
		if p.opt.maxLabels > 0 && len(row.Tags)+len(p.opt.extraTags) > p.opt.maxLabels {
			return fmt.Errorf("prometheus label limit exceeded: %d", p.opt.maxLabels)
		}
	}
	return nil
}

func (p *PromScraper) callbackForRow(rows []Row) error {
	var pts []*point.Point
	opts := point.DefaultMetricOptions()

	for _, row := range rows {
		measurementName, metricName := p.splitMetricName(row.Metric)
		var kvs point.KVs
		kvs = kvs.Set(metricName, row.Value)

		for _, tag := range row.Tags {
			kvs = kvs.AddTag(tag.Key, tag.Value)
		}
		for key, value := range p.opt.extraTags {
			kvs = kvs.AddTag(key, value)
		}

		timeNs := p.timestamp
		if p.opt.honorTimestamps && row.Timestamp > 0 {
			// Convert it to nanoseconds.
			timeNs = row.Timestamp * int64(time.Millisecond)
		}

		pt := point.NewPoint(measurementName, kvs, append(opts, point.WithTimestamp(timeNs))...)
		pts = append(pts, pt)
	}

	return p.opt.callback(pts)
}

func (p *PromScraper) newRequest(ctx context.Context, u string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/plain;version=0.0.4;q=1,*/*;q=0.1")
	for k, v := range p.opt.httpHeaders {
		req.Header.Set(k, v)
	}

	if req.Header.Get("Authorization") == "" && p.opt.bearerTokenSource != nil {
		token, err := p.opt.bearerTokenSource.Token()
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

var errBodySizeLimit = errors.New("prometheus response body size limit exceeded")

type limitReader struct {
	reader    io.Reader
	remaining int64
}

func (r *limitReader) Read(buf []byte) (int, error) {
	if r.remaining <= 0 {
		var probe [1]byte
		n, err := r.reader.Read(probe[:])
		if n > 0 {
			return 0, errBodySizeLimit
		}
		return 0, err
	}
	if int64(len(buf)) > r.remaining {
		buf = buf[:r.remaining]
	}
	n, err := r.reader.Read(buf)
	r.remaining -= int64(n)
	return n, err
}

func (p *PromScraper) resetBearerToken() {
	if p.opt.bearerTokenSource != nil {
		p.opt.bearerTokenSource.ResetToken()
	}
}

func (p *PromScraper) splitMetricName(name string) (measurementName, metricName string) {
	// By default, measurement name and metric name are split according to the first '_' met.
	index := strings.Index(name, "_")

	switch index {
	case -1, 0, len(name) - 1:
		measurementName = "unknown"
		metricName = "unknown"
		return
	}

	measurementName = name[:index]
	metricName = name[index+1:]

	if p.opt.measurement != "" {
		measurementName = p.opt.measurement
	}

	// If the keepExistMetricName is true, keep the raw value for field names.
	if p.opt.keepExistMetricName {
		metricName = name
		return
	}
	return
}
