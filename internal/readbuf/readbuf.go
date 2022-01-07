// Package readbuf wraps buffer and ReadLine functions
package readbuf

import (
	"bytes"
	"io"
)

// README
//

type ReadBuffer struct {
	io.Reader
	buf           []byte
	previousBlock []byte
}

func NewReadBuffer(reader io.Reader, bufSize int) *ReadBuffer {
	return &ReadBuffer{
		Reader: reader,
		buf:    make([]byte, bufSize),
	}
}

func (r *ReadBuffer) ReadLines() ([][]byte, error) {
	n, err := r.Read(r.buf)
	if err != nil {
		return nil, err
	}

	// 以换行符 split
	lines := bytes.Split(r.buf[:n], []byte{'\n'})
	if len(lines) == 0 {
		return nil, nil
	}

	// block 不为空时，将其内容添加到 lines 首元素前端
	// block 置空
	if len(r.previousBlock) != 0 {
		lines[0] = append(r.previousBlock, lines[0]...)
		r.previousBlock = r.previousBlock[:0]
	}

	// 当 lines 最后一个元素不为空时，说明这段内容并不包含换行符，将其暂存到 previousBlock
	if len(lines[len(lines)-1]) != 0 {
		// 将 lines 尾元素 append previousBlock，避免占用此 slice 造成内存泄漏
		r.previousBlock = append(r.previousBlock, lines[len(lines)-1]...)
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return nil, nil
	}

	// 最后一个元素可能是空，可以选择剔除
	// 不这么做的原因是，这是功能函数，不做多余操作
	if len(lines[len(lines)-1]) == 0 {
		return lines[:len(lines)-1], nil
	}
	return lines, nil
}
