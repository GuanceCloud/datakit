// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	a := [3]int{}
	arr := a[:]
	queue := NewCircularQueue(arr)

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)
	queue.Enqueue(4)
	queue.Enqueue(5)
	queue.Enqueue(6)
	queue.Enqueue(7)
	for i := 5; i < 8; i++ {
		_, ok := queue.ContainsFunc(func(i int) bool {
			return i == 1
		})
		assert.False(t, ok)
	}
	for j := 5; j < 8; j++ {
		v, ok := queue.ContainsFunc(func(i int) bool {
			return i == j
		})
		assert.True(t, ok)
		assert.Equal(t, j, v)
		v, ok = queue.DeleteFunc(func(i int) bool {
			return i == j
		})
		assert.True(t, ok)
		assert.Equal(t, j, v)
		_, ok = queue.ContainsFunc(func(i int) bool {
			return i == j
		})
		assert.False(t, ok)
	}
}
