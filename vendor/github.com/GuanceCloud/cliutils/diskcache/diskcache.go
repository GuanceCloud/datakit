// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diskcache is a simple local-disk cache implements.
//
// The diskcache package is a local-disk cache, it implements following functions:
//
//  1. Concurrent Put()/Get().
//  2. Recoverable last-read-position on restart.
//  3. Exclusive Open() on same path.
//  4. Errors during Get() are retriable.
//  5. Auto-rotate on batch size.
//  6. Drop in FIFO policy when max capacity reached.
//  7. We can configure various specifics in environments without to modify options source code.
package diskcache

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	dataHeaderLen = 4

	// EOFHint labels a file's end.
	EOFHint = uint32(0xdeadbeef)
)

// Generic diskcache errors.
var (
	// Invalid read size.
	ErrUnexpectedReadSize = errors.New("unexpected read size")

	// Data send to Put() exceed the maxDataSize.
	ErrTooLargeData = errors.New("too large data")

	// Get on no data cache.
	ErrEOF = errors.New("EOF")

	// Invalid cache filename.
	ErrInvalidDataFileName       = errors.New("invalid datafile name")
	ErrInvalidDataFileNameSuffix = errors.New("invalid datafile name suffix")

	// Invalid file header.
	ErrBadHeader = errors.New("bad header")
)

// DiskCache is the representation of a disk cache.
// A DiskCache is safe for concurrent use by multiple goroutines.
// Do not Open the same-path diskcache among goroutines.
type DiskCache struct {
	path string

	dataFiles []string

	// current writing/reading file.
	curWriteFile,
	curReadfile string

	// current write/read fd
	wfd, rfd *os.File

	// If current write file go nothing put for a
	// long time(wakeup), we rotate it manually.
	wfdLastWrite time.Time

	// how long to wakeup a sleeping write-file
	wakeup time.Duration

	wlock, // used to exclude concurrent Put.
	rlock *sync.Mutex // used to exclude concurrent Get.
	rwlock *sync.Mutex // used to exclude switch/rotate/drop/Close

	flock *flock // disabled multi-Open on same path
	pos   *pos   // current read fd position info

	// specs of current diskcache
	size, // current byte size
	curBatchSize, // current writing file's size
	batchSize, // current batch size(static)
	capacity int64 // capacity of the diskcache
	maxDataSize int32 // max data size of single Put()

	// File permission, default 0750/0640
	dirPerms,
	filePerms os.FileMode

	// various flags
	noSync, // NoSync if enabled, may cause data missing, default false
	noFallbackOnError, // ignore Fn() error
	noPos, // no position
	noLock bool // no file lock

	// labels used to export prometheus flags
	labels []string
}

func (c *DiskCache) String() string {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	// if there too many files(>10), only print file count
	if n := len(c.dataFiles); n > 10 {
		return fmt.Sprintf("%s/[size: %d][fallback: %v][nosync: %v][nopos: %v][nolock: %v][files: %d][maxDataSize: %d][batchSize: %d][capacity: %d][dataFiles: %d]",
			c.path, c.size, c.noFallbackOnError, c.noSync, c.noPos, c.noLock, len(c.dataFiles), c.maxDataSize, c.batchSize, c.capacity, n,
		)
	} else {
		return fmt.Sprintf("%s/[size: %d][fallback: %v][nosync: %v][nopos: %v][nolock: %v][files: %d][maxDataSize: %d][batchSize: %d][capacity: %d][dataFiles: %v]",
			c.path, c.size, c.noFallbackOnError, c.noSync, c.noLock, c.noPos, len(c.dataFiles), c.maxDataSize, c.batchSize, c.capacity, c.dataFiles,
		)
	}
}
