// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"

// IOConf configure io module in datakit.conf.
type IOConf struct {
	FeedChanSize int `toml:"feed_chan_size,omitzero"`

	MaxCacheCount                  int `toml:"max_cache_count"`
	MaxDynamicCacheCountDeprecated int `toml:"max_dynamic_cache_count,omitzero"`

	FlushInterval string `toml:"flush_interval"`
	FlushWorkers  int    `toml:"flush_workers"`

	EnableCache        bool   `toml:"enable_cache"`
	CacheAll           bool   `toml:"cache_all"`
	CacheSizeGB        int    `toml:"cache_max_size_gb"`
	CacheCleanInterval string `toml:"cache_clean_interval"`

	Filters map[string]filter.FilterConditions `toml:"filters"`
}
