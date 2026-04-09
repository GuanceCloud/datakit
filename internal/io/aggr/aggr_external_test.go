// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
)

const (
	defaultDatawayURL          = "http://127.0.0.1:9528"
	defaultDatawayMetricsURL   = "http://127.0.0.1:9090/metrics"
	defaultReceiverAddr        = "127.0.0.1:19528"
	externalTestTokenPrefix    = "tkn_external_aggr"
	externalMetricMeasurement  = "aggr_metric_case"
	externalLoggingMeasurement = "aggr_logging_case"
	externalTracingMeasurement = "aggr_trace_case"

	datawayAPIAggregate    = "/v1/aggregate"
	datawayAPITailSampling = "/v1/tail_sampling_v2"
	datawayAPIWriteMetric  = "/v1/write/metric"

	externalForwardWaitTimeout       = 90 * time.Second
	externalDerivedMetricWaitTimeout = 120 * time.Second
	externalWaitLogInterval          = 10 * time.Second
)

func TestProcessExternalLPTestdata(t *testing.T) {
	if testing.Short() {
		t.Skip("skip external lp testdata case in short mode")
	}

	datawayURL := getEnvOrDefault("AGGR_TEST_DATAWAY_URL", defaultDatawayURL)
	metricsURL := getEnvOrDefault("AGGR_TEST_DATAWAY_METRICS_URL", defaultDatawayMetricsURL)
	receiverAddr := getEnvOrDefault("AGGR_TEST_RECEIVER_ADDR", defaultReceiverAddr)
	configVersion := time.Now().UnixNano()
	token := strings.TrimSpace(os.Getenv("AGGR_TEST_TOKEN"))
	if token == "" {
		token = fmt.Sprintf("%s_%d", externalTestTokenPrefix, configVersion)
	}

	if !isEndpointReachable(datawayURL) {
		t.Skipf("dataway not reachable at %s, please start dataway first", datawayURL)
	}
	if !isEndpointReachable(metricsURL) {
		t.Skipf("dataway metrics not reachable at %s, please expose /metrics first", metricsURL)
	}

	metricsBefore, err := scrapePromMetrics(metricsURL)
	require.NoError(t, err)

	receiver := newForwardReceiver(t, receiverAddr, token)
	defer receiver.Close()
	t.Logf("receiver listening at %s, ensure dataway remote_host points to this URL", receiver.URL())
	t.Logf("external aggr test config: dataway=%s metrics=%s token=%s version=%d ttl=%s", datawayURL, metricsURL, token, configVersion, newExternalLPTailSamplingConfig(configVersion).Tracing.DataTTL)

	ag := &Aggregator{
		Endpoints:           []string{datawayURL + "?token=" + token},
		metricConfig:        newExternalLPMetricConfig(),
		metricEnabled:       true,
		tailSamplingConfig:  newExternalLPTailSamplingConfig(configVersion),
		tailSamplingEnabled: true,
		MaxRawBodySize:      1024 * 1024,
		Timeout:             5 * time.Second,
		Internal:            time.Second,
	}
	require.NoError(t, ag.metricConfig.Setup())
	require.NoError(t, ag.tailSamplingConfig.Init())
	ag.initHTTP()
	ag.sendTSConfigToDW()
	t.Cleanup(func() {
		t.Logf("receiver summary: %s", receiver.DebugSummary())
		if metricsAfter, err := scrapePromMetrics(metricsURL); err == nil {
			t.Logf("forward metric deltas at cleanup:\n%s", metricDeltaReport(metricsBefore, metricsAfter, forwardMetricChecks()))
		} else {
			t.Logf("scrape metrics at cleanup failed: %v", err)
		}
	})

	now := time.Now()
	cases := []struct {
		name     string
		category point.Category
		input    string
		pts      []*point.Point
		aggrCalc int
	}{
		{
			name:     "metric",
			category: point.Metric,
			input:    "synthetic-metric",
			pts:      buildExternalMetricPoints(now),
			aggrCalc: 6,
		},
		{
			name:     "logging",
			category: point.Logging,
			input:    "synthetic-logging",
			pts:      buildExternalLoggingPoints(now),
			aggrCalc: 3,
		},
		{
			name:     "tracing",
			category: point.Tracing,
			input:    "synthetic-tracing",
			pts:      buildExternalTracingPoints(now),
			aggrCalc: 5,
		},
	}
	expectedAggregatePointTotal := 0.0
	expectedTailSamplingInputTraceTotal := 0.0
	expectedTailSamplingKeptTraceTotal := 2.0
	expectedTailSamplingInputSpanTotal := 0.0
	expectedTailSamplingKeptSpanTotal := 4.0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pts := clonePoints(tc.pts)
			require.NotEmpty(t, pts)
			t.Logf("generated %d synthetic points for %s", len(pts), tc.name)
			for _, pt := range pts {
				pt.SetTime(now)
			}
			res, err := ag.Process(tc.category, tc.input, pts)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Greater(t, res.SelectedPoints, 0)
			t.Logf("process result: category=%s input=%s selected=%d consumed=%v batch_packages=%d tail_sampling_packages=%d",
				tc.category, tc.input, res.SelectedPoints, res.Consumed, res.BatchPackages, res.TailSamplingPackages)
			expectedAggregatePointTotal += float64(tc.aggrCalc)

			if tc.category == point.Tracing {
				expectedTailSamplingInputTraceTotal += float64(res.TailSamplingPackages)
				expectedTailSamplingInputSpanTotal += float64(len(pts))
				assert.True(t, res.Consumed)
				assert.Nil(t, res.Points)
				assert.Greater(t, res.TailSamplingPackages, 0)
				return
			}

			assert.False(t, res.Consumed)
			assert.NotNil(t, res.Points)
		})
	}

	waitFor(t, externalForwardWaitTimeout, 200*time.Millisecond, "wait dataway forwarded metric or tracing data", func() (bool, string) {
		failOnReceiverError(t, receiver, "waiting forwarded data")
		return receiver.RequestCount(datawayAPIWriteMetric) > 0 || receiver.RequestCount(point.URLTracing) > 0,
			receiver.DebugSummary()
	})
	t.Logf("receiver got forwarded requests: %s", receiver.DebugSummary())
	waitFor(t, externalDerivedMetricWaitTimeout, time.Second, "wait dataway forwarded tail-sampling derived metrics", func() (bool, string) {
		failOnReceiverError(t, receiver, "waiting tail-sampling derived metrics")
		done := receiver.FieldCount("trace_total_count") > 0 &&
			receiver.FieldCount("span_total_count") > 0 &&
			receiver.FieldCount("trace_kept_count") > 0 &&
			receiver.MeasurementCount("tail_sampling") > 0
		return done, receiver.DebugSummary()
	})

	waitFor(t, externalDerivedMetricWaitTimeout, time.Second, "wait dataway special metrics forwarded as points", func() (bool, string) {
		failOnReceiverError(t, receiver, "waiting special metric points")
		return hasAllFields(receiver, specialMetricFieldChecks()), receiver.DebugSummary()
	})

	forwardChecks := forwardMetricChecks()
	waitFor(t, externalDerivedMetricWaitTimeout, time.Second, "wait dataway forward metrics increase", func() (bool, string) {
		metricsAfter, err := scrapePromMetrics(metricsURL)
		if err != nil {
			return false, fmt.Sprintf("scrape metrics failed: %v", err)
		}
		return metricDeltasPositive(metricsBefore, metricsAfter, forwardChecks), metricDeltaReport(metricsBefore, metricsAfter, forwardChecks)
	})

	metricsAfter, err := scrapePromMetrics(metricsURL)
	require.NoError(t, err)
	assertMetricDeltasPositive(t, metricsBefore, metricsAfter, forwardChecks)
	t.Logf("forward metric deltas:\n%s", metricDeltaReport(metricsBefore, metricsAfter, forwardChecks))

	waitFor(t, externalDerivedMetricWaitTimeout, time.Second, "wait dataway forwarded aggregated metric/logging/tracing fields", func() (bool, string) {
		failOnReceiverError(t, receiver, "waiting aggregated forwarded fields")
		done := receiver.FieldCount("latency.count") > 0 &&
			receiver.FieldCount("message.count") > 0 &&
			receiver.FieldCount("span_id.count") > 0
		return done, receiver.DebugSummary()
	})

	receiver.AssertNoError(t)
	t.Logf("receiver summary: %s", receiver.DebugSummary())
	assert.Greater(t, receiver.PointCount(datawayAPIWriteMetric), 0)
	assert.Equal(t, int(expectedTailSamplingKeptSpanTotal), receiver.PointCount(point.URLTracing), "unexpected kept tracing point count")
	assert.Greater(t, receiver.FieldCount("latency.max"), 0, "missing metric aggregation field latency.max")
	assert.Greater(t, receiver.FieldCount("latency.count"), 0, "missing metric aggregation field latency.count")
	assert.Greater(t, receiver.FieldCount("message.count"), 0, "missing logging aggregation field")
	assert.Greater(t, receiver.FieldCount("span_id.count"), 0, "missing tracing aggregation field")
	assert.Greater(t, receiver.FieldCount("trace_total_count"), 0, "missing tail-sampling derived field trace_total_count")
	assert.Greater(t, receiver.FieldCount("trace_kept_count"), 0, "missing tail-sampling derived field trace_kept_count")
	assert.Greater(t, receiver.MeasurementCount("tail_sampling"), 0, "missing tail_sampling derived measurement")
	for _, field := range specialMetricFieldChecks() {
		assert.Greaterf(t, receiver.FieldCount(field), 0, "missing forwarded special metric field %s", field)
	}
	assertReceiverPointFieldValue(t, receiver, "dataway_aggregate", "dataway_http_aggr_point_total",
		map[string]string{"api": datawayAPIAggregate, "token": token}, expectedAggregatePointTotal)
	assertReceiverPointFieldValue(t, receiver, "dataway_aggregate", "dataway_http_tail_sampling_trace_total",
		map[string]string{"token": token}, expectedTailSamplingInputTraceTotal)
	assertReceiverPointFieldValue(t, receiver, "dataway_aggregate", "dataway_http_tail_sampling_span_total",
		map[string]string{"token": token}, expectedTailSamplingInputSpanTotal)
	assertReceiverPointFieldValue(t, receiver, "dataway_aggregate", "dataway_http_tail_sampling_packet_send_total",
		map[string]string{"token": token, "data_type": point.Tracing.String(), "result": "success"}, expectedTailSamplingKeptTraceTotal)
	assertReceiverPointFieldPositive(t, receiver, "dataway_aggregate", "dataway_http_api_body_size_bytes_total",
		map[string]string{"api": datawayAPIAggregate, "token": token})
	assertReceiverPointFieldPositive(t, receiver, "dataway_aggregate", "dataway_http_api_body_size_bytes_total",
		map[string]string{"api": datawayAPITailSampling, "token": token})
	assertReceiverPointFieldValue(t, receiver, externalMetricMeasurement, "latency.max",
		map[string]string{"service_name": "svc-a", "id": "1"}, 20)
	assertReceiverPointFieldValue(t, receiver, externalMetricMeasurement, "latency.max",
		map[string]string{"service_name": "svc-b", "id": "2"}, 7)
	assertReceiverPointFieldValue(t, receiver, externalMetricMeasurement, "latency.count",
		map[string]string{"service_name": "svc-a", "id": "1"}, 2)
	assertReceiverPointFieldValue(t, receiver, externalMetricMeasurement, "latency.count",
		map[string]string{"service_name": "svc-b", "id": "2"}, 1)
	assertReceiverPointFieldValue(t, receiver, externalLoggingMeasurement, "message.count",
		map[string]string{"session_id": "session-a"}, 2)
	assertReceiverPointFieldValue(t, receiver, externalLoggingMeasurement, "message.count",
		map[string]string{"session_id": "session-b"}, 1)
	assertReceiverPointFieldValue(t, receiver, externalTracingMeasurement, "span_id.count",
		map[string]string{"trace_id": "trace-error"}, 2)
	assertReceiverPointFieldValue(t, receiver, externalTracingMeasurement, "span_id.count",
		map[string]string{"trace_id": "trace-slow"}, 2)
	assertReceiverPointFieldValue(t, receiver, externalTracingMeasurement, "span_id.count",
		map[string]string{"trace_id": "trace-drop"}, 1)
	assertReceiverPointFieldValue(t, receiver, "tail_sampling", "trace_total_count", nil, 3)
	assertReceiverPointFieldValue(t, receiver, "tail_sampling", "trace_kept_count", nil, 2)
	assertReceiverPointFieldValue(t, receiver, "tail_sampling", "span_total_count", nil, expectedTailSamplingInputSpanTotal)
}

type forwardReceiver struct {
	srv   *http.Server
	ln    net.Listener
	base  string
	token string
	done  chan struct{}
	// decodeTracing controls whether /v1/write/tracing payloads are protobuf-decoded.
	decodeTracing bool

	closeOnce sync.Once

	mu         sync.Mutex
	reqCount   map[string]int
	pointCount map[string]int
	fieldCount map[string]int
	msrCount   map[string]int
	tokenCount map[string]int
	points     []*receivedPoint
	samples    []string
	errs       []string
}

type receivedPoint struct {
	path   string
	token  string
	name   string
	tags   map[string]string
	attrs  map[string]string
	fields map[string]float64
}

func newForwardReceiver(t *testing.T, addr, token string) *forwardReceiver {
	t.Helper()

	rec := &forwardReceiver{
		token:         token,
		done:          make(chan struct{}),
		decodeTracing: true,
		reqCount:      map[string]int{},
		pointCount:    map[string]int{},
		fieldCount:    map[string]int{},
		msrCount:      map[string]int{},
		tokenCount:    map[string]int{},
		points:        []*receivedPoint{},
		samples:       []string{},
		errs:          []string{},
	}

	ln, err := net.Listen("tcp", addr)
	require.NoErrorf(t, err, "listen receiver on %s failed", addr)
	rec.ln = ln
	rec.base = "http://" + ln.Addr().String()
	rec.srv = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec.handle(w, r)
		}),
	}
	go func() {
		defer close(rec.done)
		if err := rec.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			rec.appendErr(fmt.Sprintf("receiver serve failed: %v", err))
		}
	}()

	return rec
}

func (r *forwardReceiver) URL() string {
	return r.base
}

func (r *forwardReceiver) Close() {
	r.closeOnce.Do(func() {
		if r.ln != nil {
			_ = r.ln.Close()
		}

		if r.srv != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := r.srv.Shutdown(ctx); err != nil &&
				errors.Is(err, http.ErrServerClosed) {
				r.appendErr(fmt.Sprintf("receiver shutdown failed: %v", err))
			}
		}

		if r.done != nil {
			<-r.done
		}
	})
}

func (r *forwardReceiver) RequestCount(path string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reqCount[path]
}

func (r *forwardReceiver) PointCount(path string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pointCount[path]
}

func (r *forwardReceiver) FieldCount(field string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.fieldCount[field]
}

func (r *forwardReceiver) MeasurementCount(msr string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.msrCount[msr]
}

func (r *forwardReceiver) DebugSummary() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	keys := make([]string, 0, len(r.fieldCount))
	for key := range r.fieldCount {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, r.fieldCount[key]))
	}

	samples := append([]string(nil), r.samples...)

	msrKeys := make([]string, 0, len(r.msrCount))
	for key := range r.msrCount {
		msrKeys = append(msrKeys, key)
	}
	sort.Strings(msrKeys)

	msrParts := make([]string, 0, len(msrKeys))
	for _, key := range msrKeys {
		msrParts = append(msrParts, fmt.Sprintf("%s=%d", key, r.msrCount[key]))
	}

	tokenKeys := make([]string, 0, len(r.tokenCount))
	for key := range r.tokenCount {
		tokenKeys = append(tokenKeys, key)
	}
	sort.Strings(tokenKeys)

	tokenParts := make([]string, 0, len(tokenKeys))
	for _, key := range tokenKeys {
		tokenParts = append(tokenParts, fmt.Sprintf("%s=%d", key, r.tokenCount[key]))
	}

	return fmt.Sprintf("requests=%v points=%v tokens={%s} fields={%s} measurements={%s} samples=%v",
		r.reqCount, r.pointCount, strings.Join(tokenParts, ", "), strings.Join(parts, ", "), strings.Join(msrParts, ", "), samples)
}

func (r *forwardReceiver) ErrorSummary() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strings.Join(r.errs, "; ")
}

func (r *forwardReceiver) AssertNoError(t *testing.T) {
	t.Helper()

	r.mu.Lock()
	defer r.mu.Unlock()
	require.Empty(t, r.errs, "receiver errors: %s", strings.Join(r.errs, "; "))
}

func (r *forwardReceiver) handle(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.appendErr(fmt.Sprintf("read body failed: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	path := req.URL.Path
	gotToken := req.URL.Query().Get("token")
	r.mu.Lock()
	r.reqCount[path]++
	if gotToken != "" {
		r.tokenCount[path+"?token="+gotToken]++
	}
	r.mu.Unlock()

	if strings.HasPrefix(path, "/v1/write/") {
		if gotToken == "" {
			r.appendErr(fmt.Sprintf("missing token on path %s", path))
		}
		if path == point.URLTracing && gotToken != r.token {
			r.appendErr(fmt.Sprintf("unexpected token on path %s: got %q, want %q", path, gotToken, r.token))
		}
		if path == point.URLTracing && !r.decodeTracing {
			w.WriteHeader(http.StatusOK)
			return
		}

		dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
		pts, err := dec.Decode(body)
		point.PutDecoder(dec)
		if err != nil {
			r.appendErr(fmt.Sprintf("decode protobuf failed on %s: %v", path, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		r.mu.Lock()
		r.pointCount[path] += len(pts)
		for _, pt := range pts {
			r.msrCount[pt.Name()]++
			r.points = append(r.points, newReceivedPoint(path, gotToken, pt))

			for _, f := range pt.Fields() {
				r.fieldCount[f.Key]++
			}

			if len(r.samples) < 5 {
				r.samples = append(r.samples, formatPointSample(pt))
			}
		}
		r.mu.Unlock()
	}

	w.WriteHeader(http.StatusOK)
}

func (r *forwardReceiver) appendErr(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errs = append(r.errs, msg)
}

func getEnvOrDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func isEndpointReachable(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}

	conn, err := net.DialTimeout("tcp", u.Host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func newExternalLPMetricConfig() *aggregate.AggregatorConfigure {
	return &aggregate.AggregatorConfigure{
		DefaultWindow: 10 * time.Second,
		AggregateRules: []*aggregate.AggregateRule{
			{
				Name: "external-metric-max-count",
				Selector: &aggregate.RuleSelector{
					Category:     point.SMetric,
					Measurements: []string{externalMetricMeasurement},
					MetricName:   []string{"latency"},
				},
				Groupby: []string{"service_name", "id"},
				Algorithms: map[string]*aggregate.AggregationAlgoConfig{
					"latency.max": {
						Method:      string(aggregate.MAX),
						SourceField: "latency",
					},
					"latency.count": {
						Method:      string(aggregate.COUNT),
						SourceField: "latency",
					},
				},
			},
			{
				Name: "external-logging-count",
				Selector: &aggregate.RuleSelector{
					Category:     point.SLogging,
					Measurements: []string{externalLoggingMeasurement},
					MetricName:   []string{"message"},
				},
				Groupby: []string{"session_id"},
				Algorithms: map[string]*aggregate.AggregationAlgoConfig{
					"message.count": {
						Method:      string(aggregate.COUNT),
						SourceField: "message",
					},
				},
			},
			{
				Name: "external-tracing-count",
				Selector: &aggregate.RuleSelector{
					Category:     point.STracing,
					Measurements: []string{externalTracingMeasurement},
					MetricName:   []string{"span_id"},
				},
				Groupby: []string{"trace_id"},
				Algorithms: map[string]*aggregate.AggregationAlgoConfig{
					"span_id.count": {
						Method:      string(aggregate.COUNT),
						SourceField: "span_id",
					},
				},
			},
		},
	}
}

func newExternalLPTailSamplingConfig(version int64) *aggregate.TailSamplingConfigs {
	return &aggregate.TailSamplingConfigs{
		Version: version,
		Tracing: &aggregate.TraceTailSampling{
			DataTTL:  10 * time.Second,
			GroupKey: "trace_id",
			Pipelines: []*aggregate.SamplingPipeline{
				{
					Name:      "external-keep-error",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ status = "error" }`,
					Action:    aggregate.PipelineActionKeep,
				},
				{
					Name:      "external-keep-slow-resource",
					Type:      aggregate.PipelineTypeCondition,
					Condition: `{ resource = "/slow" }`,
					Action:    aggregate.PipelineActionKeep,
				},
				{
					Name:      "external-drop-default",
					Type:      aggregate.PipelineTypeCondition,
					Condition: "{ 1 = 1 }",
					Action:    aggregate.PipelineActionDrop,
				},
			},
		},
	}
}

func buildExternalMetricPoints(now time.Time) []*point.Point {
	opts := append(point.DefaultMetricOptions(), point.WithTime(now))
	return []*point.Point{
		point.NewPoint(externalMetricMeasurement, point.KVs{}.
			Add("latency", float64(10)).
			AddTag("service_name", "svc-a").
			AddTag("id", "1"),
			opts...),
		point.NewPoint(externalMetricMeasurement, point.KVs{}.
			Add("latency", float64(20)).
			AddTag("service_name", "svc-a").
			AddTag("id", "1"),
			opts...),
		point.NewPoint(externalMetricMeasurement, point.KVs{}.
			Add("latency", float64(7)).
			AddTag("service_name", "svc-b").
			AddTag("id", "2"),
			opts...),
	}
}

func buildExternalLoggingPoints(now time.Time) []*point.Point {
	opts := append(point.DefaultLoggingOptions(), point.WithTime(now))
	return []*point.Point{
		point.NewPoint(externalLoggingMeasurement, point.KVs{}.
			Add("message", "login ok").
			Add("session_id", "session-a"),
			opts...),
		point.NewPoint(externalLoggingMeasurement, point.KVs{}.
			Add("message", "cart open").
			Add("session_id", "session-a"),
			opts...),
		point.NewPoint(externalLoggingMeasurement, point.KVs{}.
			Add("message", "checkout submit").
			Add("session_id", "session-b"),
			opts...),
	}
}

func buildExternalTracingPoints(now time.Time) []*point.Point {
	baseStart := now.UnixMicro()
	return []*point.Point{
		newExternalTraceSpan(now, "trace-error", "span-error-root", "0", "/error", "ok", 1000, baseStart+0),
		newExternalTraceSpan(now, "trace-error", "span-error-child", "span-error-root", "/db", "error", 500, baseStart+1000),
		newExternalTraceSpan(now, "trace-slow", "span-slow-root", "0", "/slow", "ok", 6000, baseStart+2000),
		newExternalTraceSpan(now, "trace-slow", "span-slow-child", "span-slow-root", "/cache", "ok", 700, baseStart+3000),
		newExternalTraceSpan(now, "trace-drop", "span-drop-root", "0", "/drop", "ok", 900, baseStart+4000),
	}
}

func newExternalTraceSpan(now time.Time, traceID, spanID, parentID, resource, status string, durationMicro, startTimeMicro int64) *point.Point {
	opts := append(point.CommonLoggingOptions(), point.WithTime(now))
	return point.NewPoint(externalTracingMeasurement, point.NewKVs(map[string]any{
		"trace_id":   traceID,
		"span_id":    spanID,
		"parent_id":  parentID,
		"resource":   resource,
		"status":     status,
		"duration":   durationMicro,
		"start_time": startTimeMicro,
		"service":    "checkout",
	}), opts...)
}

type promMetricSample struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type promMetricCheck struct {
	Name   string
	Labels map[string]string
}

func scrapePromMetrics(rawURL string) ([]promMetricSample, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scrape metrics got status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	out := make([]promMetricSample, 0, len(lines))
	for _, line := range lines {
		s, ok := parsePromMetricLine(line)
		if ok {
			out = append(out, s)
		}
	}

	return out, nil
}

func parsePromMetricLine(line string) (promMetricSample, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return promMetricSample{}, false
	}

	metricPart, valuePart, ok := splitPromMetricLine(line)
	if !ok {
		return promMetricSample{}, false
	}

	name := metricPart
	labels := map[string]string{}
	if lidx := strings.Index(metricPart, "{"); lidx >= 0 {
		if !strings.HasSuffix(metricPart, "}") {
			return promMetricSample{}, false
		}

		name = metricPart[:lidx]
		labels = parsePromLabels(metricPart[lidx+1 : len(metricPart)-1])
	}

	v, err := strconv.ParseFloat(valuePart, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return promMetricSample{}, false
	}

	return promMetricSample{
		Name:   name,
		Labels: labels,
		Value:  v,
	}, true
}

func splitPromMetricLine(line string) (metricPart, valuePart string, ok bool) {
	var braces int
	for idx, ch := range line {
		switch ch {
		case '{':
			braces++
		case '}':
			if braces > 0 {
				braces--
			}
		case ' ', '\t':
			if braces != 0 {
				continue
			}

			metricPart = strings.TrimSpace(line[:idx])
			rest := strings.TrimSpace(line[idx+1:])
			if metricPart == "" || rest == "" {
				return "", "", false
			}
			valuePart = strings.Fields(rest)[0]
			return metricPart, valuePart, true
		}
	}

	return "", "", false
}

func parsePromLabels(raw string) map[string]string {
	labels := map[string]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return labels
	}

	for _, pair := range splitPromLabelPairs(raw) {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		unquoted, err := strconv.Unquote(val)
		if err != nil {
			continue
		}
		labels[key] = unquoted
	}

	return labels
}

func splitPromLabelPairs(raw string) []string {
	parts := make([]string, 0, 4)
	start := 0
	inQuote := false
	escaped := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		switch {
		case escaped:
			escaped = false
		case ch == '\\':
			escaped = true
		case ch == '"':
			inQuote = !inQuote
		case ch == ',' && !inQuote:
			parts = append(parts, raw[start:i])
			start = i + 1
		}
	}

	parts = append(parts, raw[start:])
	return parts
}

func metricDeltasPositive(before, after []promMetricSample, checks []promMetricCheck) bool {
	for _, chk := range checks {
		if promMetricValue(after, chk.Name, chk.Labels)-promMetricValue(before, chk.Name, chk.Labels) <= 0 {
			return false
		}
	}

	return true
}

func assertMetricDeltasPositive(t *testing.T, before, after []promMetricSample, checks []promMetricCheck) {
	t.Helper()

	for _, chk := range checks {
		delta := promMetricValue(after, chk.Name, chk.Labels) - promMetricValue(before, chk.Name, chk.Labels)
		assert.Greaterf(t, delta, 0.0, "metric delta <= 0: %s labels=%v report=%s", chk.Name, chk.Labels, metricDeltaReport(before, after, checks))
	}
}

func metricDeltaReport(before, after []promMetricSample, checks []promMetricCheck) string {
	lines := make([]string, 0, len(checks))
	for _, chk := range checks {
		beforeVal := promMetricValue(before, chk.Name, chk.Labels)
		afterVal := promMetricValue(after, chk.Name, chk.Labels)
		lines = append(lines, fmt.Sprintf("%s labels=%v before=%.3f after=%.3f delta=%.3f",
			chk.Name, chk.Labels, beforeVal, afterVal, afterVal-beforeVal))
	}

	return strings.Join(lines, "\n")
}

func forwardMetricChecks() []promMetricCheck {
	return []promMetricCheck{
		{Name: "dataway_http_api_send_points_count", Labels: map[string]string{"api": datawayAPIWriteMetric}},
		{Name: "dataway_http_api_recv_points_count", Labels: map[string]string{"api": datawayAPIWriteMetric}},
		{Name: "dataway_http_api_elapsed_seconds_count", Labels: map[string]string{"api": datawayAPIWriteMetric, "method": "POST", "sinked": "false", "status": "OK"}},
		{Name: "dataway_http_api_send_points_count", Labels: map[string]string{"api": point.URLTracing}},
		{Name: "dataway_http_api_recv_points_count", Labels: map[string]string{"api": point.URLTracing}},
		{Name: "dataway_http_api_elapsed_seconds_count", Labels: map[string]string{"api": point.URLTracing, "method": "POST", "sinked": "false", "status": "OK"}},
	}
}

func specialMetricFieldChecks() []string {
	return []string{
		"dataway_http_aggr_point_total",
		"dataway_http_api_body_size_bytes_total",
		"dataway_http_tail_sampling_trace_total",
		"dataway_http_tail_sampling_span_total",
		"dataway_http_tail_sampling_packet_send_total",
	}
}

func hasAllFields(receiver *forwardReceiver, fields []string) bool {
	for _, field := range fields {
		if receiver.FieldCount(field) <= 0 {
			return false
		}
	}

	return true
}

func formatPointSample(pt *point.Point) string {
	if pt == nil {
		return "<nil>"
	}

	fieldKeys := make([]string, 0, len(pt.Fields()))
	for _, field := range pt.Fields() {
		fieldKeys = append(fieldKeys, field.Key)
	}
	sort.Strings(fieldKeys)

	return fmt.Sprintf("%s tags=%d fields=%v", pt.Name(), len(pt.Tags()), fieldKeys)
}

func newReceivedPoint(path, token string, pt *point.Point) *receivedPoint {
	if pt == nil {
		return nil
	}

	rp := &receivedPoint{
		path:   path,
		token:  token,
		name:   pt.Name(),
		tags:   pt.MapTags(),
		attrs:  map[string]string{},
		fields: map[string]float64{},
	}

	for _, field := range pt.Fields() {
		raw := pt.Get(field.Key)
		if str, ok := raw.(string); ok {
			rp.attrs[field.Key] = str
		}
		if val, ok := numericFieldValue(raw); ok {
			rp.fields[field.Key] = val
		}
	}

	return rp
}

func numericFieldValue(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	default:
		return 0, false
	}
}

func (r *forwardReceiver) sumFieldValue(name, field string, tags map[string]string) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	var total float64
	for _, pt := range r.points {
		if pt == nil || pt.name != name {
			continue
		}
		if !labelsContainFlexible(pt.tags, pt.attrs, tags) {
			continue
		}
		total += pt.fields[field]
	}

	return total
}

func assertReceiverPointFieldValue(t *testing.T, receiver *forwardReceiver, name, field string, tags map[string]string, expected float64) {
	t.Helper()

	got := receiver.sumFieldValue(name, field, tags)
	assert.InDeltaf(t, expected, got, 0.0001, "unexpected receiver point field value: name=%s field=%s tags=%v summary=%s",
		name, field, tags, receiver.DebugSummary())
}

func assertReceiverPointFieldPositive(t *testing.T, receiver *forwardReceiver, name, field string, tags map[string]string) {
	t.Helper()

	got := receiver.sumFieldValue(name, field, tags)
	assert.Greaterf(t, got, 0.0, "receiver point field not positive: name=%s field=%s tags=%v summary=%s",
		name, field, tags, receiver.DebugSummary())
}

func promMetricValue(samples []promMetricSample, name string, labels map[string]string) float64 {
	var total float64
	for _, s := range samples {
		if s.Name != name {
			continue
		}
		if !labelsContain(s.Labels, labels) {
			continue
		}
		total += s.Value
	}

	return total
}

func labelsContain(got, expected map[string]string) bool {
	for key, exp := range expected {
		if got[key] != exp {
			return false
		}
	}

	return true
}

func labelsContainFlexible(tags, attrs, expected map[string]string) bool {
	for key, exp := range expected {
		if tags[key] == exp {
			continue
		}
		if attrs[key] == exp {
			continue
		}
		return false
	}

	return true
}

func loadLPPointsFromDir(t *testing.T, dir string) []*point.Point {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
	defer point.PutDecoder(dec)

	now := time.Now()
	pts := make([]*point.Point, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".lp" {
			continue
		}

		fp := filepath.Join(dir, entry.Name())
		body, err := os.ReadFile(fp)
		require.NoError(t, err)

		filePts, err := dec.Decode(body)
		require.NoErrorf(t, err, "decode lp file failed: %s", fp)
		for _, pt := range filePts {
			pt.SetTime(now)
		}
		pts = append(pts, filePts...)
	}

	return pts
}

func clonePoints(pts []*point.Point) []*point.Point {
	out := make([]*point.Point, 0, len(pts))
	for _, pt := range pts {
		if pt == nil {
			continue
		}
		out = append(out, point.FromPB(pt.PBPoint()))
	}

	return out
}

func failOnReceiverError(t *testing.T, receiver *forwardReceiver, stage string) {
	t.Helper()

	if errSummary := receiver.ErrorSummary(); errSummary != "" {
		t.Fatalf("receiver error while %s: %s", stage, errSummary)
	}
}

func waitFor(t *testing.T, timeout, interval time.Duration, reason string, cond func() (bool, string)) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	nextLogAt := time.Now().Add(externalWaitLogInterval)
	var lastState string
	for time.Now().Before(deadline) {
		done, state := cond()
		if state != "" {
			lastState = state
		}
		if done {
			return
		}
		if time.Now().After(nextLogAt) {
			t.Logf("waiting: %s; state=%s", reason, lastState)
			nextLogAt = time.Now().Add(externalWaitLogInterval)
		}
		time.Sleep(interval)
	}

	t.Fatalf("timeout after %s: %s; last_state=%s", timeout, reason, lastState)
}
