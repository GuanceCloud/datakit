// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// LockType represents different types of locks in diskcache.
type LockType string

const (
	LockTypeWrite LockType = "write"
	LockTypeRead  LockType = "read"
	LockTypeRW    LockType = "rw"
)

// InstrumentedMutex wraps sync.Mutex with contention tracking.
type InstrumentedMutex struct {
	mu           sync.Mutex
	lockType     LockType
	path         string
	lockWaitTime *prometheus.HistogramVec
	contention   *prometheus.CounterVec
}

// NewInstrumentedMutex creates a new instrumented mutex.
func NewInstrumentedMutex(lockType LockType,
	path string,
	lockWaitTime *prometheus.HistogramVec,
	contention *prometheus.CounterVec,
) *InstrumentedMutex {
	return &InstrumentedMutex{
		lockType:     lockType,
		path:         path,
		lockWaitTime: lockWaitTime,
		contention:   contention,
	}
}

// Lock acquires the mutex with contention tracking.
func (im *InstrumentedMutex) Lock() {
	start := time.Now()

	// Check if mutex is already locked (contention)
	if im.mu.TryLock() {
		// No contention - immediate acquisition
		im.observeLockTime(start, false)
		return
	}

	// Contention occurred - wait for lock
	im.observeContention()
	im.mu.Lock()
	im.observeLockTime(start, true)
}

// TryLock attempts to acquire mutex without blocking.
func (im *InstrumentedMutex) TryLock() bool {
	start := time.Now() // nolint:ifshort
	acquired := im.mu.TryLock()
	if acquired {
		im.observeLockTime(start, false)
	} else {
		im.observeContention()
	}
	return acquired
}

// Unlock releases the mutex.
func (im *InstrumentedMutex) Unlock() {
	im.mu.Unlock()
}

// observeLockTime records the total time to acquire the lock.
func (im *InstrumentedMutex) observeLockTime(start time.Time, hadContention bool) {
	duration := time.Since(start).Seconds()

	// Record wait time
	im.lockWaitTime.WithLabelValues(string(im.lockType), im.path).Observe(duration)

	// If this was a contention scenario, also record it
	if hadContention {
		im.contention.WithLabelValues(string(im.lockType), im.path).Inc()
	}
}

// observeContention records a contention event.
func (im *InstrumentedMutex) observeContention() {
	im.contention.WithLabelValues(string(im.lockType), im.path).Inc()
}

// InstrumentedRWMutex wraps sync.RWMutex with contention tracking.
type InstrumentedRWMutex struct {
	mu           sync.RWMutex
	path         string
	lockWaitTime *prometheus.HistogramVec
	contention   *prometheus.CounterVec
}

// NewInstrumentedRWMutex creates a new instrumented RWMutex.
func NewInstrumentedRWMutex(path string, lockWaitTime *prometheus.HistogramVec, contention *prometheus.CounterVec) *InstrumentedRWMutex {
	return &InstrumentedRWMutex{
		path:         path,
		lockWaitTime: lockWaitTime,
		contention:   contention,
	}
}

// RLock acquires read lock with contention tracking.
func (irm *InstrumentedRWMutex) RLock() {
	start := time.Now()

	if irm.mu.TryRLock() {
		// No contention
		irm.lockWaitTime.WithLabelValues(string(LockTypeRead), irm.path).Observe(time.Since(start).Seconds())
		return
	}

	// Contention occurred
	irm.contention.WithLabelValues(string(LockTypeRead), irm.path).Inc()
	irm.mu.RLock()
	irm.lockWaitTime.WithLabelValues(string(LockTypeRead), irm.path).Observe(time.Since(start).Seconds())
}

// TryRLock attempts to acquire read lock without blocking.
func (irm *InstrumentedRWMutex) TryRLock() bool {
	start := time.Now()
	acquired := irm.mu.TryRLock()
	if acquired {
		irm.lockWaitTime.WithLabelValues(string(LockTypeRead), irm.path).Observe(time.Since(start).Seconds())
	} else {
		irm.contention.WithLabelValues(string(LockTypeRead), irm.path).Inc()
	}
	return acquired
}

// RUnlock releases read lock.
func (irm *InstrumentedRWMutex) RUnlock() {
	irm.mu.RUnlock()
}

// Lock acquires write lock with contention tracking.
func (irm *InstrumentedRWMutex) Lock() {
	start := time.Now()

	if irm.mu.TryLock() {
		// No contention
		irm.lockWaitTime.WithLabelValues(string(LockTypeWrite), irm.path).Observe(time.Since(start).Seconds())
		return
	}

	// Contention occurred
	irm.contention.WithLabelValues(string(LockTypeWrite), irm.path).Inc()
	irm.mu.Lock()
	irm.lockWaitTime.WithLabelValues(string(LockTypeWrite), irm.path).Observe(time.Since(start).Seconds())
}

// TryLock attempts to acquire write lock without blocking.
func (irm *InstrumentedRWMutex) TryLock() bool {
	start := time.Now()
	acquired := irm.mu.TryLock()
	if acquired {
		irm.lockWaitTime.WithLabelValues(string(LockTypeWrite), irm.path).Observe(time.Since(start).Seconds())
	} else {
		irm.contention.WithLabelValues(string(LockTypeWrite), irm.path).Inc()
	}
	return acquired
}

// Unlock releases write lock.
func (irm *InstrumentedRWMutex) Unlock() {
	irm.mu.Unlock()
}
