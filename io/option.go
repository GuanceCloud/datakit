// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"os"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils/diskcache"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
)

// IOOption used to add various options to setup io module.
type IOOption func(x *dkIO)

// WithDataway used to setup where data write to(dataway).
func WithDataway(dw dataway.IDataway) IOOption {
	return func(x *dkIO) {
		x.dw = dw
	}
}

// WithFilters used to setup point filter.
func WithFilters(filters map[string]filter.FilterConditions) IOOption {
	return func(x *dkIO) {
		x.filters = filters
	}
}

// WithDiskCacheCleanInterval used to control clean(retry-on-failed-data)
// interval of disk cache.
func WithDiskCacheCleanInterval(du time.Duration) IOOption {
	return func(x *dkIO) {
		if int64(du) > 0 {
			x.cacheCleanInterval = du
		}
	}
}

// WithDiskCacheSize used to set max disk cache(in GB bytes).
func WithDiskCacheSize(gb int) IOOption {
	return func(x *dkIO) {
		if gb > 0 {
			x.cacheSizeGB = gb
		}
	}
}

// WithCacheAll will cache all categories.
// By default, metric(M), object(CO/O) and dial-testing data point not cached.
func WithCacheAll(on bool) IOOption {
	return func(x *dkIO) {
		x.cacheAll = on
	}
}

// WithDiskCache used to set/unset disk cache on failed data.
func WithDiskCache(on bool) IOOption {
	return func(x *dkIO) {
		x.enableCache = on

		// TODO: init disk cache for each categories
		for _, c := range point.AllCategories() {
			p := filepath.Join(datakit.CacheDir, c.String())
			capacity := int64(x.cacheSizeGB * 1024 * 1024 * 1024)

			cache, err := diskcache.Open(
				diskcache.WithPath(p),
				diskcache.WithCapacity(capacity),
				diskcache.WithWakeup(30*time.Second), // to disable generate too many files under cache
			)
			if err != nil {
				log.Warnf("NewWALCache to %s with capacity %d: %s", p, capacity, err.Error())
				continue
			} else {
				x.fcs[c.URL()] = cache
			}
		}
	}
}

// WithOutputFile used to set a local file, the points will write
// to the file(in the form line-protocol).
func WithOutputFile(fpath string) IOOption {
	return func(x *dkIO) {
		if fpath == "" {
			return
		}

		x.outputFile = fpath

		// if file open failed, ignored.
		f, err := os.OpenFile(filepath.Clean(fpath), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
		if err != nil {
			log.Warnf("OpenFile: %s, ignored", err)
		}
		x.fd = f
	}
}

// WithOutputFileOnInputs set inputs that their point write to file.
func WithOutputFileOnInputs(inputs []string) IOOption {
	return func(x *dkIO) {
		x.outputFileInputs = inputs
	}
}

// WithFlushWorkers set IO flush workers.
func WithFlushWorkers(n int) IOOption {
	return func(x *dkIO) {
		if n > 0 {
			x.flushWorkers = n
		}
	}
}

// WithFlushInterval used to contol when to flush cached data.
func WithFlushInterval(d time.Duration) IOOption {
	return func(x *dkIO) {
		if int64(d) > 0 {
			x.flushInterval = d
		}
	}
}

// WithFeedCap used to set capacity of feed chan.
func WithFeedCap(size int) IOOption {
	return func(x *dkIO) {
		if size > 0 {
			x.feedChanSize = size
		}
	}
}

// WithMaxCacheCount used to set max cache size.
// The count used to control when to send the cached data.
func WithMaxCacheCount(count int) IOOption {
	return func(x *dkIO) {
		if count > 0 {
			x.maxCacheCount = count
		}
	}
}
