// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jit

import (
	"context"
	"errors"
	"testing"
)

func TestCancelledRunnerRequestsDoNotCompile(t *testing.T) {
	runtime := &checkRuntime{program: &checkProgram{}}
	runner := &Runner{runtime: runtime, profile: "pipeline-go-1.4.3-datakit", max: 4, cache: make(map[[32]byte]*cacheEntry)}
	defer runner.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runner.ProcessContext(ctx, "exit()", nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("error: %v", err)
	}
	snapshot, err := NewModuleSnapshot("main.p", map[string]string{"main.p": "exit()"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runner.ProcessModulesContext(ctx, snapshot, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("module error: %v", err)
	}
	if runtime.compileCount != 0 || len(runner.cache) != 0 {
		t.Fatal("canceled request compiled or populated cache")
	}
}

func TestCancellableRequestDoesNotSilentlyUseLegacyProgram(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := processProgramContext(ctx, &checkProgram{}, nil); err == nil {
		t.Fatal("ignored cancellation requirement")
	}
	if _, err := processProgramContext(context.Background(), &checkProgram{}, nil); err != nil {
		t.Fatal(err)
	}
}
