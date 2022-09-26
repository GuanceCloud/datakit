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
