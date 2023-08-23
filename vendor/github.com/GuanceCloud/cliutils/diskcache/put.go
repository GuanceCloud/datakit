// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"encoding/binary"
	"time"
)

// Put write @data to disk cache, if reached batch size, a new batch is rotated.
// Put is safe to call concurrently with other operations and will
// block until all other operations finish.
func (c *DiskCache) Put(data []byte) error {
	start := time.Now() // count time before lock

	c.wlock.Lock()
	defer c.wlock.Unlock()

	defer func() {
		putVec.WithLabelValues(c.path).Inc()
		putBytesVec.WithLabelValues(c.path).Add(float64(len(data)))
		putLatencyVec.WithLabelValues(c.path).Observe(float64(time.Since(start) / time.Microsecond))
		sizeVec.WithLabelValues(c.path).Set(float64(c.size))
	}()

	if c.capacity > 0 && c.size+int64(len(data)) > c.capacity {
		if err := c.dropBatch(); err != nil {
			return err
		}
	}

	if c.maxDataSize > 0 && int32(len(data)) > c.maxDataSize {
		return ErrTooLargeData
	}

	hdr := make([]byte, dataHeaderLen)

	binary.LittleEndian.PutUint32(hdr, uint32(len(data)))
	if _, err := c.wfd.Write(hdr); err != nil {
		return err
	}

	if _, err := c.wfd.Write(data); err != nil {
		return err
	}

	if !c.noSync {
		if err := c.wfd.Sync(); err != nil {
			return err
		}
	}

	c.curBatchSize += int64(len(data) + dataHeaderLen)
	c.size += int64(len(data) + dataHeaderLen)
	c.wfdLastWrite = time.Now()

	// rotate new file
	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return err
		}
	}

	return nil
}
