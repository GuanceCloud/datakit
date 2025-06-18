// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Fn is the handler to eat cache from diskcache.
type Fn func([]byte) error

func (c *DiskCache) switchNextFile() error {
	if c.curReadfile != "" {
		if err := c.removeCurrentReadingFile(); err != nil {
			return fmt.Errorf("removeCurrentReadingFile: %w", err)
		}
	}

	// reopen next file to read
	return c.doSwitchNextFile()
}

func (c *DiskCache) skipBadFile() error {
	defer func() {
		droppedDataVec.WithLabelValues(c.path, reasonBadDataFile).Observe(float64(c.curReadSize))
	}()

	return c.switchNextFile()
}

// Get fetch new data from disk cache, then passing to fn
//
// Get is safe to call concurrently with other operations and will
// block until all other operations finish.
func (c *DiskCache) Get(fn Fn) error {
	return c.doGet(nil, fn)
}

// BufGet fetch new data from disk cache, and read into buf.
func (c *DiskCache) BufGet(buf []byte, fn Fn) error {
	return c.doGet(buf, fn)
}

func (c *DiskCache) doGet(buf []byte, fn Fn) error {
	var (
		n, nbytes int
		err       error
	)

	c.rlock.Lock()
	defer c.rlock.Unlock()

	start := time.Now()

	defer func() {
		if uint32(nbytes) != EOFHint {
			getBytesVec.WithLabelValues(c.path).Observe(float64(nbytes))

			// get on EOF not counted as a real Get
			getLatencyVec.WithLabelValues(c.path).Observe(float64(time.Since(start)) / float64(time.Second))
		}
	}()

	// wakeup sleeping write file, rotate it for succession reading!
	if time.Since(c.wfdLastWrite) > c.wakeup && c.curBatchSize > 0 {
		wakeupVec.WithLabelValues(c.path).Inc()

		if err = func() error {
			c.wlock.Lock()
			defer c.wlock.Unlock()
			return c.rotate()
		}(); err != nil {
			return err
		}
	}

	if c.rfd == nil {
		if err = c.switchNextFile(); err != nil {
			return err
		}
	}

retry:
	if c.rfd == nil {
		return ErrNoData
	}

	if n, err = c.rfd.Read(c.batchHeader); err != nil || n != dataHeaderLen {
		// On bad datafile, just ignore and delete the file.
		if err = c.skipBadFile(); err != nil {
			return err
		}

		goto retry // read next new file to save another Get() calling.
	}

	// how many bytes of current data?
	nbytes = int(binary.LittleEndian.Uint32(c.batchHeader))

	if uint32(nbytes) == EOFHint { // EOF
		if err := c.switchNextFile(); err != nil {
			return fmt.Errorf("switchNextFile: %w", err)
		}

		goto retry // read next new file to save another Get() calling.
	}

	if buf == nil {
		buf = make([]byte, nbytes)
	}

	if len(buf) < nbytes {
		// seek to next read position
		if _, err := c.rfd.Seek(int64(nbytes), io.SeekCurrent); err != nil {
			return fmt.Errorf("rfd.Seek(%d): %w", nbytes, err)
		}

		droppedDataVec.WithLabelValues(c.path, reasonTooSmallReadBuffer).Observe(float64(nbytes))
		return ErrTooSmallReadBuf
	}

	if n, err := c.rfd.Read(buf[:nbytes]); err != nil {
		return fmt.Errorf("rfd.Read(%d buf): %w", len(buf[:nbytes]), err)
	} else if n != nbytes {
		return ErrUnexpectedReadSize
	}

	if fn == nil {
		goto __updatePos
	}

	if err = fn(buf[:nbytes]); err != nil {
		// seek back
		if !c.noFallbackOnError {
			if _, serr := c.rfd.Seek(-int64(dataHeaderLen+nbytes), io.SeekCurrent); serr != nil {
				return fmt.Errorf("c.rfd.Seek(%d) on FallbackOnError: %w", -int64(dataHeaderLen+nbytes), serr)
			}

			seekBackVec.WithLabelValues(c.path).Inc()
			goto __end // do not update .pos
		}
	}

__updatePos:
	// update seek position
	if !c.noPos && nbytes > 0 {
		c.pos.Seek += int64(dataHeaderLen + nbytes)
		if derr := c.pos.dumpFile(); derr != nil {
			return derr
		}

		posUpdatedVec.WithLabelValues("get", c.path).Inc()
	}

__end:
	return err
}
