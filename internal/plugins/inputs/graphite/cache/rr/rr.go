// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rr contains graphite rr cache
package rr

import (
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/cache"
)

type metricMapperRRCache struct {
	lock    sync.RWMutex
	size    int
	items   map[string]interface{}
	metrics *cache.CacheMetrics
}

func NewMetricMapperRRCache(size int) (*metricMapperRRCache, error) {
	if size <= 0 {
		return nil, nil
	}

	metrics := cache.NewCacheMetrics()
	c := &metricMapperRRCache{
		items:   make(map[string]interface{}, size+1),
		size:    size,
		metrics: metrics,
	}
	return c, nil
}

func (m *metricMapperRRCache) Get(metricKey string) (interface{}, bool) {
	m.lock.RLock()
	result, ok := m.items[metricKey]
	m.lock.RUnlock()

	return result, ok
}

func (m *metricMapperRRCache) Add(metricKey string, result interface{}) {
	go m.trackCacheLength()
	m.lock.Lock()

	m.items[metricKey] = result

	// evict an item if needed
	if len(m.items) > m.size {
		for k := range m.items {
			delete(m.items, k)
			break
		}
	}

	m.lock.Unlock()
}

func (m *metricMapperRRCache) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.items = make(map[string]interface{}, m.size+1)
	m.metrics.CacheLength.Set(0)
}

func (m *metricMapperRRCache) trackCacheLength() {
	m.lock.RLock()
	length := len(m.items)
	m.lock.RUnlock()
	m.metrics.CacheLength.Set(float64(length))
}
