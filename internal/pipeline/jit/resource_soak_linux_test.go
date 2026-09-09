// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type processResources struct {
	rssBytes      uint64
	fds           int
	threads       int
	nativeThreads int
	goroutines    int
}

func readProcessResources() (processResources, error) {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return processResources{}, err
	}
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return processResources{}, err
	}
	result := processResources{fds: len(entries), goroutines: runtime.NumGoroutine()}
	tasks, err := os.ReadDir("/proc/self/task")
	if err != nil {
		return processResources{}, err
	}
	for _, task := range tasks {
		name, err := os.ReadFile("/proc/self/task/" + task.Name() + "/comm")
		if err == nil && strings.HasPrefix(strings.TrimSpace(string(name)), "platypus-") {
			result.nativeThreads++
		}
	}
	for _, line := range strings.Split(string(status), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "VmRSS:":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return processResources{}, err
			}
			result.rssBytes = value * 1024
		case "Threads:":
			value, err := strconv.Atoi(fields[1])
			if err != nil {
				return processResources{}, err
			}
			result.threads = value
		}
	}
	return result, nil
}

// This is the bounded CI form of the production replacement soak. Set
// PLATYPUS_JIT_SOAK_ITERATIONS higher (or PLATYPUS_JIT_SOAK_DURATION, e.g.
// "1h") for release qualification. Every generation opens the real library,
// compiles machine code, processes a batch and closes all native resources.
func TestNativeRunnerReplacementResourceSoak(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real PLATYPUS_JIT_RUNTIME")
	}
	iterations := 200
	if text := os.Getenv("PLATYPUS_JIT_SOAK_ITERATIONS"); text != "" {
		value, err := strconv.Atoi(text)
		if err != nil || value < 1 {
			t.Fatalf("invalid PLATYPUS_JIT_SOAK_ITERATIONS %q", text)
		}
		iterations = value
	}
	duration := time.Duration(0)
	if text := os.Getenv("PLATYPUS_JIT_SOAK_DURATION"); text != "" {
		value, err := time.ParseDuration(text)
		if err != nil || value <= 0 {
			t.Fatalf("invalid PLATYPUS_JIT_SOAK_DURATION %q", text)
		}
		duration = value
	}

	source := strings.Repeat(`replace(message,"[a-z]+","x")`+"\n", 16) +
		`cache_set("seen",message); add_key(done,cache_get("seen")==message)`
	input, err := EncodeFlatPoints([]Point{
		{Version: 1, Category: "logging", Measurement: "soak", Fields: map[string]any{"message": "alpha beta gamma"}},
		{Version: 1, Category: "logging", Measurement: "soak", Fields: map[string]any{"message": "delta epsilon"}},
		{Version: 1, Category: "logging", Measurement: "soak", Fields: map[string]any{"message": "zeta eta theta"}},
		{Version: 1, Category: "logging", Measurement: "soak", Fields: map[string]any{"message": "iota kappa"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	runGeneration := func(index int) {
		t.Helper()
		runner, err := NewRunnerWithServices(path, "pipeline-go-1.4.3-datakit", 4, nil, ServiceConfig{Version: 1})
		if err != nil {
			t.Fatalf("generation %d open: %v", index, err)
		}
		if check := runner.Check(source); check.Route != RouteJITNative {
			_ = runner.Close()
			t.Fatalf("generation %d route: %+v", index, check)
		}
		batch, err := runner.Process(source, input)
		if err != nil || len(batch.Records) != 4 {
			_ = runner.Close()
			t.Fatalf("generation %d process: records=%d err=%v", index, len(batch.Records), err)
		}
		for record, value := range batch.Records {
			if value.Status != TerminalOK {
				_ = runner.Close()
				t.Fatalf("generation %d record %d: %+v", index, record, value)
			}
		}
		if err := runner.Close(); err != nil {
			t.Fatalf("generation %d close: %v", index, err)
		}
	}

	for i := 0; i < 10; i++ {
		runGeneration(-i - 1)
	}
	runtime.GC()
	baseline, err := readProcessResources()
	if err != nil {
		t.Fatal(err)
	}
	peak := baseline
	started := time.Now()
	completed := 0
	for completed < iterations || (duration > 0 && time.Since(started) < duration) {
		runGeneration(completed)
		completed++
		if completed%25 == 0 {
			current, err := readProcessResources()
			if err != nil {
				t.Fatal(err)
			}
			if current.rssBytes > peak.rssBytes {
				peak.rssBytes = current.rssBytes
			}
			if current.fds > peak.fds {
				peak.fds = current.fds
			}
			if current.threads > peak.threads {
				peak.threads = current.threads
			}
			if current.nativeThreads > peak.nativeThreads {
				peak.nativeThreads = current.nativeThreads
			}
			if current.goroutines > peak.goroutines {
				peak.goroutines = current.goroutines
			}
		}
	}
	runtime.GC()
	final, err := readProcessResources()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("generations=%d elapsed=%s baseline=%+v peak=%+v final=%+v", completed, time.Since(started), baseline, peak, final)
	if final.fds > baseline.fds+4 {
		t.Errorf("FDs retained: baseline=%d final=%d", baseline.fds, final.fds)
	}
	// The Go scheduler may retain extra M threads after sustained CPU pressure;
	// track Rust-owned named workers separately so this gate detects native
	// generation leaks without treating normal Go runtime adaptation as one.
	if final.threads > baseline.threads+8 {
		t.Errorf("threads retained: baseline=%d final=%d", baseline.threads, final.threads)
	}
	if final.nativeThreads > baseline.nativeThreads {
		t.Errorf("native threads retained: baseline=%d final=%d", baseline.nativeThreads, final.nativeThreads)
	}
	if final.goroutines > baseline.goroutines+4 {
		t.Errorf("goroutines retained: baseline=%d final=%d", baseline.goroutines, final.goroutines)
	}
	const rssAllowance = 32 << 20
	if final.rssBytes > baseline.rssBytes+rssAllowance {
		t.Errorf("RSS retained beyond allowance: baseline=%s final=%s delta=%s", byteSize(baseline.rssBytes), byteSize(final.rssBytes), byteSize(final.rssBytes-baseline.rssBytes))
	}
}

func byteSize(value uint64) string {
	return fmt.Sprintf("%.2fMiB", float64(value)/(1<<20))
}
