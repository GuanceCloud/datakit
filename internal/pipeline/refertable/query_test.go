// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinSearch(t *testing.T) {
	cases := []struct {
		li  []int
		v   int
		ok  bool
		idx int
	}{
		{
			[]int{1, 2, 5, 6},
			2,
			true,
			1,
		},
		{
			[]int{1, 2, 5, 6},
			5,
			true,
			2,
		},
		{
			[]int{1, 2, 5, 6},
			1,
			true,
			0,
		},
		{
			[]int{1, 2, 5, 6},
			6,
			true,
			3,
		},
		{
			[]int{1, 2, 5, 6, 11},
			2,
			true,
			1,
		},
		{
			[]int{1, 2, 5, 6, 11},
			11,
			true,
			4,
		},
		{
			[]int{1},
			6,
			false,
			0,
		},
		{
			[]int{1},
			1,
			true,
			0,
		},
	}

	for _, v := range cases {
		if idx, ok := binSearch(v.li, v.v); ok {
			assert.Equal(t, v.ok, ok)
			assert.Equal(t, v.idx, idx)
		} else {
			assert.Equal(t, v.ok, ok)
		}
	}
}

func TestQuery(t *testing.T) {
	cases := []struct {
		index  map[string]map[any][]int
		keys   []string
		values []any
		count  int
		ok     bool
		ret    []int
	}{
		{
			index: map[string]map[any][]int{
				"key1": {
					"1": {1, 3, 5, 11, 18},
					"3": {2, 4, 6, 8, 12},
				},
				"key2": {
					1: {1, 2, 5, 8, 12, 17},
					2: {3, 4, 6, 11},
				},
			},
			keys:   []string{"key2", "key1"},
			values: []any{1, "1"},
			count:  0,
			ok:     true,
			ret:    []int{1, 5},
		},
		{
			index: map[string]map[any][]int{
				"key1": {
					"1": {3, 5, 11, 18},
					"3": {2, 4, 6, 8, 12},
				},
				"key2": {
					1: {1, 2, 3, 5, 8, 12, 17},
					2: {4, 6, 11},
				},
				"key3": {
					1.1: {2, 3, 5},
					2.2: {1, 4, 6, 7},
				},
			},
			keys:   []string{"key3", "key2", "key1"},
			values: []any{1.1, 1, "1"},
			count:  1,
			ok:     true,
			ret:    []int{3},
		},
		{
			index:  map[string]map[any][]int{},
			keys:   []string{"key5", "key2", "key1"},
			values: []any{1.1, 1, "1"},
			count:  1,
			ok:     false,
			ret:    []int{3},
		},
		{
			index: map[string]map[any][]int{
				"key1": {},
				"key3": {
					1.1: {2, 3, 5},
					2.2: {1, 4, 6, 7},
				},
			},
			keys:   []string{"key3", "key1"},
			values: []any{1.1, "1"},
			count:  1,
			ok:     false,
			ret:    []int{3},
		},
		{
			index: map[string]map[any][]int{
				"key1": {
					"1": {1, 3, 5, 11, 18},
					"3": {2, 4, 6, 8, 12},
				},
				"key2": {
					1: {2, 8, 12, 17},
					2: {3, 4, 6, 11},
				},
			},
			keys:   []string{"key2", "key1"},
			values: []any{1, "1"},
			count:  0,
			ok:     false,
			ret:    []int{1, 5},
		},
		{
			index: map[string]map[any][]int{
				"key1": {
					"1": {1, 3, 5, 11, 18},
					"3": {2, 4, 6, 8, 12},
				},
				"key2": {
					1: {2, 8, 12, 17},
					2: {3, 4, 6, 11},
				},
			},
			keys:   []string{"key2", "key1"},
			values: []any{1},
			count:  0,
			ok:     false,
			ret:    []int{1, 5},
		},
	}
	for idx, v := range cases {
		t.Run(fmt.Sprint("index_", idx), func(t *testing.T) {
			if ret, ok := query(v.index, v.keys, v.values, v.count); ok {
				assert.Equal(t, v.ok, ok)
				assert.Equal(t, v.ret, ret)
			} else {
				assert.Equal(t, v.ok, ok)
			}
		})
	}
}
