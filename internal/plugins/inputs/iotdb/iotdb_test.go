// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package iotdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const testMetrics = `# HELP iotdb_requests_total Number of requests.
# TYPE iotdb_requests_total counter
iotdb_requests_total{node="dn-1",region="east"} 12 1700000000000
# HELP iotdb_latency_seconds Request latency.
# TYPE iotdb_latency_seconds summary
iotdb_latency_seconds{quantile="0.5",node="dn-1"} 0.25
iotdb_latency_seconds{quantile="0.9",node="dn-1"} 0.8
iotdb_latency_seconds_sum{node="dn-1"} 2.5
iotdb_latency_seconds_count{node="dn-1"} 5
# HELP iotdb_size_bytes Storage size.
# TYPE iotdb_size_bytes gauge
iotdb_size_bytes{node="dn-1"} 42.5
# HELP iotdb_request_duration_seconds Request duration.
# TYPE iotdb_request_duration_seconds histogram
iotdb_request_duration_seconds_bucket{le="0.5",node="dn-1"} 2
iotdb_request_duration_seconds_bucket{le="+Inf",node="dn-1"} 5
iotdb_request_duration_seconds_sum{node="dn-1"} 2.5
iotdb_request_duration_seconds_count{node="dn-1"} 5
`

func TestInputRegistered(t *testing.T) {
	creator, ok := inputs.AllInputs[inputName]
	if !ok {
		t.Fatalf("input %q was not registered", inputName)
	}
	if _, ok := creator().(*Input); !ok {
		t.Fatalf("input %q factory returned an unexpected type", inputName)
	}
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default", want: defaultEndpoint},
		{name: "host adds scheme and metrics path", input: "iotdb.local:9091", want: "http://iotdb.local:9091/metrics"},
		{name: "preserves configured path", input: "https://iotdb.local/custom", want: "https://iotdb.local/custom"},
		{name: "normalizes scheme case", input: "HTTPS://iotdb.local/metrics", want: "https://iotdb.local/metrics"},
		{name: "rejects unsupported scheme", input: "ftp://iotdb.local/metrics", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEndpoint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePrometheusMetrics(t *testing.T) {
	scrapedAt := time.Unix(1_700_000_001, 0)
	samples, err := parsePrometheusMetrics([]byte(testMetrics), scrapedAt)
	if err != nil {
		t.Fatalf("parsePrometheusMetrics() error = %v", err)
	}
	if len(samples) != 10 {
		t.Fatalf("got %d samples, want 10", len(samples))
	}

	byName := make(map[string][]metricSample)
	for _, sample := range samples {
		byName[sample.name] = append(byName[sample.name], sample)
	}
	counter := byName["iotdb_requests_total"]
	if len(counter) != 1 || counter[0].value != 12 || counter[0].labels["node"] != "dn-1" {
		t.Fatalf("counter sample = %#v", counter)
	}
	if want := time.UnixMilli(1_700_000_000_000); !counter[0].timestamp.Equal(want) {
		t.Fatalf("counter timestamp = %s, want %s", counter[0].timestamp, want)
	}
	if len(byName["iotdb_latency_seconds"]) != 2 || byName["iotdb_latency_seconds"][0].labels["quantile"] == "" {
		t.Fatalf("summary quantile samples = %#v", byName["iotdb_latency_seconds"])
	}
	if len(byName["iotdb_latency_seconds_sum"]) != 1 || byName["iotdb_latency_seconds_sum"][0].value != 2.5 {
		t.Fatalf("summary sum = %#v", byName["iotdb_latency_seconds_sum"])
	}
	if len(byName["iotdb_latency_seconds_count"]) != 1 || byName["iotdb_latency_seconds_count"][0].value != 5 {
		t.Fatalf("summary count = %#v", byName["iotdb_latency_seconds_count"])
	}
	if len(byName["iotdb_size_bytes"]) != 1 || byName["iotdb_size_bytes"][0].timestamp != scrapedAt {
		t.Fatalf("gauge sample = %#v", byName["iotdb_size_bytes"])
	}
	buckets := byName["iotdb_request_duration_seconds_bucket"]
	if len(buckets) != 2 || buckets[0].labels["le"] != "0.5" || buckets[1].labels["le"] != "+Inf" {
		t.Fatalf("histogram buckets = %#v", buckets)
	}
	if len(byName["iotdb_request_duration_seconds_count"]) != 1 || byName["iotdb_request_duration_seconds_count"][0].value != 5 {
		t.Fatalf("histogram count = %#v", byName["iotdb_request_duration_seconds_count"])
	}
}

func TestScrapeMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metrics" {
			t.Errorf("request path = %q, want /metrics", r.URL.Path)
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != "proxy-user" || password != "proxy-password" {
			t.Errorf("request did not use configured Basic Auth")
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(testMetrics))
	}))
	defer server.Close()

	samples, err := scrapeMetrics(context.Background(), server.Client(), server.URL, "proxy-user", "proxy-password", time.Now())
	if err != nil {
		t.Fatalf("scrapeMetrics() error = %v", err)
	}
	if len(samples) != 10 {
		t.Fatalf("got %d samples, want 10", len(samples))
	}
}

type captureFeeder struct {
	category point.Category
	points   []*point.Point
}

func (f *captureFeeder) Feed(category point.Category, points []*point.Point, _ ...dkio.FeedOption) error {
	f.category = category
	f.points = append(f.points, points...)
	return nil
}

func (*captureFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestCollectFeedsIoTDBSamples(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testMetrics))
	}))
	defer server.Close()

	feeder := &captureFeeder{}
	ipt := &Input{
		Endpoint: server.URL,
		Tags:     map[string]string{"environment": "test"},
		Tagger:   datakit.DefaultGlobalTagger(),
		feeder:   feeder,
		client:   server.Client(),
	}
	if err := ipt.collect(context.Background()); err != nil {
		t.Fatalf("collect() error = %v", err)
	}
	if feeder.category != point.Metric || len(feeder.points) != 10 {
		t.Fatalf("Feed received category=%v points=%d, want metric category and 10 points", feeder.category, len(feeder.points))
	}
	for _, pt := range feeder.points {
		line := pt.LineProto()
		if strings.Contains(line, "metric_name=iotdb_requests_total") {
			if !strings.Contains(line, "node=dn-1") || !strings.Contains(line, "environment=test") || !strings.Contains(line, "value=12") {
				t.Fatalf("counter point is missing expected labels or value: %s", line)
			}
			return
		}
	}
	t.Fatal("feed did not contain the IoTDB counter sample")
}

func TestScrapeMetricsRejectsHTTPErrorAndOversizedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/failure" {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(strings.Repeat("x", maxMetricsResponseBytes+1)))
	}))
	defer server.Close()

	if _, err := scrapeMetrics(context.Background(), server.Client(), server.URL+"/failure", "", "", time.Now()); err == nil {
		t.Fatal("expected non-2xx response to fail")
	}
	if _, err := scrapeMetrics(context.Background(), server.Client(), server.URL+"/oversized", "", "", time.Now()); err == nil {
		t.Fatal("expected oversized metrics response to fail")
	}
}

func TestParsePrometheusMetricsRejectsMalformedInput(t *testing.T) {
	if _, err := parsePrometheusMetrics([]byte("# TYPE x gauge\nx nope\n"), time.Now()); err == nil {
		t.Fatal("expected malformed exposition to fail")
	}
}

func TestRemoteReporter(t *testing.T) {
	endpoint := os.Getenv("DATAKIT_IOTDB_METRICS_URL")
	if endpoint == "" {
		t.Skip("set DATAKIT_IOTDB_METRICS_URL to the IoTDB Prometheus reporter endpoint")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	feeder := &captureFeeder{}
	ipt := &Input{Endpoint: endpoint, Tagger: datakit.DefaultGlobalTagger(), feeder: feeder, client: &http.Client{Timeout: 12 * time.Second}}
	defer ipt.client.CloseIdleConnections()
	if err := ipt.collect(ctx); err != nil {
		t.Fatal(err)
	}
	if feeder.category != point.Metric || len(feeder.points) == 0 {
		t.Fatalf("no metric points delivered from IoTDB (category %v)", feeder.category)
	}
	for _, pt := range feeder.points {
		line := pt.LineProto()
		if strings.Contains(line, "metric_name=") && strings.Contains(line, "value=") && strings.Contains(line, "nodeType=") {
			t.Logf("collected %d points, including a labelled IoTDB node metric", len(feeder.points))
			return
		}
	}
	t.Fatal("no labelled IoTDB node metrics delivered")
}
