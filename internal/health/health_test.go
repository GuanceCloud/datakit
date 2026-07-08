// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package health

import (
	"testing"
	"time"
)

func TestRegistryLiveness(t *testing.T) {
	now := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	registry := NewRegistry()
	registry.now = func() time.Time { return now }
	reporter := registry.Register("container-runtime", "containerd.sock", Options{
		FailureTimeout: 10 * time.Minute,
		SilenceTimeout: time.Hour,
	})

	assertLive(t, registry.Snapshot(), true)
	reporter.Failure()
	now = now.Add(11 * time.Minute)
	assertLive(t, registry.Snapshot(), false)
	reporter.Alive()
	assertLive(t, registry.Snapshot(), true)

	reporter.Failure()
	now = now.Add(11 * time.Minute)
	reporter.Close()
	assertLive(t, registry.Snapshot(), true)
}

func TestSilenceFailsLiveness(t *testing.T) {
	now := time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)
	registry := NewRegistry()
	registry.now = func() time.Time { return now }
	registry.Register("collector", "", Options{SilenceTimeout: 10 * time.Minute})

	now = now.Add(11 * time.Minute)
	assertLive(t, registry.Snapshot(), false)
}

func assertLive(t *testing.T, snapshot Snapshot, want bool) {
	t.Helper()
	if snapshot.Live != want {
		t.Fatalf("live: got %t, want %t", snapshot.Live, want)
	}
	if want && len(snapshot.FailedComponents) != 0 {
		t.Fatalf("healthy snapshot has %d failed components", len(snapshot.FailedComponents))
	}
	if !want && len(snapshot.FailedComponents) == 0 {
		t.Fatal("unhealthy snapshot has no failed component")
	}
	if snapshot.CheckedAt.IsZero() {
		t.Fatal("checked_at is empty")
	}
}
