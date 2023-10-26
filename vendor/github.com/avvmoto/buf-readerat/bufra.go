// Copyright 2016 avvmoto. All rights reserved.

// Package buf-readerat implements buffered io.ReaderAt. It wraps an io.ReaderAt
// object, creating another io.ReaderAt object that also implements the interface
// but provides buffering.
package bufra

import "io"

const sizeNotInitialized = -1
const minBufferSize = 4

// BufReaderAt implements buffering for an io.ReaderAt object.
type BufReaderAt struct {
	offset     int64
	buf        []byte // TODO: ring buffer may be better
	readerAt   io.ReaderAt
	sizeTilEOF int
	lastErr    error
}

// NewBufReaderAt returns a new BufReaderAt whose buffer has the specified size.
func NewBufReaderAt(readerAt io.ReaderAt, size int) *BufReaderAt {
	if size < minBufferSize {
		size = minBufferSize
	}

	return &BufReaderAt{
		offset:     0,
		readerAt:   readerAt,
		buf:        make([]byte, size),
		sizeTilEOF: sizeNotInitialized,
	}
}

// ReadAt implements bufferd io.ReadAt. If the under underlying ReadAt return some error, this ReadAt return the same error at corresponding offset.
func (r *BufReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	bufEnd := off + int64(len(b))

	if (off < r.expectedChaceOffset(off) && r.expectedChaceOffset(off) < bufEnd) || (off < r.expectedChaceEnd(off) && r.expectedChaceEnd(off) < bufEnd) {
		// TODO: should be renew chache at here to decrease calling the original io.ReadAt().
		return r.readerAt.ReadAt(b, off)
	} else if !r.isInitialized() || r.expectedChaceEnd(off) <= r.offset || r.cacheEnd() <= r.expectedChaceOffset(off) {
		n, err = r.readAtAndRenewCache(off)
		if n == 0 {
			return
		}

		return r.copySafe(b, off)
	} else if r.offset <= off && bufEnd <= r.cacheEnd() {
		return r.copySafe(b, off)
	}

	return
}

func (r *BufReaderAt) bufSize() int64 {
	return int64(len(r.buf))
}

func (r *BufReaderAt) expectedChaceOffset(off int64) (i int64) {
	return (off / r.bufSize()) * r.bufSize()
}

func (r *BufReaderAt) expectedChaceEnd(off int64) (i int64) {
	return r.expectedChaceOffset(off) + r.bufSize()
}

func (r *BufReaderAt) cacheEnd() int64 {
	return r.offset + r.bufSize()
}

func (r *BufReaderAt) readAtAndRenewCache(off int64) (n int, err error) {
	eOff := r.expectedChaceOffset(off)

	// check whether this offset has been cached
	if r.isInitialized() && eOff == r.offset {
		return
	}

	// not cached, so read
	r.reset(eOff)
	n, err = r.readerAt.ReadAt(r.buf, eOff)

	r.sizeTilEOF = n
	r.lastErr = err

	return
}

// Copy c.buf to b, regarding c.sizeTilEOF.
// return last happend error as err when buffer ends.
func (r *BufReaderAt) copySafe(b []byte, off int64) (n int, err error) {
	bufEnd := off + int64(len(b))
	s := int(off - r.expectedChaceOffset(off))

	var e int
	switch {
	case r.dataEnd() <= off:
		return 0, io.EOF
	case off < r.dataEnd() && r.dataEnd() < bufEnd:
		e = r.sizeTilEOF
		err = r.lastErr
	case bufEnd <= r.dataEnd():
		e = s + len(b)
	}

	copy(b, r.buf[s:e])

	return e - s, err
}

func (r *BufReaderAt) reset(offset int64) {
	r.offset = offset
	r.sizeTilEOF = sizeNotInitialized
	r.lastErr = nil
}

func (r *BufReaderAt) dataEnd() int64 {
	return r.offset + int64(r.sizeTilEOF)
}

func (r *BufReaderAt) isInitialized() bool {
	return 0 <= r.sizeTilEOF
}
