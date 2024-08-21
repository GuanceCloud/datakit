// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package reader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithOptions(t *testing.T) {
	t.Run("with-bufSize", func(t *testing.T) {
		opt := defaultOption()
		WithBufSize(10)(opt)
		assert.Equal(t, 10, opt.bufSize)
	})

	t.Run("with-maxLineLength", func(t *testing.T) {
		opt := defaultOption()
		WithMaxLineLength(20)(opt)
		assert.Equal(t, 20, opt.maxLineLength)
	})

	t.Run("disablePreviousBlock", func(t *testing.T) {
		opt := defaultOption()
		DisablePreviousBlock()(opt)
		assert.Equal(t, true, opt.disablePreviousBlock)
	})
}
