package aggregate

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
)

type derivedMetricKey struct {
	token       string
	measurement string
	metricName  string
	stage       DerivedMetricStage
	decision    DerivedMetricDecision
	tagsHash    uint64
}

type derivedMetricValue struct {
	token       string
	measurement string
	metricName  string
	stage       DerivedMetricStage
	decision    DerivedMetricDecision
	tags        map[string]string
	sum         float64
}

type derivedHistogramValue struct {
	token       string
	measurement string
	metricName  string
	stage       DerivedMetricStage
	decision    DerivedMetricDecision
	tags        map[string]string
	buckets     []float64
	counts      []float64
	sum         float64
	count       float64
}

type derivedMetricBucket struct {
	sums map[derivedMetricKey]*derivedMetricValue
	hist map[derivedMetricKey]*derivedHistogramValue
}

type DerivedMetricCollector struct {
	mu      sync.Mutex
	window  time.Duration
	buckets map[int64]*derivedMetricBucket
}

const DefaultDerivedMetricFlushWindow = 30 * time.Second

func NewDerivedMetricCollector(window time.Duration) *DerivedMetricCollector {
	if window <= 0 {
		window = DefaultDerivedMetricFlushWindow
	}

	return &DerivedMetricCollector{
		window:  window,
		buckets: make(map[int64]*derivedMetricBucket),
	}
}

func (c *DerivedMetricCollector) Add(records []DerivedMetricRecord) {
	if len(records) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, record := range records {
		timestamp := record.Time
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		exp := AlignNextWallTime(timestamp, c.window)
		key := newDerivedMetricKey(record)
		bucket := c.buckets[exp]
		if bucket == nil {
			bucket = &derivedMetricBucket{
				sums: make(map[derivedMetricKey]*derivedMetricValue),
				hist: make(map[derivedMetricKey]*derivedHistogramValue),
			}
			c.buckets[exp] = bucket
		}

		switch record.Kind {
		case DerivedMetricKindHistogram:
			current := bucket.hist[key]
			if current == nil {
				current = newDerivedHistogramValue(record)
				bucket.hist[key] = current
			}
			current.observe(record.Value)
		case "", DerivedMetricKindSum:
			current := bucket.sums[key]
			if current == nil {
				current = &derivedMetricValue{
					token:       record.Token,
					measurement: record.measurement(),
					metricName:  record.MetricName,
					stage:       record.Stage,
					decision:    record.Decision,
					tags:        cloneTags(record.Tags),
				}
				bucket.sums[key] = current
			}
			current.sum += record.Value
		default:
			continue
		}
	}
}

func (c *DerivedMetricCollector) Flush(now time.Time) []*DerivedMetricPoints {
	c.mu.Lock()
	defer c.mu.Unlock()

	nowUnix := now.Unix()
	grouped := make(map[string][]*point.Point)

	for exp, bucket := range c.buckets {
		if exp > nowUnix {
			continue
		}

		for _, val := range bucket.sums {
			grouped[val.token] = append(grouped[val.token], val.toPoint(exp))
		}
		for _, val := range bucket.hist {
			grouped[val.token] = append(grouped[val.token], val.toPoints(exp)...)
		}

		delete(c.buckets, exp)
	}

	res := make([]*DerivedMetricPoints, 0, len(grouped))
	for token, pts := range grouped {
		res = append(res, &DerivedMetricPoints{
			Token: token,
			PTS:   pts,
		})
	}

	return res
}

func newDerivedMetricKey(record DerivedMetricRecord) derivedMetricKey {
	return derivedMetricKey{
		token:       record.Token,
		measurement: record.measurement(),
		metricName:  record.MetricName,
		stage:       record.Stage,
		decision:    record.Decision,
		tagsHash:    hashDerivedMetricTags(record.Tags),
	}
}

func (v *derivedMetricValue) toPoint(exp int64) *point.Point {
	kvs := point.KVs{}
	kvs = kvs.Add(v.metricName, v.sum)

	if v.stage != "" {
		kvs = kvs.AddTag("stage", string(v.stage))
	}

	if v.decision != "" {
		kvs = kvs.AddTag("decision", string(v.decision))
	}

	keys := sortedTagKeys(v.tags)
	for _, key := range keys {
		kvs = kvs.AddTag(key, v.tags[key])
	}

	return point.NewPoint(v.measurement, kvs, point.WithTimestamp(exp*int64(time.Second)))
}

func newDerivedHistogramValue(record DerivedMetricRecord) *derivedHistogramValue {
	buckets := append([]float64(nil), record.Buckets...)
	sort.Float64s(buckets)

	return &derivedHistogramValue{
		token:       record.Token,
		measurement: record.measurement(),
		metricName:  record.MetricName,
		stage:       record.Stage,
		decision:    record.Decision,
		tags:        cloneTags(record.Tags),
		buckets:     buckets,
		counts:      make([]float64, len(buckets)+1), // last bucket is +Inf
	}
}

func (v *derivedHistogramValue) observe(val float64) {
	v.sum += val
	v.count += 1
	idx := len(v.buckets) // +Inf
	for i, le := range v.buckets {
		if val <= le {
			idx = i
			break
		}
	}
	v.counts[idx] += 1
}

func (v *derivedHistogramValue) toPoints(exp int64) []*point.Point {
	ts := exp * int64(time.Second)
	pts := make([]*point.Point, 0, len(v.counts)+2)

	cumulative := 0.0
	for i := range v.counts {
		cumulative += v.counts[i]
		le := "+Inf"
		if i < len(v.buckets) {
			le = trimFloat(v.buckets[i])
		}
		pts = append(pts, point.NewPoint(v.measurement, v.buildKVs(v.metricName+"_bucket", cumulative, "le", le), point.WithTimestamp(ts)))
	}

	pts = append(pts, point.NewPoint(v.measurement, v.buildKVs(v.metricName+"_sum", v.sum), point.WithTimestamp(ts)))
	pts = append(pts, point.NewPoint(v.measurement, v.buildKVs(v.metricName+"_count", v.count), point.WithTimestamp(ts)))

	return pts
}

func (v *derivedHistogramValue) buildKVs(field string, value float64, extraTags ...string) point.KVs {
	kvs := point.KVs{}
	kvs = kvs.Add(field, value)

	if v.stage != "" {
		kvs = kvs.AddTag("stage", string(v.stage))
	}
	if v.decision != "" {
		kvs = kvs.AddTag("decision", string(v.decision))
	}

	keys := sortedTagKeys(v.tags)
	for _, key := range keys {
		kvs = kvs.AddTag(key, v.tags[key])
	}

	if len(extraTags) == 2 {
		kvs = kvs.AddTag(extraTags[0], extraTags[1])
	}

	return kvs
}

func hashDerivedMetricTags(tags map[string]string) uint64 {
	if len(tags) == 0 {
		return 0
	}

	keys := sortedTagKeys(tags)
	hash := Seed1
	for _, key := range keys {
		hash = HashCombine(hash, xxhash.Sum64(cliutils.ToUnsafeBytes(key)))
		hash = HashCombine(hash, xxhash.Sum64(cliutils.ToUnsafeBytes(tags[key])))
	}

	return hash
}

func sortedTagKeys(tags map[string]string) []string {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return nil
	}

	res := make(map[string]string, len(tags))
	for key, value := range tags {
		res[key] = value
	}
	return res
}

func (c *DerivedMetricCollector) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return "derived-metric-collector buckets=" + strconv.Itoa(len(c.buckets))
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
