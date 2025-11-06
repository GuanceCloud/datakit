// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// IsFull test if reach max capacity limit after put newData into cache.
func (c *DiskCache) IsFull(newData []byte) bool {
	return c.capacity > 0 && c.size.Load()+int64(len(newData)) > c.capacity
}

// Put write @data to disk cache, if reached batch size, a new batch is rotated.
// Put is safe to call concurrently with other operations and will
// block until all other operations finish.
func (c *DiskCache) Put(data []byte) error {
	start := time.Now() // count time before lock

	c.wlock.Lock()
	defer c.wlock.Unlock()

	defer func() {
		putLatencyVec.WithLabelValues(c.path).Observe(time.Since(start).Seconds())
	}()

	if c.IsFull(data) {
		if c.noDrop {
			return ErrCacheFull
		}

		if c.filoDrop { // do not accept new data
			droppedDataVec.WithLabelValues(c.path, reasonExceedCapacity).Observe(float64(len(data)))
			return ErrCacheFull
		}

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
	c.wfdLastWrite = time.Now()

	// rotate new file
	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return err
		}
	}

	return nil
}

// StreamPut read from r for bytes and write to storage.
//
// If we read the data from some network stream(such as HTTP response body),
// we can use StreamPut to avoid a intermediate buffer to accept the huge(may be) body.
func (c *DiskCache) StreamPut(r io.Reader, size int) error {
	var (
		//nolint:ineffassign
		total       int64
		err         error
		startOffset int64
		start       = time.Now()
	)

	if size <= 0 {
		return ErrInvalidStreamSize
	}

	c.wlock.Lock()
	defer c.wlock.Unlock()

	if c.capacity > 0 && c.size.Load()+int64(size) > c.capacity {
		return ErrCacheFull
	}

	if c.maxDataSize > 0 && size > int(c.maxDataSize) {
		return ErrTooLargeData
	}

	if startOffset, err = c.wfd.Seek(0, io.SeekCurrent); err != nil {
		return fmt.Errorf("Seek(0, SEEK_CUR): %w", err)
	}

	defer func() {
		if total > 0 && err != nil { // fallback to origin position
			if _, serr := c.wfd.Seek(startOffset, io.SeekStart); serr != nil {
				c.LastErr = serr
			}
		}

		putLatencyVec.WithLabelValues(c.path).Observe(time.Since(start).Seconds())
	}()

	if size > 0 {
		binary.LittleEndian.PutUint32(c.batchHeader, uint32(size))
		if _, err := c.wfd.Write(c.batchHeader); err != nil {
			return err
		}
	}

	total, err = io.CopyN(c.wfd, r, int64(size))
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	c.curBatchSize += (total + dataHeaderLen)

	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return err
		}
	}

	return nil
}
