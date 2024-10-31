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
	SetReader(io.Reader)
	ReadLines() ([][]byte, int, error)
	ReadLineBlock() ([]byte, int, error)
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

func (r *reader) SetReader(rd io.Reader) {
	// 避免再次 NewReader 导致 previousBlock 失效
	r.rd = rd
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
	return r.splitLines(dst), n, nil
}

func (r *reader) ReadLineBlock() ([]byte, int, error) {
	n, err := r.rd.Read(r.buf)
	if err != nil && err != io.EOF {
		return nil, n, err
	}
	if n == 0 {
		return nil, n, ErrReadEmpty
	}

	return r.splitLineBlock(r.buf[:n]), n, nil
}

func (r *reader) splitLines(b []byte) [][]byte {
	lines := SplitLines(b)
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

var splitCharacter = []byte{'\n'}

func (r *reader) splitLineBlock(b []byte) []byte {
	var block []byte

	index := bytes.LastIndex(b, splitCharacter)
	if index == -1 {
		r.previousBlock = append(r.previousBlock, b...)
	} else {
		block = make([]byte, len(r.previousBlock)+index+1)

		copy(block, r.previousBlock)
		previousBlockLen := len(r.previousBlock)
		r.previousBlock = nil

		copy(block[previousBlockLen:], b[:index+1])
		r.previousBlock = append(r.previousBlock, b[index+1:]...)
	}

	if len(r.previousBlock) > r.opt.maxLineLength {
		if len(block) == 0 {
			block = make([]byte, len(r.previousBlock))
			copy(block, r.previousBlock)
		} else {
			// FIXME: lint error?
			// nolint
			block = append(block, r.previousBlock...)
		}
		r.previousBlock = nil
	}

	return block
}

func SplitLines(b []byte) [][]byte {
	return bytes.Split(b, splitCharacter)
}
