// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package lru contains graphite lru cache
package lru

import (
	"sync"

	"github.com/golang/groupcache/lru"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/cache"
)

type metricMapperLRUCache struct {
	cache   *lruCache
	metrics *cache.CacheMetrics
}

func NewMetricMapperLRUCache(size int) (*metricMapperLRUCache, error) {
	if size <= 0 {
		return nil, nil
	}

	metrics := cache.NewCacheMetrics()
	cache := newLruCache(size)

	return &metricMapperLRUCache{cache: cache, metrics: metrics}, nil
}

func (m *metricMapperLRUCache) Get(metricKey string) (interface{}, bool) {
	m.metrics.CacheGetsTotal.Inc()
	if result, ok := m.cache.Get(metricKey); ok {
		m.metrics.CacheHitsTotal.Inc()
		return result, true
	} else {
		return nil, false
	}
}

func (m *metricMapperLRUCache) Add(metricKey string, result interface{}) {
	go m.trackCacheLength()
	m.cache.Add(metricKey, result)
}

func (m *metricMapperLRUCache) trackCacheLength() {
	m.metrics.CacheLength.Set(float64(m.cache.Len()))
}

func (m *metricMapperLRUCache) Reset() {
	m.cache.Clear()
	m.metrics.CacheLength.Set(0)
}

type lruCache struct {
	cache *lru.Cache
	lock  sync.RWMutex
}

func newLruCache(maxEntries int) *lruCache {
	return &lruCache{
		cache: lru.New(maxEntries),
	}
}

func (l *lruCache) Get(key string) (interface{}, bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.cache.Get(key)
}

func (l *lruCache) Add(key string, value interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.cache.Add(key, value)
}

func (l *lruCache) Len() int {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.cache.Len()
}

func (l *lruCache) Clear() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.cache.Clear()
}
