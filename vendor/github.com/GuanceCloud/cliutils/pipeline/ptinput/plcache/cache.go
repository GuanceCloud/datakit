// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package plcache implements cache in pipeline
package plcache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

var (
	ErrClosed = errors.New("cache is closed")
	ErrArg    = errors.New("illegal arguments")
)

type cacheItem struct {
	key        string
	value      any
	expiration time.Duration
	circle     int
}

type Cache struct {
	mu        sync.RWMutex
	interval  time.Duration
	ticker    *time.Ticker
	slots     []*list.List
	numSlots  int
	tickedPos int

	// All cache items for fast query
	items map[string]*list.Element

	setChannel  chan cacheItem
	stopChannel chan struct{}
}

// NewCache returns a Cache.
func NewCache(interval time.Duration, numSlots int) (*Cache, error) {
	if interval <= 0 || numSlots <= 0 {
		return nil, ErrArg
	}

	return NewCacheWithTicker(interval, numSlots, time.NewTicker(interval))
}

// NewCacheWithTicker returns a Cache with the given ticker.
func NewCacheWithTicker(interval time.Duration, numSlots int, ticker *time.Ticker) (*Cache, error) {
	c := &Cache{
		interval:  interval,
		ticker:    ticker,
		slots:     make([]*list.List, numSlots),
		numSlots:  numSlots,
		tickedPos: numSlots - 1,
		items:     make(map[string]*list.Element),

		setChannel:  make(chan cacheItem),
		stopChannel: make(chan struct{}),
	}
	c.initSlot()
	go c.run()

	return c, nil
}

func (c *Cache) initSlot() {
	for i := 0; i < c.numSlots; i++ {
		c.slots[i] = list.New()
	}
}

func (c *Cache) run() {
	for {
		select {
		case <-c.ticker.C:
			c.onTick()
		case ci := <-c.setChannel:
			c.setCacheItem(ci)
		case <-c.stopChannel:
			c.ticker.Stop()
			return
		}
	}
}

func (c *Cache) onTick() {
	c.tickedPos = (c.tickedPos + 1) % c.numSlots
	lst := c.slots[c.tickedPos]
	c.scanAndRemoveExpiredCache(lst)
}

func (c *Cache) scanAndRemoveExpiredCache(lst *list.List) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for e := lst.Front(); e != nil; {
		ci := e.Value.(cacheItem)
		if ci.circle > 0 {
			ci.circle--
			e = e.Next()
			continue
		}
		lst.Remove(e)
		delete(c.items, e.Value.(cacheItem).key)
		e = e.Next()
	}
}

func (c *Cache) setCacheItem(ci cacheItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ci.expiration < c.interval {
		ci.expiration = c.interval
	}

	pos, circle := c.getPosAndCircle(ci.expiration)
	ci.circle = circle
	c.slots[pos].PushBack(ci)
	c.items[ci.key] = c.slots[pos].Back()
}

func (c *Cache) getPosAndCircle(d time.Duration) (pos, circle int) {
	step := int(d / c.interval)
	pos = (c.tickedPos + step) % c.numSlots
	circle = (step - 1) / c.numSlots

	return
}

func (c *Cache) Get(key string) (any, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	select {
	case <-c.stopChannel:
		return nil, false, ErrClosed
	default:
		if elem, exists := c.items[key]; exists {
			item := elem.Value.(cacheItem)

			return item.value, true, nil
		}
		return nil, false, nil
	}
}

func (c *Cache) Set(key string, value any, expiration time.Duration) error {
	if expiration <= 0 {
		return ErrArg
	}

	select {
	case c.setChannel <- cacheItem{
		key:        key,
		value:      value,
		expiration: expiration,
	}:
		return nil
	case <-c.stopChannel:
		return ErrClosed
	}
}

func (c *Cache) Stop() {
	close(c.stopChannel)
}
