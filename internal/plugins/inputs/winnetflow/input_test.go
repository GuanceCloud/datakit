// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package winnetflow

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type flushTestFeeder struct {
	points []*point.Point
}

func (f *flushTestFeeder) Feed(_ point.Category, pts []*point.Point, _ ...dkio.FeedOption) error {
	f.points = append(f.points, pts...)
	return nil
}

func (*flushTestFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestNewInputInitializesProductionDependencies(t *testing.T) {
	ipt := NewInput()
	if ipt.feeder == nil {
		t.Fatal("NewInput returned a nil feeder")
	}
	if ipt.tagger == nil {
		t.Fatal("NewInput returned a nil tagger")
	}
	if ipt.semStop == nil {
		t.Fatal("NewInput returned a nil stop semaphore")
	}
}

func TestHTTPStatsHasAnomaly(t *testing.T) {
	if httpStatsHasAnomaly(httpStats{}) {
		t.Fatal("zero stats should not be anomalous")
	}
	for name, st := range map[string]httpStats{
		"dropped":           {dropped: 1},
		"parse_errors":      {parseErrors: 1},
		"missed_connection": {missedConn: 1},
		"missed_request":    {missedReq: 1},
		"evicted_request":   {evictedReq: 1},
		"dropped_request":   {droppedReq: 1},
		"requests_skipped":  {requestsSkipped: 1},
		"events_lost":       {session: etwSessionStats{eventsLost: 1}},
		"buffers_lost":      {session: etwSessionStats{realTimeBuffersLost: 1}},
	} {
		t.Run(name, func(t *testing.T) {
			if !httpStatsHasAnomaly(st) {
				t.Fatal("non-zero loss counter should be anomalous")
			}
		})
	}
}

func TestFlushPointsPublishesPartialAggregates(t *testing.T) {
	feeder := &flushTestFeeder{}
	ipt := NewInput()
	ipt.feeder = feeder
	ipt.agg = newFlowAggregator(time.Minute, defaultMaxFlows)
	ipt.httpAgg = newHTTPAggregator(defaultMaxHTTPRequests)
	ipt.agg.includeLoopback = true
	ipt.httpAgg.includeLoopback = true
	ipt.agg.onEvent(&flowEvent{
		ts: time.Now(), kind: evTCPDataSend, transport: "tcp", family: "IPv4",
		direction: directionOutgoing, pid: 42,
		srcIP: "127.0.0.1", srcPort: 50000, dstIP: "127.0.0.1", dstPort: 443,
		bytes: 1024,
	})
	ipt.httpAgg.onEvent(&httpEvent{
		ts: time.Now(), start: time.Now().Add(-time.Millisecond), family: "IPv4", pid: 42,
		srcIP: "127.0.0.1", srcPort: 8080, dstIP: "127.0.0.1", dstPort: 50001,
		method: "GET", path: "/final", status: 200,
	})

	ipt.flushPoints()
	if len(feeder.points) != 2 {
		t.Fatalf("fed points = %d, want 2", len(feeder.points))
	}
	ipt.flushPoints()
	if len(feeder.points) != 2 {
		t.Fatalf("empty second flush fed more points: %d", len(feeder.points))
	}
}

func TestIntervalParsing(t *testing.T) {
	ipt := NewInput()
	if got := ipt.interval(); got != defaultInterval {
		t.Fatalf("default interval = %s, want %s", got, defaultInterval)
	}

	ipt.Interval = "2m"
	if got := ipt.interval(); got != 2*time.Minute {
		t.Fatalf("interval = %s, want 2m", got)
	}

	ipt.Interval = "bogus"
	if got := ipt.interval(); got != defaultInterval {
		t.Fatalf("invalid interval should fall back, got %s", got)
	}

	ipt.Interval = "1ms"
	if got := ipt.interval(); got != minInterval {
		t.Fatalf("too small interval should clamp to %s, got %s", minInterval, got)
	}

	ipt.Interval = "1h"
	if got := ipt.interval(); got != maxInterval {
		t.Fatalf("too large interval should clamp to %s, got %s", maxInterval, got)
	}
}

func TestReadEnv(t *testing.T) {
	ipt := NewInput()
	ipt.ReadEnv(map[string]string{
		"ENV_INPUT_WINNETFLOW_TAGS":                  "a=b,c=d",
		"ENV_INPUT_WINNETFLOW_INTERVAL":              "90s",
		"ENV_INPUT_WINNETFLOW_ETW_BUFFER_SIZE_KB":    "128",
		"ENV_INPUT_WINNETFLOW_ETW_MIN_BUFFERS":       "16",
		"ENV_INPUT_WINNETFLOW_ETW_MAX_BUFFERS":       "512",
		"ENV_INPUT_WINNETFLOW_ETW_MAX_BUFFERS_BOGUS": "x",
		"ENV_INPUT_WINNETFLOW_MAX_FLOWS":             "1024",
		"ENV_INPUT_WINNETFLOW_ENABLE_HTTPFLOW":       "false",
		"ENV_INPUT_WINNETFLOW_MAX_HTTP_REQUESTS":     "2048",
		"ENV_INPUT_WINNETFLOW_HTTPFLOW_PATH_LIMIT":   "128",
	})
	if ipt.Tags["a"] != "b" || ipt.Tags["c"] != "d" {
		t.Fatalf("tags not parsed: %v", ipt.Tags)
	}
	if got := ipt.interval(); got != 90*time.Second {
		t.Fatalf("interval = %s, want 90s", got)
	}
	if ipt.EtwBufferSizeKB != 128 || ipt.EtwMinBuffers != 16 || ipt.EtwMaxBuffers != 512 {
		t.Fatalf("etw params not parsed: %+v", ipt)
	}
	if ipt.MaxFlows != 1024 {
		t.Fatalf("max_flows = %d, want 1024", ipt.MaxFlows)
	}
	if ipt.EnableHTTPFlow {
		t.Fatal("enable_httpflow should be false")
	}
	if ipt.MaxHTTPRequests != 2048 {
		t.Fatalf("max_http_requests = %d, want 2048", ipt.MaxHTTPRequests)
	}
	if ipt.HTTPFlowPathLimit != 128 {
		t.Fatalf("httpflow_path_limit = %d, want 128", ipt.HTTPFlowPathLimit)
	}
}

func TestHTTPFlowDefaults(t *testing.T) {
	ipt := NewInput()
	if !ipt.EnableHTTPFlow {
		t.Fatal("enable_httpflow should default to true")
	}
	if ipt.MaxHTTPRequests != defaultMaxHTTPRequests {
		t.Fatalf("max_http_requests default = %d, want %d", ipt.MaxHTTPRequests, defaultMaxHTTPRequests)
	}
	if ipt.HTTPFlowPathLimit != defaultHTTPPathLimit {
		t.Fatalf("httpflow_path_limit default = %d, want %d", ipt.HTTPFlowPathLimit, defaultHTTPPathLimit)
	}
}

func TestCapacityLimits(t *testing.T) {
	ipt := NewInput()
	ipt.MaxFlows = maxMaxFlows + 1
	ipt.MaxHTTPRequests = maxMaxHTTPRequests + 1
	if got := ipt.maxFlows(); got != maxMaxFlows {
		t.Fatalf("max_flows clamp = %d, want %d", got, maxMaxFlows)
	}
	if got := ipt.httpMaxRequests(); got != maxMaxHTTPRequests {
		t.Fatalf("max_http_requests clamp = %d, want %d", got, maxMaxHTTPRequests)
	}

	for _, tc := range []struct {
		entries int
		limit   int
		want    int
	}{
		{entries: -1, limit: 10, want: 0},
		{entries: 5, limit: 10, want: 5},
		{entries: 20, limit: 10, want: 10},
	} {
		if got := boundedMapHint(tc.entries, tc.limit); got != tc.want {
			t.Errorf("boundedMapHint(%d, %d) = %d, want %d", tc.entries, tc.limit, got, tc.want)
		}
	}
}

func TestETWConfigDefaults(t *testing.T) {
	cfg, adjusted := normalizeETWConfig(0, 0, 0)
	if adjusted {
		t.Fatal("default ETW config should not require adjustment")
	}
	if cfg.bufferSizeKB != defaultEtwBufferSizeKB || cfg.minBuffers != defaultEtwMinBuffers || cfg.maxBuffers != defaultEtwMaxBuffers {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestNormalizeETWConfig(t *testing.T) {
	tests := []struct {
		name                string
		buffer, min, max    int
		wantBuffer, wantMin uint32
		wantMax             uint32
		wantAdjusted        bool
	}{
		{name: "valid", buffer: 128, min: 16, max: 512, wantBuffer: 128, wantMin: 16, wantMax: 512},
		{name: "negative uses defaults", buffer: -1, min: -1, max: -1, wantBuffer: 64, wantMin: 8, wantMax: 256, wantAdjusted: true},
		{name: "lower bounds", buffer: 1, min: 1, max: 1, wantBuffer: 4, wantMin: 2, wantMax: 2, wantAdjusted: true},
		{name: "upper bounds and memory", buffer: int(^uint(0) >> 1), min: int(^uint(0) >> 1), max: int(^uint(0) >> 1), wantBuffer: 1024, wantMin: 256, wantMax: 256, wantAdjusted: true},
		{name: "inverted", buffer: 64, min: 512, max: 16, wantBuffer: 64, wantMin: 512, wantMax: 512, wantAdjusted: true},
		{name: "memory budget", buffer: 1024, min: 8, max: 4096, wantBuffer: 1024, wantMin: 8, wantMax: 256, wantAdjusted: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, adjusted := normalizeETWConfig(tc.buffer, tc.min, tc.max)
			if cfg.bufferSizeKB != tc.wantBuffer || cfg.minBuffers != tc.wantMin || cfg.maxBuffers != tc.wantMax {
				t.Fatalf("config = %+v, want buffer=%d min=%d max=%d", cfg, tc.wantBuffer, tc.wantMin, tc.wantMax)
			}
			if adjusted != tc.wantAdjusted {
				t.Fatalf("adjusted = %t, want %t", adjusted, tc.wantAdjusted)
			}
		})
	}
}

func TestNextCollectorRetryDelay(t *testing.T) {
	tests := []struct {
		current time.Duration
		want    time.Duration
	}{
		{collectorRestartDelay, 10 * time.Second},
		{30 * time.Second, collectorMaxRetryDelay},
		{collectorMaxRetryDelay, collectorMaxRetryDelay},
	}
	for _, tc := range tests {
		if got := nextCollectorRetryDelay(tc.current); got != tc.want {
			t.Errorf("nextCollectorRetryDelay(%s) = %s, want %s", tc.current, got, tc.want)
		}
	}
}

func TestCollectorSurvivedInterval(t *testing.T) {
	started := time.Unix(100, 0)
	interval := time.Minute
	if collectorSurvivedInterval(time.Time{}, started.Add(interval), interval) {
		t.Fatal("zero start time must not be healthy")
	}
	if collectorSurvivedInterval(started, started.Add(interval-time.Nanosecond), interval) {
		t.Fatal("collector became healthy before a full interval")
	}
	if !collectorSurvivedInterval(started, started.Add(interval), interval) {
		t.Fatal("collector should be healthy after a full interval")
	}
}
