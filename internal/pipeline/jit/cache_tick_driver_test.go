// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"testing"
	"time"
)

func TestCacheTickPhaseAndDelayedDelivery(t *testing.T) {
	base := time.Unix(100, 0)
	entry := cacheTickRegistration{interval: time.Second, next: base.Add(time.Second)}
	for _, tc := range []struct {
		elapsed time.Duration
		due     bool
		next    time.Duration
	}{
		{999 * time.Millisecond, false, time.Second},
		{time.Second, true, 2 * time.Second},
		{time.Second, false, 2 * time.Second},
		{5500 * time.Millisecond, true, 6 * time.Second},
		{5500 * time.Millisecond, false, 6 * time.Second},
		{6 * time.Second, true, 7 * time.Second},
	} {
		if got := entry.takeDue(base.Add(tc.elapsed)); got != tc.due {
			t.Fatalf("elapsed=%v due=%v want=%v", tc.elapsed, got, tc.due)
		}
		if want := base.Add(tc.next); !entry.next.Equal(want) {
			t.Fatalf("next=%v want=%v", entry.next, want)
		}
	}
	// Independent registration phase survives the same delayed dispatch.
	other := cacheTickRegistration{interval: time.Second, next: base.Add(1250 * time.Millisecond)}
	if !other.takeDue(base.Add(5500*time.Millisecond)) || !other.next.Equal(base.Add(6250*time.Millisecond)) {
		t.Fatal("registration phase lost")
	}
}

func TestCacheTickDriverSharesWorkerAndDrainsClose(t *testing.T) {
	d := newCacheTickDriver()
	defer d.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	if _, ok := d.register(time.Millisecond, func() { close(entered); <-release }); !ok {
		t.Fatal("register")
	}
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not start")
	}
	closed := make(chan struct{})
	go func() { d.Close(); close(closed) }()
	// Wait until Close has marked the driver closed without waiting for the
	// blocked callback. Registration refusal establishes that ordering.
	deadline := time.After(2 * time.Second)
	for {
		id, ok := d.register(time.Hour, func() {})
		if !ok {
			break
		}
		d.unregister(id)
		select {
		case <-deadline:
			t.Fatal("Close did not start")
		default:
		}
	}
	select {
	case <-closed:
		t.Fatal("Close returned with live callback")
	default:
	}
	// Release via a separate channel send; deferred close handles early failure.
	release <- struct{}{}
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not drain")
	}
	if _, ok := d.register(time.Second, func() {}); ok {
		t.Fatal("registered after close")
	}
}

func TestCacheTickDriverMultipleRegistrations(t *testing.T) {
	d := newCacheTickDriver()
	defer d.Close()
	seen := make(chan int, 32)
	for i := 0; i < 3; i++ {
		i := i
		if _, ok := d.register(time.Duration(i+1)*time.Millisecond, func() {
			select {
			case seen <- i:
			default:
			}
		}); !ok {
			t.Fatal("register")
		}
	}
	got := map[int]bool{}
	deadline := time.After(2 * time.Second)
	for len(got) < 3 {
		select {
		case i := <-seen:
			got[i] = true
		case <-deadline:
			t.Fatal("missing callback")
		}
	}
	d.Close()
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.entries) != 0 {
		t.Fatal("closed driver retains registrations")
	}
}
