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
)

func TestNativeCacheConcurrentRefreshTickAndClose(t *testing.T) {
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
	for round := 0; round < 20; round++ {
		compiled, err := rt.Compile(`add_key(result,cache_set("key",message,2) == nil)`, "pipeline-go-1.4.3-datakit")
		if err != nil {
			t.Fatal(err)
		}
		p := compiled.(*nativeProgram)
		if err := p.enableCacheTicks(uint64(time.Second)); err != nil {
			t.Fatal(err)
		}
		input, err := EncodeFlatPoints([]Point{{Version: 1, Category: "logging", Measurement: "race", Fields: map[string]any{"message": "refresh"}}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.ProcessIndexed(input); err != nil {
			t.Fatal(err)
		}
		start := make(chan struct{})
		ready := make(chan struct{}, 5)
		errors := make(chan error, 5)
		var workers sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				for i := 0; i < 64; i++ {
					batch, err := p.ProcessIndexed(input)
					if i == 0 {
						ready <- struct{}{}
					}
					if err != nil {
						if err.Error() != "JIT program is closed" {
							errors <- err
						}
						return
					}
					if len(batch.Records) != 1 || batch.Records[0].Status != TerminalOK {
						errors <- fmt.Errorf("unexpected batch: %+v", batch)
						return
					}
				}
			}()
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for i := 0; i < 64; i++ {
				p.scheduledCacheTick()
				if i == 0 {
					ready <- struct{}{}
				}
			}
		}()
		close(start)
		for i := 0; i < 5; i++ {
			<-ready
		}
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
		workers.Wait()
		close(errors)
		for err := range errors {
			t.Errorf("round=%d: %v", round, err)
		}
		if err := p.cacheClockError(); err != nil {
			t.Fatal(err)
		}
		if _, err := p.ProcessIndexed(input); err == nil {
			t.Fatal("closed program executed")
		}
		if err := p.Close(); err != nil {
			t.Fatal("repeat close:", err)
		}
	}
}
