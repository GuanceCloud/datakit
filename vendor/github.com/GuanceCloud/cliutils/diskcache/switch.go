// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"errors"
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
		return NewCacheError(OpPos, err, "failed_to_load_position_file").
			WithPath(c.path).WithFile(c.pos.fname)
	}

	if pos == nil {
		return nil
	}

	// check file's healty
	if _, err := os.Stat(string(pos.Name)); err != nil { // not exist
		if err := c.pos.reset(); err != nil {
			return NewCacheError(OpPos, err, "failed_to_reset_position_after_missing_file").
				WithPath(c.path).WithFile(c.pos.fname)
		}

		return nil
	}

	// invalid .pos, ignored
	if pos.Seek <= 0 && pos.Name == nil {
		return nil
	}

	fd, err := os.OpenFile(string(pos.Name), os.O_RDONLY, c.filePerms)
	if err != nil {
		return WrapFileOperationError(OpOpen, err, c.path, string(pos.Name)).
			WithDetails(fmt.Sprintf("failed_to_open_position_file: seek=%d", pos.Seek))
	}

	if _, err := fd.Seek(pos.Seek, io.SeekStart); err != nil {
		return WrapFileOperationError(OpSeek, err, c.path, string(pos.Name)).
			WithDetails(fmt.Sprintf("failed_to_seek_to_position: seek=%d", pos.Seek))
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

	// clear .pos: prepare for new .pos for next new file.
	if !c.noPos {
		if err := c.pos.reset(); err != nil {
			return NewCacheError(OpSwitch, err, "failed_to_reset_position_for_switch").
				WithPath(c.path)
		}
	}

	if len(c.dataFiles) == 0 {
		return nil
	} else {
		c.curReadfile = c.dataFiles[0]
	}

	fd, err := os.OpenFile(c.curReadfile, os.O_RDONLY, c.filePerms)
	if err != nil {
		return WrapFileOperationError(OpOpen, err, c.path, c.curReadfile).
			WithDetails(fmt.Sprintf("failed_to_open_next_read_file: available_files=%v", c.dataFiles))
	}

	c.rfd = fd

	if fi, err := c.rfd.Stat(); err != nil {
		return WrapFileOperationError(OpStat, err, c.path, c.curReadfile).
			WithDetails("failed_to_stat_read_file")
	} else {
		c.curReadSize = fi.Size()
	}

	if !c.noPos {
		c.pos.Name = []byte(c.curReadfile)
		c.pos.Seek = 0
		if err := c.pos.doDumpFile(); err != nil {
			return NewCacheError(OpSwitch, err, "failed_to_dump_position_after_switch").
				WithPath(c.path).WithFile(c.curReadfile)
		}

		posUpdatedVec.WithLabelValues("switch", c.path).Inc()
	}

	return nil
}

// open write file.
func (c *DiskCache) openWriteFile() error {
	if fi, err := os.Stat(c.curWriteFile); err == nil { // file exists
		if fi.IsDir() {
			return NewCacheError(OpCreate, errors.New("data file should not be dir"), "").
				WithPath(c.path).WithFile(c.curWriteFile)
		}

		c.curBatchSize = fi.Size()
	} else {
		// file not exists
		c.curBatchSize = 0
	}

	// write append fd, always write to the same-name file
	wfd, err := os.OpenFile(c.curWriteFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, c.filePerms)
	if err != nil {
		return WrapFileOperationError(OpCreate, err, c.path, c.curWriteFile).
			WithDetails("failed_to_open_write_file")
	}

	c.wfdLastWrite = time.Now()
	c.wfd = wfd
	return nil
}
