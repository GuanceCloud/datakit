// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import "os"

func (c *DiskCache) dropBatch() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if len(c.dataFiles) == 0 {
		return nil
	}

	fname := c.dataFiles[0]

	if c.rfd != nil && c.curReadfile == fname {
		if err := c.rfd.Close(); err != nil {
			return err
		}

		c.rfd = nil
	}

	if fi, err := os.Stat(fname); err == nil {
		if err := os.Remove(fname); err != nil {
			return err
		}

		c.size -= fi.Size()

		c.dataFiles = c.dataFiles[1:]

		droppedBatchVec.WithLabelValues(c.labels...).Inc()
		droppedBytesVec.WithLabelValues(c.labels...).Add(float64(fi.Size()))
		datafilesVec.WithLabelValues(c.labels...).Set(float64(len(c.dataFiles)))
		sizeVec.WithLabelValues(c.labels...).Set(float64(c.size))
	}

	return nil
}
