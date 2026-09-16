// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"fmt"
	"sync"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func TestQPSAggregatorHTTPFilterAndGrouping(t *testing.T) {
	base := time.Unix(1_700_000_000, 500)
	agg := newQPSAggregator([]string{"service", "span.kind", "http_method", "http_status_code"}, nil, 100)

	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc-a", map[string]string{"http.method": "GET", "span.kind": "server"}),
		qpsTestSpan("svc-a", map[string]string{"http.status_code": "200", "span.kind": "client"}),
		qpsTestSpan("svc-a", map[string]string{"http.method": "GET", "span.kind": "server"}),
		qpsTestSpan("svc-a", map[string]string{"span.kind": "server"}),
		nil,
	}, base, "10.0.0.1")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 2)

	counts := make(map[string]int64)
	for _, pt := range pts {
		value, ok := pt.GetI("qps")
		require.True(t, ok)
		counts[pt.GetTag(itrace.TagSpanKind)+"/"+pt.GetTag(itrace.TagHttpMethod)+"/"+pt.GetTag(itrace.TagHttpStatusCode)] = value
		assert.Equal(t, base.Truncate(time.Second), pt.Time())
		assert.Equal(t, inputName, pt.GetTag(itrace.TagSource))
	}
	assert.Equal(t, int64(2), counts["server/GET/"])
	assert.Equal(t, int64(1), counts["client//200"])
}

func TestQPSAggregatorAcceptsNormalizedHTTPMeta(t *testing.T) {
	base := time.Unix(1_700_000_100, 0)
	agg := newQPSAggregator([]string{"http_method", "http_status_code"}, nil, 100)
	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc", map[string]string{"http_method": "POST"}),
		qpsTestSpan("svc", map[string]string{"http_status_code": "503"}),
	}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 2)
	assert.ElementsMatch(t, []string{"POST", ""}, []string{
		pts[0].GetTag(itrace.TagHttpMethod), pts[1].GetTag(itrace.TagHttpMethod),
	})
}

func TestQPSAggregatorSecondBuckets(t *testing.T) {
	base := time.Unix(1_700_000_200, 0)
	agg := newQPSAggregator([]string{"service"}, map[string]string{"region": "cn"}, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET"})

	agg.ObserveTrace(DDTrace{span}, base, "")
	agg.ObserveTrace(DDTrace{span}, base.Add(1500*time.Millisecond), "")

	// The second containing the cutoff is partial and must be retained.
	first := agg.Drain(base.Add(time.Second))
	require.Len(t, first, 1)
	assert.Equal(t, base, first[0].Time())
	assert.Equal(t, "cn", first[0].GetTag("region"))

	second := agg.Drain(base.Add(2 * time.Second))
	require.Len(t, second, 1)
	assert.Equal(t, base.Add(time.Second), second[0].Time())
	assert.Empty(t, agg.Drain(base.Add(3*time.Second)))
}

func TestQPSAggregatorDoesNotReopenDrainedSecond(t *testing.T) {
	base := time.Unix(1_700_000_250, 0)
	agg := newQPSAggregator([]string{"service"}, nil, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET"})

	agg.ObserveTrace(DDTrace{span}, base, "")
	first := agg.Drain(base.Add(time.Second))
	require.Len(t, first, 1)
	assert.Equal(t, base, first[0].Time())

	// Simulate a request timestamped before a worker-pool backlog. Since that
	// second is already emitted, its count moves to the oldest open second.
	agg.ObserveTrace(DDTrace{span}, base, "")
	second := agg.Drain(base.Add(2 * time.Second))
	require.Len(t, second, 1)
	assert.Equal(t, base.Add(time.Second), second[0].Time())
}

func TestQPSAggregatorCardinalityOverflow(t *testing.T) {
	base := time.Unix(1_700_000_300, 0)
	agg := newQPSAggregator([]string{"service"}, nil, 1)
	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc-a", map[string]string{"http.method": "GET"}),
		qpsTestSpan("svc-b", map[string]string{"http.method": "GET"}),
		qpsTestSpan("svc-c", map[string]string{"http.status_code": "500"}),
	}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 2)
	var regular, overflow *point.Point
	for _, pt := range pts {
		if pt.GetTag(qpsOverflowTag) == "true" {
			overflow = pt
		} else {
			regular = pt
		}
	}
	require.NotNil(t, regular)
	require.NotNil(t, overflow)
	regularCount, ok := regular.GetI("qps")
	require.True(t, ok)
	overflowCount, ok := overflow.GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(1), regularCount)
	assert.Equal(t, int64(2), overflowCount)
}

func TestQPSAggregatorFixedTagOverridesGroupingTag(t *testing.T) {
	base := time.Unix(1_700_000_400, 0)
	agg := newQPSAggregator([]string{"service"}, map[string]string{"service": "fixed"}, 100)
	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc-a", map[string]string{"http.method": "GET"}),
		qpsTestSpan("svc-b", map[string]string{"http.method": "GET"}),
	}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	assert.Equal(t, "fixed", pts[0].GetTag(itrace.TagService))
	value, ok := pts[0].GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(2), value)
}

func TestQPSAggregatorDirectionIsOnlyADimension(t *testing.T) {
	base := time.Unix(1_700_000_500, 0)
	agg := newQPSAggregator([]string{"http_direction"}, nil, 100)
	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc", map[string]string{"http.method": "GET", "span.kind": "server"}),
		qpsTestSpan("svc", map[string]string{"http.method": "POST", "span.kind": "client"}),
		qpsTestSpan("svc", map[string]string{"http.status_code": "200"}),
	}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 3)
	directions := make([]string, 0, len(pts))
	for _, pt := range pts {
		directions = append(directions, pt.GetTag("http_direction"))
	}
	assert.ElementsMatch(t, []string{"inbound", "outbound", "unknown"}, directions)
}

func TestQPSAggregatorConcurrentObserveAndDrain(t *testing.T) {
	base := time.Unix(1_700_000_600, 0)
	agg := newQPSAggregator([]string{"service"}, nil, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET"})

	const goroutines = 32
	const iterations = 100
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				agg.ObserveTrace(DDTrace{span}, base, "")
			}
		}()
	}
	wg.Wait()

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	value, ok := pts[0].GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(goroutines*iterations), value)
}

func TestQPSAggregatorDefaultTags(t *testing.T) {
	base := time.Unix(1_700_000_700, 0)
	agg := newQPSAggregator(defaultQPSTagNames(), nil, 100)
	agg.ObserveTrace(DDTrace{qpsTestSpan("svc", map[string]string{
		"env":              "prod",
		"version":          "1.2.3",
		"span.kind":        "client",
		"http.method":      "GET",
		"http.status_code": "200",
	})}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	assert.Equal(t, "svc", pts[0].GetTag(itrace.TagService))
	assert.Equal(t, "prod", pts[0].GetTag(itrace.TagEnv))
	assert.Equal(t, "1.2.3", pts[0].GetTag(itrace.TagVersion))
	assert.Equal(t, "client", pts[0].GetTag(itrace.TagSpanKind))
	assert.Equal(t, "GET", pts[0].GetTag(itrace.TagHttpMethod))
	assert.Equal(t, "200", pts[0].GetTag(itrace.TagHttpStatusCode))
}

func TestQPSAggregatorAllowsNoGroupingTags(t *testing.T) {
	base := time.Unix(1_700_000_720, 0)
	agg := newQPSAggregator([]string{}, nil, 100)
	agg.ObserveTrace(DDTrace{
		qpsTestSpan("svc-a", map[string]string{"http.method": "GET"}),
		qpsTestSpan("svc-b", map[string]string{"http.status_code": "200"}),
	}, base, "")

	pts := agg.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	value, ok := pts[0].GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(2), value)
}

func TestQPSAggregationRunsBeforeTraceTruncation(t *testing.T) {
	base := time.Unix(1_700_000_750, 0)
	ipt := defaultInput()
	ipt.traceMaxSpans = 1
	ipt.qpsAggregator = newQPSAggregator([]string{"service"}, nil, 100)
	ipt.customTagsX = itrace.NewCustomTags(nil, ddTags)

	trace := DDTrace{
		qpsTestSpan("svc", map[string]string{"db.type": "sql"}),
		qpsTestSpan("svc", map[string]string{"http.method": "GET"}),
	}
	dktrace := ipt.ddtraceToDkTrace(trace, nil, "127.0.0.1", base)
	require.Len(t, dktrace, 1)

	pts := ipt.qpsAggregator.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	value, ok := pts[0].GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(1), value)
}

func TestQPSAggregationRunsBeforePriorityDrop(t *testing.T) {
	base := time.Unix(1_700_000_760, 0)
	ipt := defaultInput()
	ipt.qpsAggregator = newQPSAggregator([]string{"service"}, nil, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET"})
	span.Metrics = map[string]float64{keyPriority: -1}

	dktrace := ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", base)
	assert.Empty(t, dktrace)

	pts := ipt.qpsAggregator.Drain(base.Add(time.Second))
	require.Len(t, pts, 1)
	value, ok := pts[0].GetI("qps")
	require.True(t, ok)
	assert.Equal(t, int64(1), value)
}

func TestQPSAggregatorBoundsSecondBucketsBeforePriorityDrop(t *testing.T) {
	droppedBefore := qpsDroppedSpanCount(t)
	feeder := &blockingQPSFeeder{
		MockedFeeder: dkio.NewMockedFeeder(),
		entered:      make(chan struct{}),
		release:      make(chan struct{}),
	}
	ipt := defaultInput()
	ipt.feeder = feeder
	ipt.qpsAggregator = newQPSAggregator([]string{"service"}, nil, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET"})
	span.Metrics = map[string]float64{keyPriority: -1}

	// Seed a completed bucket so gatherMetrics reaches the blocking Feed.
	ipt.qpsAggregator.ObserveTrace(DDTrace{span}, time.Now().Add(-2*time.Second), "127.0.0.1")
	gatherDone := make(chan struct{})
	var releaseOnce sync.Once
	releaseFeed := func() {
		releaseOnce.Do(func() { close(feeder.release) })
	}
	t.Cleanup(releaseFeed)
	go func() {
		ipt.gatherMetrics()
		close(gatherDone)
	}()
	select {
	case <-feeder.entered:
	case <-time.After(time.Second):
		t.Fatal("gatherMetrics did not reach metric Feed")
	}

	base := time.Now().Add(time.Second).Truncate(time.Second)
	for second := 0; second <= qpsMaxPendingSeconds; second++ {
		dktrace := ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", base.Add(time.Duration(second)*time.Second))
		assert.Empty(t, dktrace)
	}
	// An observation older than the retained window must not reopen an
	// already-evicted bucket.
	assert.Empty(t, ipt.ddtraceToDkTrace(DDTrace{span}, nil, "127.0.0.1", base))

	buckets := 0
	entries := 0
	for i := range ipt.qpsAggregator.shards {
		shard := &ipt.qpsAggregator.shards[i]
		buckets += len(shard.buckets)
		for _, groups := range shard.buckets {
			entries += len(groups)
		}
	}
	assert.LessOrEqual(t, buckets, qpsMaxPendingSeconds)
	assert.LessOrEqual(t, entries, qpsMaxPendingSeconds)
	assert.Equal(t, float64(2), qpsDroppedSpanCount(t)-droppedBefore)

	pts := ipt.qpsAggregator.Drain(base.Add((qpsMaxPendingSeconds + 1) * time.Second))
	require.Len(t, pts, qpsMaxPendingSeconds)
	assert.Equal(t, base.Add(time.Second), pts[0].Time())
	assert.Equal(t, base.Add(qpsMaxPendingSeconds*time.Second), pts[len(pts)-1].Time())

	releaseFeed()
	select {
	case <-gatherDone:
	case <-time.After(time.Second):
		t.Fatal("gatherMetrics did not return after Feed was released")
	}
}

func qpsDroppedSpanCount(t *testing.T) float64 {
	t.Helper()
	metric := &dto.Metric{}
	require.NoError(t, droppedQPSSpans.WithLabelValues(qpsDropReasonRetentionLimit).Write(metric))
	return metric.GetCounter().GetValue()
}

func TestQPSAggregatorEvictionReleasesSeries(t *testing.T) {
	base := time.Unix(1_700_000_780, 0)
	serviceA := "svc-a"
	serviceB := qpsTestServiceInSameShard(t, serviceA)
	agg := newQPSAggregator([]string{"service"}, nil, 1)

	agg.ObserveTrace(DDTrace{qpsTestSpan(serviceA, map[string]string{"http.method": "GET"})}, base, "")
	agg.ObserveTrace(DDTrace{qpsTestSpan(serviceB, map[string]string{"http.method": "GET"})},
		base.Add(qpsMaxPendingSeconds*time.Second), "")

	pts := agg.Drain(base.Add((qpsMaxPendingSeconds + 1) * time.Second))
	require.Len(t, pts, 1)
	assert.Equal(t, serviceB, pts[0].GetTag(itrace.TagService))
	assert.Empty(t, pts[0].GetTag(qpsOverflowTag))
}

func qpsTestServiceInSameShard(t *testing.T, service string) string {
	t.Helper()
	wantShard := qpsValuesHash([]string{service}).first % qpsShardCount
	for i := 0; i < 1000; i++ {
		candidate := fmt.Sprintf("svc-same-shard-%d", i)
		if qpsValuesHash([]string{candidate}).first%qpsShardCount == wantShard {
			return candidate
		}
	}
	t.Fatal("failed to find test service in the same QPS shard")
	return ""
}

type blockingQPSFeeder struct {
	*dkio.MockedFeeder
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

func (f *blockingQPSFeeder) Feed(point.Category, []*point.Point, ...dkio.FeedOption) error {
	f.once.Do(func() { close(f.entered) })
	<-f.release
	return nil
}

func TestGatherMetricsWithQPSOnly(t *testing.T) {
	feeder := dkio.NewMockedFeeder()
	ipt := &Input{
		feeder:                 feeder,
		TracingMetricQPSEnable: true,
		qpsAggregator:          newQPSAggregator([]string{"service"}, nil, 100),
	}
	observedAt := time.Now().Add(-2 * time.Second).Truncate(time.Second)
	ipt.qpsAggregator.ObserveTrace(DDTrace{
		qpsTestSpan("svc", map[string]string{"http.method": "GET"}),
	}, observedAt, "")

	ipt.gatherMetrics()
	pts, err := feeder.AnyPoints(time.Second)
	require.NoError(t, err)
	require.Len(t, pts, 1)
	assert.Equal(t, observedAt, pts[0].Time())
}

func BenchmarkQPSAggregatorObserveTrace(b *testing.B) {
	agg := newQPSAggregator([]string{"service", "http_method", "http_status_code"}, nil, 100)
	span := qpsTestSpan("svc", map[string]string{"http.method": "GET", "http.status_code": "200"})
	trace := DDTrace{span}
	now := time.Unix(1_700_000_800, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg.ObserveTrace(trace, now, "")
	}
}

func qpsTestSpan(service string, meta map[string]string) *DDSpan {
	return &DDSpan{
		Service:  service,
		Name:     "request",
		Resource: fmt.Sprintf("GET /%s", service),
		Meta:     meta,
	}
}
