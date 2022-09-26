package diskcache

import (
	"fmt"
	"os"
	"time"
)

// open next read file.
func (c *DiskCache) switchNextFile() error {

	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if len(c.dataFiles) == 0 {
		return nil
		//c.curReadfile = c.curWriteFile
	} else {
		c.curReadfile = c.dataFiles[0]
	}

	l.Debugf("&&&&&&&&&&&&&&&&&&&&&&& read datafile: %s => %+#v",
		c.curReadfile, c.dataFiles)
	fd, err := os.OpenFile(c.curReadfile, os.O_RDONLY, c.opt.FilePerms)
	if err != nil {
		return fmt.Errorf("under switchNextFile, OpenFile: %w, datafile: %+#v, ", err, c.dataFiles)
	}

	c.rfd = fd
	return nil
}

// open write file.
func (c *DiskCache) openWriteFile() error {
	if fi, err := os.Stat(c.curWriteFile); err == nil { // file exists
		if fi.IsDir() {
			return fmt.Errorf("data file should not be dir")
		}

		c.curBatchSize = fi.Size()
	} else {
		// file not exists
		c.curBatchSize = 0
	}

	// write append fd, always write to the same-name file
	wfd, err := os.OpenFile(c.curWriteFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, c.opt.FilePerms)
	if err != nil {
		return fmt.Errorf("under openWriteFile, OpenFile: %w", err)
	}

	c.wfdCreated = time.Now()
	c.wfd = wfd
	return nil
}
