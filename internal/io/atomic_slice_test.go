// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package io

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppend(t *testing.T) {
	const size = 1000
	n := size - 1
	maxSize := n * (1 + n) / 2
	// slice := NewSafeSlice[int](int64(maxSize))
	slice := NewSafeSlice[[]int](0)
	for c := 0; c < 10; c++ {
		m := [size]int{}
		wg := sync.WaitGroup{}
		for i := 0; i < size; i++ {
			wg.Add(1)
			i := i
			go func() {
				ints := make([]int, 0)
				for j := 0; j < i; j++ {
					ints = append(ints, i)
				}

				slice.Append(ints)
				wg.Done()
			}()
		}
		wg.Wait()
		arr := slice.Reset(false)
		assert.Equal(t, maxSize, len(arr))
		for _, v := range arr {
			m[v]++
		}
		for i, v := range m {
			assert.Equal(t, i, v)
		}
	}
}

func TestReset(t *testing.T) {
	const size = 50000
	slice := NewSafeSlice[[]int](0)
	wg := sync.WaitGroup{}
	for c := 0; c < 10; c++ {
		wg.Add(1)
		go func() {
			for i := 0; i < size; i++ {
				if rand.Int()%2 == 0 {
					d := slice.Reset(false)
					m := map[int]int{}
					for _, v := range d {
						m[v]++
					}
					for k, v := range m {
						assert.Equal(t, 0, v%k)
					}
				} else {
					n := rand.Int63n(size)
					arr := make([]int, n)
					for i := range arr {
						arr[i] = int(n)
					}
					slice.Append(arr)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
