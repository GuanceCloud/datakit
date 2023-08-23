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
	"sync"
	"time"
)

// Open init and create a new disk cache. We can set other options with various options.
func Open(opts ...CacheOption) (*DiskCache, error) {
	c := defaultInstance()

	// apply extra options
	for _, x := range opts {
		if x != nil {
			x(c)
		}
	}

	if err := c.doOpen(); err != nil {
		return nil, err
	}

	defer func() {
		c.labels = append(c.labels,
			fmt.Sprintf("%v", c.noFallbackOnError),
			fmt.Sprintf("%v", c.noLock),
			fmt.Sprintf("%v", c.noPos),
			fmt.Sprintf("%v", c.noSync),
			c.path,
		)

		openTimeVec.WithLabelValues(c.labels...).Set(float64(time.Now().Unix()))
	}()

	return c, nil
}

func defaultInstance() *DiskCache {
	return &DiskCache{
		noSync: false,

		batchSize:   20 * 1024 * 1024,
		maxDataSize: 0, // not set

		wlock:  &sync.Mutex{},
		rlock:  &sync.Mutex{},
		rwlock: &sync.Mutex{},

		wakeup:    time.Second * 3,
		dirPerms:  0o750,
		filePerms: 0o640,
		pos: &pos{
			Seek: 0,
			Name: nil,
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
		return err
	}

	// disable open multiple times
	if !c.noLock {
		fl := newFlock(c.path)
		if err := fl.lock(); err != nil {
			return fmt.Errorf("lock: %w", err)
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

	// write append fd, always write to the same-name file
	if err := c.openWriteFile(); err != nil {
		return err
	}

	// list files under @path
	if err := filepath.Walk(c.path,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if fi.IsDir() {
				return nil
			}

			switch filepath.Base(path) {
			case ".lock", ".pos": // ignore them
			case "data": // count on size
				c.size += fi.Size()
			default:
				c.size += fi.Size()
				c.dataFiles = append(c.dataFiles, path)
			}

			return nil
		}); err != nil {
		return err
	}

	sort.Strings(c.dataFiles) // make file-name sorted for FIFO Get()

	// first get, try load .pos
	if !c.noPos {
		if err := c.loadUnfinishedFile(); err != nil {
			return err
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
			return err
		}
		c.rfd = nil
	}

	if !c.noLock {
		if c.flock != nil {
			if err := c.flock.unlock(); err != nil {
				return err
			}
		}
	}

	if c.wfd != nil {
		if err := c.wfd.Close(); err != nil {
			return err
		}
		c.wfd = nil
	}

	return nil
}
