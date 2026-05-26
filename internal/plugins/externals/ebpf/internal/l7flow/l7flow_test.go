//go:build linux
// +build linux

package l7flow

import (
	"testing"
	"time"
)

func TestPerfLostWarningLimiterAggregatesWithinWindow(t *testing.T) {
	now := time.Unix(1700000000, 0)
	limiter := newPerfLostWarningLimiter(func() time.Time { return now })

	msg := limiter.format(3, 10)
	if msg != "lost 10 events on cpu 3" {
		t.Fatalf("unexpected first warning: %q", msg)
	}

	if msg = limiter.format(5, 20); msg != "" {
		t.Fatalf("expected warning suppression within window, got %q", msg)
	}

	now = now.Add(apiflowPerfLostLogInterval + time.Second)
	msg = limiter.format(7, 30)
	want := "lost 30 events on cpu 7 (aggregated 20 additional lost events over 10s)"
	if msg != want {
		t.Fatalf("unexpected aggregated warning: %q != %q", msg, want)
	}
}

func TestPerfLostWarningLimiterSkipsZeroCount(t *testing.T) {
	limiter := newPerfLostWarningLimiter(time.Now)
	if msg := limiter.format(1, 0); msg != "" {
		t.Fatalf("expected zero-count updates to stay silent, got %q", msg)
	}
}
