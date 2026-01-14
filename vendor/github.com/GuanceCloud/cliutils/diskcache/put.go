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
			return WrapPutError(ErrCacheFull, c.path, len(data)).WithDetails("no_drop_enabled")
		}

		if c.filoDrop { // do not accept new data
			droppedDataVec.WithLabelValues(c.path, reasonExceedCapacity).Observe(float64(len(data)))
			return WrapPutError(ErrCacheFull, c.path, len(data)).WithDetails("filo_drop_policy")
		}

		if err := c.dropBatch(); err != nil {
			return WrapPutError(err, c.path, len(data)).WithDetails("failed_to_drop_batch")
		}
	}

	if c.maxDataSize > 0 && int32(len(data)) > c.maxDataSize {
		return WrapPutError(ErrTooLargeData, c.path, len(data)).WithDetails(
			fmt.Sprintf("max_size=%d, actual_size=%d", c.maxDataSize, len(data)))
	}

	hdr := make([]byte, dataHeaderLen)

	binary.LittleEndian.PutUint32(hdr, uint32(len(data)))
	if _, err := c.wfd.Write(hdr); err != nil {
		return WrapFileOperationError(OpWrite, err, c.path, c.wfd.Name()).
			WithDetails("failed_to_write_header")
	}

	if _, err := c.wfd.Write(data); err != nil {
		return WrapFileOperationError(OpWrite, err, c.path, c.wfd.Name()).
			WithDetails("failed_to_write_data")
	}

	if !c.noSync {
		if err := c.wfd.Sync(); err != nil {
			return WrapFileOperationError(OpSync, err, c.path, c.wfd.Name()).
				WithDetails("failed_to_sync_write")
		}
	}

	c.curBatchSize += int64(len(data) + dataHeaderLen)
	c.wfdLastWrite = time.Now()

	// rotate new file
	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return WrapPutError(err, c.path, len(data)).WithDetails("failed_to_rotate_batch")
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
		return NewCacheError(OpStreamPut, ErrInvalidStreamSize,
			fmt.Sprintf("invalid_size=%d", size)).WithPath(c.path)
	}

	c.wlock.Lock()
	defer c.wlock.Unlock()

	if c.capacity > 0 && c.size.Load()+int64(size) > c.capacity {
		return NewCacheError(OpStreamPut, ErrCacheFull,
			fmt.Sprintf("capacity_exceeded: current=%d, new=%d, max=%d",
				c.size.Load(), size, c.capacity)).WithPath(c.path)
	}

	if c.maxDataSize > 0 && size > int(c.maxDataSize) {
		return NewCacheError(OpStreamPut, ErrTooLargeData,
			fmt.Sprintf("size_exceeded: max=%d, actual=%d", c.maxDataSize, size)).WithPath(c.path)
	}

	if startOffset, err = c.wfd.Seek(0, io.SeekCurrent); err != nil {
		return WrapFileOperationError(OpSeek, err, c.path, c.wfd.Name()).
			WithDetails("failed_to_get_current_position")
	}

	defer func() {
		if total > 0 && err != nil { // fallback to origin position
			if _, serr := c.wfd.Seek(startOffset, io.SeekStart); serr != nil {
				c.LastErr = WrapFileOperationError(OpSeek, serr, c.path, c.wfd.Name()).
					WithDetails(fmt.Sprintf("failed_to_fallback_to_position_%d", startOffset))
			}
		}

		putLatencyVec.WithLabelValues(c.path).Observe(time.Since(start).Seconds())
	}()

	if size > 0 {
		binary.LittleEndian.PutUint32(c.batchHeader, uint32(size))
		if _, err := c.wfd.Write(c.batchHeader); err != nil {
			return WrapFileOperationError(OpWrite, err, c.path, c.wfd.Name()).
				WithDetails("failed_to_write_stream_header")
		}
	}

	total, err = io.CopyN(c.wfd, r, int64(size))
	if err != nil && !errors.Is(err, io.EOF) {
		return NewCacheError(OpStreamPut, err,
			fmt.Sprintf("failed_to_copy_stream_data: expected=%d, copied=%d", size, total)).
			WithPath(c.path)
	}

	c.curBatchSize += (total + dataHeaderLen)

	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return NewCacheError(OpStreamPut, err, "failed_to_rotate_after_stream_put").
				WithPath(c.path)
		}
	}

	return nil
}
