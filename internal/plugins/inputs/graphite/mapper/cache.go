// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package mapper contains graphite mapper cache type.
package mapper

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/cache/lru"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/cache/rr"
)

type CacheType string

const (
	CacheLRU         CacheType = "lru"
	CacheRR          CacheType = "rr"
	CacheUnknown     CacheType = "unknown"
	DefaultCacheSize           = 1000
)

func getCache(cacheSize int, cacheType CacheType) (MetricMapperCache, error) {
	var cache MetricMapperCache
	var err error
	if cacheSize == 0 {
		return nil, nil
	} else {
		//nolint:exhaustive
		switch cacheType {
		case CacheLRU:
			cache, err = lru.NewMetricMapperLRUCache(cacheSize)
		case CacheRR:
			cache, err = rr.NewMetricMapperRRCache(cacheSize)
		default:
			cache, err = lru.NewMetricMapperLRUCache(cacheSize)
		}

		if err != nil {
			return nil, err
		}
	}
	return cache, nil
}
