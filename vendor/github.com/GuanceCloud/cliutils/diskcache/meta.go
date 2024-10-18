// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

// Size return current size of the cache.
func (c *DiskCache) Size() int64 {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if len(c.dataFiles) > 0 { // there are files waiting to be read
		return c.size.Load()
	} else {
		return 0
	}
}

// RawSize return current size plus current writing file(`data') of the cache.
func (c *DiskCache) RawSize() int64 {
	return c.size.Load()
}

// Capacity return max capacity of the cache.
func (c *DiskCache) Capacity() int64 {
	return c.capacity
}

// MaxDataSize return max single data piece size of the cache.
func (c *DiskCache) MaxDataSize() int32 {
	return c.maxDataSize
}

// MaxBatchSize return max single data file size of the cache.
//
// With proper data file size(default is 20MB), we can make the switch/rotate
// and garbage collection more quickly when all piece of data wthin the data
// file has been Get() out of the file.
func (c *DiskCache) MaxBatchSize() int64 {
	return c.batchSize
}

// Path return dir of current diskcache.
func (c *DiskCache) Path() string {
	return c.path
}
