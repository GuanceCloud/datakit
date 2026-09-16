// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

const (
	defaultQPSMaxSeries         = 1000
	qpsMaxPendingSeconds        = 120
	qpsShardCount               = 16
	qpsStackTagCount            = 16
	qpsOverflowTag              = "qps_overflow"
	qpsDropReasonRetentionLimit = "retention_limit"
)

func defaultQPSTagNames() []string {
	return []string{
		itrace.TagService,
		itrace.TagEnv,
		itrace.TagVersion,
		itrace.TagSpanKind,
		itrace.TagHttpMethod,
		itrace.TagHttpStatusCode,
	}
}

type qpsTagSpec struct {
	name     string
	metaKeys []string
}

type qpsSeriesHash struct {
	first  uint64
	second uint64
}

type qpsSeries struct {
	values   []string
	overflow bool
}

type qpsShard struct {
	mu sync.Mutex

	buckets        map[int64]map[*qpsSeries]uint64
	activeSeries   map[qpsSeriesHash][]*qpsSeries
	overflowSeries *qpsSeries
	newestSecond   int64
	hasNewest      bool
}

// qpsAggregator keeps filtering, grouping, cardinality protection and second
// buckets behind a small concurrent interface. Drain excludes observations;
// observations proceed concurrently through independent shards.
type qpsAggregator struct {
	gate sync.RWMutex

	tags       []qpsTagSpec
	fixedTags  map[string]string
	maxSeries  int64
	seriesUsed int64
	// oldestOpenSecond prevents worker-pool backlog from reopening a second
	// whose gauge point has already been emitted.
	oldestOpenSecond int64
	shards           [qpsShardCount]qpsShard
}

func newQPSAggregator(tags []string, fixedTags map[string]string, maxSeries int) *qpsAggregator {
	if maxSeries <= 0 {
		maxSeries = defaultQPSMaxSeries
	}

	a := &qpsAggregator{
		tags:      newQPSTagSpecs(tags, fixedTags),
		fixedTags: make(map[string]string, len(fixedTags)),
		maxSeries: int64(maxSeries),
	}
	for k, v := range fixedTags {
		a.fixedTags[k] = v
	}
	for i := range a.shards {
		a.shards[i].buckets = make(map[int64]map[*qpsSeries]uint64)
		a.shards[i].activeSeries = make(map[qpsSeriesHash][]*qpsSeries)
		a.shards[i].overflowSeries = &qpsSeries{overflow: true}
	}

	return a
}

func newQPSTagSpecs(tags []string, fixedTags map[string]string) []qpsTagSpec {
	seen := make(map[string]struct{}, len(tags))
	result := make([]qpsTagSpec, 0, len(tags))

	for _, configuredName := range tags {
		configuredName = strings.TrimSpace(configuredName)
		if configuredName == "" {
			continue
		}

		name := strings.ReplaceAll(configuredName, ".", "_")
		if _, ok := seen[name]; ok {
			continue
		}
		if _, fixed := fixedTags[name]; fixed {
			// A fixed input tag wins when points are built. Grouping by the
			// span value would create duplicate points with identical tags.
			continue
		}
		seen[name] = struct{}{}

		keys := []string{configuredName}
		if configuredName != name {
			keys = append(keys, name)
		}
		var mapped []string
		for rawName, normalizedName := range ddTags {
			if normalizedName == name {
				mapped = append(mapped, rawName)
			}
		}
		sort.Strings(mapped)
		for _, key := range mapped {
			if !stringSliceContains(keys, key) {
				keys = append(keys, key)
			}
		}

		result = append(result, qpsTagSpec{name: name, metaKeys: keys})
	}

	return result
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// ObserveTrace counts all HTTP spans. A span is HTTP when either the request
// method or response status code is present; span kind and topology are only
// grouping dimensions and deliberately do not filter client calls.
func (a *qpsAggregator) ObserveTrace(trace DDTrace, observedAt time.Time, remoteIP string) {
	if a == nil || len(trace) == 0 {
		return
	}
	if observedAt.IsZero() {
		observedAt = time.Now()
	}
	second := observedAt.Unix()

	a.gate.RLock()
	defer a.gate.RUnlock()
	if second < a.oldestOpenSecond {
		second = a.oldestOpenSecond
	}

	for _, span := range trace {
		if !isHTTPSpanForQPS(span) {
			continue
		}
		a.observeSpan(span, second, remoteIP)
	}
}

func (a *qpsAggregator) observeSpan(span *DDSpan, second int64, remoteIP string) {
	var stackValues [qpsStackTagCount]string
	values := stackValues[:0]
	if len(a.tags) > len(stackValues) {
		values = make([]string, 0, len(a.tags))
	}
	for _, spec := range a.tags {
		values = append(values, qpsTagValue(spec, span, remoteIP))
	}
	hash := qpsValuesHash(values)
	shard := &a.shards[hash.first%qpsShardCount]

	shard.mu.Lock()
	dropped, accepted := a.prepareSecond(shard, second)
	if !accepted {
		dropped++
		shard.mu.Unlock()
		droppedQPSSpans.WithLabelValues(qpsDropReasonRetentionLimit).Add(float64(dropped))
		return
	}
	series := findQPSSeries(shard.activeSeries[hash], values)
	if series == nil {
		if a.reserveSeries() {
			series = &qpsSeries{values: append([]string(nil), values...)}
			shard.activeSeries[hash] = append(shard.activeSeries[hash], series)
		} else {
			series = shard.overflowSeries
		}
	}

	bucket := shard.buckets[second]
	if bucket == nil {
		bucket = make(map[*qpsSeries]uint64)
		shard.buckets[second] = bucket
	}
	bucket[series]++
	shard.mu.Unlock()
	if dropped > 0 {
		droppedQPSSpans.WithLabelValues(qpsDropReasonRetentionLimit).Add(float64(dropped))
	}
}

// prepareSecond bounds the number of pending seconds retained by one shard.
// Keeping the newest data is more useful after a blocked metric Feed recovers,
// while dropping old buckets prevents memory from growing with stall duration.
// The caller must hold shard.mu.
func (a *qpsAggregator) prepareSecond(shard *qpsShard, second int64) (uint64, bool) {
	if shard.hasNewest && second <= shard.newestSecond-qpsMaxPendingSeconds {
		return 0, false
	}
	if shard.hasNewest && second <= shard.newestSecond {
		return 0, true
	}

	shard.newestSecond = second
	shard.hasNewest = true
	oldestRetained := second - qpsMaxPendingSeconds + 1
	var dropped uint64
	for bucketSecond, groups := range shard.buckets {
		if bucketSecond >= oldestRetained {
			continue
		}
		for _, count := range groups {
			dropped += count
		}
		delete(shard.buckets, bucketSecond)
	}
	if dropped == 0 {
		return 0, true
	}

	activeBefore := qpsActiveSeriesCount(shard)
	activeAfter := rebuildQPSActiveSeries(shard)
	atomic.AddInt64(&a.seriesUsed, int64(activeAfter-activeBefore))
	return dropped, true
}

func qpsActiveSeriesCount(shard *qpsShard) int {
	count := 0
	for _, candidates := range shard.activeSeries {
		count += len(candidates)
	}
	return count
}

func findQPSSeries(candidates []*qpsSeries, values []string) *qpsSeries {
	for _, candidate := range candidates {
		if len(candidate.values) != len(values) {
			continue
		}
		matched := true
		for i := range values {
			if candidate.values[i] != values[i] {
				matched = false
				break
			}
		}
		if matched {
			return candidate
		}
	}
	return nil
}

func (a *qpsAggregator) reserveSeries() bool {
	for {
		used := atomic.LoadInt64(&a.seriesUsed)
		if used >= a.maxSeries {
			return false
		}
		if atomic.CompareAndSwapInt64(&a.seriesUsed, used, used+1) {
			return true
		}
	}
}

type qpsDetachedBucket struct {
	second int64
	groups map[*qpsSeries]uint64
}

// Drain returns completed seconds only. The current partial second remains in
// memory so one timestamp is never emitted twice as two partial gauge values.
func (a *qpsAggregator) Drain(before time.Time) []*point.Point {
	if a == nil {
		return nil
	}
	if before.IsZero() {
		before = time.Now()
	}
	cutoff := before.Truncate(time.Second).Unix()

	var detached []qpsDetachedBucket
	a.gate.Lock()
	if cutoff > a.oldestOpenSecond {
		a.oldestOpenSecond = cutoff
	}
	remainingSeries := int64(0)
	for i := range a.shards {
		shard := &a.shards[i]
		shard.mu.Lock()
		for second, groups := range shard.buckets {
			if second < cutoff {
				detached = append(detached, qpsDetachedBucket{second: second, groups: groups})
				delete(shard.buckets, second)
			}
		}
		remainingSeries += int64(rebuildQPSActiveSeries(shard))
		shard.mu.Unlock()
	}
	atomic.StoreInt64(&a.seriesUsed, remainingSeries)
	a.gate.Unlock()

	return a.detachedPoints(detached)
}

func rebuildQPSActiveSeries(shard *qpsShard) int {
	active := make(map[qpsSeriesHash][]*qpsSeries)
	seen := make(map[*qpsSeries]struct{})
	for _, groups := range shard.buckets {
		for series := range groups {
			if series.overflow {
				continue
			}
			if _, ok := seen[series]; ok {
				continue
			}
			seen[series] = struct{}{}
			hash := qpsValuesHash(series.values)
			active[hash] = append(active[hash], series)
		}
	}
	shard.activeSeries = active
	return len(seen)
}

func (a *qpsAggregator) detachedPoints(detached []qpsDetachedBucket) []*point.Point {
	if len(detached) == 0 {
		return nil
	}

	type record struct {
		second int64
		series *qpsSeries
		count  uint64
	}
	overflow := make(map[int64]uint64)
	var records []record
	for _, bucket := range detached {
		for series, count := range bucket.groups {
			if series.overflow {
				overflow[bucket.second] += count
			} else {
				records = append(records, record{second: bucket.second, series: series, count: count})
			}
		}
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].second != records[j].second {
			return records[i].second < records[j].second
		}
		return compareQPSValues(records[i].series.values, records[j].series.values) < 0
	})

	points := make([]*point.Point, 0, len(records)+len(overflow))
	for _, item := range records {
		points = append(points, a.point(item.second, item.series.values, item.count, false))
	}
	seconds := make([]int64, 0, len(overflow))
	for second := range overflow {
		seconds = append(seconds, second)
	}
	sort.Slice(seconds, func(i, j int) bool { return seconds[i] < seconds[j] })
	for _, second := range seconds {
		points = append(points, a.point(second, nil, overflow[second], true))
	}

	return points
}

func compareQPSValues(left, right []string) int {
	for i := 0; i < len(left) && i < len(right); i++ {
		if comparison := strings.Compare(left[i], right[i]); comparison != 0 {
			return comparison
		}
	}
	return len(left) - len(right)
}

func (a *qpsAggregator) point(second int64, values []string, count uint64, overflow bool) *point.Point {
	kvs := point.NewTags(a.fixedTags)
	kvs = kvs.AddTag(itrace.TagSource, inputName)
	if overflow {
		kvs = kvs.SetTag(qpsOverflowTag, "true")
	} else {
		for i, spec := range a.tags {
			if values[i] != "" {
				kvs = kvs.AddTag(spec.name, values[i])
			}
		}
	}
	kvs = kvs.Add("qps", int64(count))
	opts := append(point.DefaultMetricOptions(), point.WithTime(time.Unix(second, 0)))

	return point.NewPoint(itrace.TracingMetricName, kvs, opts...)
}

func isHTTPSpanForQPS(span *DDSpan) bool {
	if span == nil {
		return false
	}
	return firstMetaValue(span.Meta, "http.method", "http_method") != "" ||
		firstMetaValue(span.Meta, "http.status_code", "http_status_code") != ""
}

func qpsTagValue(spec qpsTagSpec, span *DDSpan, remoteIP string) string {
	switch spec.name {
	case itrace.TagService:
		return span.Service
	case itrace.TagOperation:
		return span.Name
	case itrace.FieldResource:
		return strings.ReplaceAll(span.Resource, "\n", " ")
	case itrace.TagSource:
		return inputName
	case itrace.TagSpanStatus:
		if span.Error != 0 {
			return itrace.StatusErr
		}
		return itrace.StatusOk
	case itrace.TagHttpStatusClass:
		return itrace.GetClass(firstMetaValue(span.Meta, "http.status_code", "http_status_code"))
	case itrace.TagCollectorSourceIP:
		return remoteIP
	case itrace.TagSourceType:
		return itrace.GetSpanSourceType(span.Type)
	case "http_direction":
		switch firstMetaValue(span.Meta, "span.kind", "span_kind") {
		case "server":
			return "inbound"
		case "client":
			return "outbound"
		default:
			return "unknown"
		}
	default:
		return firstMetaValue(span.Meta, spec.metaKeys...)
	}
}

func firstMetaValue(meta map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := meta[key]; value != "" {
			return value
		}
	}
	return ""
}

func qpsValuesHash(values []string) qpsSeriesHash {
	const (
		offsetA = uint64(14695981039346656037)
		offsetB = uint64(7809847782465536322)
		primeA  = uint64(1099511628211)
		primeB  = uint64(14029467366897019727)
	)

	result := qpsSeriesHash{first: offsetA, second: offsetB}
	for _, value := range values {
		length := uint64(len(value))
		for shift := 0; shift < 64; shift += 8 {
			part := byte(length >> shift)
			result.first = (result.first ^ uint64(part)) * primeA
			result.second = (result.second ^ uint64(part)) * primeB
		}
		for i := 0; i < len(value); i++ {
			result.first = (result.first ^ uint64(value[i])) * primeA
			result.second = (result.second ^ uint64(value[i])) * primeB
		}
	}
	return result
}
