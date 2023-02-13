// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

func (c *DiskCache) Size() int64 {
	return c.size
}

func (c *DiskCache) DroppedBatch() int {
	return c.droppedBatch
}

func (c *DiskCache) RotateCount() int {
	return c.rotateCount
}

func (c *DiskCache) FileCount() int {
	return len(c.dataFiles)
}
