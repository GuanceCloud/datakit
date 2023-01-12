// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"os"
	"strconv"
)

func (opt *Option) syncEnv() {
	if v, ok := os.LookupEnv("ENV_DISKCACHE_BATCH_SIZE"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			opt.BatchSize = i
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_MAX_DATA_SIZE"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			opt.MaxDataSize = i
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_CAPACITY"); ok && v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			opt.Capacity = i
		}
	}

	if v, ok := os.LookupEnv("ENV_DISKCACHE_NO_SYNC"); ok && v != "" {
		opt.NoSync = true
	}
}
