// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"sync"
	"time"
)

type cacheTickRegistration struct {
	interval time.Duration
	next     time.Time
	tick     func()
}

// Returns one delivery at most, even when multiple intervals elapsed. Keeping
// the phase calculation separate makes delayed scheduling testable without
// sleeps or assumptions about the host scheduler.
func (entry *cacheTickRegistration) takeDue(now time.Time) bool {
	if entry.next.After(now) {
		return false
	}
	entry.next = now.Add(entry.interval - now.Sub(entry.next)%entry.interval)
	return true
}

// One worker owns all timer dispatch for a runtime. Registrations preserve
// their own phase; a delayed dispatch does not replay every missed tick.
// Callbacks must finish before Close returns and must not call Close themselves.
type cacheTickDriver struct {
	mu      sync.Mutex
	entries map[uint64]cacheTickRegistration
	nextID  uint64
	closed  bool
	wake    chan struct{}
	stop    chan struct{}
	done    chan struct{}
}

func newCacheTickDriver() *cacheTickDriver {
	d := &cacheTickDriver{entries: make(map[uint64]cacheTickRegistration), wake: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{})}
	go d.run()
	return d
}

func (d *cacheTickDriver) register(interval time.Duration, tick func()) (uint64, bool) {
	if interval <= 0 || tick == nil {
		return 0, false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.nextID == ^uint64(0) {
		return 0, false
	}
	d.nextID++
	d.entries[d.nextID] = cacheTickRegistration{interval: interval, next: time.Now().Add(interval), tick: tick}
	select {
	case d.wake <- struct{}{}:
	default:
	}
	return d.nextID, true
}

// A callback already selected for dispatch may still run. The program's own
// lease/closed guard must reject stale calls before touching the native handle.
func (d *cacheTickDriver) unregister(id uint64) {
	d.mu.Lock()
	delete(d.entries, id)
	d.mu.Unlock()
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

func (d *cacheTickDriver) Close() {
	d.mu.Lock()
	if !d.closed {
		d.closed = true
		close(d.stop)
		clear(d.entries)
	}
	d.mu.Unlock()
	<-d.done
}

func (d *cacheTickDriver) run() {
	defer close(d.done)
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		d.mu.Lock()
		if d.closed {
			d.mu.Unlock()
			return
		}
		now := time.Now()
		wait := time.Hour
		var ready []func()
		for id, entry := range d.entries {
			if entry.takeDue(now) {
				ready = append(ready, entry.tick)
				d.entries[id] = entry
			}
			if delay := entry.next.Sub(now); delay < wait {
				wait = delay
			}
		}
		d.mu.Unlock()
		nextWake := now.Add(wait)
		for _, tick := range ready {
			select {
			case <-d.stop:
				return
			default:
				tick()
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(time.Until(nextWake))
		select {
		case <-d.stop:
			return
		case <-d.wake:
		case <-timer.C:
		}
	}
}
