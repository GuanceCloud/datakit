// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// Copyright 2026-present Guance, Inc.

package jit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCancellationPoolReusesOnlyUncancelledTokens(t *testing.T) {
	var created, destroyed, requested atomic.Uint64
	n := &nativeRuntime{
		cancelCreate:  func(out *uintptr) int32 { *out = uintptr(created.Add(1)); return 0 },
		cancelDestroy: func(uintptr) int32 { destroyed.Add(1); return 0 },
		cancelRequest: func(uintptr) int32 { requested.Add(1); return 0 },
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < 100; i++ {
		token, release, err := n.watchCancellation(ctx)
		if err != nil || token != 1 {
			t.Fatal("healthy token was not reused", token, err)
		}
		release()
	}
	if created.Load() != 1 || destroyed.Load() != 0 {
		t.Fatal("unbounded healthy token churn")
	}
	token, release, err := n.watchCancellation(ctx)
	if err != nil {
		t.Fatal(err)
	}
	done := n.cancelPool // Pool must be empty while its single lease is in flight.
	if len(done) != 0 {
		t.Fatal("active token still pooled")
	}
	cancel()
	// Wait on an independent AfterFunc is insufficient to establish that our
	// callback started. The fake request is the observable cancellation boundary.
	deadline := time.Now().Add(time.Second)
	for requested.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("cancel callback stalled")
		}
		time.Sleep(time.Millisecond)
	}
	release()
	if destroyed.Load() != 1 {
		t.Fatal("cancelled token recycled")
	}
	next, stop := context.WithCancel(context.Background())
	defer stop()
	newToken, finish, err := n.watchCancellation(next)
	if err != nil || newToken == token {
		t.Fatal("late cancellation can poison next call")
	}
	finish()
	if err := n.closeCancellationPool(); err != nil {
		t.Fatal(err)
	}
	if created.Load() != destroyed.Load() {
		t.Fatal("token leak")
	}
}

func TestCancellationPoolBoundsIdleNativeResources(t *testing.T) {
	var created, destroyed atomic.Uint64
	n := &nativeRuntime{
		cancelCreate:  func(out *uintptr) int32 { *out = uintptr(created.Add(1)); return 0 },
		cancelDestroy: func(uintptr) int32 { destroyed.Add(1); return 0 },
		cancelRequest: func(uintptr) int32 { return 0 },
	}
	var releases []func()
	for i := 0; i < maxIdleCancellations*2; i++ {
		_, release, err := n.watchCancellation(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		releases = append(releases, release)
	}
	for _, release := range releases {
		release()
	}
	if len(n.cancelPool) != maxIdleCancellations {
		t.Fatal("unbounded idle pool")
	}
	if err := n.closeCancellationPool(); err != nil {
		t.Fatal(err)
	}
	if created.Load() != destroyed.Load() {
		t.Fatal("native pool not drained")
	}
}
