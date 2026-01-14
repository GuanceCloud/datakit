// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

func setupLogger() {
	once.Do(func() {
		l = logger.SLogger("diskcache")
	})
}

// Open init and create a new disk cache. We can set other options with various options.
func Open(opts ...CacheOption) (*DiskCache, error) {
	setupLogger()

	c := defaultInstance()

	// apply extra options
	for _, x := range opts {
		if x != nil {
			x(c)
		}
	}

	if err := c.doOpen(); err != nil {
		return nil, WrapOpenError(err, c.path).WithDetails("failed_to_open_diskcache")
	}

	defer func() {
		c.labels = append(c.labels,
			strconv.FormatBool(c.noFallbackOnError),
			strconv.FormatBool(c.noLock),
			strconv.FormatBool(c.noPos),
			strconv.FormatBool(c.noSync),
			c.path,
		)

		openTimeVec.WithLabelValues(c.labels...).Set(float64(time.Now().Unix()))
	}()

	return c, nil
}

func defaultInstance() *DiskCache {
	return &DiskCache{
		noSync: false,

		batchHeader: make([]byte, dataHeaderLen),

		batchSize:   20 * 1024 * 1024,
		maxDataSize: 0, // not set

		wlock:  nil, // Will be initialized in doOpen() when path is known
		rlock:  nil, // Will be initialized in doOpen() when path is known
		rwlock: nil, // Will be initialized in doOpen() when path is known

		wakeup:    time.Second * 3,
		dirPerms:  0o750,
		filePerms: 0o640,
		pos: &pos{
			Seek: 0,
			Name: nil,

			// dump position each 100ms or 100 update
			dumpInterval: time.Millisecond * 100,
			dumpCount:    100,
		},
	}
}

func (c *DiskCache) doOpen() error {
	if c.dirPerms == 0 {
		c.dirPerms = 0o755
	}

	if c.filePerms == 0 {
		c.filePerms = 0o640
	}

	if c.batchSize == 0 {
		c.batchSize = 20 * 1024 * 1024
	}

	if int64(c.maxDataSize) > c.batchSize {
		// reset max-data-size to half of batch size
		c.maxDataSize = int32(c.batchSize / 2)
	}

	if err := os.MkdirAll(c.path, c.dirPerms); err != nil {
		return NewCacheError(OpCreate, err, fmt.Sprintf("failed_to_create_directory: perms=%o", c.dirPerms)).
			WithPath(c.path)
	}

	// disable open multiple times
	if !c.noLock {
		fl := newFlock(c.path)
		if err := fl.lock(); err != nil {
			return WrapLockError(err, c.path, 0).WithDetails("failed_to_acquire_directory_lock")
		} else {
			c.flock = fl
		}
	}

	if !c.noPos {
		// use `.pos' file to remember the reading position.
		c.pos.fname = filepath.Join(c.path, ".pos")
	}
	c.curWriteFile = filepath.Join(c.path, "data")

	c.syncEnv()

	// set stable metrics
	capVec.WithLabelValues(c.path).Set(float64(c.capacity))
	maxDataVec.WithLabelValues(c.path).Set(float64(c.maxDataSize))
	batchSizeVec.WithLabelValues(c.path).Set(float64(c.batchSize))

	// Initialize instrumented locks now that we have the path
	if c.wlock == nil {
		c.wlock = NewInstrumentedMutex(LockTypeWrite, c.path, lockWaitTimeVec, lockContentionVec)
	}
	if c.rlock == nil {
		c.rlock = NewInstrumentedMutex(LockTypeRead, c.path, lockWaitTimeVec, lockContentionVec)
	}
	if c.rwlock == nil {
		c.rwlock = NewInstrumentedMutex(LockTypeRW, c.path, lockWaitTimeVec, lockContentionVec)
	}

	// write append fd, always write to the same-name file
	if err := c.openWriteFile(); err != nil {
		return NewCacheError(OpOpen, err, "failed_to_open_write_file").
			WithPath(c.path).WithFile(c.curWriteFile)
	}

	// list files under @path
	if err := filepath.Walk(c.path,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return NewCacheError(OpOpen, err, "failed_to_walk_directory").
					WithPath(c.path).WithFile(path)
			}

			if fi.IsDir() {
				return nil
			}

			switch filepath.Base(path) {
			case ".lock", ".pos": // ignore them
			case "data": // not rotated writing file, do not count on sizeVec.
			default:
				c.size.Add(fi.Size())
				sizeVec.WithLabelValues(c.path).Add(float64(fi.Size()))
				c.dataFiles = append(c.dataFiles, path)
			}

			return nil
		}); err != nil {
		return err
	}

	sort.Strings(c.dataFiles) // make file-name sorted for FIFO Get()
	l.Infof("on open loaded %d files", len(c.dataFiles))
	datafilesVec.WithLabelValues(c.path).Set(float64(len(c.dataFiles)))

	// first get, try load .pos
	if !c.noPos {
		if err := c.loadUnfinishedFile(); err != nil {
			return NewCacheError(OpOpen, err, "failed_to_load_position_file").
				WithPath(c.path)
		}
	}

	return nil
}

// Close reclame fd resources.
// Close is safe to call concurrently with other operations and will
// block until all other operations finish.
func (c *DiskCache) Close() error {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	defer func() {
		lastCloseTimeVec.WithLabelValues(c.path).Set(float64(time.Now().Unix()))
	}()

	if c.rfd != nil {
		if err := c.rfd.Close(); err != nil {
			return WrapCloseError(err, c.path, "read_fd")
		}
		c.rfd = nil
	}

	if !c.noLock {
		if c.flock != nil {
			if err := c.flock.unlock(); err != nil {
				return WrapLockError(err, c.path, 0).WithDetails("failed_to_release_directory_lock")
			}
		}
	}

	if c.wfd != nil {
		if err := c.wfd.Close(); err != nil {
			return WrapCloseError(err, c.path, "write_fd")
		}
		c.wfd = nil
	}

	if c.pos != nil {
		if err := c.pos.close(); err != nil {
			return WrapPosError(err, c.path, c.pos.Seek).WithDetails("failed_to_close_position_file")
		}
	}

	return nil
}
