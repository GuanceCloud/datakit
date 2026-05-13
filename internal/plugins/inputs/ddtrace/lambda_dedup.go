// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package ddtrace

import "sync"

type lambdaSpanKey struct {
	traceID uint64
	spanID  uint64
}

type lambdaSpanDeduper struct {
	mu       sync.Mutex
	capacity int
	queue    []lambdaSpanKey
	seen     map[lambdaSpanKey]struct{}
}

func newLambdaSpanDeduper(capacity int) *lambdaSpanDeduper {
	if capacity <= 0 {
		capacity = 1024
	}
	return &lambdaSpanDeduper{
		capacity: capacity,
		queue:    make([]lambdaSpanKey, 0, capacity),
		seen:     make(map[lambdaSpanKey]struct{}, capacity),
	}
}

func (d *lambdaSpanDeduper) ShouldKeep(traceID, spanID uint64) bool {
	if d == nil {
		return true
	}

	key := lambdaSpanKey{traceID: traceID, spanID: spanID}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.seen[key]; ok {
		return false
	}

	if len(d.queue) == d.capacity {
		oldest := d.queue[0]
		delete(d.seen, oldest)
		copy(d.queue, d.queue[1:])
		d.queue = d.queue[:len(d.queue)-1]
	}

	d.queue = append(d.queue, key)
	d.seen[key] = struct{}{}

	return true
}
