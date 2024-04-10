// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Beats (https://github.com/elastic/beats).

//go:build windows
// +build windows

package winevent

import (
	"context"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

// handleCallback is a callback function to free the handle.
type handleCallback func(string, EvtHandle)

// element is a cache element.
type element struct {
	expiration time.Time
	timeout    time.Duration
	value      EvtHandle
}

// IsExpired checks whether the element is expired.
func (e *element) IsExpired() bool {
	return time.Now().After(e.expiration)
}

// UpdateLastAccessTime updates the last access time of the element.
func (e *element) UpdateLastAccessTime() {
	e.expiration = time.Now().Add(e.timeout)
}

// handleCache is a cache of handles.
type handleCache struct {
	sync.RWMutex
	elements        map[string]*element
	timeout         time.Duration
	removalCallback handleCallback
	postCleanUp     func()
	stopCh          chan struct{}
}

// Get gets the handle from the cache.
func (c *handleCache) Get(k string) EvtHandle {
	c.RLock()
	defer c.RUnlock()
	if v, ok := c.elements[k]; ok && !v.IsExpired() {
		v.UpdateLastAccessTime()
		return v.value
	}

	return NilHandle
}

// Put puts the handle into the cache.
func (c *handleCache) Put(k string, v EvtHandle) {
	c.Lock()
	defer c.Unlock()
	c.elements[k] = &element{
		expiration: time.Now().Add(c.timeout),
		timeout:    c.timeout,
		value:      v,
	}
}

// CleanUp cleans up the expired handles.
func (c *handleCache) CleanUp() {
	defer func() {
		if c.postCleanUp != nil {
			c.postCleanUp()
		}
	}()
	c.Lock()
	defer c.Unlock()
	for k, v := range c.elements {
		if v.IsExpired() {
			delete(c.elements, k)
			if c.removalCallback != nil {
				c.removalCallback(k, v.value)
			}
		}
	}
}

// StartCleanWorker starts a worker to clean up the expired handles.
func (c *handleCache) StartCleanWorker(interval time.Duration) {
	c.stopCh = make(chan struct{})
	ticker := time.NewTicker(interval)
	g := goroutine.NewGroup(goroutine.Option{Name: "winevent_cache"})
	g.Go(func(ctx context.Context) error {
		for {
			select {
			case <-c.stopCh:
				ticker.Stop()
				return nil
			case <-ticker.C:
				c.CleanUp()
			}
		}
	})
}

// Size gets the size of the cache.
func (c *handleCache) Size() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.elements)
}

// StopCleanWorker stops the worker.
func (c *handleCache) StopCleanWorker() {
	if c.stopCh != nil {
		close(c.stopCh)
	}
}

// newHandleCache creates a new handleCache.
func newHandleCache(d time.Duration, initialSize int, removalCallback handleCallback, postCleanUp func()) *handleCache {
	return &handleCache{
		timeout:         d,
		elements:        make(map[string]*element, initialSize),
		removalCallback: removalCallback,
		postCleanUp:     postCleanUp,
	}
}
