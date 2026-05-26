// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestReadConfig(t *testing.T) {
	ag := &Aggregator{}
	ag.loadMetricConfigFromFile("./testdata/aggr.toml")

	require.NotNil(t, ag.metricConfig)
	require.NoError(t, ag.metricConfig.Setup())
	assert.Equal(t, 15*time.Second, ag.metricConfig.DefaultWindow)
	require.Len(t, ag.metricConfig.AggregateRules, 6)
	assert.Equal(t, "otel-jvm-memory", ag.metricConfig.AggregateRules[0].Name)
	assert.Equal(t, "trace_root_span_count", ag.metricConfig.AggregateRules[1].Name)
	assert.Equal(t, "otel_logging_count", ag.metricConfig.AggregateRules[2].Name)
	assert.Equal(t, "otel_jvm_class_loaded_sum", ag.metricConfig.AggregateRules[3].Name)
	assert.Equal(t, "otel_jvm_class_unloaded_sum", ag.metricConfig.AggregateRules[4].Name)
	assert.Equal(t, "otel_jvm_threads_live_sum", ag.metricConfig.AggregateRules[5].Name)
}

func TestReadTailSamplingConfig(t *testing.T) {
	ag := &Aggregator{}
	ag.loadTailSamplingConfigFromFile("./testdata/tail-sampling.toml")

	require.NotNil(t, ag.tailSamplingConfig)
	require.NotNil(t, ag.tailSamplingConfig.Tracing)
	require.NotNil(t, ag.tailSamplingConfig.Logging)
	assert.Equal(t, int64(2), ag.tailSamplingConfig.Version)
	assert.Equal(t, "trace_id", ag.tailSamplingConfig.Tracing.GroupKey)
	require.Len(t, ag.tailSamplingConfig.Tracing.Pipelines, 6)
	assert.Equal(t, "server_key", ag.tailSamplingConfig.Tracing.Pipelines[0].Name)
	assert.Equal(t, "sample-rest", ag.tailSamplingConfig.Tracing.Pipelines[5].Name)
	require.Len(t, ag.tailSamplingConfig.Logging.GroupDimensions, 1)
	assert.Equal(t, "trace_id", ag.tailSamplingConfig.Logging.GroupDimensions[0].GroupKey)
	require.Len(t, ag.tailSamplingConfig.Logging.GroupDimensions[0].Pipelines, 1)
	assert.Equal(t, "keep-logs-with-trace-id", ag.tailSamplingConfig.Logging.GroupDimensions[0].Pipelines[0].Name)
}

func TestDefaultTokenPrefersAggrEndpoints(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_aggr"},
		DW: &dataway.Dataway{
			Token: "tkn_dw",
		},
	}
	ag.initHTTP()

	assert.Equal(t, "tkn_aggr", ag.defaultToken())
}

func TestPickMetric(t *testing.T) {
	ag := &Aggregator{
		metricConfig: newMetricConfig(),
	}
	require.NoError(t, ag.metricConfig.Setup())

	batches := ag.PickMetric(point.SMetric, newMetricPoints())
	require.Len(t, batches, 1)

	for _, batch := range batches {
		require.Len(t, batch.Batchs, 1)
		assert.NotZero(t, batch.Batchs[0].RoutingKey)
	}
}

func TestPickMetricWithAggrTomlRules(t *testing.T) {
	ag := &Aggregator{}
	ag.loadMetricConfigFromFile("./testdata/aggr.toml")
	require.NotNil(t, ag.metricConfig)
	require.NoError(t, ag.metricConfig.Setup())

	for _, rule := range ag.metricConfig.AggregateRules {
		for s, algo := range rule.Algorithms {
			t.Logf("%s = %+v", s, algo)
		}
	}
}

func TestPickTrace(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_trace"},
		tailSamplingConfig: &aggregate.TailSamplingConfigs{
			Version: 3,
			Tracing: &aggregate.TraceTailSampling{},
		},
	}

	packages := ag.PickTrace("ddtrace", newAPMPoints())
	require.Len(t, packages, 1)

	for _, pkg := range packages {
		assert.Equal(t, "tkn_trace", pkg.Token)
		assert.Equal(t, int64(3), pkg.ConfigVersion)
		assert.Equal(t, point.STracing, pkg.DataType)
	}
}

func TestPickLogging(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_logging"},
		tailSamplingConfig: &aggregate.TailSamplingConfigs{
			Version: 1,
			Logging: &aggregate.LoggingTailSampling{
				GroupDimensions: []*aggregate.LoggingGroupDimension{
					{GroupKey: "session_id"},
				},
			},
		},
	}

	packages, passed := ag.PickLogging("logging", newLoggingPoints())
	require.Len(t, packages, 1)
	assert.Len(t, passed, 1)

	for _, pkg := range packages {
		assert.Equal(t, "tkn_logging", pkg.Token)
		assert.Equal(t, int64(1), pkg.ConfigVersion)
		assert.Equal(t, point.SLogging, pkg.DataType)
		assert.Equal(t, "session_id", pkg.GroupKey)
	}
}

func TestPickLoggingDifferentGroupDimensionsDoNotCollide(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_logging"},
		tailSamplingConfig: &aggregate.TailSamplingConfigs{
			Version: 1,
			Logging: &aggregate.LoggingTailSampling{
				GroupDimensions: []*aggregate.LoggingGroupDimension{
					{GroupKey: "session_id"},
					{GroupKey: "user_id"},
				},
			},
		},
	}

	packages, passed := ag.PickLogging("logging", newLoggingPointsForDimensionCollision())
	require.Len(t, packages, 2)
	assert.Empty(t, passed)

	groupKeys := make([]string, 0, len(packages))
	for _, pkg := range packages {
		groupKeys = append(groupKeys, pkg.GroupKey)
		assert.Equal(t, "123", pkg.RawGroupId)
	}
	assert.ElementsMatch(t, []string{"session_id", "user_id"}, groupKeys)
}

func TestPickRUM(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_rum"},
		tailSamplingConfig: &aggregate.TailSamplingConfigs{
			Version: 2,
			RUM: &aggregate.RUMTailSampling{
				GroupDimensions: []*aggregate.RUMGroupDimension{
					{GroupKey: "session_id"},
				},
			},
		},
	}

	packages, passed := ag.PickRUM("rum", newRUMPoints())
	require.Len(t, packages, 1)
	assert.Len(t, passed, 1)

	for _, pkg := range packages {
		assert.Equal(t, "tkn_rum", pkg.Token)
		assert.Equal(t, int64(2), pkg.ConfigVersion)
		assert.Equal(t, point.SRUM, pkg.DataType)
		assert.Equal(t, "session_id", pkg.GroupKey)
	}
}

func TestPickRUMDifferentGroupDimensionsDoNotCollide(t *testing.T) {
	ag := &Aggregator{
		Endpoints: []string{"http://127.0.0.1:9528?token=tkn_rum"},
		tailSamplingConfig: &aggregate.TailSamplingConfigs{
			Version: 2,
			RUM: &aggregate.RUMTailSampling{
				GroupDimensions: []*aggregate.RUMGroupDimension{
					{GroupKey: "session_id"},
					{GroupKey: "user_id"},
				},
			},
		},
	}

	packages, passed := ag.PickRUM("rum", newRUMPointsForDimensionCollision())
	require.Len(t, packages, 2)
	assert.Empty(t, passed)

	groupKeys := make([]string, 0, len(packages))
	for _, pkg := range packages {
		groupKeys = append(groupKeys, pkg.GroupKey)
		assert.Equal(t, "123", pkg.RawGroupId)
	}
	assert.ElementsMatch(t, []string{"session_id", "user_id"}, groupKeys)
}

func TestProcessTracingConsumed(t *testing.T) {
	var metricReqs int32
	var tsReqs int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case datakit.Aggregate:
			atomic.AddInt32(&metricReqs, 1)
		case datakit.TailSampling:
			atomic.AddInt32(&tsReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:           []string{ts.URL + "?token=tkn_trace"},
		metricConfig:        newTracingMetricConfig(),
		metricEnabled:       true,
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, Tracing: &aggregate.TraceTailSampling{}},
		tailSamplingEnabled: true,
	}
	require.NoError(t, ag.metricConfig.Setup())
	ag.initHTTP()

	result, err := ag.Process(point.Tracing, "ddtrace", newAPMPoints())
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Points)
	assert.True(t, result.Consumed)
	assert.Equal(t, 4, result.SelectedPoints)
	assert.Equal(t, 2, result.BatchPackages)
	assert.Equal(t, 1, result.TailSamplingPackages)
	assert.Equal(t, int32(1), atomic.LoadInt32(&metricReqs))
	assert.Equal(t, int32(1), atomic.LoadInt32(&tsReqs))
}

func TestProcessLoggingReturnsPassthrough(t *testing.T) {
	var metricReqs int32
	var tsReqs int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case datakit.Aggregate:
			atomic.AddInt32(&metricReqs, 1)
		case datakit.TailSampling:
			atomic.AddInt32(&tsReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:           []string{ts.URL + "?token=tkn_logging"},
		metricConfig:        newLoggingMetricConfig(),
		metricEnabled:       true,
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, Logging: &aggregate.LoggingTailSampling{GroupDimensions: []*aggregate.LoggingGroupDimension{{GroupKey: "session_id"}}}},
		tailSamplingEnabled: true,
	}
	require.NoError(t, ag.metricConfig.Setup())
	ag.initHTTP()

	result, err := ag.Process(point.Logging, "logging", newLoggingPoints())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Points, 1)
	assert.False(t, result.Consumed)
	assert.Equal(t, 3, result.SelectedPoints)
	assert.Equal(t, 2, result.BatchPackages)
	assert.Equal(t, 1, result.TailSamplingPackages)
	assert.Equal(t, int32(1), atomic.LoadInt32(&metricReqs))
	assert.Equal(t, int32(1), atomic.LoadInt32(&tsReqs))
}

func TestProcessRUMReturnsPassthrough(t *testing.T) {
	var tsReqs int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.TailSampling {
			atomic.AddInt32(&tsReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:           []string{ts.URL + "?token=tkn_rum"},
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, RUM: &aggregate.RUMTailSampling{GroupDimensions: []*aggregate.RUMGroupDimension{{GroupKey: "session_id"}}}},
		tailSamplingEnabled: true,
	}
	ag.initHTTP()

	result, err := ag.Process(point.RUM, "rum", newRUMPoints())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Points, 1)
	assert.False(t, result.Consumed)
	assert.Equal(t, 1, result.SelectedPoints)
	assert.Equal(t, 0, result.BatchPackages)
	assert.Equal(t, 1, result.TailSamplingPackages)
	assert.Equal(t, int32(1), atomic.LoadInt32(&tsReqs))
}

func TestSendTailSamplingPackagesSplit(t *testing.T) {
	var (
		mu      sync.Mutex
		reqLens []int
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.TailSampling {
			w.WriteHeader(http.StatusOK)
			return
		}
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read tail sampling request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		reqLens = append(reqLens, len(bs))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:           []string{ts.URL + "?token=tkn_trace"},
		MaxRawBodySize:      1024,
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, Tracing: &aggregate.TraceTailSampling{}},
		tailSamplingEnabled: true,
	}
	ag.initHTTP()

	packages := ag.PickTrace("ddtrace", newLargeAPMPoints(128, 300))
	require.NoError(t, ag.SendTailSamplingPackages(packages))

	require.Greater(t, len(reqLens), 1)
	for _, n := range reqLens {
		assert.LessOrEqual(t, n, ag.MaxRawBodySize)
	}
}

func TestSendMetricBatchesSplit(t *testing.T) {
	var (
		mu      sync.Mutex
		reqLens []int
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.Aggregate {
			w.WriteHeader(http.StatusOK)
			return
		}

		bs, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read aggregate request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		reqLens = append(reqLens, len(bs))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:      []string{ts.URL + "?token=tkn_metric"},
		MaxRawBodySize: 2048,
		metricConfig:   newMetricConfig(),
	}
	require.NoError(t, ag.metricConfig.Setup())
	ag.initHTTP()

	batchMap := ag.PickMetric(point.SMetric, newLargeMetricPoints(200, 128))
	ag.SendMetricBatches(point.SMetric, batchMap)

	require.Greater(t, len(reqLens), 1)
	for _, n := range reqLens {
		assert.LessOrEqual(t, n, ag.MaxRawBodySize)
	}
}

func newMetricConfig() *aggregate.AggregatorConfigure {
	return &aggregate.AggregatorConfigure{
		DefaultWindow: 10 * time.Second,
		AggregateRules: []*aggregate.AggregateRule{
			{
				Name: "otel-jvm-memory",
				Selector: &aggregate.RuleSelector{
					Category:     point.SMetric,
					Measurements: []string{"opentelemetry"},
					MetricName:   []string{"jvm.buffer.memory.used"},
				},
				Groupby: []string{"service_name", "id"},
				Algorithms: map[string]*aggregate.AggregationAlgoConfig{
					"jvm.buffer.memory.used.max": {
						Method:      string(aggregate.MAX),
						SourceField: "jvm.buffer.memory.used",
						AddTags:     map[string]string{"method": "max"},
					},
				},
			},
		},
	}
}

func newTracingMetricConfig() *aggregate.AggregatorConfigure {
	return &aggregate.AggregatorConfigure{
		DefaultWindow: 10 * time.Second,
		AggregateRules: []*aggregate.AggregateRule{
			{
				Name: "trace-resource-count",
				Selector: &aggregate.RuleSelector{
					Category:     point.STracing,
					Measurements: []string{"ddtrace"},
					MetricName:   []string{"duration"},
				},
				Groupby: []string{"resource"},
				Algorithms: map[string]*aggregate.AggregationAlgoConfig{
					"duration.max": {
						Method:      string(aggregate.MAX),
						SourceField: "duration",
					},
				},
			},
		},
	}
}

func newLoggingMetricConfig() *aggregate.AggregatorConfigure {
	return &aggregate.AggregatorConfigure{
		DefaultWindow: 10 * time.Second,
		AggregateRules: []*aggregate.AggregateRule{
			{
				Name: "logging-message-count",
				Selector: &aggregate.RuleSelector{
					Category:     point.SLogging,
					Measurements: []string{"test_log"},
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
		},
	}
}

func newMetricPoints() []*point.Point {
	kvs := point.KVs{}
	kvs = kvs.Add("jvm.buffer.memory.used", float64(100)).
		AddTag("service_name", "test").
		AddTag("id", "1")

	return []*point.Point{
		point.NewPoint("opentelemetry", kvs, point.DefaultMetricOptions()...),
	}
}

func newMetricPointsForAggrToml() []*point.Point {
	legacyPoint := point.NewPoint("otel_service", point.KVs{}.
		Add("jvm.buffer.memory.used", float64(128)).
		AddTag("service_name", "svc-a").
		AddTag("id", "1"),
		point.DefaultMetricOptions()...)

	otelPoint := point.NewPoint("otel_service", point.KVs{}.
		Add("jvm.class.loaded", float64(320)).
		Add("jvm.class.unloaded", float64(11)).
		Add("jvm.threads.live", float64(27)).
		AddTag("service_name", "svc-a"),
		point.DefaultMetricOptions()...)

	unmatchedPoint := point.NewPoint("otel_service", point.KVs{}.
		Add("process.runtime.jvm.threads.count", float64(9)).
		AddTag("service_name", "svc-a"),
		point.DefaultMetricOptions()...)
	span_count := point.NewPoint("opentelemetry", point.KVs{}.
		Add("span_id", "123123123").
		AddTag("service_name", "svc-a").
		AddTag("trace_id", "9876543321").
		Add("parent_id", "0"),
		point.DefaultLoggingOptions()...,
	)

	return []*point.Point{legacyPoint, otelPoint, unmatchedPoint, span_count}
}

func selectedMetricFieldCounts(batchMap map[uint64]*aggregate.Batchs) map[string]int {
	fieldCounts := map[string]int{}
	for _, batchs := range batchMap {
		if batchs == nil {
			continue
		}

		for _, batch := range batchs.Batchs {
			if batch == nil || batch.Points == nil {
				continue
			}

			for _, pb := range batch.Points.Arr {
				pt := point.FromPB(pb)
				for _, field := range pt.Fields() {
					fieldCounts[field.Key]++
				}
			}
		}
	}

	return fieldCounts
}

func newLargeMetricPoints(count int, tagPayload int) []*point.Point {
	pts := make([]*point.Point, 0, count)
	payload := strings.Repeat("x", tagPayload)
	for i := 0; i < count; i++ {
		kvs := point.KVs{}
		kvs = kvs.Add("jvm.buffer.memory.used", float64(100+i)).
			AddTag("service_name", "svc-"+payload).
			AddTag("id", "id-1")
		pts = append(pts, point.NewPoint("opentelemetry", kvs, point.DefaultMetricOptions()...))
	}
	return pts
}

func newAPMPoints() []*point.Point {
	now := time.Now()

	pt1 := point.NewPoint("ddtrace", point.NewKVs(map[string]interface{}{
		"resource":   "/resource",
		"trace_id":   "1000000000",
		"span_id":    "123456789",
		"start_time": now.Unix(),
		"duration":   int64(1000),
	}), point.CommonLoggingOptions()...)
	pt1.SetTime(now)

	pt2 := point.NewPoint("ddtrace", point.NewKVs(map[string]interface{}{
		"resource":   "/resource",
		"trace_id":   "1000000000",
		"span_id":    "987654321",
		"start_time": now.Unix(),
		"duration":   int64(2000),
		"status":     "error",
	}), point.CommonLoggingOptions()...)
	pt2.SetTime(now)

	return []*point.Point{pt1, pt2}
}

func newLargeAPMPoints(count int, payloadSize int) []*point.Point {
	pts := make([]*point.Point, 0, count)
	now := time.Now()
	payload := strings.Repeat("y", payloadSize)
	for i := 0; i < count; i++ {
		pt := point.NewPoint("ddtrace", point.NewKVs(map[string]interface{}{
			"resource":   "/resource/" + payload,
			"trace_id":   "2000000000",
			"span_id":    strconv.Itoa(100000000 + i),
			"start_time": now.Unix(),
			"duration":   int64(1000 + i),
			"message":    payload,
		}), point.CommonLoggingOptions()...)
		pt.SetTime(now)
		pts = append(pts, pt)
	}

	return pts
}

func newLoggingPoints() []*point.Point {
	now := time.Now()
	grouped := point.NewPoint("test_log", point.KVs{}.
		Add("message", "grouped").
		Add("session_id", "s-1"),
		point.DefaultLoggingOptions()...)
	grouped.SetTime(now)

	passed := point.NewPoint("test_log", point.KVs{}.
		Add("message", "passthrough"),
		point.DefaultLoggingOptions()...)
	passed.SetTime(now)

	return []*point.Point{grouped, passed}
}

func newRUMPoints() []*point.Point {
	now := time.Now()
	grouped := point.NewPoint("test_rum", point.KVs{}.
		Add("view", "/home").
		Add("session_id", "rum-1"),
		point.CommonLoggingOptions()...)
	grouped.SetTime(now)

	passed := point.NewPoint("test_rum", point.KVs{}.
		Add("view", "/help"),
		point.CommonLoggingOptions()...)
	passed.SetTime(now)

	return []*point.Point{grouped, passed}
}

func newLoggingPointsForDimensionCollision() []*point.Point {
	now := time.Now()

	sessionGrouped := point.NewPoint("test_log", point.KVs{}.
		Add("message", "session-grouped").
		Add("session_id", "123"),
		point.DefaultLoggingOptions()...)
	sessionGrouped.SetTime(now)

	userGrouped := point.NewPoint("test_log", point.KVs{}.
		Add("message", "user-grouped").
		Add("user_id", "123"),
		point.DefaultLoggingOptions()...)
	userGrouped.SetTime(now)

	return []*point.Point{sessionGrouped, userGrouped}
}

func newRUMPointsForDimensionCollision() []*point.Point {
	now := time.Now()

	sessionGrouped := point.NewPoint("test_rum", point.KVs{}.
		Add("view", "/session").
		Add("session_id", "123"),
		point.CommonLoggingOptions()...)
	sessionGrouped.SetTime(now)

	userGrouped := point.NewPoint("test_rum", point.KVs{}.
		Add("view", "/user").
		Add("user_id", "123"),
		point.CommonLoggingOptions()...)
	userGrouped.SetTime(now)

	return []*point.Point{sessionGrouped, userGrouped}
}
