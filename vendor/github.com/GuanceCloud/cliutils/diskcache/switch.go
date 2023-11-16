// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"fmt"
	"io"
	"os"
	"time"
)

// switch to next file remembered in .pos file.
func (c *DiskCache) loadUnfinishedFile() error {
	if _, err := os.Stat(c.pos.fname); err != nil {
		return nil // .pos file not exist
	}

	pos, err := posFromFile(c.pos.fname)
	if err != nil {
		return fmt.Errorf("posFromFile: %w", err)
	}

	// check file's healty
	if _, err := os.Stat(string(pos.Name)); err != nil { // not exist
		if err := c.pos.reset(); err != nil {
			return err
		}

		return nil
	}

	// invalid .pos, ignored
	if pos.Seek <= 0 {
		return nil
	}

	fd, err := os.OpenFile(string(pos.Name), os.O_RDONLY, c.filePerms)
	if err != nil {
		return fmt.Errorf("OpenFile: %w", err)
	}

	if _, err := fd.Seek(pos.Seek, io.SeekStart); err != nil {
		return fmt.Errorf("Seek(%q: %d, 0): %w", pos.Name, pos.Seek, err)
	}

	c.rfd = fd
	c.curReadfile = string(pos.Name)
	c.pos.Name = pos.Name
	c.pos.Seek = pos.Seek

	return nil
}

// open next read file.
func (c *DiskCache) doSwitchNextFile() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if len(c.dataFiles) == 0 {
		return nil
	} else {
		c.curReadfile = c.dataFiles[0]
	}

	fd, err := os.OpenFile(c.curReadfile, os.O_RDONLY, c.filePerms)
	if err != nil {
		return fmt.Errorf("under switchNextFile, OpenFile: %w, datafile: %+#v, ", err, c.dataFiles)
	}

	c.rfd = fd

	if !c.noPos {
		c.pos.Name = []byte(c.curReadfile)
		c.pos.Seek = 0
		if err := c.pos.dumpFile(); err != nil {
			return err
		}
	}

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
	wfd, err := os.OpenFile(c.curWriteFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, c.filePerms)
	if err != nil {
		return fmt.Errorf("under openWriteFile, OpenFile: %w", err)
	}

	c.wfdLastWrite = time.Now()
	c.wfd = wfd
	return nil
}
