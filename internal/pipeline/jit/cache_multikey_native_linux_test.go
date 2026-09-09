// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
)

// Supplemental deterministic coverage of the multiple-items-in-slot behavior
// in plcache/cache_test.go, including a selective refresh and native batches.
func TestNativeCacheSameSlotSelectiveRefresh(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	rt.(*nativeRuntime).manualCacheTicks = true
	for _, count := range []int{2, 10, 128} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			compiled, err := rt.Compile(`if seed { stored=(cache_set(key,message,ttl) == nil) }; add_key(result,cache_get(key))`, "pipeline-go-1.4.3-datakit")
			if err != nil {
				t.Fatal(err)
			}
			p := compiled.(*nativeProgram)
			defer p.Close()
			p.static = nil
			if err := p.enableCacheTicks(uint64(time.Second)); err != nil {
				t.Fatal(err)
			}
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
					t.Fatal("Go tick worker stalled")
				}
				if err := cache.Set("barrier", true, 100*time.Second); err != nil {
					t.Fatal(err)
				}
				if err := p.advanceCacheTick(); err != nil {
					t.Fatal(err)
				}
			}
			run := func(tick int, seedAll, refresh bool) {
				t.Helper()
				points := make([]Point, count)
				wants := make([]any, count)
				for i := range points {
					key := fmt.Sprintf("key-%d", i)
					seed, value, ttl := seedAll || (refresh && i == 0), "initial", int64(2)
					if refresh && i == 0 {
						value, ttl = "refreshed", 4
					}
					if seed {
						if err := cache.Set(key, value, time.Duration(ttl)*time.Second); err != nil {
							t.Fatal(err)
						}
					}
					got, exists, err := cache.Get(key)
					if err != nil {
						t.Fatal(err)
					}
					var expected any
					if tick < 2 {
						expected = "initial"
					}
					if i == 0 && tick >= 1 && tick < 5 {
						expected = "refreshed"
					}
					if !exists {
						got = nil
					}
					if got != expected {
						t.Fatalf("Go tick=%d key=%s got=%v expected=%v", tick, key, got, expected)
					}
					wants[i] = expected
					points[i] = Point{Version: 1, Category: "logging", Measurement: "multi-key", Fields: map[string]any{"key": key, "seed": seed, "message": value, "ttl": ttl}}
				}
				input, err := EncodeFlatPoints(points)
				if err != nil {
					t.Fatal(err)
				}
				batch, err := p.ProcessIndexed(input)
				if err != nil {
					t.Fatal(err)
				}
				if len(batch.Records) != count {
					t.Fatal("record count mismatch")
				}
				for i, record := range batch.Records {
					if record.Status != TerminalOK {
						t.Fatalf("tick=%d record=%d: %+v", tick, i, record)
					}
					deltas, err := record.MutationDeltas()
					if err != nil {
						t.Fatal(err)
					}
					for _, delta := range deltas {
						if err := delta.Apply(&points[i]); err != nil {
							t.Fatal(err)
						}
					}
					if got := points[i].Fields["result"]; got != wants[i] {
						t.Fatalf("native tick=%d record=%d got=%v expected=%v", tick, i, got, wants[i])
					}
				}
			}
			run(0, true, false)
			for tick := 1; tick <= 7; tick++ {
				advance()
				run(tick, false, tick == 1)
			}
		})
	}
}
