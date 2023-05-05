// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	T "testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownMatch(t *T.T) {
	t.Run("basic", func(t *T.T) {
		assert.Len(t, doMatch("中文haha"), 1)
	})

	t.Run("basic-2", func(t *T.T) {
		assert.Len(t, doMatch("中文haha哈哈"), 2)
	})

	t.Run("match-chinese-punctuation", func(t *T.T) {
		assert.Len(t, doMatch("中文，haha，中文"), 0)
	})

	t.Run("match-bad-inline-code", func(t *T.T) {
		assert.Len(t, doMatch("中文`haha`中文"), 2)
	})

	t.Run("match-valid-inline-code", func(t *T.T) {
		assert.Len(t, doMatch("中文 `haha` 中文"), 0)
	})
}
