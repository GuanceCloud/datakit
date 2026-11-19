// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package iotdb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const maxMetricsResponseBytes = 16 << 20

const (
	defaultEndpoint = "http://127.0.0.1:9091/metrics"
	defaultTimeout  = 10 * time.Second
)

type metricSample struct {
	name       string
	metricType string
	labels     map[string]string
	value      float64
	timestamp  time.Time
}

func normalizeEndpoint(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse endpoint: %w", err)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("endpoint must be an absolute HTTP or HTTPS URL")
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "/metrics"
	}
	return u.String(), nil
}

func scrapeMetrics(ctx context.Context, client *http.Client, endpoint, username, password string, scrapedAt time.Time) ([]metricSample, error) {
	url, err := normalizeEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create metrics request: %w", err)
	}
	request.Header.Set("Accept", "text/plain; version=0.0.4, application/openmetrics-text; version=1.0.0")
	if username != "" || password != "" {
		request.SetBasicAuth(username, password)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request IoTDB metrics endpoint: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("IoTDB metrics endpoint returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxMetricsResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read IoTDB metrics response: %w", err)
	}
	if len(body) > maxMetricsResponseBytes {
		return nil, fmt.Errorf("IoTDB metrics response exceeds %d bytes", maxMetricsResponseBytes)
	}

	samples, err := parsePrometheusMetrics(body, scrapedAt)
	if err != nil {
		return nil, fmt.Errorf("parse IoTDB metrics response: %w", err)
	}
	return samples, nil
}

func parsePrometheusMetrics(body []byte, scrapedAt time.Time) ([]metricSample, error) {
	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(families))
	for name := range families {
		names = append(names, name)
	}
	sort.Strings(names)

	samples := make([]metricSample, 0)
	for _, name := range names {
		family := families[name]
		if family == nil {
			continue
		}
		metricType := strings.ToLower(family.GetType().String())
		for _, metric := range family.GetMetric() {
			if metric == nil {
				continue
			}
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				if label != nil && label.Name != nil && label.Value != nil {
					labels[label.GetName()] = label.GetValue()
				}
			}
			ts := scrapedAt
			if metric.TimestampMs != nil {
				ts = time.UnixMilli(metric.GetTimestampMs())
			}

			switch family.GetType() {
			case dto.MetricType_COUNTER:
				samples = append(samples, metricSample{name, metricType, labels, metric.GetCounter().GetValue(), ts})
			case dto.MetricType_GAUGE:
				samples = append(samples, metricSample{name, metricType, labels, metric.GetGauge().GetValue(), ts})
			case dto.MetricType_UNTYPED:
				samples = append(samples, metricSample{name, metricType, labels, metric.GetUntyped().GetValue(), ts})
			case dto.MetricType_HISTOGRAM:
				histogram := metric.GetHistogram()
				if histogram == nil {
					return nil, fmt.Errorf("metric family %q has an empty histogram", name)
				}
				for _, bucket := range histogram.GetBucket() {
					bucketLabels := cloneLabels(labels)
					bucketLabels["le"] = formatPrometheusFloat(bucket.GetUpperBound())
					samples = append(samples, metricSample{name + "_bucket", metricType, bucketLabels, float64(bucket.GetCumulativeCount()), ts})
				}
				samples = append(samples,
					metricSample{name + "_sum", metricType, cloneLabels(labels), histogram.GetSampleSum(), ts},
					metricSample{name + "_count", metricType, cloneLabels(labels), float64(histogram.GetSampleCount()), ts})
			case dto.MetricType_SUMMARY:
				summary := metric.GetSummary()
				if summary == nil {
					return nil, fmt.Errorf("metric family %q has an empty summary", name)
				}
				for _, quantile := range summary.GetQuantile() {
					quantileLabels := cloneLabels(labels)
					quantileLabels["quantile"] = formatPrometheusFloat(quantile.GetQuantile())
					samples = append(samples, metricSample{name, metricType, quantileLabels, quantile.GetValue(), ts})
				}
				samples = append(samples,
					metricSample{name + "_sum", metricType, cloneLabels(labels), summary.GetSampleSum(), ts},
					metricSample{name + "_count", metricType, cloneLabels(labels), float64(summary.GetSampleCount()), ts})
			default:
				return nil, fmt.Errorf("metric family %q uses unsupported type %s", name, family.GetType().String())
			}
		}
	}
	return samples, nil
}

func cloneLabels(labels map[string]string) map[string]string {
	cloned := make(map[string]string, len(labels)+1)
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
}

func formatPrometheusFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
