// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Rotate force diskcache switch(rotate) from current write file(cwf) to next
// new file, leave cwf become readble on successive Get().
//
// NOTE: You do not need to call Rotate() during daily usage, we export
// that function for testing cases.
func (c *DiskCache) Rotate() error {
	return c.rotate()
}

// rotate to next new file, append to reading list.
func (c *DiskCache) rotate() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	defer func() {
		rotateVec.WithLabelValues(c.labels...).Inc()
		sizeVec.WithLabelValues(c.labels...).Set(float64(c.size))
		datafilesVec.WithLabelValues(c.labels...).Set(float64(len(c.dataFiles)))
	}()

	eof := make([]byte, dataHeaderLen)
	binary.LittleEndian.PutUint32(eof, EOFHint)
	if _, err := c.wfd.Write(eof); err != nil { // append EOF to file end
		return err
	}

	// NOTE: EOF bytes do not count to size

	// rotate file
	var newfile string
	if len(c.dataFiles) == 0 {
		newfile = filepath.Join(c.path, fmt.Sprintf("data.%032d", 0)) // first rotate file
	} else {
		// parse last file's name, such as `data.000003', the new rotate file is `data.000004`
		last := c.dataFiles[len(c.dataFiles)-1]
		arr := strings.Split(filepath.Base(last), ".")
		if len(arr) != 2 {
			return ErrInvalidDataFileName
		}
		x, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return ErrInvalidDataFileNameSuffix
		}

		// data.0003 -> data.0004
		newfile = filepath.Join(c.path, fmt.Sprintf("data.%032d", x+1))
	}

	// close current writing file
	if err := c.wfd.Close(); err != nil {
		return err
	}
	c.wfd = nil

	// rename data -> data.0004
	if err := os.Rename(c.curWriteFile, newfile); err != nil {
		return err
	}

	c.dataFiles = append(c.dataFiles, newfile)
	sort.Strings(c.dataFiles)

	// reopen new write file
	if err := c.openWriteFile(); err != nil {
		return err
	}

	return nil
}

// after file read on EOF, remove the file.
func (c *DiskCache) removeCurrentReadingFile() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	defer func() {
		sizeVec.WithLabelValues(c.labels...).Set(float64(c.size))
		removeVec.WithLabelValues(c.labels...).Inc()
		datafilesVec.WithLabelValues(c.labels...).Set(float64(len(c.dataFiles)))
	}()

	if c.rfd != nil {
		if err := c.rfd.Close(); err != nil {
			return err
		}
		c.rfd = nil
	}

	if fi, err := os.Stat(c.curReadfile); err == nil {
		c.size -= (fi.Size() - dataHeaderLen) // EOF bytes do not counted in size
	}

	if err := os.Remove(c.curReadfile); err != nil {
		return fmt.Errorf("removeCurrentReadingFile: %s: %w", c.curReadfile, err)
	}

	if len(c.dataFiles) > 0 {
		c.dataFiles = c.dataFiles[1:] // first file removed
	}

	return nil
}
