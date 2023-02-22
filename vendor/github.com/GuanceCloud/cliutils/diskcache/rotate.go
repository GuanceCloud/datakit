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

// rotate to next new file, append to reading list.
func (c *DiskCache) rotate() error {
	l.Debugf("try rotate...")

	eof := make([]byte, dataHeaderLen)
	binary.LittleEndian.PutUint32(eof, eofHint)

	if _, err := c.wfd.Write(eof); err != nil {
		return err
	}

	c.size += dataHeaderLen // eof hint also count to size

	// rotate file
	var newfile string
	if len(c.dataFiles) == 0 {
		newfile = filepath.Join(c.path, fmt.Sprintf("data.%032d", 0)) // first rotate file
	} else {
		// parse last file's name, i.e., `data.000003', the new rotate file is `data.000004`
		last := c.dataFiles[len(c.dataFiles)-1]
		arr := strings.Split(filepath.Base(last), ".")
		if len(arr) != 2 {
			return ErrInvalidDataFileName
		}
		x, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return ErrInvalidDataFileNameSuffix
		}

		newfile = filepath.Join(c.path, fmt.Sprintf("data.%032d", x+1))
	}

	// close current writing file
	if err := c.wfd.Close(); err != nil {
		return err
	}

	if err := os.Rename(c.curWriteFile, newfile); err != nil {
		return err
	}

	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	c.dataFiles = append(c.dataFiles, newfile)
	sort.Strings(c.dataFiles)

	l.Debugf("+++++++++++++++++++++++ add datafile: %s => %s | %+#v",
		c.curWriteFile, newfile, c.dataFiles)

	// reopen new write file
	if err := c.openWriteFile(); err != nil {
		return err
	}
	c.rotateCount++
	return nil
}

// after file read on EOF, remove the file.
func (c *DiskCache) removeCurrentReadingFile() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if c.rfd != nil {
		if err := c.rfd.Close(); err != nil {
			return err
		}
		c.rfd = nil
	}

	if fi, err := os.Stat(c.curReadfile); err == nil {
		c.size -= fi.Size()
	}

	if err := os.Remove(c.curReadfile); err != nil {
		return fmt.Errorf("removeCurrentReadingFile: %s: %w", c.curReadfile, err)
	}

	if len(c.dataFiles) > 0 {
		c.dataFiles = c.dataFiles[1:] // first file removed
		l.Debugf("----------------------- remove datafile: %s => %+#v",
			c.curReadfile, c.dataFiles)
	}

	return nil
}
