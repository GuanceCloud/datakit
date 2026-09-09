// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package pipeline

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
	pljit "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/jit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

type generationTestRunner struct {
	closed     atomic.Bool
	closeCalls atomic.Int32
	closeDone  chan struct{}
}

func TestPipelineLeaseCapturesReferInstance(t *testing.T) {
	originalManager, _ := plval.GetManager()
	originalTable, _ := plval.GetRefTb()
	t.Cleanup(func() { plval.SetManager(originalManager); plval.SetRefTb(originalTable) })
	plval.SetManager(plval.NewScriptManager(nil, nil))
	old, err := refertable.NewReferTable(refertable.RefTbCfg{URL: "http://127.0.0.1/unused"})
	if err != nil {
		t.Fatal(err)
	}
	next, err := refertable.NewReferTable(refertable.RefTbCfg{URL: "http://127.0.0.1/unused"})
	if err != nil {
		t.Fatal(err)
	}
	plval.SetRefTb(old)
	manager, runner, offload, snapshot, _, err := acquirePipelineLeasesContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Release()
	if runner != nil {
		defer runner.release()
	}
	if offload != nil {
		defer offload.Release()
	}
	plval.SetRefTb(next)
	if snapshot != old.Tables() {
		t.Fatal("batch lost old refer instance")
	}
	manager2, runner2, offload2, snapshot2, _, err := acquirePipelineLeasesContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	manager2.Release()
	if runner2 != nil {
		runner2.release()
	}
	if offload2 != nil {
		offload2.Release()
	}
	if snapshot2 != next.Tables() {
		t.Fatal("new batch did not capture new refer instance")
	}
}

func TestRunPlCancellationWhileInitializationLocked(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			jitInitMu.Lock()
			defer jitInitMu.Unlock()
			var ctx context.Context
			var cancel context.CancelFunc
			want := context.Canceled
			if mode == "deadline" {
				ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
				want = context.DeadlineExceeded
			} else {
				ctx, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			started := make(chan struct{})
			done := make(chan error, 1)
			go func() {
				close(started)
				result, err := RunPlContext(ctx, point.Logging, nil, nil)
				if result != nil {
					done <- errors.New("canceled call returned a result")
					return
				}
				done <- err
			}()
			<-started
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatalf("got %v want %v", err, want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("cancellation waited for initialization lock")
			}
		})
	}
}

func TestPipelineLeasesNeverMixPublishedPairs(t *testing.T) {
	original, _ := plval.GetManager()
	first, second := plval.NewScriptManager(nil, nil), plval.NewScriptManager(nil, nil)
	a, _ := first.Acquire()
	firstRaw := a.Manager()
	a.Release()
	b, _ := second.Acquire()
	secondRaw := b.Manager()
	b.Release()
	one, two := &generationTestRunner{}, &generationTestRunner{}
	if previous := jitRunners.replace(one); previous != nil {
		t.Fatal("unexpected active runner")
	}
	plval.SetManager(first)
	t.Cleanup(func() { plval.SetManager(original); closeJITGeneration(jitRunners.replace(nil)) })
	start := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-start
		for i := 0; i < 500; i++ {
			jitInitMu.Lock()
			if i%2 == 0 {
				plval.SetManager(second)
				jitRunners.replace(two)
			} else {
				plval.SetManager(first)
				jitRunners.replace(one)
			}
			jitInitMu.Unlock()
		}
	}()
	close(start)
	// Processor doubles deliberately stay alive: this test isolates pair
	// acquisition, while native retirement/draining has separate tests.
	for i := 0; i < 1000; i++ {
		manager, runner, ok := acquirePipelineLeases()
		if !ok {
			<-done
			t.Fatal("missing generation")
		}
		matched := (manager.Manager() == firstRaw && runner.runner() == one) || (manager.Manager() == secondRaw && runner.runner() == two)
		manager.Release()
		runner.release()
		if !matched {
			<-done
			t.Fatal("manager and runner came from different publications")
		}
	}
	<-done
}

func TestPipelineJITPreflightFailurePreservesManagerAndHTTP(t *testing.T) {
	manager, hadManager := plval.GetManager()
	originalHTTP := funcs.FuncsMap["http_request"]
	t.Cleanup(func() { plval.SetManager(manager); funcs.FuncsMap["http_request"] = originalHTTP })
	// Explicitly select strict initialization; the default policy falls back to Go.
	err := InitPipeline(&plval.PipelineCfg{
		DisableHTTPRequestFunc: true,
		JIT:                    &plval.JITCfg{Enabled: true, OnInitError: "error", RuntimePath: filepath.Join(t.TempDir(), "missing.so")},
	}, nil, nil, t.TempDir())
	if err == nil {
		t.Fatal("missing runtime accepted")
	}
	after, hasManager := plval.GetManager()
	if after != manager || hasManager != hadManager {
		t.Fatal("preflight failure replaced manager")
	}
	if reflect.ValueOf(funcs.FuncsMap["http_request"]).Pointer() != reflect.ValueOf(originalHTTP).Pointer() {
		t.Fatal("preflight failure changed HTTP function")
	}
}

func TestJITInitFailureRetainsRunner(t *testing.T) {
	// Do not replace any unrelated global generation in this test.
	jitRunners.mu.Lock()
	if jitRunners.current != nil {
		jitRunners.mu.Unlock()
		t.Fatal("unexpected active runner")
	}
	jitRunners.mu.Unlock()
	old := &generationTestRunner{}
	jitRunners.replace(old)
	t.Cleanup(func() {
		if err := closeJITGeneration(jitRunners.replace(nil)); err != nil {
			t.Error(err)
		}
	})
	cfg := &plval.PipelineCfg{JIT: &plval.JITCfg{Enabled: true, RuntimePath: filepath.Join(t.TempDir(), "missing.so"), MaxCachedPrograms: 2}}
	if err := initJIT(cfg, ""); err == nil {
		t.Fatal("JIT initialization failure was hidden")
	}
	lease := jitRunners.acquire()
	if lease == nil {
		t.Fatal("failed candidate unpublished active runner")
	}
	defer lease.release()
	if lease.runner() != old || old.closed.Load() {
		t.Fatal("failed candidate replaced or closed active runner")
	}
}

func (*generationTestRunner) Prepare(string) error { return nil }

func (*generationTestRunner) Check(string) pljit.CheckResult {
	return pljit.CheckResult{Route: pljit.RouteJITNative}
}

func (*generationTestRunner) Invalidate(string) {}

func (*generationTestRunner) Projection(string) (pljit.InputProjection, error) {
	return pljit.InputProjection{}, nil
}

func (*generationTestRunner) Process(string, []byte) (pljit.Batch, error) {
	return pljit.Batch{}, nil
}

func (*generationTestRunner) MinBatchSize() int { return 0 }

func (runner *generationTestRunner) Close() error {
	runner.closeCalls.Add(1)
	if !runner.closed.CompareAndSwap(false, true) {
		return errors.New("runner closed more than once")
	}
	if runner.closeDone != nil {
		close(runner.closeDone)
	}
	return nil
}

func TestJITRunnerAsyncRetirementPublishesBeforeOldDrain(t *testing.T) {
	var registry jitRunnerRegistry
	oldRunner := &generationTestRunner{closeDone: make(chan struct{})}
	newRunner := &generationTestRunner{}
	registry.replace(oldRunner)
	oldLease := registry.acquire()
	if oldLease == nil || oldLease.runner() != oldRunner {
		t.Fatal("old batch did not acquire its generation")
	}

	retired := registry.replace(newRunner)
	retireJITGeneration(retired)
	select {
	case <-oldRunner.closeDone:
		t.Fatal("old runner closed while its batch was active")
	default:
	}
	newLease := registry.acquire()
	if newLease == nil || newLease.runner() != newRunner {
		t.Fatal("new batch could not enter the published generation")
	}
	newLease.release()
	oldLease.release()
	select {
	case <-oldRunner.closeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("old runner was not reclaimed after its final lease")
	}
	if oldRunner.closeCalls.Load() != 1 || newRunner.closed.Load() {
		t.Fatalf("retirement close state old=%d new=%v", oldRunner.closeCalls.Load(), newRunner.closed.Load())
	}
	if err := closeJITGeneration(registry.replace(nil)); err != nil {
		t.Fatal(err)
	}
}

func TestJITRunnerAsyncRetirementReclaimsReplacementBurst(t *testing.T) {
	var registry jitRunnerRegistry
	const replacements = 500
	runners := make([]*generationTestRunner, 0, replacements+1)
	first := &generationTestRunner{closeDone: make(chan struct{})}
	runners = append(runners, first)
	registry.replace(first)

	for i := 0; i < replacements; i++ {
		// Hold a lease across publication so every retirement exercises the
		// asynchronous drain path rather than closing immediately by accident.
		lease := registry.acquire()
		if lease == nil {
			t.Fatalf("replacement %d could not acquire current generation", i)
		}
		next := &generationTestRunner{closeDone: make(chan struct{})}
		runners = append(runners, next)
		retireJITGeneration(registry.replace(next))
		lease.release()
	}

	for index, runner := range runners[:len(runners)-1] {
		select {
		case <-runner.closeDone:
		case <-time.After(5 * time.Second):
			t.Fatalf("retired runner %d was not reclaimed", index)
		}
		if runner.closeCalls.Load() != 1 {
			t.Fatalf("retired runner %d closed %d times", index, runner.closeCalls.Load())
		}
	}
	final := runners[len(runners)-1]
	if final.closed.Load() {
		t.Fatal("active runner closed during replacement burst")
	}
	if err := closeJITGeneration(registry.replace(nil)); err != nil {
		t.Fatal(err)
	}
	if final.closeCalls.Load() != 1 {
		t.Fatalf("final runner closed %d times", final.closeCalls.Load())
	}
}

func TestJITRunnerGenerationDefersCloseUntilBatchRelease(t *testing.T) {
	var registry jitRunnerRegistry
	oldRunner := &generationTestRunner{}
	newRunner := &generationTestRunner{}
	if retired := registry.replace(oldRunner); retired != nil {
		t.Fatal("first runner replacement unexpectedly retired a generation")
	}

	oldLease := registry.acquire()
	if oldLease == nil || oldLease.runner() != oldRunner {
		t.Fatal("batch did not acquire the old runner generation")
	}
	retired := registry.replace(newRunner)
	if retired == nil {
		t.Fatal("runner replacement did not retire the old generation")
	}
	newLease := registry.acquire()
	if newLease == nil || newLease.runner() != newRunner {
		t.Fatal("new batch did not acquire the published runner generation")
	}

	closeDone := make(chan error, 1)
	go func() { closeDone <- closeJITGeneration(retired) }()
	if oldRunner.closed.Load() {
		t.Fatal("old runner closed while its batch lease was active")
	}
	oldLease.release()
	if err := <-closeDone; err != nil {
		t.Fatalf("close retired runner: %v", err)
	}
	if !oldRunner.closed.Load() || oldRunner.closeCalls.Load() != 1 {
		t.Fatalf("old runner close state = %v, calls = %d", oldRunner.closed.Load(), oldRunner.closeCalls.Load())
	}
	if newRunner.closed.Load() {
		t.Fatal("replacement runner was closed with the retired generation")
	}

	newLease.release()
	if err := closeJITGeneration(registry.replace(nil)); err != nil {
		t.Fatalf("close replacement runner: %v", err)
	}
}

func TestJITRunnerGenerationConcurrentAcquireReplace(t *testing.T) {
	var registry jitRunnerRegistry
	registry.replace(&generationTestRunner{})

	const (
		workers      = 16
		replacements = 100
	)
	start := make(chan struct{})
	errs := make(chan error, workers)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			<-start
			for range replacements {
				lease := registry.acquire()
				if lease == nil {
					errs <- errors.New("runner generation disappeared during replacement")
					return
				}
				runner := lease.runner().(*generationTestRunner)
				if runner.closed.Load() {
					errs <- errors.New("batch acquired a closed runner generation")
					lease.release()
					return
				}
				lease.release()
			}
		}()
	}
	close(start)

	for range replacements {
		retired := registry.replace(&generationTestRunner{})
		if err := closeJITGeneration(retired); err != nil {
			t.Fatalf("close retired runner: %v", err)
		}
	}
	workersDone.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if err := closeJITGeneration(registry.replace(nil)); err != nil {
		t.Fatalf("close final runner: %v", err)
	}
}
