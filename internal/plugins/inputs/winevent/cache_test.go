// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows
// +build windows

package winevent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var val = EvtHandle(1)
var key = "a"

func TestCachePutAndGet(t *testing.T) {
	c := newHandleCache(time.Millisecond, 10, nil, nil)
	c.Put(key, val)
	assert.Equal(t, val, c.Get(key))
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, NilHandle, c.Get(key))
}

func TestCacheCleanUp(t *testing.T) {
	c := newHandleCache(1*time.Millisecond, 10, nil, nil)
	c.Put(key, val)
	assert.Equal(t, 1, c.Size())
	time.Sleep(10 * time.Millisecond)
	c.CleanUp()
	assert.Equal(t, 0, c.Size())

	c.Put(key, val)
	assert.Equal(t, 1, c.Size())
	c.StartCleanWorker(1 * time.Millisecond)
	defer c.StopCleanWorker()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, c.Size())
}

func TestCacheRemoveCallback(t *testing.T) {
	removeCnt := 0
	c := newHandleCache(time.Millisecond, 10, func(s string, eh EvtHandle) {
		removeCnt++
	}, nil)

	c.Put("1", val)
	c.Put("2", val)
	c.Put("3", val)
	c.Put("4", val)
	time.Sleep(10 * time.Millisecond)

	c.CleanUp()
	assert.Equal(t, 4, removeCnt)
}
