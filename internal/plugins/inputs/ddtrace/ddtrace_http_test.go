// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package ddtrace

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	p8s "github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	msgpHandler codec.MsgpackHandle
	encoder     = codec.NewEncoder(nil, &msgpHandler)
	decoder     = codec.NewDecoder(nil, &msgpHandler)
)

func msgpackEncoder(ddtraces DDTraces) ([]byte, error) {
	return Marshal(ddtraces)
}

func Marshal(src interface{}) ([]byte, error) {
	buf := bufpool.GetBuffer()
	encoder.Reset(buf)
	err := encoder.Encode(src)
	b := buf.Bytes()
	bufpool.PutBuffer(buf)

	return b, err
}

type ddHandler struct {
	ipt *Input
}

func (d *ddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ipt := defaultInput()
	ipt.customTagsX = itrace.NewCustomTags([]string{}, ddTags)
	ipt.handleDDTraces(w, r)
}

func Test_handleDDTraces(t *testing.T) {
	mockFeed := dkio.NewMockedFeeder()
	afterGatherRun = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	ts := httptest.NewServer(&ddHandler{})
	buf, err := jsonEncoder(randomDDTraces(10, 10))
	if err != nil {
		t.Error(err.Error())

		return
	}

	req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code =%d", resp.StatusCode)
		return
	}
	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		t.Logf("point len= %d", len(pts))
	}
}

func TestHandleDDInfo(t *testing.T) {
	ipt := defaultInput()
	ipt.Endpoints = []string{v2, v3, stats, v5, "/unknown"}
	ipt.maxTraceBody = 32 << 20

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, info, nil)
	ipt.handleDDInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got struct {
		Version                string   `json:"version"`
		GitCommit              string   `json:"git_commit"`
		Endpoints              []string `json:"endpoints"`
		FeatureFlags           []string `json:"feature_flags"`
		ClientDropP0s          bool     `json:"client_drop_p0s"`
		SpanMetaStructs        bool     `json:"span_meta_structs"`
		LongRunningSpans       bool     `json:"long_running_spans"`
		SpanEvents             bool     `json:"span_events"`
		EvpProxyAllowedHeaders []string `json:"evp_proxy_allowed_headers"`
		PeerTags               []string `json:"peer_tags"`
		SpanKindsStatsComputed []string `json:"span_kinds_stats_computed"`
		Config                 struct {
			MaxRequestBytes        int64                         `json:"max_request_bytes"`
			AnalyzedSpansByService map[string]map[string]float64 `json:"analyzed_spans_by_service"`
		} `json:"config"`
		FilterTags struct {
			Require []string `json:"require"`
			Reject  []string `json:"reject"`
		} `json:"filter_tags"`
		FilterTagsRegex struct {
			Require []string `json:"require"`
			Reject  []string `json:"reject"`
		} `json:"filter_tags_regex"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))

	assert.Equal(t, ddInfoAgentVersion, got.Version)
	assert.NotEmpty(t, got.GitCommit)
	assert.Equal(t, []string{v3, v5}, got.Endpoints)
	assert.NotContains(t, got.Endpoints, stats)
	assert.False(t, got.ClientDropP0s)
	assert.False(t, got.SpanMetaStructs)
	assert.False(t, got.LongRunningSpans)
	assert.False(t, got.SpanEvents)
	assert.Equal(t, int64(32<<20), got.Config.MaxRequestBytes)
	assert.NotNil(t, got.FeatureFlags)
	assert.NotNil(t, got.EvpProxyAllowedHeaders)
	assert.NotNil(t, got.PeerTags)
	assert.NotNil(t, got.SpanKindsStatsComputed)
	assert.NotNil(t, got.Config.AnalyzedSpansByService)
	assert.NotNil(t, got.FilterTags.Require)
	assert.NotNil(t, got.FilterTags.Reject)
	assert.NotNil(t, got.FilterTagsRegex.Require)
	assert.NotNil(t, got.FilterTagsRegex.Reject)
}

func BenchmarkDDTrace_Msgsize(b *testing.B) {
	mockFeed := dkio.NewMockedFeeder()
	afterGatherRun = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	ts := httptest.NewServer(&ddHandler{})
	buf, err := msgpackEncoder(randomDDTraces(10, 10))
	if err != nil {
		b.Error(err.Error())

		return
	}
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
		if err != nil {
			b.Error(err.Error())

			return
		}
		// req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", "application/msgpack")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Error(err.Error())

			return
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b.Logf("request failed with status code %d\n", resp.StatusCode)
		}
	}

	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		b.Logf("point len= %d", len(pts))
	}
}

func TestDDTrace(t *testing.T) {
	mockFeed := dkio.NewMockedFeeder()
	gather := itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	sample := &itrace.Sampler{
		SamplingRateGlobal: 1,
	}
	sample.Init()
	gather.AppendFilter(sample.Sample)
	gather.AppendFilter(itrace.PenetrateErrorTracing)
	afterGatherRun = gather
	tagSpan := randomDDSpan()
	tagSpan.Meta["http.url"] = "/tmall"
	tagSpan.Meta["span.kind"] = "server"
	tagSpan.Meta["process.env"] = "env"
	tagSpan.Meta["error.msg"] = "error message"
	tagSpan.Meta["error.type"] = "type"

	ts := httptest.NewServer(&ddHandler{})
	trace := DDTraces{DDTrace{tagSpan}}
	buf, err := msgpackEncoder(trace)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/msgpack")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("request failed with status code %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		t.Logf("point len= %d", len(pts))
		if len(pts) != 1 {
			t.Error("points len !=1")
			return
		}
		assert.Equal(t, pts[0].GetTag("http_url"), "/tmall", "must be /tmall")
		assert.Equal(t, pts[0].GetTag("error_message"), "error message", "error_message")
		assert.NotEqual(t, pts[0].GetTag("process_env"), "env", "must empty")
	}

	sampleSpan := randomDDSpan()
	sampleSpan.Metrics[keyPriority] = -1
	sampleSpan2 := randomDDSpan()
	sampleSpan2.Metrics[keyPriority] = 2
	sampleTrace := DDTraces{DDTrace{sampleSpan, sampleSpan2}}

	buf, err = msgpackEncoder(sampleTrace)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	req, err = http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/msgpack")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Logf("request failed with status code %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	pts, err = mockFeed.AnyPoints(time.Second)
	if err == nil {
		assert.Len(t, pts, 1)
	}
}

func TestSamplingPriorityDecision(t *testing.T) {
	tests := []struct {
		name                string
		priorities          []float64
		excludes            []int
		wantSpans           int
		wantPriority        string
		wantAction          string
		wantSamplingRateTag string
	}{
		{
			name:         "rule sampler reject",
			priorities:   []float64{-3},
			wantPriority: "-3",
			wantAction:   samplingPriorityActionDrop,
		},
		{
			name:         "user reject",
			priorities:   []float64{-1},
			wantPriority: "-1",
			wantAction:   samplingPriorityActionDrop,
		},
		{
			name:         "auto reject",
			priorities:   []float64{0},
			wantPriority: "0",
			wantAction:   samplingPriorityActionDrop,
		},
		{
			name:         "excluded auto reject bypasses once",
			priorities:   []float64{0, 0},
			excludes:     []int{0},
			wantSpans:    2,
			wantPriority: "0",
			wantAction:   samplingPriorityActionBypass,
		},
		{
			name:         "non-excluded reject wins",
			priorities:   []float64{0, -1},
			excludes:     []int{0},
			wantPriority: "-1",
			wantAction:   samplingPriorityActionDrop,
		},
		{
			name:                "keep priority is unchanged",
			priorities:          []float64{1},
			wantSpans:           1,
			wantSamplingRateTag: itrace.SamplerKeep,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := "sampling-priority-test-" + strconv.Itoa(i)
			trace := make(DDTrace, 0, len(tc.priorities))
			for j, priority := range tc.priorities {
				trace = append(trace, &DDSpan{
					Service:  service,
					Name:     "operation",
					Resource: "resource",
					TraceID:  uint64(i + 1),
					SpanID:   uint64(j + 1),
					Start:    time.Now().UnixNano(),
					Duration: int64(time.Millisecond),
					Meta:     map[string]string{},
					Metrics:  map[string]float64{keyPriority: priority},
				})
			}

			ipt := defaultInput()
			ipt.SamplingPriorityDropExcludes = tc.excludes
			ipt.customTagsX = itrace.NewCustomTags(nil, ddTags)
			if tc.wantAction != "" {
				samplingPriorityTraceDecision.DeleteLabelValues(tc.wantPriority, tc.wantAction, service)
				samplingPrioritySpanDecision.DeleteLabelValues(tc.wantPriority, tc.wantAction, service)
				t.Cleanup(func() {
					samplingPriorityTraceDecision.DeleteLabelValues(tc.wantPriority, tc.wantAction, service)
					samplingPrioritySpanDecision.DeleteLabelValues(tc.wantPriority, tc.wantAction, service)
				})
			}

			got := ipt.ddtraceToDkTrace(trace, nil, "127.0.0.1", time.Time{})
			assert.Len(t, got, tc.wantSpans)
			if tc.wantSamplingRateTag != "" {
				require.NotEmpty(t, got)
				assert.Equal(t, tc.wantSamplingRateTag, got[0].GetTag(itrace.SampleRateKey))
			}
			if tc.wantAction == samplingPriorityActionBypass {
				for _, span := range got {
					assert.Empty(t, span.GetTag(itrace.SampleRateKey))
				}
			}

			if tc.wantAction != "" {
				assert.Equal(t, float64(1), prometheusCounterValue(t,
					samplingPriorityTraceDecision, tc.wantPriority, tc.wantAction, service))
				assert.Equal(t, float64(len(trace)), prometheusCounterValue(t,
					samplingPrioritySpanDecision, tc.wantPriority, tc.wantAction, service))
			}
		})
	}
}

func TestNormalizeSamplingPriorityDropExcludes(t *testing.T) {
	ipt := &Input{SamplingPriorityDropExcludes: []int{0, -1, 0, -3, 1, 2}}
	ipt.normalizeSamplingPriorityDropExcludes()
	assert.Equal(t, []int{0, -1, -3}, ipt.SamplingPriorityDropExcludes)
}

func TestExcludedSamplingPriorityStillUsesDataKitSampler(t *testing.T) {
	span := &DDSpan{
		Service:  "sampling-priority-datakit-sampler",
		Name:     "operation",
		Resource: "resource",
		TraceID:  1,
		SpanID:   1,
		Start:    time.Now().UnixNano(),
		Duration: int64(time.Millisecond),
		Meta:     map[string]string{},
		Metrics:  map[string]float64{keyPriority: 0},
	}
	ipt := defaultInput()
	ipt.SamplingPriorityDropExcludes = []int{0}
	ipt.customTagsX = itrace.NewCustomTags(nil, ddTags)

	dktrace := ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", time.Time{})
	require.Len(t, dktrace, 1)
	assert.Empty(t, dktrace[0].GetTag(itrace.SampleRateKey))

	sampler := (&itrace.Sampler{SamplingRateGlobal: 0}).Init()
	_, dropped := sampler.Sample(log, dktrace)
	assert.True(t, dropped)
}

func prometheusCounterValue(t *testing.T, counter *p8s.CounterVec, labels ...string) float64 {
	t.Helper()

	metric := &dto.Metric{}
	require.NoError(t, counter.WithLabelValues(labels...).Write(metric))

	return metric.GetCounter().GetValue()
}

func TestDDTraceCustomerMetrics(t *testing.T) {
	ipt := defaultInput()
	ipt.CustomerTags = []string{
		"custom.meta",
		"custom.metric",
		"duration",
		"same.key",
		`reg:^business\.`,
	}
	ipt.customTagsX = itrace.NewCustomTags(ipt.CustomerTags, cloneStringMap(ddTags))

	span := &DDSpan{
		Service:  "service",
		Name:     "operation",
		Resource: "resource",
		TraceID:  1,
		SpanID:   2,
		Start:    time.Unix(1, 0).UnixNano(),
		Duration: int64(10 * time.Millisecond),
		Meta: map[string]string{
			"custom.meta": "meta-value",
			"same.key":    "meta-wins",
		},
		Metrics: map[string]float64{
			"custom.metric":    12.5,
			"business.count":   3,
			"duration":         99,
			"same.key":         2,
			"unmatched.metric": 7,
		},
	}

	got := ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", time.Time{})
	require.Len(t, got, 1)
	pt := got[0].Point

	assert.Equal(t, "meta-value", pt.GetTag("custom_meta"))
	assert.Equal(t, "meta-wins", pt.GetTag("same_key"))
	assert.Empty(t, pt.GetTag("custom_metric"))
	assert.Equal(t, 12.5, pointFloat(t, pt, "custom_metric"))
	assert.Equal(t, float64(3), pointFloat(t, pt, "business_count"))
	assert.Equal(t, int64(10*time.Millisecond)/int64(time.Microsecond), pointInt(t, pt, itrace.FieldDuration))

	message, ok := pt.GetS(itrace.FieldMessage)
	require.True(t, ok)
	var remaining wrapMessage
	require.NoError(t, json.Unmarshal([]byte(message), &remaining))
	assert.Equal(t, map[string]float64{
		"duration":         99,
		"same.key":         2,
		"unmatched.metric": 7,
	}, remaining.Metrics)
	assert.Empty(t, remaining.Meta)
}

func TestDDTraceCustomerMetricsReservedKeys(t *testing.T) {
	tests := []struct {
		name         string
		customerTags []string
	}{
		{
			name:         "exact match",
			customerTags: []string{itrace.TagSpanStatus, itrace.FieldMessage, "http.status.class"},
		},
		{
			name:         "regexp match",
			customerTags: []string{`reg:^(status|message|http\.status\.class)$`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.CustomerTags = tc.customerTags
			ipt.customTagsX = itrace.NewCustomTags(ipt.CustomerTags, cloneStringMap(ddTags))

			span := &DDSpan{
				Service:  "service",
				Name:     "operation",
				Resource: "resource",
				TraceID:  1,
				SpanID:   2,
				Start:    time.Unix(1, 0).UnixNano(),
				Duration: int64(10 * time.Millisecond),
				Error:    1,
				Meta: map[string]string{
					"http.status_code": "200",
				},
				Metrics: map[string]float64{
					itrace.TagSpanStatus: 42,
					itrace.FieldMessage:  7,
					"http.status.class":  9,
				},
			}

			got := ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", time.Time{})
			require.Len(t, got, 1)
			pt := got[0].Point

			assert.Equal(t, itrace.StatusErr, pt.GetTag(itrace.TagSpanStatus))
			assert.Equal(t, "2xx", pt.GetTag(itrace.TagHttpStatusClass))
			message, ok := pt.GetS(itrace.FieldMessage)
			require.True(t, ok)

			var remaining wrapMessage
			require.NoError(t, json.Unmarshal([]byte(message), &remaining))
			assert.Equal(t, map[string]float64{
				itrace.TagSpanStatus: 42,
				itrace.FieldMessage:  7,
				"http.status.class":  9,
			}, remaining.Metrics)
		})
	}
}

func cloneStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}

	return dst
}

func pointFloat(t *testing.T, pt *point.Point, key string) float64 {
	t.Helper()
	value, ok := pt.GetF(key)
	require.True(t, ok)
	return value
}

func pointInt(t *testing.T, pt *point.Point, key string) int64 {
	t.Helper()
	value, ok := pt.GetI(key)
	require.True(t, ok)
	return value
}

func randomDDSpan() *DDSpan {
	return &DDSpan{
		Service:  testutils.RandString(10),
		Name:     testutils.RandString(10),
		Resource: testutils.RandString(10),
		TraceID:  uint64(testutils.RandInt64(10)),
		SpanID:   uint64(testutils.RandInt64(10)),
		ParentID: uint64(testutils.RandInt64(10)),
		Start:    testutils.RandTime().UnixNano(),
		Duration: testutils.RandInt64(6),
		Meta:     testutils.RandTags(10, 10, 20),
		Metrics:  testutils.RandMetrics(10, 10),
		Type: testutils.RandWithinStrings([]string{
			"consul", "cache", "memcached", "redis", "aerospike", "cassandra", "db", "elasticsearch", "leveldb",
			"", "mongodb", "sql", "http", "web", "benchmark", "build", "custom", "datanucleus", "dns", "graphql", "grpc", "hibernate", "queue", "rpc", "soap", "template", "test", "worker",
		}),
	}
}

func randomDDTrace(n int) DDTrace {
	ddtrace := make(DDTrace, n)
	for i := 0; i < n; i++ {
		ddtrace[i] = randomDDSpan()
	}

	return ddtrace
}

func randomDDTraces(n, m int) DDTraces {
	ddtraces := make(DDTraces, n)
	for i := 0; i < n; i++ {
		ddtraces[i] = randomDDTrace(m)
	}

	return ddtraces
}

func jsonEncoder(ddtraces DDTraces) ([]byte, error) {
	return json.Marshal(ddtraces)
}

func BenchmarkDecodeRequest(b *testing.B) {
	cases := []struct {
		name          string
		traceCount    int
		spansPerTrace int
	}{
		{
			name:          "100_spans",
			traceCount:    10,
			spansPerTrace: 10,
		},
		{
			name:          "4096_spans",
			traceCount:    128,
			spansPerTrace: 32,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			buf, err := msgpackEncoder(randomDDTraces(tc.traceCount, tc.spansPerTrace))
			if err != nil {
				b.Fatal(err)
			}

			param := &itrace.TraceParameters{
				URLPath: v4,
				Media:   "application/msgpack",
				Encode:  "",
				Body:    bytes.NewBuffer(buf),
			}

			b.Run("with_pool", func(b *testing.B) {
				ddtracePoolT := &sync.Pool{
					New: func() interface{} {
						return DDTraces{}
					},
				}

				b.SetBytes(int64(len(buf)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					dt := ddtracePoolT.Get().(DDTraces)
					if err := decodeRequest(param, &dt); err != nil {
						b.Fatal(err)
					}

					keepInPool := dt.shouldKeepInPool()
					dt.reset(keepInPool)
					if keepInPool {
						ddtracePoolT.Put(dt) //nolint
					}
				}
			})

			b.Run("no_pool", func(b *testing.B) {
				b.SetBytes(int64(len(buf)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var dt DDTraces
					if err := decodeRequest(param, &dt); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

func TestDDTraceMemoryUnderLoad(t *testing.T) {
	if os.Getenv("DDTRACE_MEMORY_LOAD_TEST") != "1" {
		t.Skip("set DDTRACE_MEMORY_LOAD_TEST=1 to run the 3-minute ddtrace memory load test")
	}

	oldLog := log
	if err := logger.InitRoot(&logger.Option{Level: logger.WARN, Flags: logger.OPT_DEFAULT}); err != nil {
		t.Fatal(err)
	}
	log = logger.SLogger(inputName)
	defer func() {
		logger.Reset()
		log = oldLog
	}()

	duration := ddtraceLoadEnvDuration(t, "DDTRACE_MEMORY_LOAD_DURATION", 3*time.Minute)
	sampleInterval := ddtraceLoadEnvDuration(t, "DDTRACE_MEMORY_LOAD_SAMPLE_INTERVAL", 10*time.Second)
	workers := ddtraceLoadEnvInt(t, "DDTRACE_MEMORY_LOAD_WORKERS", 1)
	payloadCount := ddtraceLoadEnvInt(t, "DDTRACE_MEMORY_LOAD_PAYLOADS", 16)
	traceCount := ddtraceLoadEnvInt(t, "DDTRACE_MEMORY_LOAD_TRACES", 128)
	spansPerTrace := ddtraceLoadEnvInt(t, "DDTRACE_MEMORY_LOAD_SPANS_PER_TRACE", 32)

	payloads := make([][]byte, payloadCount)
	var payloadBytes int
	for i := range payloads {
		buf, err := msgpackEncoder(randomDDTraces(traceCount, spansPerTrace))
		if err != nil {
			t.Fatal(err)
		}
		payloads[i] = append([]byte(nil), buf...)
		payloadBytes += len(payloads[i])
	}

	t.Logf("payloads=%d total_payload_bytes=%d traces_per_payload=%d spans_per_trace=%d workers=%d duration=%s sample_interval=%s",
		len(payloads), payloadBytes, traceCount, spansPerTrace, workers, duration, sampleInterval)

	cases := []struct {
		name    string
		recycle func(DDTraces)
	}{
		{
			name:    "current",
			recycle: recycleDDTraces,
		},
	}
	if ddtraceLoadEnvBool(t, "DDTRACE_MEMORY_LOAD_COMPARE_LEGACY", true) {
		cases = append(cases, struct {
			name    string
			recycle func(DDTraces)
		}{
			name:    "legacy_recycle",
			recycle: legacyRecycleDDTracesForTest,
		})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDDTraceMemoryLoad(t, payloads, workers, duration, sampleInterval, tc.recycle)
		})
	}
}

func TestInt64ToPaddedString(t *testing.T) {
	type args struct {
		num uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "123",
			args: args{num: 1234567890},
			want: 16,
		},
		{
			name: "1234",
			args: args{num: 6450066879287049030},
			want: 16,
		},
		{
			name: "max",
			args: args{num: math.MaxUint64},
			want: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tid := Int64ToPaddedString(tt.args.num)
			t.Logf("tid=%s", tid)
			assert.Equalf(t, tt.want, len(tid), "Int64ToPaddedString(%v)", tt.args.num)
		})
	}
}

func runDDTraceMemoryLoad(
	t *testing.T,
	payloads [][]byte,
	workers int,
	duration time.Duration,
	sampleInterval time.Duration,
	recycle func(DDTraces),
) {
	t.Helper()

	oldPool := ddtracePool
	oldRecycle := recycleDDTracesFunc
	ddtracePool = &sync.Pool{
		New: func() interface{} {
			return DDTraces{}
		},
	}
	recycleDDTracesFunc = recycle
	defer func() {
		recycleDDTracesFunc = oldRecycle
		ddtracePool = oldPool
	}()

	mockFeed := dkio.NewMockedFeeder()
	afterGatherRun = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))

	ipt := defaultInput()
	ipt.customTagsX = itrace.NewCustomTags([]string{}, ddTags)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipt.handleDDTraces(w, r)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	drainCtx, drainCancel := context.WithCancel(context.Background())
	defer drainCancel()

	var (
		drainWG  sync.WaitGroup
		points   uint64
		requests uint64
		failures uint64
		payloadN uint64
	)

	drainWG.Add(1)
	go func() {
		defer drainWG.Done()
		for {
			select {
			case <-drainCtx.Done():
				return
			default:
			}

			pts, err := mockFeed.AnyPoints(10 * time.Millisecond)
			if err == nil {
				atomic.AddUint64(&points, uint64(len(pts)))
			}
		}
	}()

	runtime.GC()
	start := time.Now()
	samples := []ddtraceMemorySample{readDDTraceMemorySample(start)}

	sampleDone := make(chan struct{})
	go func() {
		defer close(sampleDone)
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				samples = append(samples, readDDTraceMemorySample(start))
				return
			case <-ticker.C:
				samples = append(samples, readDDTraceMemorySample(start))
			}
		}
	}()

	client := ts.Client()
	client.Timeout = 30 * time.Second

	var sendWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		sendWG.Add(1)
		go func() {
			defer sendWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				idx := int(atomic.AddUint64(&payloadN, 1)-1) % len(payloads)
				payload := payloads[idx]
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+v4, bytes.NewReader(payload))
				if err != nil {
					atomic.AddUint64(&failures, 1)
					continue
				}
				req.Header.Set("Content-Type", "application/msgpack")
				req.Header.Set("X-Datadog-Trace-Count", "1")

				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() == nil {
						atomic.AddUint64(&failures, 1)
					}
					continue
				}

				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					atomic.AddUint64(&failures, 1)
					continue
				}
				atomic.AddUint64(&requests, 1)
			}
		}()
	}

	sendWG.Wait()
	<-sampleDone

	drainCancel()
	drainWG.Wait()

	runtime.GC()
	samples = append(samples, readDDTraceMemorySample(start))

	first := samples[0]
	last := samples[len(samples)-1]
	for _, sample := range samples {
		t.Logf("t=%s heap_alloc=%d heap_inuse=%d heap_sys=%d heap_released=%d rss=%d num_gc=%d",
			sample.Elapsed.Round(time.Second),
			sample.HeapAlloc,
			sample.HeapInuse,
			sample.HeapSys,
			sample.HeapReleased,
			sample.RSS,
			sample.NumGC,
		)
	}

	t.Logf("summary requests=%d points=%d failures=%d heap_alloc_delta=%d heap_inuse_delta=%d rss_delta=%d num_gc_delta=%d",
		atomic.LoadUint64(&requests),
		atomic.LoadUint64(&points),
		atomic.LoadUint64(&failures),
		int64(last.HeapAlloc)-int64(first.HeapAlloc),
		int64(last.HeapInuse)-int64(first.HeapInuse),
		int64(last.RSS)-int64(first.RSS),
		last.NumGC-first.NumGC,
	)

	if failures != 0 {
		t.Fatalf("memory load test saw %d failed requests", failures)
	}
}

type ddtraceMemorySample struct {
	Elapsed      time.Duration
	HeapAlloc    uint64
	HeapInuse    uint64
	HeapSys      uint64
	HeapReleased uint64
	RSS          uint64
	NumGC        uint32
}

func readDDTraceMemorySample(start time.Time) ddtraceMemorySample {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return ddtraceMemorySample{
		Elapsed:      time.Since(start),
		HeapAlloc:    mem.HeapAlloc,
		HeapInuse:    mem.HeapInuse,
		HeapSys:      mem.HeapSys,
		HeapReleased: mem.HeapReleased,
		RSS:          readRSSBytesForTest(),
		NumGC:        mem.NumGC,
	}
}

func readRSSBytesForTest() uint64 {
	bts, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}

	fields := strings.Fields(string(bts))
	if len(fields) < 2 {
		return 0
	}

	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}

	return pages * uint64(os.Getpagesize())
}

func ddtraceLoadEnvDuration(t *testing.T, key string, fallback time.Duration) time.Duration {
	t.Helper()

	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	dur, err := time.ParseDuration(val)
	if err != nil {
		t.Fatalf("invalid %s=%q: %s", key, val, err)
	}
	if dur <= 0 {
		t.Fatalf("%s must be greater than 0, got %s", key, dur)
	}

	return dur
}

func ddtraceLoadEnvInt(t *testing.T, key string, fallback int) int {
	t.Helper()

	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	n, err := strconv.Atoi(val)
	if err != nil {
		t.Fatalf("invalid %s=%q: %s", key, val, err)
	}
	if n <= 0 {
		t.Fatalf("%s must be greater than 0, got %d", key, n)
	}

	return n
}

func ddtraceLoadEnvBool(t *testing.T, key string, fallback bool) bool {
	t.Helper()

	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	ok, err := strconv.ParseBool(val)
	if err != nil {
		t.Fatalf("invalid %s=%q: %s", key, val, err)
	}

	return ok
}

func legacyRecycleDDTracesForTest(traces DDTraces) {
	legacyResetDDTracesForTest(traces)
	ddtracePool.Put(traces) //nolint
}

func legacyResetDDTracesForTest(traces DDTraces) {
	for _, trace := range traces {
		for _, span := range trace {
			if span == nil {
				continue
			}

			span.Service = ""
			span.Name = ""
			span.Resource = ""
			span.TraceID = 0
			span.SpanID = 0
			span.ParentID = 0
			span.Start = 0
			span.Duration = 0
			span.Error = 0
			for k := range span.Meta {
				span.Meta[k] = ""
			}
			for k := range span.Metrics {
				span.Metrics[k] = 0
			}
			span.Type = ""
		}
	}
}
