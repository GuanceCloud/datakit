// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build pipeline_jit && linux && (amd64 || arm64)

package jit

import (
	"errors"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestServicesHostFactoryLifecycle(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	called := 0
	host := NewPipelineGoHost(nil)
	factory := func() (*HostCompat, error) { called++; return host, nil }
	if _, err := NewRunnerWithServicesHostFactory(filepath.Join(t.TempDir(), "missing.so"), "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, factory); err == nil || called != 0 {
		t.Fatalf("load failure invoked factory: %d %v", called, err)
	}
	if _, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 99}, factory); err == nil || called != 0 {
		t.Fatalf("policy failure invoked factory: %d %v", called, err)
	}
	want := errors.New("host preparation failed")
	if runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, func() (*HostCompat, error) { return nil, want }); runner != nil || !errors.Is(err, want) {
		t.Fatalf("factory error: %v %v", runner, err)
	}
	runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, factory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { runner.Close() })
	native := runner.runtime.(*nativeRuntime)
	if called != 1 || native.host == nil || native.host.host != host {
		t.Fatal("host not attached to prepared runtime")
	}
	token := native.host.token
	if _, ok := hostCompatHosts.Load(token); !ok {
		t.Fatal("host registration missing")
	}
	if check := runner.Check(`add_key(version,"prepared")`); check.Route != RouteJITNative {
		t.Fatalf("prepared runtime unusable: %+v", check)
	}
	if err := runner.Close(); err != nil {
		t.Fatal(err)
	}
	if _, ok := hostCompatHosts.Load(token); ok {
		t.Fatal("host registration retained after close")
	}
}

func TestServicesHostFactoryCloseWaitsForCallback(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, func() (*HostCompat, error) {
		return NewPipelineGoHostObserved(nil, func(op string) {
			if op == "geoip" {
				close(entered)
				<-release
			}
		}), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { unblock(); runner.Close() })
	token := runner.runtime.(*nativeRuntime).host.token
	input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "close", Fields: map[string]any{"ip": "203.0.113.1"}}})
	if err != nil {
		t.Fatal(err)
	}
	type result struct {
		batch Batch
		err   error
	}
	done := make(chan result, 1)
	go func() {
		batch, err := runner.Process(`geoip(ip); add_key(after,true)`, input)
		done <- result{batch, err}
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("callback did not start")
	}
	closed := make(chan error, 1)
	go func() { closed <- runner.Close() }()
	// Observe the actual close transition, not a sleep-based assumption.
	deadline := time.After(5 * time.Second)
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		runner.mu.Lock()
		closing := runner.closed
		runner.mu.Unlock()
		if closing {
			break
		}
		select {
		case <-tick.C:
		case <-deadline:
			t.Fatal("close did not start")
		}
	}
	select {
	case err := <-closed:
		t.Fatalf("close returned during callback: %v", err)
	default:
	}
	if _, ok := hostCompatHosts.Load(token); !ok {
		t.Fatal("host removed during callback")
	}
	if _, err := runner.Process(`add_key(unexpected,true)`, input); err == nil {
		t.Fatal("closing runner accepted new process")
	}
	unblock()
	select {
	case result := <-done:
		if result.err != nil || len(result.batch.Records) != 1 || result.batch.Records[0].Status != TerminalOK {
			t.Fatalf("in-flight result: %+v", result)
		}
		pt := Point{Version: 1, Category: "logging", Fields: map[string]any{"ip": "203.0.113.1"}}
		applyHostCompatRecord(t, result.batch, 0, &pt)
		if pt.Fields["after"] != true {
			t.Fatalf("in-flight script did not complete: %+v", pt.Fields)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("process did not drain")
	}
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("close did not drain")
	}
	if _, ok := hostCompatHosts.Load(token); ok {
		t.Fatal("host registration leaked")
	}
}

func TestServicesHostFactoryExecutesIsolatedCallbacks(t *testing.T) {
	path := os.Getenv("PLATYPUS_JIT_RUNTIME")
	if path == "" {
		t.Fatal("requires real runtime")
	}
	const source = `geoip(ip); add_key(after,true)`
	var calls [2]atomic.Int32
	var runners [2]*Runner
	for i, city := range []string{"first-city", "second-city"} {
		i, city := i, city
		runner, err := NewRunnerWithServicesHostFactory(path, "pipeline-go-1.4.3-datakit", 2, ServiceConfig{Version: 1}, func() (*HostCompat, error) {
			return NewPipelineGoHostObserved(&hostCompatIPDB{record: &ipdb.IPdbRecord{City: city, Region: "region", Country: "CN", Isp: "isp"}}, func(operation string) {
				if operation == "geoip" {
					calls[i].Add(1)
				}
			}), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		runners[i] = runner
		t.Cleanup(func() { runner.Close() })
		if check := runner.Check(source); check.Route != RouteJITWithHost {
			t.Fatalf("expected host route: %+v", check)
		}
	}
	run := func(i int, wantCity string) {
		t.Helper()
		got := runHostCompatPoint(t, runners[i], source, Point{Version: 1, Category: "logging", Measurement: "factory", TimeUnixNano: 1700000000000000123, Fields: map[string]any{"ip": "203.0.113.1"}})
		if got.Fields["city"] != wantCity || got.Fields["after"] != true {
			t.Fatalf("wrong host output: %+v", got.Fields)
		}
	}
	run(0, "first-city")
	run(1, "second-city")
	if calls[0].Load() != 1 || calls[1].Load() != 1 {
		t.Fatalf("callback counts: %d/%d", calls[0].Load(), calls[1].Load())
	}
	if err := runners[0].Close(); err != nil {
		t.Fatal(err)
	}
	run(1, "second-city")
	if calls[0].Load() != 1 || calls[1].Load() != 2 {
		t.Fatalf("callback isolation after close: %d/%d", calls[0].Load(), calls[1].Load())
	}
}
