// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
)

func TestNativeCacheTickExpirationAndClose(t *testing.T) {
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
	compiled, err := rt.Compile(`if seed { cache_set("key","stored",2) }; add_key(result,cache_get("key"))`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	p := compiled.(*nativeProgram)
	defer p.Close()
	// Exercise indexed delta transport so the result is independently decoded.
	p.static = nil
	if err := p.advanceCacheTick(); err == nil {
		t.Fatal("tick before enable accepted")
	}
	if err := p.enableCacheTicks(0); err == nil {
		t.Fatal("zero interval accepted")
	}
	if err := p.enableCacheTicks(1_000_000_000); err != nil {
		t.Fatal(err)
	}
	if err := p.enableCacheTicks(1_000_000_000); err == nil {
		t.Fatal("duplicate enable accepted")
	}
	run := func(seed bool, want any) {
		t.Helper()
		point := Point{Version: 1, Category: "logging", Measurement: "clock", Fields: map[string]any{"seed": seed}}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := p.ProcessIndexed(input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("bad batch: %+v", batch)
		}
		deltas, err := batch.Records[0].MutationDeltas()
		if err != nil {
			t.Fatal(err)
		}
		for _, delta := range deltas {
			if err := delta.Apply(&point); err != nil {
				t.Fatal(err)
			}
		}
		if point.Fields["result"] != want {
			t.Fatalf("result=%v want=%v", point.Fields["result"], want)
		}
	}
	run(true, "stored")
	if err := p.advanceCacheTick(); err != nil {
		t.Fatal(err)
	}
	run(false, "stored")
	if err := p.advanceCacheTick(); err != nil {
		t.Fatal(err)
	}
	run(false, nil)
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.advanceCacheTick(); err == nil {
		t.Fatal("tick after close accepted")
	}
	if err := p.enableCacheTicks(1); err == nil {
		t.Fatal("enable after close accepted")
	}
}

func TestNativeCacheConcurrentTickAndWrites(t *testing.T) {
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
	compiled, err := rt.Compile(`if seed { cache_set(message,"value",2) }; add_key(result,cache_get(message))`, "pipeline-go-1.4.3-datakit")
	if err != nil {
		t.Fatal(err)
	}
	p := compiled.(*nativeProgram)
	defer p.Close()
	p.static = nil
	if err := p.enableCacheTicks(uint64(time.Second)); err != nil {
		t.Fatal(err)
	}
	process := func(worker int, seed bool, want *string) error {
		points := make([]Point, 8)
		for i := range points {
			points[i] = Point{Version: 1, Category: "logging", Measurement: "concurrent", Fields: map[string]any{"seed": seed, "message": fmt.Sprintf("%d/%d", worker, i)}}
		}
		input, err := EncodeFlatPoints(points)
		if err != nil {
			return err
		}
		batch, err := p.ProcessIndexed(input)
		if err != nil {
			return err
		}
		if len(batch.Records) != len(points) {
			return fmt.Errorf("record count %d", len(batch.Records))
		}
		for i, record := range batch.Records {
			if record.Status != TerminalOK {
				return fmt.Errorf("terminal: %+v", record)
			}
			deltas, err := record.MutationDeltas()
			if err != nil {
				return err
			}
			for _, delta := range deltas {
				if err := delta.Apply(&points[i]); err != nil {
					return err
				}
			}
			value := points[i].Fields["result"]
			if want != nil && value != *want {
				return fmt.Errorf("key %d/%d: got %v want %s", worker, i, value, *want)
			}
			if !seed && want == nil && value != nil {
				return fmt.Errorf("key %d/%d retained %v", worker, i, value)
			}
		}
		return nil
	}
	start := make(chan struct{})
	errors := make(chan error, 5)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for round := 0; round < 30; round++ {
				if err := process(worker, true, nil); err != nil {
					errors <- err
					return
				}
			}
		}(worker)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 300; i++ {
			if err := p.advanceCacheTick(); err != nil {
				errors <- err
				return
			}
		}
	}()
	close(start)
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
	if t.Failed() {
		return
	}
	// With all writers/ticks joined, seed and read every key independently of
	// race ordering, then expire and verify all 32 keys rather than one probe.
	value := "value"
	for worker := 0; worker < 4; worker++ {
		if err := process(worker, true, &value); err != nil {
			t.Fatal(err)
		}
	}
	for worker := 0; worker < 4; worker++ {
		if err := process(worker, false, &value); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := p.advanceCacheTick(); err != nil {
			t.Fatal(err)
		}
	}
	for worker := 0; worker < 4; worker++ {
		if err := process(worker, false, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNativeCacheTickMatchesGoWheel(t *testing.T) {
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
	for _, tc := range []struct {
		name           string
		oldTTL, newTTL int64
		nested         bool
	}{
		{"one-tick", 1, 2, false},
		{"full-wheel", 3, 4, false},
		{"wheel-plus-one", 4, 5, false},
		{"two-wheels", 6, 7, false},
		{"shorten-across-wheel", 5, 1, false},
		{"nested-extend-across-wheel", 2, 5, true},
		{"nested-shorten-across-wheel", 5, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := `if seed { cache_set("key",message,ttl) }; add_key(result,cache_get("key"))`
			if tc.nested {
				source = `if seed { stored=(cache_set("key",message,ttl) == nil) }; add_key(result,cache_get("key"))`
			}
			compiled, err := rt.Compile(source, "pipeline-go-1.4.3-datakit")
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
			run := func(seed bool, message string, expiry int64, expected any) {
				t.Helper()
				if seed && expiry > 0 {
					if err := cache.Set("key", message, time.Duration(expiry)*time.Second); err != nil {
						t.Fatal(err)
					}
				}
				want, exists, err := cache.Get("key")
				if err != nil {
					t.Fatal(err)
				}
				if !exists {
					want = nil
				}
				if want != expected {
					t.Fatalf("Go cache timeline: got=%v want=%v", want, expected)
				}
				point := Point{Version: 1, Category: "logging", Measurement: "paired-clock", Fields: map[string]any{"seed": seed, "message": message, "ttl": expiry}}
				input, err := EncodeFlatPoints([]Point{point})
				if err != nil {
					t.Fatal(err)
				}
				batch, err := p.ProcessIndexed(input)
				if err != nil {
					t.Fatal(err)
				}
				if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
					t.Fatalf("bad batch: %+v", batch)
				}
				deltas, err := batch.Records[0].MutationDeltas()
				if err != nil {
					t.Fatal(err)
				}
				for _, delta := range deltas {
					if err := delta.Apply(&point); err != nil {
						t.Fatal(err)
					}
				}
				if point.Fields["result"] != want {
					t.Fatalf("Rust=%v Go=%v", point.Fields["result"], want)
				}
			}
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
			run(true, "initial", tc.oldTTL, "initial")
			advance()
			var initial any = "initial"
			if tc.oldTTL == 1 {
				initial = nil
			}
			run(false, "", 0, initial)
			run(true, "replacement", tc.newTTL, "replacement")
			run(true, "invalid", 0, "replacement")
			// Continue past both old and replacement deadlines: shortening must
			// expire promptly; extending must survive the stale old deadline.
			for tick := int64(1); tick <= max(tc.oldTTL, tc.newTTL)+1; tick++ {
				advance()
				var expected any
				if tick < tc.newTTL {
					expected = "replacement"
				}
				run(false, "", 0, expected)
			}
		})
	}
}

func TestNativeAutomaticCacheDriver(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	rt, err := openRuntime(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close()
	n := rt.(*nativeRuntime)
	var programs []*nativeProgram
	for i := 0; i < 3; i++ {
		compiled, err := rt.Compile(`if seed { cache_set("key","stored",1) }; add_key(result,cache_get("key"))`, "pipeline-go-1.4.3-datakit")
		if err != nil {
			t.Fatal(err)
		}
		p := compiled.(*nativeProgram)
		defer p.Close()
		p.static = nil
		programs = append(programs, p)
		if p.clockRegistration == 0 {
			t.Fatal("no automatic clock registration")
		}
	}
	read := func(p *nativeProgram, seed bool) any {
		t.Helper()
		point := Point{Version: 1, Category: "logging", Measurement: "auto", Fields: map[string]any{"seed": seed}}
		input, err := EncodeFlatPoints([]Point{point})
		if err != nil {
			t.Fatal(err)
		}
		batch, err := p.ProcessIndexed(input)
		if err != nil {
			t.Fatal(err)
		}
		if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
			t.Fatalf("bad batch: %+v", batch)
		}
		deltas, err := batch.Records[0].MutationDeltas()
		if err != nil {
			t.Fatal(err)
		}
		for _, delta := range deltas {
			if err := delta.Apply(&point); err != nil {
				t.Fatal(err)
			}
		}
		return point.Fields["result"]
	}
	for _, p := range programs {
		if read(p, true) != "stored" {
			t.Fatal("seed not stored")
		}
	}
	poll := time.NewTicker(20 * time.Millisecond)
	defer poll.Stop()
	deadline := time.After(4 * time.Second)
	for {
		allExpired := true
		for _, p := range programs {
			if read(p, false) != nil {
				allExpired = false
			}
		}
		if allExpired {
			break
		}
		select {
		case <-poll.C:
		case <-deadline:
			t.Fatal("automatic ticks did not expire caches")
		}
	}
	for _, p := range programs {
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
	}
	n.clockDriver.mu.Lock()
	remaining := len(n.clockDriver.entries)
	n.clockDriver.mu.Unlock()
	if remaining != 0 {
		t.Fatal("closed programs retained by driver")
	}
	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-n.clockDriver.done:
	default:
		t.Fatal("library closed before driver exit")
	}
}

func TestScheduledCacheErrorStopsExecution(t *testing.T) {
	p := &nativeProgram{handle: 1, runtime: &nativeRuntime{cacheTick: func(uintptr) int32 { return 2 }}}
	// A running ProcessIndexed holds this read lease. Tick must still execute,
	// otherwise a long batch delays expiration for every registered script.
	p.mu.RLock()
	done := make(chan struct{})
	go func() { p.scheduledCacheTick(); close(done) }()
	select {
	case <-done:
		p.mu.RUnlock()
	case <-time.After(2 * time.Second):
		p.mu.RUnlock()
		t.Fatal("tick blocked by process read lease")
	}
	if p.cacheClockError() == nil {
		t.Fatal("tick error lost")
	}
	if _, err := p.ProcessIndexed(nil); err == nil {
		t.Fatal("execution continued after tick failure")
	}
}

func TestCacheDriverRegistrationFollowsGenerationLease(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	runner, err := NewRunner(path, "pipeline-go-1.4.3-datakit", 4)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Close()
	const source = `cache_set("key","value"); add_key(result,cache_get("key"))`
	for generation := 0; generation < 25; generation++ {
		old, err := runner.acquireProgram(source)
		if err != nil {
			t.Fatal(err)
		}
		p := old.program.(*nativeProgram)
		d := p.runtime.clockDriver
		count := func() int { d.mu.Lock(); defer d.mu.Unlock(); return len(d.entries) }
		if count() != 1 {
			t.Fatalf("generation %d: expected one registration", generation)
		}
		runner.Invalidate(source)
		if count() != 1 {
			t.Fatal("in-flight generation unregistered early")
		}
		fresh, err := runner.acquireProgram(source)
		if err != nil {
			old.Release()
			t.Fatal(err)
		}
		q := fresh.program.(*nativeProgram)
		if q == p || q.runtime.clockDriver != d || count() != 2 {
			t.Fatal("new generation did not share driver independently")
		}
		if err := p.advanceCacheTick(); err != nil {
			t.Fatal("leased old generation stopped:", err)
		}
		old.Release()
		if count() != 1 {
			t.Fatal("released old generation retained by scheduler")
		}
		if err := p.advanceCacheTick(); err == nil {
			t.Fatal("released old handle still callable")
		}
		if err := q.advanceCacheTick(); err != nil {
			t.Fatal("new generation damaged:", err)
		}
		runner.Invalidate(source)
		fresh.Release()
		if count() != 0 {
			t.Fatal("registration leaked after both generations drained")
		}
	}
}
