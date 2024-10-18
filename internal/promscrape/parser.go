// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promscrape

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	// The maximum size of a single line returned by ReadLinesBlock.
	maxLineSize = 256 * 1024

	// Default size in bytes of a single block returned by ReadLinesBlock.
	defaultBlockSize = 64 * 1024
)

func ParseStream(r io.Reader, defaultTimestamp int64, isGzipped bool, callback func(rows []Row) error) error {
	ctx := getStreamContext(r)
	defer putStreamContext(ctx)
	for ctx.Read() {
		uw := getUnmarshalWork()
		uw.ctx = ctx
		uw.callback = callback
		uw.defaultTimestamp = defaultTimestamp
		uw.reqBuf, ctx.reqBuf = ctx.reqBuf, uw.reqBuf
		if err := uw.Unmarshal(); err != nil {
			ctx.err = err
		}
		putUnmarshalWork(uw)
	}
	return ctx.Error()
}

var (
	unmarshalWorkPool sync.Pool
	streamContextPool sync.Pool
)

func getUnmarshalWork() *unmarshalWork {
	v := unmarshalWorkPool.Get()
	if v == nil {
		return &unmarshalWork{}
	}
	return v.(*unmarshalWork)
}

func putUnmarshalWork(uw *unmarshalWork) {
	uw.reset()
	unmarshalWorkPool.Put(uw)
}

func getStreamContext(r io.Reader) *streamContext {
	if v := streamContextPool.Get(); v != nil {
		ctx := v.(*streamContext)
		ctx.br.Reset(r)
		return ctx
	}
	return &streamContext{
		br: bufio.NewReaderSize(r, defaultBlockSize),
	}
}

func putStreamContext(ctx *streamContext) {
	ctx.reset()
	streamContextPool.Put(ctx)
}

type streamContext struct {
	br      *bufio.Reader
	reqBuf  []byte
	tailBuf []byte
	err     error
}

func (ctx *streamContext) Read() bool {
	if ctx.err != nil {
		return false
	}
	ctx.reqBuf, ctx.tailBuf, ctx.err = ReadLinesBlock(ctx.br, ctx.reqBuf, ctx.tailBuf)
	if ctx.err != nil {
		if errors.Is(ctx.err, io.EOF) {
			ctx.err = fmt.Errorf("cannot read Prometheus exposition data: %w", ctx.err)
		}
		return false
	}
	return true
}

func (ctx *streamContext) Error() error {
	if errors.Is(ctx.err, io.EOF) {
		return nil
	}
	return ctx.err
}

func (ctx *streamContext) reset() {
	ctx.br.Reset(nil)
	ctx.reqBuf = ctx.reqBuf[:0]
	ctx.tailBuf = ctx.tailBuf[:0]
	ctx.err = nil
}

type unmarshalWork struct {
	rows             Rows
	ctx              *streamContext
	callback         func(rows []Row) error
	defaultTimestamp int64
	reqBuf           []byte
}

func (uw *unmarshalWork) reset() {
	uw.rows.Reset()
	uw.ctx = nil
	uw.callback = nil
	uw.defaultTimestamp = 0
	uw.reqBuf = uw.reqBuf[:0]
}

func (uw *unmarshalWork) runCallback(rows []Row) error {
	return uw.callback(rows)
}

func (uw *unmarshalWork) Unmarshal() error {
	if err := uw.rows.Unmarshal(string(uw.reqBuf)); err != nil {
		return err
	}

	rows := uw.rows.Rows

	defaultTimestamp := uw.defaultTimestamp
	if defaultTimestamp <= 0 {
		defaultTimestamp = time.Now().UnixNano() / 1e6
	}
	for i := range rows {
		r := &rows[i]
		if r.Timestamp == 0 {
			r.Timestamp = defaultTimestamp
		}
	}

	return uw.runCallback(rows)
}

// ReadLinesBlock reads a block of lines delimited by '\n' from tailBuf and r into dstBuf.
//
// Trailing chars after the last newline are put into tailBuf.
//
// Returns (dstBuf, tailBuf).
//
// It is expected that read timeout on r exceeds 1 second.
func ReadLinesBlock(r io.Reader, dstBuf, tailBuf []byte) ([]byte, []byte, error) {
	return ReadLinesBlockExt(r, dstBuf, tailBuf, maxLineSize)
}

// ReadLinesBlockExt reads a block of lines delimited by '\n' from tailBuf and r into dstBuf.
//
// Trailing chars after the last newline are put into tailBuf.
//
// Returns (dstBuf, tailBuf).
//
// maxLineLen limits the maximum length of a single line.
//
// It is expected that read timeout on r exceeds 1 second.
func ReadLinesBlockExt(r io.Reader, dstBuf, tailBuf []byte, maxLineLen int) ([]byte, []byte, error) {
	if cap(dstBuf) < defaultBlockSize {
		dstBuf = ResizeNoCopyNoOverallocate(dstBuf, defaultBlockSize)
	}
	dstBuf = append(dstBuf[:0], tailBuf...)
	tailBuf = tailBuf[:0]
again:
	n, err := r.Read(dstBuf[len(dstBuf):cap(dstBuf)])
	// Check for error only if zero bytes read from r, i.e. no forward progress made.
	// Otherwise process the read data.
	if n == 0 {
		if err == nil {
			return dstBuf, tailBuf, fmt.Errorf("no forward progress made")
		}
		isEOF := isEOFLikeError(err)
		if isEOF && len(dstBuf) > 0 {
			// Missing newline in the end of stream. This is OK,
			// so suppress io.EOF for now. It will be returned during the next
			// call to ReadLinesBlock.
			// This fixes https://github.com/VictoriaMetrics/VictoriaMetrics/issues/60 .
			return dstBuf, tailBuf, nil
		}
		if !isEOF {
			err = fmt.Errorf("cannot read a block of data: %w", err)
		} else {
			err = io.EOF
		}
		return dstBuf, tailBuf, err
	}
	dstBuf = dstBuf[:len(dstBuf)+n]

	// Search for the last newline in dstBuf and put the rest into tailBuf.
	nn := bytes.LastIndexByte(dstBuf[len(dstBuf)-n:], '\n')
	if nn < 0 {
		// Didn't found at least a single line.
		if len(dstBuf) > maxLineLen {
			return dstBuf, tailBuf, fmt.Errorf("too long line: more than %d bytes", maxLineLen)
		}
		if cap(dstBuf) < 2*len(dstBuf) {
			// Increase dsbBuf capacity, so more data could be read into it.
			dstBufLen := len(dstBuf)
			dstBuf = ResizeWithCopyNoOverallocate(dstBuf, 2*cap(dstBuf))
			dstBuf = dstBuf[:dstBufLen]
		}
		goto again
	}

	// Found at least a single line. Return it.
	nn += len(dstBuf) - n
	tailBuf = append(tailBuf[:0], dstBuf[nn+1:]...)
	dstBuf = dstBuf[:nn]
	return dstBuf, tailBuf, nil
}

func isEOFLikeError(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "reset by peer")
}

// ResizeNoCopyNoOverallocate resizes b to exactly n bytes and returns the resized buffer (which may be newly allocated).
//
// If newly allocated buffer is returned then b contents isn't copied to it.
func ResizeNoCopyNoOverallocate(b []byte, n int) []byte {
	if n <= cap(b) {
		return b[:n]
	}
	return make([]byte, n)
}

// ResizeWithCopyNoOverallocate resizes b to exactly n bytes and returns the resized buffer (which may be newly allocated).
//
// If newly allocated buffer is returned then b contents is copied to it.
func ResizeWithCopyNoOverallocate(b []byte, n int) []byte {
	if n <= cap(b) {
		return b[:n]
	}
	bNew := make([]byte, n)
	copy(bNew, b)
	return bNew
}
