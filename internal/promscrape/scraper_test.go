// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

const (
	mockHeader = `
# HELP datakit_http_worker_number The number of the worker
# TYPE datakit_http_worker_number gauge
`
	mockBody = `
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 11.0 1755681983
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 12.2 1755681983
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 13.0 1755681983
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 14.2 1755681983
datakit_http_worker_number{category="metric",domain="dataway.testing.com",status="%d", } 15.0 1755681983
`
)

func TestParseStream(t *testing.T) {
	count := 0
	run := func() {
		var buf bytes.Buffer
		buf.WriteString(mockHeader)
		for i := 0; i < 100; i++ {
			buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
		}
		p := &PromScraper{
			opt: &option{
				measurement:     "testing-meas",
				honorTimestamps: true,
				extraTags:       map[string]string{"key-01": "value-01"},
				callback: func(pts []*point.Point) error {
					for _, pt := range pts {
						t.Log(pt.Pretty())
					}
					count += len(pts)
					return nil
				},
			},
		}
		err := p.ParserStream(&buf)
		assert.NoError(t, err)
	}

	run()
	t.Logf("count: %d\n", count)
}

func BenchmarkParseStream(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(mockHeader)
	for i := 0; i < 10000; i++ {
		buf.WriteString(fmt.Sprintf(mockBody, i, i, i, i, i))
	}

	p := &PromScraper{
		opt: &option{
			measurement: "testing-meas",
			extraTags:   map[string]string{"key-01": "value-01"},
			callback: func(pts []*point.Point) error {
				return nil
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := p.ParserStream(&buf)
		assert.NoError(b, err)
	}
}

func TestSplitMetricName(t *testing.T) {
	cases := []struct {
		inMeasurement         string
		inName                string
		inKeepExistMetricName bool

		outMeasurementName string
		outFieldName       string
	}{
		{
			inMeasurement:         "",
			inName:                "etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "etcd",
			outFieldName:          "write_bytes_total",
		},
		{
			inMeasurement:         "set-measurement",
			inName:                "etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "set-measurement",
			outFieldName:          "write_bytes_total",
		},
		{
			inMeasurement:         "set-measurement",
			inName:                "etcd_write_bytes_total",
			inKeepExistMetricName: true,
			outMeasurementName:    "set-measurement",
			outFieldName:          "etcd_write_bytes_total",
		},
		{
			inName:                "_",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "__",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "etcd_",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "_etcd",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
		{
			inName:                "_etcd_write_bytes_total",
			inKeepExistMetricName: false,
			outMeasurementName:    "unknown",
			outFieldName:          "unknown",
		},
	}

	for _, tc := range cases {
		p := PromScraper{
			opt: &option{
				measurement:         tc.inMeasurement,
				keepExistMetricName: tc.inKeepExistMetricName,
			},
		}

		measurementName, fieldName := p.splitMetricName(tc.inName)

		assert.Equal(t, tc.outMeasurementName, measurementName)
		assert.Equal(t, tc.outFieldName, fieldName)
	}
}

func TestScrapeLimits(t *testing.T) {
	t.Run("body size", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprintln(w, "sample_metric 1")
			_, _ = fmt.Fprintln(w, "sample_metric 2")
		}))
		defer server.Close()

		scraper, err := NewPromScraper(
			WithMaxBodySize(10),
			WithCallback(func([]*point.Point) error { return nil }),
		)
		assert.NoError(t, err)
		err = scraper.ScrapeURL(server.URL)
		assert.ErrorContains(t, err, errBodySizeLimit.Error())
	})

	t.Run("sample count", func(t *testing.T) {
		scraper, err := NewPromScraper(
			WithMaxSamples(1),
			WithCallback(func([]*point.Point) error { return nil }),
		)
		assert.NoError(t, err)
		err = scraper.ParserStream(strings.NewReader("sample_metric 1\nsample_metric 2\n"))
		assert.ErrorContains(t, err, "sample limit exceeded")
	})

	t.Run("label count", func(t *testing.T) {
		scraper, err := NewPromScraper(
			WithMaxLabels(1),
			WithCallback(func([]*point.Point) error { return nil }),
		)
		assert.NoError(t, err)
		err = scraper.ParserStream(strings.NewReader("sample_metric{a=\"1\",b=\"2\"} 1\n"))
		assert.ErrorContains(t, err, "label limit exceeded")
	})
}

func TestScrapeTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
			_, _ = fmt.Fprintln(w, "sample_metric 1")
		}
	}))
	defer server.Close()

	scraper, err := NewPromScraper(
		WithRequestTimeout(50*time.Millisecond),
		WithCallback(func([]*point.Point) error { return nil }),
	)
	assert.NoError(t, err)

	start := time.Now()
	err = scraper.ScrapeURL(server.URL)
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 500*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = scraper.ScrapeURLWithContext(ctx, server.URL)
	assert.ErrorIs(t, err, context.Canceled)
}
