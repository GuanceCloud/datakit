// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plmap

import (
	"sync"
	"time"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

type AggBuckets struct {
	uploadDataFn UploadFunc
	// key: [hash(name), hash(...tagName)]
	data map[string]*bucket

	sync.RWMutex
}

func NewAggBuks(upFn UploadFunc) *AggBuckets {
	return &AggBuckets{
		uploadDataFn: upFn,
		data:         map[string]*bucket{},
	}
}

func (a *AggBuckets) CreateBucket(name string, interval time.Duration, count int,
	keepValue bool, constTags map[string]string,
) {
	a.Lock()
	defer a.Unlock()

	if a.data == nil {
		a.data = map[string]*bucket{}
	}

	buk, ok := a.data[name]
	if !ok {
		buk = newBucket(name, interval, count, keepValue, a.uploadDataFn)
		a.data[name] = buk
		buk.startScan()
	}

	buk.setExtraTag(constTags)
}

func (a *AggBuckets) SetUploadFunc(fn UploadFunc) {
	a.Lock()
	defer a.Unlock()
	a.uploadDataFn = fn
}

func (a *AggBuckets) StopAllBukScanner() {
	a.Lock()
	defer a.Unlock()

	for name, b := range a.data {
		b.stopScan()
		delete(a.data, name)
	}
}

func (a *AggBuckets) GetBucket(name string) (*bucket, bool) {
	a.RLock()
	defer a.RUnlock()

	if a.data == nil {
		return nil, false
	}
	v, ok := a.data[name]
	return v, ok
}

type aggFields struct {
	tags   []string
	fields map[string]aggMetric
}

type ptsGroup struct {
	timeline map[uint64]*aggFields

	countLimit int
}

func (g *ptsGroup) addMetric(tagsValue []string, name, action string, value any) bool {
	if g.timeline == nil {
		g.timeline = map[uint64]*aggFields{}
	}

	tagsHash := hash.Fnv1aHash(tagsValue)

	agg, ok := g.timeline[tagsHash]
	if !ok {
		agg = &aggFields{
			tags:   tagsValue,
			fields: map[string]aggMetric{},
		}
		g.timeline[tagsHash] = agg
	}

	m, ok := agg.fields[name]
	if !ok {
		m, ok = NewAggMetric(name, action)
		if !ok {
			return false
		}
		agg.fields[name] = m
	}
	m.Append(value)

	return true
}

type bucket struct {
	bukName string

	interval   time.Duration
	keepValue  bool
	countLimit int
	curCount   int

	// tagsNameHash: tagsName
	by map[uint64][]string
	// tagsNameHash: pts
	group map[uint64]*ptsGroup

	extraTags map[string]string

	stop chan struct{}

	uploadFn UploadFunc

	sync.Mutex
}

func (buk *bucket) startScan() {
	if buk.stop != nil || buk.interval <= 0 {
		return
	}

	stop := make(chan struct{})
	buk.stop = stop

	go func() {
		ticker := time.NewTicker(buk.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				buk.Lock()
				pts := endAgg(buk)
				if len(pts) > 0 && buk.uploadFn != nil {
					_ = buk.uploadFn(buk.bukName, pts)
				}
				buk.Unlock()
			case <-stop:
				return
			}
		}
	}()
}

func (buk *bucket) stopScan() {
	buk.Lock()
	defer buk.Unlock()

	if buk.stop == nil {
		return
	}
	close(buk.stop)
	buk.stop = nil

	if buk.uploadFn != nil {
		pts := endAgg(buk)
		_ = buk.uploadFn(buk.bukName, pts)
	}
}

func (buk *bucket) setExtraTag(extra map[string]string) {
	buk.Lock()
	defer buk.Unlock()

	buk.extraTags = extra
}

func (buk *bucket) AddMetric(fieldName, action string, tagsName,
	tagsValue []string, aggField any,
) bool {
	tagNameHash := hash.Fnv1aHash(tagsName)

	buk.Lock()
	defer buk.Unlock()

	if buk.by == nil {
		buk.by = map[uint64][]string{}
	}

	if buk.group == nil {
		buk.group = map[uint64]*ptsGroup{}
	}

	if _, ok := buk.by[tagNameHash]; !ok {
		t := make([]string, len(tagsValue))
		copy(t, tagsName)
		buk.by[tagNameHash] = t
	}

	group, ok := buk.group[tagNameHash]
	if !ok {
		group = &ptsGroup{
			countLimit: buk.countLimit,
		}
		buk.group[tagNameHash] = group
	}

	if ok := group.addMetric(tagsValue, fieldName, action, aggField); ok {
		if buk.countLimit > 0 {
			buk.curCount++
			if buk.curCount >= buk.countLimit {
				if buk.uploadFn != nil {
					pts := endAgg(buk)
					_ = buk.uploadFn(buk.bukName, pts)
				}
				buk.curCount = 0
			}
		}
		return true
	}

	return false
}

func newBucket(name string, interval time.Duration, count int,
	keepValue bool, uploadFn UploadFunc,
) *bucket {
	return &bucket{
		bukName:    name,
		interval:   interval,
		keepValue:  keepValue,
		countLimit: count,
		by:         map[uint64][]string{},
		group:      map[uint64]*ptsGroup{},
		uploadFn:   uploadFn,
	}
}

func conv2Pt(name string, tagsName []string, aggTF *aggFields,
	extraTags map[string]string,
) (*point.Point, bool) {
	if len(tagsName) != len(aggTF.tags) {
		return nil, false
	}
	tags := map[string]string{}

	for k, v := range extraTags {
		tags[k] = v
	}

	for i := 0; i < len(tagsName); i++ {
		tags[tagsName[i]] = aggTF.tags[i]
	}

	fields := map[string]any{}
	for k, v := range aggTF.fields {
		if v != nil {
			fields[k] = v.Value()
		}
	}
	if pt, err := point.NewPoint(name, tags, fields, point.MOpt()); err != nil {
		return nil, false
	} else {
		return pt, true
	}
}

// 结束聚合.
func endAgg(b *bucket) []*point.Point {
	pts := []*point.Point{}

	for tagNameHash, group := range b.group {
		if group == nil {
			continue
		}
		tagsName, ok := b.by[tagNameHash]
		if !ok {
			continue
		}
		for _, tl := range group.timeline {
			if pt, ok := conv2Pt(b.bukName, tagsName, tl, b.extraTags); ok {
				pts = append(pts, pt)
			}
		}
	}

	if !b.keepValue {
		b.by = map[uint64][]string{}
		b.group = map[uint64]*ptsGroup{}
	}

	return pts
}

type aggMetric interface {
	Append(any)
	Value() any
}

func NewAggMetric(name, action string) (aggMetric, bool) {
	switch action {
	case "avg":
		return &avgMetric{}, true
	case "sum":
		return &sumMetric{}, true
	case "min":
		return &minMetric{}, true
	case "max":
		return &maxMetric{}, true
	case "set":
		return &setMetric{}, true
	default:
		return nil, false
	}
}

type avgMetric struct {
	sum   float64
	count float64
}

func (f *avgMetric) Append(v any) {
	f.sum += cast.ToFloat64(v)
	f.count++
}

func (f *avgMetric) Value() any {
	return f.sum / f.count
}

type sumMetric struct {
	sum float64
}

func (f *sumMetric) Append(v any) {
	f.sum += cast.ToFloat64(v)
}

func (f *sumMetric) Value() any {
	return f.sum
}

type minMetric struct {
	inserted bool
	min      float64
}

func (f *minMetric) Append(v any) {
	min := cast.ToFloat64(v)

	if f.inserted {
		if f.min > min {
			f.min = min
		}
	} else {
		f.min = min
		f.inserted = false
	}
}

func (f *minMetric) Value() any {
	return f.min
}

type maxMetric struct {
	max float64
}

func (f *maxMetric) Append(v any) {
	if max := cast.ToFloat64(v); f.max < max {
		f.max = max
	}
}

func (f *maxMetric) Value() any {
	return f.max
}

type setMetric struct {
	set float64
}

func (f *setMetric) Append(v any) {
	f.set = cast.ToFloat64(v)
}

func (f *setMetric) Value() any {
	return f.set
}
