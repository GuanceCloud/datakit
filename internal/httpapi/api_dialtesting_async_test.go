// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	dt "github.com/GuanceCloud/cliutils/dialtesting"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDialtestingDebugConfig() dialtestingDebugConfig {
	return dialtestingDebugConfig{
		maxConcurrentRuns:       1,
		maxQueuedRuns:           2,
		maxQueueWait:            time.Second,
		resultTTL:               time.Minute,
		maxLongPollWait:         10 * time.Second,
		maxRetainedTerminalRuns: 10,
	}
}

func successfulDialtestingDebugExecutor() (*dialtestingDebugResponse, error) {
	return &dialtestingDebugResponse{Status: "success", Fields: map[string]interface{}{"ok": true}}, nil
}

func submitTestDialtestingDebugRun(
	t *testing.T,
	m *dialtestingDebugManager,
	owner, requestID, hash string,
	execute func() (*dialtestingDebugResponse, error),
) *dialtestingDebugRunResponse {
	t.Helper()
	response, err := m.submit(context.Background(), owner, requestID, hash, "HTTP", "trace-test",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) { return execute, nil })
	require.NoError(t, err)
	return response
}

func waitForDialtestingDebugStatus(
	t *testing.T, m *dialtestingDebugManager, owner, runID, status string,
) *dialtestingDebugRunResponse {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		response, err := m.get(context.Background(), owner, runID, 0)
		require.NoError(t, err)
		if response.Status == status {
			return response
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("run %s did not reach status %s", runID, status)
	return nil
}

func dialtestingDebugErrorCode(t *testing.T, err error) string {
	t.Helper()
	switch typed := err.(type) {
	case *uhttp.MsgError:
		return typed.ErrCode
	case *uhttp.HttpError:
		return typed.ErrCode
	default:
		t.Fatalf("unexpected error type %T", err)
		return ""
	}
}

func TestDialtestingDebugManagerStateAndLongPollWakeup(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	release := make(chan struct{})
	started := make(chan struct{})
	response := submitTestDialtestingDebugRun(t, m, "owner-a", "req-a", "hash-a",
		func() (*dialtestingDebugResponse, error) {
			close(started)
			<-release
			return &dialtestingDebugResponse{Status: "fail", Fields: map[string]interface{}{"fail_reason": "unreachable"}}, nil
		})
	<-started
	waitForDialtestingDebugStatus(t, m, "owner-a", response.RunID, dialtestingDebugStatusRunning)

	done := make(chan *dialtestingDebugRunResponse, 1)
	go func() {
		result, _ := m.get(context.Background(), "owner-a", response.RunID, 5*time.Second)
		done <- result
	}()
	close(release)
	select {
	case result := <-done:
		assert.Equal(t, dialtestingDebugStatusCompleted, result.Status)
		require.NotNil(t, result.Result)
		assert.Equal(t, "fail", result.Result.Status)
	case <-time.After(time.Second):
		t.Fatal("long poll was not woken by completion")
	}
}

func TestDialtestingDebugManagerOwnerAndIdempotency(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	release := make(chan struct{})
	var prepared atomic.Int32
	prepare := func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
		prepared.Add(1)
		return func() (*dialtestingDebugResponse, error) {
			<-release
			return successfulDialtestingDebugExecutor()
		}, nil
	}
	first, err := m.submit(context.Background(), "owner-a", "req-a", "hash-a", "HTTP", "trace", prepare)
	require.NoError(t, err)
	second, err := m.submit(context.Background(), "owner-a", "req-a", "hash-a", "HTTP", "trace", prepare)
	require.NoError(t, err)
	assert.Equal(t, first.RunID, second.RunID)
	assert.Equal(t, int32(1), prepared.Load())

	_, err = m.submit(context.Background(), "owner-a", "req-a", "different", "HTTP", "trace", prepare)
	assert.ErrorContains(t, err, "request_id already used")
	_, err = m.get(context.Background(), "owner-b", first.RunID, 0)
	assert.ErrorContains(t, err, "run not found")
	close(release)
}

func TestDialtestingDebugManagerConcurrentIdempotency(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	prepareRelease := make(chan struct{})
	executeRelease := make(chan struct{})
	var prepared atomic.Int32
	var executed atomic.Int32
	prepare := func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
		prepared.Add(1)
		<-prepareRelease
		return func() (*dialtestingDebugResponse, error) {
			executed.Add(1)
			<-executeRelease
			return successfulDialtestingDebugExecutor()
		}, nil
	}

	const count = 32
	results := make(chan string, count)
	errs := make(chan error, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := m.submit(context.Background(), "owner-a", "req-a", "hash-a", "HTTP", "trace", prepare)
			if err != nil {
				errs <- err
				return
			}
			results <- response.RunID
		}()
	}
	require.Eventually(t, func() bool { return prepared.Load() == 1 }, time.Second, time.Millisecond)
	close(prepareRelease)
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var runID string
	for got := range results {
		if runID == "" {
			runID = got
		}
		assert.Equal(t, runID, got)
	}
	assert.Equal(t, int32(1), prepared.Load())
	require.Eventually(t, func() bool { return executed.Load() == 1 }, time.Second, time.Millisecond)
	close(executeRelease)
}

func TestDialtestingDebugManagerPreparingPayloadConflict(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	prepareStarted := make(chan struct{})
	prepareRelease := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := m.submit(context.Background(), "owner-a", "req-a", "hash-a", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				close(prepareStarted)
				<-prepareRelease
				return successfulDialtestingDebugExecutor, nil
			})
		firstDone <- err
	}()
	<-prepareStarted

	var conflictingPrepared atomic.Int32
	_, err := m.submit(context.Background(), "owner-a", "req-a", "different", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			conflictingPrepared.Add(1)
			return successfulDialtestingDebugExecutor, nil
		})
	assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
	assert.Equal(t, "datakit.InvalidTask", dialtestingDebugErrorCode(t, err))
	assert.Zero(t, conflictingPrepared.Load())
	close(prepareRelease)
	require.NoError(t, <-firstDone)
}

func TestDialtestingDebugManagerQueueFullAndQueueTimeout(t *testing.T) {
	cfg := testDialtestingDebugConfig()
	cfg.maxQueuedRuns = 2
	m := newDialtestingDebugManager(cfg)
	t.Cleanup(m.close)
	now := time.Now()
	m.now = func() time.Time { return now }
	release := make(chan struct{})
	started := make(chan struct{})
	first := submitTestDialtestingDebugRun(t, m, "owner", "req-1", "hash-1",
		func() (*dialtestingDebugResponse, error) {
			close(started)
			<-release
			return successfulDialtestingDebugExecutor()
		})
	<-started
	waitForDialtestingDebugStatus(t, m, "owner", first.RunID, dialtestingDebugStatusRunning)
	var staleExecuted atomic.Int32
	staleExecutor := func() (*dialtestingDebugResponse, error) {
		staleExecuted.Add(1)
		return successfulDialtestingDebugExecutor()
	}
	second := submitTestDialtestingDebugRun(t, m, "owner", "req-2", "hash-2", staleExecutor)
	third := submitTestDialtestingDebugRun(t, m, "owner", "req-3", "hash-3", staleExecutor)
	var fullPrepared atomic.Int32
	_, err := m.submit(context.Background(), "owner", "req-full", "hash-full", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			fullPrepared.Add(1)
			return successfulDialtestingDebugExecutor, nil
		})
	assert.ErrorIs(t, err, errDialtestingDebugQueueFull)
	assert.Zero(t, fullPrepared.Load())
	_, err = m.submit(context.Background(), "owner", "req-invalid", "hash-invalid", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			fullPrepared.Add(1)
			return nil, uhttp.Error(ErrDialtestingInvalidTask, "invalid task")
		})
	assert.ErrorIs(t, err, errDialtestingDebugQueueFull)
	assert.Zero(t, fullPrepared.Load())

	now = now.Add(2 * time.Second)
	executionOrder := make(chan string, 2)
	fourth := submitTestDialtestingDebugRun(t, m, "owner", "req-4", "hash-4",
		func() (*dialtestingDebugResponse, error) {
			executionOrder <- "4"
			return successfulDialtestingDebugExecutor()
		})
	fifth := submitTestDialtestingDebugRun(t, m, "owner", "req-5", "hash-5",
		func() (*dialtestingDebugResponse, error) {
			executionOrder <- "5"
			return successfulDialtestingDebugExecutor()
		})
	waitForDialtestingDebugStatus(t, m, "owner", second.RunID, dialtestingDebugStatusTimedOut)
	waitForDialtestingDebugStatus(t, m, "owner", third.RunID, dialtestingDebugStatusTimedOut)
	close(release)
	waitForDialtestingDebugStatus(t, m, "owner", fourth.RunID, dialtestingDebugStatusCompleted)
	waitForDialtestingDebugStatus(t, m, "owner", fifth.RunID, dialtestingDebugStatusCompleted)
	assert.Equal(t, "4", <-executionOrder)
	assert.Equal(t, "5", <-executionOrder)
	assert.Zero(t, staleExecuted.Load())
}

func TestDialtestingDebugManagerPreparationReservesAndReleasesCapacity(t *testing.T) {
	cfg := testDialtestingDebugConfig()
	cfg.maxQueuedRuns = 1
	m := newDialtestingDebugManager(cfg)
	t.Cleanup(m.close)

	prepareStarted := make(chan struct{})
	prepareRelease := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := m.submit(context.Background(), "owner", "req-preparing", "hash-preparing", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				close(prepareStarted)
				<-prepareRelease
				return nil, uhttp.Error(ErrDialtestingInvalidTask, "invalid task")
			})
		firstDone <- err
	}()
	<-prepareStarted

	var rejectedPrepared atomic.Int32
	_, err := m.submit(context.Background(), "owner", "req-rejected", "hash-rejected", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			rejectedPrepared.Add(1)
			return successfulDialtestingDebugExecutor, nil
		})
	assert.ErrorIs(t, err, errDialtestingDebugQueueFull)
	assert.Zero(t, rejectedPrepared.Load())

	close(prepareRelease)
	assert.Equal(t, http.StatusBadRequest, getStatusCode(<-firstDone))
	accepted, err := m.submit(context.Background(), "owner", "req-accepted", "hash-accepted", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			return successfulDialtestingDebugExecutor, nil
		})
	require.NoError(t, err)
	require.NotEmpty(t, accepted.RunID)
}

func TestDialtestingDebugManagerPreparationPanicReleasesCapacity(t *testing.T) {
	cfg := testDialtestingDebugConfig()
	cfg.maxQueuedRuns = 1
	m := newDialtestingDebugManager(cfg)
	t.Cleanup(m.close)

	panicResult := make(chan error, 1)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				panicResult <- fmt.Errorf("panic escaped: %v", recovered)
			}
		}()
		_, err := m.submit(context.Background(), "owner", "req-panic", "hash-panic", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				panic("prepare failed")
			})
		panicResult <- err
	}()
	assert.ErrorContains(t, <-panicResult, "prepare failed")

	accepted, err := m.submit(context.Background(), "owner", "req-after-panic", "hash-after-panic", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			return successfulDialtestingDebugExecutor, nil
		})
	require.NoError(t, err)
	require.NotEmpty(t, accepted.RunID)
}

func TestDialtestingDebugManagerPreparingWaiterHonorsContextCancellation(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	prepareStarted := make(chan struct{})
	prepareRelease := make(chan struct{})
	var releaseOnce sync.Once
	releasePreparation := func() {
		releaseOnce.Do(func() { close(prepareRelease) })
	}
	defer releasePreparation()

	firstDone := make(chan *dialtestingDebugRunResponse, 1)
	go func() {
		response, _ := m.submit(context.Background(), "owner", "req-shared", "hash-shared", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				close(prepareStarted)
				<-prepareRelease
				return successfulDialtestingDebugExecutor, nil
			})
		firstDone <- response
	}()
	<-prepareStarted

	waiterCtx, cancelWaiter := context.WithCancel(context.Background())
	waiterDone := make(chan error, 1)
	var waiterPrepared atomic.Int32
	go func() {
		_, err := m.submit(waiterCtx, "owner", "req-shared", "hash-shared", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				waiterPrepared.Add(1)
				return successfulDialtestingDebugExecutor, nil
			})
		waiterDone <- err
	}()
	cancelWaiter()

	select {
	case err := <-waiterDone:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("preparation waiter did not stop after context cancellation")
	}
	assert.Zero(t, waiterPrepared.Load())
	select {
	case <-firstDone:
		t.Fatal("canceling a waiter canceled the shared preparation")
	default:
	}

	releasePreparation()
	first := <-firstDone
	require.NotNil(t, first)
	retry, err := m.submit(context.Background(), "owner", "req-shared", "hash-shared", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			return successfulDialtestingDebugExecutor, nil
		})
	require.NoError(t, err)
	assert.Equal(t, first.RunID, retry.RunID)
}

func TestDialtestingDebugManagerRejectsSubmitAfterClose(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	m.close()

	var prepared atomic.Int32
	response, err := m.submit(context.Background(), "owner", "req", "hash", "HTTP", "trace",
		func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
			prepared.Add(1)
			return successfulDialtestingDebugExecutor, nil
		})
	assert.Nil(t, response)
	assert.ErrorIs(t, err, errDialtestingDebugManagerStopped)
	assert.Zero(t, prepared.Load())
}

func TestDialtestingDebugManagerCloseDuringPreparation(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	prepareStarted := make(chan struct{})
	prepareRelease := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		_, err := m.submit(context.Background(), "owner", "req", "hash", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				close(prepareStarted)
				<-prepareRelease
				return successfulDialtestingDebugExecutor, nil
			})
		firstDone <- err
	}()
	<-prepareStarted

	waiterStarted := make(chan struct{})
	waiterDone := make(chan error, 1)
	go func() {
		close(waiterStarted)
		_, err := m.submit(context.Background(), "owner", "req", "hash", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				return successfulDialtestingDebugExecutor, nil
			})
		waiterDone <- err
	}()
	<-waiterStarted

	m.close()
	select {
	case err := <-waiterDone:
		assert.ErrorIs(t, err, errDialtestingDebugManagerStopped)
	case <-time.After(time.Second):
		t.Fatal("preparation waiter was not woken by close")
	}

	close(prepareRelease)
	assert.ErrorIs(t, <-firstDone, errDialtestingDebugManagerStopped)
	m.mu.Lock()
	assert.Empty(t, m.preparing)
	assert.Empty(t, m.queue)
	assert.Empty(t, m.runs)
	m.mu.Unlock()
}

func TestDialtestingDebugManagerCloseWakesQueuedLongPoll(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	runningRelease := make(chan struct{})
	runningStarted := make(chan struct{})
	running := submitTestDialtestingDebugRun(t, m, "owner", "running", "running-hash",
		func() (*dialtestingDebugResponse, error) {
			close(runningStarted)
			<-runningRelease
			return successfulDialtestingDebugExecutor()
		})
	<-runningStarted
	waitForDialtestingDebugStatus(t, m, "owner", running.RunID, dialtestingDebugStatusRunning)
	queued := submitTestDialtestingDebugRun(t, m, "owner", "queued", "queued-hash",
		successfulDialtestingDebugExecutor)

	longPollDone := make(chan *dialtestingDebugRunResponse, 1)
	longPollErr := make(chan error, 1)
	go func() {
		response, err := m.get(context.Background(), "owner", queued.RunID, 5*time.Second)
		longPollDone <- response
		longPollErr <- err
	}()

	m.close()
	select {
	case response := <-longPollDone:
		require.NoError(t, <-longPollErr)
		require.NotNil(t, response)
		assert.Equal(t, dialtestingDebugStatusFailed, response.Status)
		assert.Equal(t, "ServerUnavailable", response.Error)
	case <-time.After(time.Second):
		t.Fatal("queued run long poll was not woken by close")
	}
	close(runningRelease)
}

func TestDialtestingDebugManagerDoesNotExecuteDequeuedRunAfterClose(t *testing.T) {
	cfg := testDialtestingDebugConfig()
	cfg.maxConcurrentRuns = 0
	m := newDialtestingDebugManager(cfg)
	t.Cleanup(m.close)
	var executed atomic.Int32
	run := &dialtestingDebugRun{
		id:        "run",
		owner:     "owner",
		requestID: "request",
		taskType:  "HTTP",
		createdAt: m.now(),
		status:    dialtestingDebugStatusPending,
		execute: func() (*dialtestingDebugResponse, error) {
			executed.Add(1)
			return successfulDialtestingDebugExecutor()
		},
		notify: make(chan struct{}),
	}
	m.mu.Lock()
	m.runs[run.id] = run
	m.mu.Unlock()

	m.close()
	m.executeRun(run)

	assert.Zero(t, executed.Load())
	response, err := m.get(context.Background(), run.owner, run.id, 0)
	require.NoError(t, err)
	assert.Equal(t, dialtestingDebugStatusFailed, response.Status)
	assert.Equal(t, "ServerUnavailable", response.Error)
}

func TestDialtestingDebugPrepareDoesNotHoldManagerLock(t *testing.T) {
	m := newDialtestingDebugManager(testDialtestingDebugConfig())
	t.Cleanup(m.close)
	runRelease := make(chan struct{})
	running := submitTestDialtestingDebugRun(t, m, "owner", "req-running", "hash-running",
		func() (*dialtestingDebugResponse, error) {
			<-runRelease
			return successfulDialtestingDebugExecutor()
		})
	waitForDialtestingDebugStatus(t, m, "owner", running.RunID, dialtestingDebugStatusRunning)

	prepareStarted := make(chan struct{})
	prepareRelease := make(chan struct{})
	submitDone := make(chan error, 1)
	go func() {
		_, err := m.submit(context.Background(), "owner", "req-preparing", "hash-preparing", "HTTP", "trace",
			func(context.Context) (func() (*dialtestingDebugResponse, error), error) {
				close(prepareStarted)
				<-prepareRelease
				return successfulDialtestingDebugExecutor, nil
			})
		submitDone <- err
	}()
	<-prepareStarted

	getDone := make(chan error, 1)
	go func() {
		_, err := m.get(context.Background(), "owner", running.RunID, 0)
		getDone <- err
	}()
	select {
	case err := <-getDone:
		require.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("get was blocked by slow prepare")
	}
	close(runRelease)
	waitForDialtestingDebugStatus(t, m, "owner", running.RunID, dialtestingDebugStatusCompleted)
	close(prepareRelease)
	require.NoError(t, <-submitDone)
}

func TestDialtestingDebugManagerFailedTTLAndTerminalLimit(t *testing.T) {
	cfg := testDialtestingDebugConfig()
	cfg.resultTTL = time.Second
	cfg.maxRetainedTerminalRuns = 1
	m := newDialtestingDebugManager(cfg)
	t.Cleanup(m.close)
	now := time.Now()
	m.now = func() time.Time { return now }

	failed := submitTestDialtestingDebugRun(t, m, "owner", "req-fail", "hash-fail",
		func() (*dialtestingDebugResponse, error) { return nil, errors.New("boom") })
	result := waitForDialtestingDebugStatus(t, m, "owner", failed.RunID, dialtestingDebugStatusFailed)
	assert.Equal(t, "ExecutionFailed", result.Error)

	now = now.Add(time.Millisecond)
	completed := submitTestDialtestingDebugRun(t, m, "owner", "req-ok", "hash-ok", successfulDialtestingDebugExecutor)
	waitForDialtestingDebugStatus(t, m, "owner", completed.RunID, dialtestingDebugStatusCompleted)
	_, err := m.get(context.Background(), "owner", failed.RunID, 0)
	assert.ErrorContains(t, err, "run not found")

	now = now.Add(2 * time.Second)
	_, err = m.get(context.Background(), "owner", completed.RunID, 0)
	assert.ErrorContains(t, err, "run not found")
	assert.Empty(t, m.idempotency)
}

func TestDialtestingDebugEnvironment(t *testing.T) {
	names := []string{
		"ENV_DIALTESTING_DEBUG_MAX_CONCURRENT_RUNS",
		"ENV_DIALTESTING_DEBUG_MAX_QUEUED_RUNS",
		"ENV_DIALTESTING_DEBUG_MAX_QUEUE_WAIT",
		"ENV_DIALTESTING_DEBUG_RESULT_TTL",
		"ENV_DIALTESTING_DEBUG_MAX_LONG_POLL_WAIT",
		"ENV_DIALTESTING_DEBUG_MAX_RETAINED_TERMINAL_RUNS",
	}
	for _, name := range names {
		t.Setenv(name, "")
	}
	assert.Equal(t, defaultDialtestingDebugConfig(), loadDialtestingDebugConfigFromEnv())

	t.Setenv(names[0], "3")
	t.Setenv(names[1], "7")
	t.Setenv(names[2], "4m")
	t.Setenv(names[3], "20m")
	t.Setenv(names[4], "30s")
	t.Setenv(names[5], "99")
	cfg := loadDialtestingDebugConfigFromEnv()
	assert.Equal(t, 3, cfg.maxConcurrentRuns)
	assert.Equal(t, 7, cfg.maxQueuedRuns)
	assert.Equal(t, 4*time.Minute, cfg.maxQueueWait)
	assert.Equal(t, 20*time.Minute, cfg.resultTTL)
	assert.Equal(t, 10*time.Second, cfg.maxLongPollWait)
	assert.Equal(t, 99, cfg.maxRetainedTerminalRuns)

	for _, name := range names {
		t.Setenv(name, "invalid")
	}
	assert.Equal(t, defaultDialtestingDebugConfig(), loadDialtestingDebugConfigFromEnv())
}

func TestParseDialtestingDebugWait(t *testing.T) {
	previous := defaultDialtestingDebugManager
	defaultDialtestingDebugManager = nil
	t.Cleanup(func() { defaultDialtestingDebugManager = previous })

	tests := []struct {
		query   string
		want    time.Duration
		wantErr bool
	}{
		{"", 5 * time.Second, false},
		{"?wait=0", 0, false},
		{"?wait=1.5", 1500 * time.Millisecond, false},
		{"?wait=30", 10 * time.Second, false},
		{"?wait=1e300", 10 * time.Second, false},
		{"?wait=-1", 0, true},
		{"?wait=NaN", 0, true},
		{"?wait=Inf", 0, true},
		{"?wait=%2BInf", 0, true},
		{"?wait=-Inf", 0, true},
		{"?wait=no", 0, true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("query_%s", test.query), func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://localhost/run"+test.query, nil)
			require.NoError(t, err)
			got, err := parseDialtestingDebugWait(req)
			if test.wantErr {
				require.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
				assert.Equal(t, "datakit.InvalidTask", dialtestingDebugErrorCode(t, err))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	cfg := testDialtestingDebugConfig()
	cfg.maxLongPollWait = time.Second
	manager := newDialtestingDebugManager(cfg)
	defaultDialtestingDebugManager = manager
	t.Cleanup(manager.close)
	for _, query := range []string{"", "?wait=5"} {
		req, err := http.NewRequest(http.MethodGet, "http://localhost/run"+query, nil)
		require.NoError(t, err)
		got, err := parseDialtestingDebugWait(req)
		require.NoError(t, err)
		assert.Equal(t, time.Second, got)
	}
}

func TestNormalizeDialtestingDebugRequestID(t *testing.T) {
	got, err := normalizeDialtestingDebugRequestID("2DF08A9C-D949-4DB9-BF6F-BC4992D4C140")
	require.NoError(t, err)
	assert.Equal(t, "2df08a9c-d949-4db9-bf6f-bc4992d4c140", got)
	_, err = normalizeDialtestingDebugRequestID("not-a-uuid")
	assert.Error(t, err)
	_, err = normalizeDialtestingDebugRequestID("f47ac10b-58cc-11cf-a447-001122334455")
	assert.Error(t, err)
}

func TestNormalizeDialtestingDebugRunID(t *testing.T) {
	got, err := normalizeDialtestingDebugRunID("2DF08A9C-D949-4DB9-BF6F-BC4992D4C140")
	require.NoError(t, err)
	assert.Equal(t, "2df08a9c-d949-4db9-bf6f-bc4992d4c140", got)
	for _, runID := range []string{"", "not-a-uuid", "f47ac10b-58cc-11cf-a447-001122334455"} {
		_, err := normalizeDialtestingDebugRunID(runID)
		assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
		assert.Equal(t, "datakit.InvalidTask", dialtestingDebugErrorCode(t, err))
	}
}

func TestDialtestingDebugGetRouteAllowsRemoteWhitelist(t *testing.T) {
	router := gin.New()
	router.Use(apiWhiteListMiddleware([]string{"/v1/dialtesting/debug/runs"}))
	router.GET("/v1/dialtesting/debug/runs", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet,
		"/v1/dialtesting/debug/runs?run_id=2df08a9c-d949-4db9-bf6f-bc4992d4c140&wait=5", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestNormalizedDialtestingDebugPayloadHash(t *testing.T) {
	first := &dialtestingDebugRequest{
		TaskType: "http",
		Task:     map[string]interface{}{"url": "https://example.com", "method": "GET"},
		Regions:  []string{"region-a"},
	}
	second := &dialtestingDebugRequest{
		TaskType: "HTTP",
		Task:     map[string]interface{}{"method": "GET", "url": "https://example.com"},
		Regions:  []string{"region-a"},
	}
	firstHash, err := normalizedDialtestingDebugPayloadHash(first)
	require.NoError(t, err)
	secondHash, err := normalizedDialtestingDebugPayloadHash(second)
	require.NoError(t, err)
	assert.Equal(t, firstHash, secondHash)
	second.Regions = []string{"region-b"}
	secondHash, err = normalizedDialtestingDebugPayloadHash(second)
	require.NoError(t, err)
	assert.NotEqual(t, firstHash, secondHash)
}

func TestDialtestingDebugStoppedManagerHTTPMapping(t *testing.T) {
	manager := newDialtestingDebugManager(testDialtestingDebugConfig())
	previousManager := defaultDialtestingDebugManager
	defaultDialtestingDebugManager = manager
	t.Cleanup(func() {
		manager.close()
		defaultDialtestingDebugManager = previousManager
	})
	manager.close()

	body, err := json.Marshal(&dialtestingDebugRequest{
		RequestID: "2df08a9c-d949-4db9-bf6f-bc4992d4c140",
		Type:      "http",
		Task:      map[string]interface{}{},
	})
	require.NoError(t, err)
	createReq, err := http.NewRequest(
		http.MethodPost, "http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
	require.NoError(t, err)
	createReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	_, err = apiCreateDialtestingDebugRun(nil, createReq)
	require.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, getStatusCode(err))
	assert.Equal(t, "datakit.ServerUnavailable", dialtestingDebugErrorCode(t, err))

	runID := "e02a201f-3885-4edd-b77c-5e7ca34808a4"
	manager.mu.Lock()
	manager.runs[runID] = &dialtestingDebugRun{
		id:        runID,
		owner:     "workspace:account",
		requestID: "request",
		status:    dialtestingDebugStatusFailed,
		err:       errDialtestingDebugManagerStopped,
		expiresAt: time.Now().Add(time.Minute),
		notify:    make(chan struct{}),
	}
	manager.mu.Unlock()
	getReq, err := http.NewRequest(
		http.MethodGet, "http://localhost/v1/dialtesting/debug/runs?run_id="+runID+"&wait=0", nil)
	require.NoError(t, err)
	getReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	_, err = apiGetDialtestingDebugRun(nil, getReq)
	require.Error(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, getStatusCode(err))
	assert.Equal(t, "datakit.ServerUnavailable", dialtestingDebugErrorCode(t, err))
}

func TestDialtestingDebugAsyncHandlers(t *testing.T) {
	manager := newDialtestingDebugManager(testDialtestingDebugConfig())
	previousManager := defaultDialtestingDebugManager
	defaultDialtestingDebugManager = manager
	previousFields := debugFields
	previousRunErr := errRun
	debugFields = map[string]interface{}{"response_time": int64(1)}
	errRun = nil
	t.Cleanup(func() {
		manager.close()
		defaultDialtestingDebugManager = previousManager
		debugFields = previousFields
		errRun = previousRunErr
	})

	body, err := json.Marshal(&dialtestingDebugRequest{
		RequestID: "2df08a9c-d949-4db9-bf6f-bc4992d4c140",
		Type:      "http",
		Task:      map[string]interface{}{},
	})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	createdRaw, err := apiCreateDialtestingDebugRun(nil, req)
	require.NoError(t, err)
	created := createdRaw.(*dialtestingDebugRunResponse)
	require.NotEmpty(t, created.RunID)
	for _, requestID := range []string{
		"2DF08A9C-D949-4DB9-BF6F-BC4992D4C140",
		"2df08a9cd9494db9bf6fbc4992d4c140",
		"urn:uuid:2df08a9c-d949-4db9-bf6f-bc4992d4c140",
		"{2df08a9c-d949-4db9-bf6f-bc4992d4c140}",
	} {
		retryBody := bytes.Replace(body,
			[]byte("2df08a9c-d949-4db9-bf6f-bc4992d4c140"), []byte(requestID), 1)
		retryReq, err := http.NewRequest(http.MethodPost,
			"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(retryBody))
		require.NoError(t, err)
		retryReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
		retryRaw, err := apiCreateDialtestingDebugRun(nil, retryReq)
		require.NoError(t, err)
		assert.Equal(t, created.RunID, retryRaw.(*dialtestingDebugRunResponse).RunID)
	}
	conflictBody, err := json.Marshal(&dialtestingDebugRequest{
		RequestID: strings.ToUpper("2df08a9c-d949-4db9-bf6f-bc4992d4c140"),
		Type:      "http",
		Task:      map[string]interface{}{"url": "https://different.example.com"},
	})
	require.NoError(t, err)
	conflictReq, err := http.NewRequest(http.MethodPost,
		"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(conflictBody))
	require.NoError(t, err)
	conflictReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	_, err = apiCreateDialtestingDebugRun(nil, conflictReq)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
	assert.Equal(t, "datakit.InvalidTask", dialtestingDebugErrorCode(t, err))
	waitForDialtestingDebugStatus(t, manager, "workspace:account", created.RunID, dialtestingDebugStatusCompleted)

	getReq, err := http.NewRequest(http.MethodGet,
		"http://localhost/v1/dialtesting/debug/runs?run_id="+created.RunID+"&wait=0", nil)
	require.NoError(t, err)
	getReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	gotRaw, err := apiGetDialtestingDebugRun(nil, getReq)
	require.NoError(t, err)
	got := gotRaw.(*dialtestingDebugRunResponse)
	assert.Equal(t, dialtestingDebugStatusCompleted, got.Status)
	require.NotNil(t, got.Result)
	assert.Equal(t, "success", got.Result.Status)

	getReq.Header.Set(dialtestingDebugOwnerHeader, "another:owner")
	_, err = apiGetDialtestingDebugRun(nil, getReq)
	assert.Equal(t, http.StatusNotFound, getStatusCode(err))

	errRun = errors.New("executor failed")
	failedBody := bytes.Replace(body,
		[]byte("2df08a9c-d949-4db9-bf6f-bc4992d4c140"),
		[]byte("e02a201f-3885-4edd-b77c-5e7ca34808a4"), 1)
	failedReq, err := http.NewRequest(http.MethodPost, "http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(failedBody))
	require.NoError(t, err)
	failedReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	failedRaw, err := apiCreateDialtestingDebugRun(nil, failedReq)
	require.NoError(t, err)
	failed := failedRaw.(*dialtestingDebugRunResponse)
	waitForDialtestingDebugStatus(t, manager, "workspace:account", failed.RunID, dialtestingDebugStatusFailed)
	failedGet, err := http.NewRequest(http.MethodGet,
		"http://localhost/v1/dialtesting/debug/runs?run_id="+failed.RunID+"&wait=0", nil)
	require.NoError(t, err)
	failedGet.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	_, err = apiGetDialtestingDebugRun(nil, failedGet)
	assert.Equal(t, http.StatusInternalServerError, getStatusCode(err))
}

func TestDialtestingDebugAsyncNetPathKeepsExecutionContext(t *testing.T) {
	manager := newDialtestingDebugManager(testDialtestingDebugConfig())
	previousManager := defaultDialtestingDebugManager
	previousSetup := dialtestingNetPathDebugTaskSetup
	previousFields := debugFields
	previousRunErr := errRun
	previousDisableInternalNetwork := DialtestingDisableInternalNetworkTask
	defaultDialtestingDebugManager = manager
	debugFields = map[string]interface{}{
		"fail_reason": "",
		"traceroute":  `{"runs":[]}`,
	}
	errRun = nil
	DialtestingDisableInternalNetworkTask = false
	var executionContext context.Context
	dialtestingNetPathDebugTaskSetup = func(ctx context.Context, _ *dt.NetPathTask) {
		executionContext = ctx
	}
	t.Cleanup(func() {
		manager.close()
		defaultDialtestingDebugManager = previousManager
		dialtestingNetPathDebugTaskSetup = previousSetup
		debugFields = previousFields
		errRun = previousRunErr
		DialtestingDisableInternalNetworkTask = previousDisableInternalNetwork
	})

	body, err := json.Marshal(&dialtestingDebugRequest{
		RequestID: "1c03835a-c8d5-49c0-80f9-6048a7f5d560",
		Type:      "netpath",
		Task: &dt.NetPathTask{
			Task:     &dt.Task{ExternalID: "netpath-debug", Name: "netpath-debug"},
			Protocol: "tcp",
			Host:     "example.com",
			Port:     "443",
		},
	})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost,
		"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")

	createdRaw, err := apiCreateDialtestingDebugRun(nil, req)
	require.NoError(t, err)
	require.NotNil(t, executionContext)
	assert.NoError(t, executionContext.Err())
	created := createdRaw.(*dialtestingDebugRunResponse)
	waitForDialtestingDebugStatus(t, manager, "workspace:account", created.RunID, dialtestingDebugStatusCompleted)
}

func TestDialtestingDebugAsyncValidatesDestinationAtExecution(t *testing.T) {
	manager := newDialtestingDebugManager(testDialtestingDebugConfig())
	previousManager := defaultDialtestingDebugManager
	previousDisableInternalNetwork := DialtestingDisableInternalNetworkTask
	previousRunErr := errRun
	defaultDialtestingDebugManager = manager
	DialtestingDisableInternalNetworkTask = true
	errRun = nil
	t.Cleanup(func() {
		manager.close()
		defaultDialtestingDebugManager = previousManager
		DialtestingDisableInternalNetworkTask = previousDisableInternalNetwork
		errRun = previousRunErr
	})

	body, err := json.Marshal(&dialtestingDebugRequest{
		RequestID: "7154c5e7-c056-4452-862f-ad6d50fb32a0",
		Type:      "icmp",
		Task:      map[string]interface{}{"host": "127.0.0.1"},
	})
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost,
		"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")

	createdRaw, err := apiCreateDialtestingDebugRun(nil, req)
	require.NoError(t, err)
	created := createdRaw.(*dialtestingDebugRunResponse)
	failed := waitForDialtestingDebugStatus(
		t, manager, "workspace:account", created.RunID, dialtestingDebugStatusFailed)
	assert.Equal(t, "InternalNetworkDenied", failed.Error)

	getReq, err := http.NewRequest(http.MethodGet,
		"http://localhost/v1/dialtesting/debug/runs?run_id="+created.RunID+"&wait=0", nil)
	require.NoError(t, err)
	getReq.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
	_, err = apiGetDialtestingDebugRun(nil, getReq)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
	assert.Equal(t, "datakit.InternalNetworkDenied", dialtestingDebugErrorCode(t, err))
	assert.ErrorContains(t, err, dialtestingInternalNetworkDeniedMessage)
}

func TestDialtestingDebugCreateHandlerErrorMapping(t *testing.T) {
	t.Run("task validation is InvalidTask", func(t *testing.T) {
		manager := newDialtestingDebugManager(testDialtestingDebugConfig())
		previousManager := defaultDialtestingDebugManager
		previousMock := defDialtestingMock
		defaultDialtestingDebugManager = manager
		defDialtestingMock = &prodDialtestingMock{}
		t.Cleanup(func() {
			manager.close()
			defaultDialtestingDebugManager = previousManager
			defDialtestingMock = previousMock
		})

		body, err := json.Marshal(&dialtestingDebugRequest{
			RequestID: "eaa04ae8-b8a8-458f-92f4-d11bf660fe78",
			Type:      "http",
			Task:      map[string]interface{}{},
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost,
			"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
		_, err = apiCreateDialtestingDebugRun(nil, req)
		require.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, getStatusCode(err))
		assert.Equal(t, "datakit.InvalidTask", dialtestingDebugErrorCode(t, err))
	})

	t.Run("internal error is ExecutionFailed", func(t *testing.T) {
		manager := newDialtestingDebugManager(testDialtestingDebugConfig())
		manager.newRunID = func() (string, error) { return "", errors.New("random source unavailable") }
		previousManager := defaultDialtestingDebugManager
		previousInitErr := errInit
		defaultDialtestingDebugManager = manager
		errInit = nil
		t.Cleanup(func() {
			manager.close()
			defaultDialtestingDebugManager = previousManager
			errInit = previousInitErr
		})

		body, err := json.Marshal(&dialtestingDebugRequest{
			RequestID: "a0ec9f2d-1522-4b54-b970-a355fb4e8529",
			Type:      "http",
			Task:      map[string]interface{}{},
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost,
			"http://localhost/v1/dialtesting/debug/runs", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set(dialtestingDebugOwnerHeader, "workspace:account")
		_, err = apiCreateDialtestingDebugRun(nil, req)
		require.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, getStatusCode(err))
		assert.Equal(t, "datakit.ExecutionFailed", dialtestingDebugErrorCode(t, err))
	})
}
