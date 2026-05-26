//go:build linux
// +build linux

package procwatch

import (
	"context"
	"reflect"
	"regexp"
	"sync/atomic"
	"testing"
	"time"
)

func TestLibraryTrackerScanDetectsChanges(t *testing.T) {
	var attached []string
	var detached []string

	tracker, err := NewLibraryTracker([]HookRule{
		{
			Re: regexp.MustCompile(`\.so$`),
			Attach: func(path string) error {
				attached = append(attached, path)
				return nil
			},
			Detach: func(path string) error {
				detached = append(detached, path)
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}

	current := map[string]struct{}{
		"/tmp/libssl.so": {},
	}
	tracker.scanFn = func(*regexp.Regexp) map[string]struct{} {
		out := make(map[string]struct{}, len(current))
		for path := range current {
			out[path] = struct{}{}
		}
		return out
	}

	if !tracker.Scan() {
		t.Fatal("expected first scan to report changes")
	}
	if got := attached; !reflect.DeepEqual(got, []string{"/tmp/libssl.so"}) {
		t.Fatalf("attached = %v, want [/tmp/libssl.so]", got)
	}

	attached = nil
	if tracker.Scan() {
		t.Fatal("expected unchanged scan to report no changes")
	}
	if len(attached) != 0 || len(detached) != 0 {
		t.Fatalf("unexpected callbacks on unchanged scan: attached=%v detached=%v", attached, detached)
	}

	current = map[string]struct{}{
		"/tmp/libcrypto.so": {},
	}
	if !tracker.Scan() {
		t.Fatal("expected changed library set to report changes")
	}
	if got := attached; !reflect.DeepEqual(got, []string{"/tmp/libcrypto.so"}) {
		t.Fatalf("attached = %v, want [/tmp/libcrypto.so]", got)
	}
	if got := detached; !reflect.DeepEqual(got, []string{"/tmp/libssl.so"}) {
		t.Fatalf("detached = %v, want [/tmp/libssl.so]", got)
	}
}

func TestNextLibraryScanIntervalBackoff(t *testing.T) {
	if got := nextLibraryScanInterval(true, time.Minute, defaultLibraryScanInterval, maxLibraryScanInterval); got != defaultLibraryScanInterval {
		t.Fatalf("changed interval = %s, want %s", got, defaultLibraryScanInterval)
	}

	if got := nextLibraryScanInterval(false, defaultLibraryScanInterval, defaultLibraryScanInterval, maxLibraryScanInterval); got != defaultLibraryScanInterval*2 {
		t.Fatalf("backoff interval = %s, want %s", got, defaultLibraryScanInterval*2)
	}

	if got := nextLibraryScanInterval(false, maxLibraryScanInterval, defaultLibraryScanInterval, maxLibraryScanInterval); got != maxLibraryScanInterval {
		t.Fatalf("capped interval = %s, want %s", got, maxLibraryScanInterval)
	}
}

func TestLibraryTrackerRunDelaysInitialScan(t *testing.T) {
	tracker, err := NewLibraryTracker([]HookRule{
		{
			Re:     regexp.MustCompile(`\.so$`),
			Attach: func(string) error { return nil },
			Detach: func(string) error { return nil },
		},
	})
	if err != nil {
		t.Fatalf("new tracker: %v", err)
	}

	var scans atomic.Int32
	tracker.scanFn = func(*regexp.Regexp) map[string]struct{} {
		scans.Add(1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracker.Run(ctx, time.Second)
	time.Sleep(50 * time.Millisecond)

	if got := scans.Load(); got != 0 {
		t.Fatalf("expected no immediate scan, got %d", got)
	}
}
