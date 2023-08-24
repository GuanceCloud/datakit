// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"os"
	"path/filepath"
	"time"
)

// A CacheOption used to set various options on DiskCache.
type CacheOption func(c *DiskCache)

// WithNoFallbackOnError disable fallback on fn() error.
//
// During Get(fn(data []btye)error{...}), if fn() failed with error,
// the next Get still get the same data from cache.
// If fallback disabled, the next read will read new data from cache,
// and the previous failed data skipped(and eventually dropped).
func WithNoFallbackOnError(on bool) CacheOption {
	return func(c *DiskCache) {
		c.noFallbackOnError = on
	}
}

// WithNoLock set .lock on or off.
//
// File '.lock' used to exclude Open() on same path.
func WithNoLock(on bool) CacheOption {
	return func(c *DiskCache) {
		c.noLock = on
	}
}

// WithNoPos set .pos on or off.
//
// The file '.pos' used to remember last Get() position, without '.pos',
// on process restart, some already-Get() data will Get() again in the
// new process, this maybe not the right action we expect.
func WithNoPos(on bool) CacheOption {
	return func(c *DiskCache) {
		c.noPos = on
	}
}

// WithWakeup set duration on wakeup(default 3s), this wakeup time
// used to shift current-writing-file to ready-to-reading-file.
//
// NOTE: without wakeup, current-writing-file maybe not read-available
// for a long time.
func WithWakeup(wakeup time.Duration) CacheOption {
	return func(c *DiskCache) {
		if int64(wakeup) > 0 {
			c.wakeup = wakeup
		}
	}
}

// WithBatchSize set file size, default 64MB.
func WithBatchSize(size int64) CacheOption {
	return func(c *DiskCache) {
		if size > 0 {
			c.batchSize = size
		}
	}
}

// WithMaxDataSize set max single data size, default 32MB.
func WithMaxDataSize(size int32) CacheOption {
	return func(c *DiskCache) {
		if size > 0 {
			c.maxDataSize = size
		}
	}
}

// WithCapacity set cache capacity, default unlimited.
func WithCapacity(size int64) CacheOption {
	return func(c *DiskCache) {
		if size > 0 {
			c.capacity = size
		}
	}
}

// WithExtraCapacity add capacity to existing cache.
func WithExtraCapacity(size int64) CacheOption {
	return func(c *DiskCache) {
		if c.capacity+size > 0 {
			c.capacity += size
			if c.path != "" {
				capVec.WithLabelValues(c.path).Set(float64(c.capacity))
			}
		}
	}
}

// WithNoSync enable/disable sync on cache write.
//
// NOTE: Without sync, the write performance 60~80 times faster for 512KB/1MB put,
// for smaller put will get more faster(1kb for 4000+ times).
func WithNoSync(on bool) CacheOption {
	return func(c *DiskCache) {
		c.noSync = on
	}
}

// WithDirPermission set disk dir permission mode.
func WithDirPermission(perms os.FileMode) CacheOption {
	return func(c *DiskCache) {
		c.dirPerms = perms
	}
}

// WithFilePermission set cache file permission mode.
func WithFilePermission(perms os.FileMode) CacheOption {
	return func(c *DiskCache) {
		c.filePerms = perms
	}
}

// WithPath set disk dirname.
func WithPath(x string) CacheOption {
	return func(c *DiskCache) {
		c.path = filepath.Clean(x)
	}
}
