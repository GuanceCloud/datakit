// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package pipeline

import (
	"errors"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
)

func TestGoCacheReplacementAndStopOracle(t *testing.T) {
	for _, tc := range []struct {
		name               string
		oldTicks, newTicks int
	}{
		{"extend-across-wheel", 2, 5},
		{"shorten-across-wheel", 5, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ticks := make(chan time.Time)
			cache, err := plcache.NewCacheWithTicker(time.Second, 3, &time.Ticker{C: ticks})
			if err != nil {
				t.Fatal(err)
			}
			defer cache.Stop()
			advance := func() {
				t.Helper()
				select {
				case ticks <- time.Time{}:
				case <-time.After(2 * time.Second):
					t.Fatal("tick worker stalled")
				}
				if err := cache.Set("barrier", true, 100*time.Second); err != nil {
					t.Fatal(err)
				}
			}
			if err := cache.Set("value", "old", time.Duration(tc.oldTicks)*time.Second); err != nil {
				t.Fatal(err)
			}
			advance()
			if err := cache.Set("value", "new", time.Duration(tc.newTicks)*time.Second); err != nil {
				t.Fatal(err)
			}
			// Invalid replacement must not remove or reschedule the existing value.
			if err := cache.Set("value", "invalid", 0); !errors.Is(err, plcache.ErrArg) {
				t.Fatalf("invalid TTL: %v", err)
			}
			for tick := 0; tick <= tc.newTicks+tc.oldTicks; tick++ {
				value, exists, err := cache.Get("value")
				if err != nil {
					t.Fatal(err)
				}
				if exists != (tick < tc.newTicks) || (exists && value != "new") {
					t.Fatalf("tick=%d got=%v exists=%v", tick, value, exists)
				}
				advance()
			}
			cache.Stop()
			cache.Stop()
			if _, exists, err := cache.Get("barrier"); exists || !errors.Is(err, plcache.ErrClosed) {
				t.Fatalf("read after stop: exists=%v err=%v", exists, err)
			}
			if err := cache.Set("value", "after-close", time.Second); !errors.Is(err, plcache.ErrClosed) {
				t.Fatalf("write after stop: %v", err)
			}
		})
	}
}

// Characterize the actual upstream time wheel independently of wall-clock
// scheduling. This is not a Rust compatibility test.
func TestGoCacheTickExpirationOracle(t *testing.T) {
	for _, tc := range []struct {
		name  string
		ttl   time.Duration
		ticks int
	}{
		{"sub-interval-rounded-up", time.Millisecond, 1},
		{"one-tick", time.Second, 1},
		{"fraction-truncated", 1500 * time.Millisecond, 1},
		{"full-wheel", 3 * time.Second, 3},
		{"wheel-plus-one", 4 * time.Second, 4},
		{"two-wheels", 6 * time.Second, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ticks := make(chan time.Time)
			cache, err := plcache.NewCacheWithTicker(time.Second, 3, &time.Ticker{C: ticks})
			if err != nil {
				t.Fatal(err)
			}
			defer cache.Stop()
			if err := cache.Set("value", "stored", tc.ttl); err != nil {
				t.Fatal(err)
			}
			for tick := 0; tick <= tc.ticks; tick++ {
				if tick > 0 {
					select {
					case ticks <- time.Unix(0, 0):
					case <-time.After(2 * time.Second):
						t.Fatal("cache tick worker did not receive tick")
					}
					// Set waits for application in the same worker: the preceding
					// unbuffered tick has finished before this barrier completes.
					if err := cache.Set("barrier", true, 100*time.Second); err != nil {
						t.Fatal(err)
					}
				}
				value, exists, err := cache.Get("value")
				if err != nil {
					t.Fatal(err)
				}
				wantExists := tick < tc.ticks
				if exists != wantExists || (exists && value != "stored") {
					t.Fatalf("tick=%d value=%v exists=%v wantExists=%v", tick, value, exists, wantExists)
				}
			}
		})
	}
}
