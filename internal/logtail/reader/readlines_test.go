// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package reader

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const mockdata = "0123456789\nabcde\nABCDE"

func TestReadLines(t *testing.T) {
	buf := bytes.Buffer{}
	buf.WriteString(mockdata)

	size := 10

	r := NewReader(&buf, WithBufSize(size))

	for i := 0; i <= len(mockdata)/size; i++ {
		switch i {
		case 0:
			res, n, err := r.ReadLines()
			assert.NoError(t, err)
			assert.Equal(t, n, size)
			assert.Nil(t, res)
		case 1:
			res, n, err := r.ReadLines()
			assert.NoError(t, err)
			assert.Equal(t, n, size)
			assert.Equal(t, 2, len(res))
			assert.Equal(t, []byte("0123456789"), res[0])
			assert.Equal(t, []byte("abcde"), res[1])
		case 2:
			res, n, err := r.ReadLines()
			assert.NoError(t, err)
			assert.Equal(t, n, len(mockdata)%size)
			assert.Nil(t, res)
		}
	}

	_, n, err := r.ReadLines()
	assert.Equal(t, n, 0)
	assert.Equal(t, ErrReadEmpty, err)
}

func TestSplit(t *testing.T) {
	r := &reader{
		opt: defaultOption(),
	}
	res := r.split([]byte(mockdata))
	assert.Equal(t, 2, len(res))
	assert.Equal(t, []byte("0123456789"), res[0])
	assert.Equal(t, []byte("abcde"), res[1])

	assert.Equal(t, []byte("ABCDE"), r.previousBlock)
}
