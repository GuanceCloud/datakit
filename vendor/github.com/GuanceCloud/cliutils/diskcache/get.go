// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// Fn is the handler to eat cache from diskcache.
type Fn func([]byte) error

func (c *DiskCache) switchNextFile() error {
	if c.curReadfile != "" {
		if err := c.removeCurrentReadingFile(); err != nil {
			return NewCacheError(OpSwitch, err, "failed_to_remove_current_reading_file").
				WithPath(c.path).WithFile(c.curReadfile)
		}
	}

	// reopen next file to read
	return c.doSwitchNextFile()
}

func (c *DiskCache) skipBadFile() error {
	defer func() {
		droppedDataVec.WithLabelValues(c.path, reasonBadDataFile).Observe(float64(c.curReadSize))
	}()

	l.Warnf("skip bad file %s with size %d bytes", c.curReadfile, c.curReadSize)

	if err := c.switchNextFile(); err != nil {
		return NewCacheError(OpGet, err, "failed_to_skip_bad_file").
			WithPath(c.path).WithFile(c.curReadfile).
			WithDetails(fmt.Sprintf("file_size=%d", c.curReadSize))
	}
	return nil
}

// Get fetch new data from disk cache, then passing to fn
//
// Get is safe to call concurrently with other operations and will
// block until all other operations finish.
func (c *DiskCache) Get(fn Fn) error {
	return c.doGet(nil, fn, nil)
}

type BufFunc func() []byte

// BufCallbackGet fetch new data from disk cache, and read into buffer that returned by bfn.
// If there is nothing to read, the bfn will not be called.
func (c *DiskCache) BufCallbackGet(bfn BufFunc, fn Fn) error {
	return c.doGet(nil, fn, bfn)
}

// BufGet fetch new data from disk cache, and read into buf.
func (c *DiskCache) BufGet(buf []byte, fn Fn) error {
	return c.doGet(buf, fn, nil)
}

func (c *DiskCache) doGet(buf []byte, fn Fn, bfn BufFunc) error {
	var (
		n, nbytes int
		err       error
	)

	c.rlock.Lock()
	defer c.rlock.Unlock()

	start := time.Now()

	defer func() {
		if uint32(nbytes) != EOFHint {
			// get on EOF not counted as a real Get
			getLatencyVec.WithLabelValues(c.path).Observe(time.Since(start).Seconds())
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
			return NewCacheError(OpGet, err, "failed_to_wakeup_sleeping_write_file").
				WithPath(c.path).
				WithDetails(fmt.Sprintf("idle_time=%v, batch_size=%d",
					time.Since(c.wfdLastWrite), c.curBatchSize))
		}
	}

	if c.rfd == nil { // no file reading, reading on the first file
		if err = c.switchNextFile(); err != nil {
			return WrapGetError(err, c.path, "")
		}
	}

retry:
	if c.rfd == nil {
		return ErrNoData
	}

	if n, err = c.rfd.Read(c.batchHeader); err != nil || n != dataHeaderLen {
		if err != nil && !errors.Is(err, io.EOF) {
			l.Errorf("read %d bytes header error: %s", dataHeaderLen, err.Error())
			err = WrapFileOperationError(OpRead, err, c.path, c.rfd.Name()).
				WithDetails(fmt.Sprintf("header_read: expected=%d, actual=%d", dataHeaderLen, n))
		} else if n > 0 && n != dataHeaderLen {
			l.Errorf("invalid header length: %d, expect %d", n, dataHeaderLen)
			err = NewCacheError(OpRead, ErrUnexpectedReadSize,
				fmt.Sprintf("header_size_mismatch: expected=%d, actual=%d", dataHeaderLen, n)).
				WithPath(c.path).WithFile(c.rfd.Name())
		}

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
			return WrapGetError(err, c.path, c.rfd.Name()).
				WithDetails("eof_encountered_during_get")
		}

		goto retry // read next new file to save another Get() calling.
	}

	var readbuf []byte

	switch {
	case buf == nil && bfn == nil: // malloc memory locally
		readbuf = make([]byte, nbytes)
	case buf == nil && bfn != nil:
		readbuf = bfn()
	default:
		readbuf = buf
	}

	if len(readbuf) < nbytes {
		// seek to next read position
		if x, err := c.rfd.Seek(int64(nbytes), io.SeekCurrent); err != nil {
			return WrapFileOperationError(OpSeek, err, c.path, c.rfd.Name()).
				WithDetails(fmt.Sprintf("failed_to_seek_past_data: data_size=%d", nbytes))
		} else {
			l.Warnf("got %d bytes to buffer with len %d, seek to new read position %d, drop %d bytes within file %s",
				nbytes, len(readbuf), x, nbytes, c.curReadfile)

			droppedDataVec.WithLabelValues(c.path, reasonTooSmallReadBuffer).Observe(float64(nbytes))
			return WrapGetError(ErrTooSmallReadBuf, c.path, c.rfd.Name()).
				WithDetails(fmt.Sprintf("buffer_too_small: required=%d, provided=%d", nbytes, len(readbuf)))
		}
	}

	if n, err := c.rfd.Read(readbuf[:nbytes]); err != nil {
		return WrapFileOperationError(OpRead, err, c.path, c.rfd.Name()).
			WithDetails(fmt.Sprintf("data_read: expected=%d, actual=%d", nbytes, n))
	} else if n != nbytes {
		return WrapGetError(ErrUnexpectedReadSize, c.path, c.rfd.Name()).
			WithDetails(fmt.Sprintf("partial_read: expected=%d, actual=%d", nbytes, n))
	}

	if fn == nil {
		goto __updatePos
	}

	if err = fn(readbuf[:nbytes]); err != nil {
		// seek back
		if !c.noFallbackOnError {
			if _, serr := c.rfd.Seek(-int64(dataHeaderLen+nbytes), io.SeekCurrent); serr != nil {
				return WrapFileOperationError(OpSeek, serr, c.path, c.rfd.Name()).
					WithDetails(fmt.Sprintf("fallback_seek_failed: offset=%d", -int64(dataHeaderLen+nbytes)))
			}

			seekBackVec.WithLabelValues(c.path).Inc()
			goto __end // do not update .pos
		}
	}

__updatePos:
	// update seek position
	if !c.noPos && nbytes > 0 {
		c.pos.Seek += int64(dataHeaderLen + nbytes)
		if do, derr := c.pos.dumpFile(); derr != nil {
			return WrapPosError(derr, c.path, c.pos.Seek).WithDetails("failed_to_update_position_after_get")
		} else if do {
			posUpdatedVec.WithLabelValues("get", c.path).Inc()
		}
	}

__end:
	return err
}
