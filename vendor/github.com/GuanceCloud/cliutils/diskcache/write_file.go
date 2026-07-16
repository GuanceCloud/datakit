// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import "os"

func (c *DiskCache) writeFileName() string {
	if c.wfd != nil {
		return c.wfd.Name()
	}

	return c.curWriteFile
}

func (c *DiskCache) readFileName() string {
	if c.rfd != nil {
		return c.rfd.Name()
	}

	return c.curReadfile
}

func (c *DiskCache) ensureWriteFile() error {
	if c.closed {
		return ErrClosed
	}

	if c.wfd != nil {
		return nil
	}

	if c.curWriteFile == "" {
		return WrapFileOperationError(OpCreate, os.ErrInvalid, c.path, c.writeFileName()).
			WithDetails("write_file_path_not_set")
	}

	return c.openWriteFile()
}
