// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"os"
	"strconv"
	"time"
)

func (c *DiskCache) syncEnv() {
	if v, ok := os.LookupEnv("ENV_DISKCACHE_BATCH_SIZE"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.batchSize = i
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_MAX_DATA_SIZE"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			c.maxDataSize = int32(i)
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_CAPACITY"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.capacity = i
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_NO_SYNC"); ok && v != "" {
		c.noSync = true
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_NO_POS"); ok && v != "" {
		c.noPos = true
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_POS_DUMP_INTERVAL"); ok && v != "" {
		if du, err := time.ParseDuration(v); err == nil && du > 0 {
			c.pos.dumpInterval = du
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_POS_DUMP_AT"); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			c.pos.dumpCount = int(n)
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_NO_LOCK"); ok && v != "" {
		c.noLock = true
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_NO_FALLBACK_ON_ERROR"); ok && v != "" {
		c.noFallbackOnError = true
	}
}
