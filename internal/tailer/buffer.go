// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import "bytes"

type buffer struct {
	buf           []byte
	previousBlock []byte
}

var splitCharacter = []byte{'\n'}

func (b *buffer) split() [][]byte {
	lines := bytes.Split(b.buf, splitCharacter)
	if len(lines) == 0 {
		return nil
	}

	var res [][]byte

	// block 不为空时，将其内容添加到 lines 首元素前端
	// block 置空
	if len(b.previousBlock) != 0 {
		lines[0] = append(b.previousBlock, lines[0]...)
		b.previousBlock = nil
	}

	// 当 lines 最后一个元素不为空时，说明这段内容并不包含换行符，将其暂存到 previousBlock
	if len(lines[len(lines)-1]) != 0 {
		// 将 lines 尾元素 append previousBlock，避免占用此 slice 造成内存泄漏
		b.previousBlock = append(b.previousBlock, lines[len(lines)-1]...)
		lines = lines[:len(lines)-1]
	}

	if len(b.previousBlock) > maxReadSize {
		tmp := make([]byte, len(b.previousBlock))
		copy(tmp, b.previousBlock)
		res = append(res, tmp)
		b.previousBlock = nil
	}

	// lines 不需要 copy，因为 split 函数已经执行过 copy
	res = append(res, lines...)
	return res
}
