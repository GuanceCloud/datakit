// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
)

// IOConf configure io module in datakit.conf.
type IOConf struct {
	FeedChanSize             int  `toml:"feed_chan_size"`
	GlobalBlockingDeprecated bool `toml:"global_blocking"`

	MaxCacheCount                  int `toml:"max_cache_count"`
	MaxDynamicCacheCountDeprecated int `toml:"max_dynamic_cache_count,omitzero"`

	CompactInterval time.Duration `toml:"flush_interval"`
	CompactWorkers  int           `toml:"flush_workers"`

	Filters map[string]filter.FilterConditions `toml:"filters"`
}
