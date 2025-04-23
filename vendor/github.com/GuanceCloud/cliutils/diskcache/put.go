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
	"os"
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
		putBytesVec.WithLabelValues(c.path).Observe(float64(len(data)))
		putLatencyVec.WithLabelValues(c.path).Observe(float64(time.Since(start)) / float64(time.Second))
		sizeVec.WithLabelValues(c.path).Set(float64(c.size.Load()))
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
	c.size.Add(int64(len(data) + dataHeaderLen))
	c.wfdLastWrite = time.Now()

	// rotate new file
	if c.curBatchSize >= c.batchSize {
		if err := c.rotate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *DiskCache) putPart(part []byte) error {
	if _, err := c.wfd.Write(part); err != nil {
		return err
	}

	if !c.noSync {
		if err := c.wfd.Sync(); err != nil {
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
		n           = 0
		total       = 0
		err         error
		startOffset int64
		start       = time.Now()
		round       = 0
	)

	c.wlock.Lock()
	defer c.wlock.Unlock()

	if c.capacity > 0 && c.size.Load()+int64(size) > c.capacity {
		return ErrCacheFull
	}

	if startOffset, err = c.wfd.Seek(0, os.SEEK_CUR); err != nil {
		return fmt.Errorf("Seek(0, SEEK_CUR): %w", err)
	}

	defer func() {
		if total > 0 && err != nil { // fallback to origin position
			if _, serr := c.wfd.Seek(startOffset, io.SeekStart); serr != nil {
				c.LastErr = serr
			}
		}

		putBytesVec.WithLabelValues(c.path).Observe(float64(size))
		putLatencyVec.WithLabelValues(c.path).Observe(float64(time.Since(start)) / float64(time.Second))
		sizeVec.WithLabelValues(c.path).Set(float64(c.size.Load()))
		streamPutVec.WithLabelValues(c.path).Observe(float64(round))
	}()

	binary.LittleEndian.PutUint32(c.batchHeader, uint32(size))
	if _, err := c.wfd.Write(c.batchHeader); err != nil {
		return err
	}

	for {
		n, err = r.Read(c.streamBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return err
			}
		}

		if n == 0 {
			break
		}

		if err = c.putPart(c.streamBuf); err != nil {
			return err
		} else {
			total += n
			round++
		}
	}

	return nil
}
