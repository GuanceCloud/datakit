// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package reader wraps readlines functions
package reader

import (
	"bytes"
	"errors"
	"io"
)

var ErrReadEmpty = errors.New("read 0")

type Reader interface {
	ReadLines() ([][]byte, int, error)
}

type reader struct {
	opt           *option
	rd            io.Reader
	buf           []byte
	previousBlock []byte
}

func NewReader(rd io.Reader, opts ...Option) Reader {
	c := defaultOption()
	for _, opt := range opts {
		opt(c)
	}
	return &reader{
		opt: c,
		rd:  rd,
		buf: make([]byte, c.bufSize),
	}
}

func (r *reader) ReadLines() ([][]byte, int, error) {
	n, err := r.rd.Read(r.buf)
	if err != nil && err != io.EOF {
		return nil, n, err
	}
	if n == 0 {
		return nil, n, ErrReadEmpty
	}

	dst := make([]byte, n)
	copy(dst, r.buf[:n])
	return r.split(dst), n, nil
}

var splitCharacter = []byte{'\n'}

func (r *reader) split(b []byte) [][]byte {
	lines := bytes.Split(b, splitCharacter)
	if len(lines) == 0 {
		return nil
	}

	if r.opt.disablePreviousBlock {
		return lines
	}

	var res [][]byte

	// block 不为空时，将其内容添加到 lines 首元素前端
	// block 置空
	if len(r.previousBlock) != 0 {
		lines[0] = append(r.previousBlock, lines[0]...)
		r.previousBlock = nil
	}

	// 当 lines 最后一个元素不为空时，说明这段内容并不包含换行符，将其暂存到 previousBlock
	if len(lines[len(lines)-1]) != 0 {
		// 将 lines 尾元素 append previousBlock，避免占用此 slice 造成内存泄漏
		r.previousBlock = append(r.previousBlock, lines[len(lines)-1]...)
		lines = lines[:len(lines)-1]
	}

	if len(r.previousBlock) > r.opt.maxLineLength {
		tmp := make([]byte, len(r.previousBlock))
		copy(tmp, r.previousBlock)
		res = append(res, tmp)
		r.previousBlock = nil
	}

	res = append(res, lines...)
	return res
}
