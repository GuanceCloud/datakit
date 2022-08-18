package diskcache

import "os"

func (c *DiskCache) dropBatch() error {

	if len(c.dataFiles) > 0 {
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

			c.rwlock.Lock()
			defer c.rwlock.Unlock()

			c.dataFiles = c.dataFiles[1:]

			l.Debugf("----------------------- drop datafile(%dth): %s(%d) => %+#v, size: %d\n",
				c.droppedBatch, fname, fi.Size(), c.dataFiles, c.size)
			c.droppedBatch++
		}
	}

	return nil
}
