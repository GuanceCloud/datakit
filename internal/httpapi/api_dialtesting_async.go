// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/google/uuid"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	dialtestingDebugOwnerHeader = "X-Dialing-Debug-Owner"

	defaultDialtestingDebugMaxConcurrentRuns       = 100
	defaultDialtestingDebugMaxQueuedRuns           = 1000
	defaultDialtestingDebugMaxQueueWait            = 3 * time.Minute
	defaultDialtestingDebugResultTTL               = 10 * time.Minute
	defaultDialtestingDebugMaxLongPollWait         = 10 * time.Second
	defaultDialtestingDebugMaxRetainedTerminalRuns = 2000
	dialtestingDebugAbsoluteMaxLongPollWait        = 10 * time.Second
	dialtestingDebugDefaultWaitSeconds             = 5
)

const (
	dialtestingDebugStatusPending   = "pending"
	dialtestingDebugStatusRunning   = "running"
	dialtestingDebugStatusCompleted = "completed"
	dialtestingDebugStatusTimedOut  = "timed_out"
	dialtestingDebugStatusFailed    = "failed"
)

var (
	errDialtestingDebugQueueFull      = errors.New("dialtesting debug queue is full")
	errDialtestingDebugManagerStopped = errors.New("dialtesting debug manager is stopped")
)

type dialtestingDebugConfig struct {
	maxConcurrentRuns       int
	maxQueuedRuns           int
	maxQueueWait            time.Duration
	resultTTL               time.Duration
	maxLongPollWait         time.Duration
	maxRetainedTerminalRuns int
}

func defaultDialtestingDebugConfig() dialtestingDebugConfig {
	return dialtestingDebugConfig{
		maxConcurrentRuns:       defaultDialtestingDebugMaxConcurrentRuns,
		maxQueuedRuns:           defaultDialtestingDebugMaxQueuedRuns,
		maxQueueWait:            defaultDialtestingDebugMaxQueueWait,
		resultTTL:               defaultDialtestingDebugResultTTL,
		maxLongPollWait:         defaultDialtestingDebugMaxLongPollWait,
		maxRetainedTerminalRuns: defaultDialtestingDebugMaxRetainedTerminalRuns,
	}
}

func loadDialtestingDebugConfigFromEnv() dialtestingDebugConfig {
	cfg := defaultDialtestingDebugConfig()
	cfg.maxConcurrentRuns = parseDialtestingDebugPositiveIntEnv(
		"ENV_DIALTESTING_DEBUG_MAX_CONCURRENT_RUNS", cfg.maxConcurrentRuns)
	cfg.maxQueuedRuns = parseDialtestingDebugPositiveIntEnv(
		"ENV_DIALTESTING_DEBUG_MAX_QUEUED_RUNS", cfg.maxQueuedRuns)
	cfg.maxQueueWait = parseDialtestingDebugDurationEnv(
		"ENV_DIALTESTING_DEBUG_MAX_QUEUE_WAIT", cfg.maxQueueWait)
	cfg.resultTTL = parseDialtestingDebugDurationEnv(
		"ENV_DIALTESTING_DEBUG_RESULT_TTL", cfg.resultTTL)
	cfg.maxLongPollWait = parseDialtestingDebugDurationEnv(
		"ENV_DIALTESTING_DEBUG_MAX_LONG_POLL_WAIT", cfg.maxLongPollWait)
	if cfg.maxLongPollWait > dialtestingDebugAbsoluteMaxLongPollWait {
		l.Warnf("ENV_DIALTESTING_DEBUG_MAX_LONG_POLL_WAIT exceeds 10s, truncated to 10s")
		cfg.maxLongPollWait = dialtestingDebugAbsoluteMaxLongPollWait
	}
	cfg.maxRetainedTerminalRuns = parseDialtestingDebugPositiveIntEnv(
		"ENV_DIALTESTING_DEBUG_MAX_RETAINED_TERMINAL_RUNS", cfg.maxRetainedTerminalRuns)
	return cfg
}

func parseDialtestingDebugPositiveIntEnv(name string, fallback int) int {
	value := datakit.GetEnv(name)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		l.Warnf("invalid %s[%s], use default %d", name, value, fallback)
		return fallback
	}
	return n
}

func parseDialtestingDebugDurationEnv(name string, fallback time.Duration) time.Duration {
	value := datakit.GetEnv(name)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		l.Warnf("invalid %s[%s], use default %s", name, value, fallback)
		return fallback
	}
	return d
}

type dialtestingDebugRun struct {
	id          string
	owner       string
	requestID   string
	payloadHash string
	taskType    string
	traceID     string
	createdAt   time.Time
	startedAt   time.Time
	completedAt time.Time
	expiresAt   time.Time
	status      string
	result      *dialtestingDebugResponse
	err         error
	execute     func() (*dialtestingDebugResponse, error)
	notify      chan struct{}
}

type dialtestingDebugRunResponse struct {
	RunID     string                    `json:"run_id"`
	Status    string                    `json:"status"`
	Phase     string                    `json:"phase,omitempty"`
	Result    *dialtestingDebugResponse `json:"result,omitempty"`
	Error     string                    `json:"error,omitempty"`
	ExpiresAt int64                     `json:"expires_at"`
}

type dialtestingDebugPreparation struct {
	payloadHash string
	done        chan struct{}
	response    *dialtestingDebugRunResponse
	err         error
}

type dialtestingDebugPrepare func(context.Context) (func() (*dialtestingDebugResponse, error), error)

type dialtestingDebugManager struct {
	mu            sync.Mutex
	cfg           dialtestingDebugConfig
	runs          map[string]*dialtestingDebugRun
	idempotency   map[string]string
	preparing     map[string]*dialtestingDebugPreparation
	queue         []*dialtestingDebugRun
	terminal      []*dialtestingDebugRun
	terminalCount int
	queueCond     *sync.Cond
	stop          chan struct{}
	stopOnce      sync.Once
	stopped       bool
	now           func() time.Time
	newRunID      func() (string, error)
}

var defaultDialtestingDebugManager *dialtestingDebugManager

func newDialtestingDebugManager(cfg dialtestingDebugConfig) *dialtestingDebugManager {
	m := &dialtestingDebugManager{
		cfg:         cfg,
		runs:        make(map[string]*dialtestingDebugRun),
		idempotency: make(map[string]string),
		preparing:   make(map[string]*dialtestingDebugPreparation),
		queue:       make([]*dialtestingDebugRun, 0, cfg.maxQueuedRuns),
		stop:        make(chan struct{}),
		now:         time.Now,
		newRunID: func() (string, error) {
			id, err := uuid.NewRandom()
			return id.String(), err
		},
	}
	m.queueCond = sync.NewCond(&m.mu)
	for i := 0; i < cfg.maxConcurrentRuns; i++ {
		go m.worker()
	}
	go m.cleaner()
	return m
}

func (m *dialtestingDebugManager) close() {
	m.stopOnce.Do(func() {
		m.mu.Lock()
		m.stopped = true
		for key, preparation := range m.preparing {
			m.completePreparationLocked(key, preparation, nil, errDialtestingDebugManagerStopped)
		}
		now := m.now()
		for _, run := range m.queue {
			if run.status == dialtestingDebugStatusPending {
				m.finishLocked(run, dialtestingDebugStatusFailed, nil, errDialtestingDebugManagerStopped, now)
			}
		}
		for i := range m.queue {
			m.queue[i] = nil
		}
		m.queue = m.queue[:0]
		dialtestingDebugQueueSize.Set(0)
		close(m.stop)
		m.queueCond.Broadcast()
		m.mu.Unlock()
	})
}

func (m *dialtestingDebugManager) submit(
	ctx context.Context,
	owner, requestID, payloadHash, taskType, traceID string,
	prepare dialtestingDebugPrepare,
) (*dialtestingDebugRunResponse, error) {
	key := owner + "\x00" + requestID

	// Existing and in-flight lookups avoid repeating preparation for retries.
	// Preparation remains outside the manager lock because task validation can do
	// DNS and other comparatively slow work.
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return nil, errDialtestingDebugManagerStopped
	}
	if runID, ok := m.idempotency[key]; ok {
		run := m.runs[runID]
		if run != nil {
			m.cleanupRunLocked(run, m.now())
			run = m.runs[runID]
		}
		if run != nil && run.payloadHash == payloadHash {
			response := snapshotDialtestingDebugRun(run)
			m.mu.Unlock()
			return response, nil
		}
		if run != nil {
			m.mu.Unlock()
			return nil, uhttp.Error(ErrDialtestingInvalidTask, "request_id already used with a different payload")
		}
		delete(m.idempotency, key)
	}
	if preparation, ok := m.preparing[key]; ok {
		if preparation.payloadHash != payloadHash {
			m.mu.Unlock()
			return nil, uhttp.Error(ErrDialtestingInvalidTask, "request_id already used with a different payload")
		}
		done := preparation.done
		m.mu.Unlock()
		select {
		case <-done:
			return preparation.response, preparation.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	m.cleanupForSubmitLocked(m.now())
	if len(m.queue)+len(m.preparing) >= m.cfg.maxQueuedRuns {
		m.mu.Unlock()
		return nil, errDialtestingDebugQueueFull
	}
	preparation := &dialtestingDebugPreparation{
		payloadHash: payloadHash,
		done:        make(chan struct{}),
	}
	m.preparing[key] = preparation
	dialtestingDebugPreparingSize.Set(float64(len(m.preparing)))
	m.mu.Unlock()

	execute, prepareErr := runDialtestingDebugPreparation(prepare)
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		m.completePreparationLocked(key, preparation, nil, errDialtestingDebugManagerStopped)
		return nil, errDialtestingDebugManagerStopped
	}
	if prepareErr != nil {
		m.completePreparationLocked(key, preparation, nil, prepareErr)
		return nil, prepareErr
	}
	m.cleanupForSubmitLocked(now)
	runID, err := m.newRunID()
	if err != nil {
		err = fmt.Errorf("generate run_id: %w", err)
		m.completePreparationLocked(key, preparation, nil, err)
		return nil, err
	}
	run := &dialtestingDebugRun{
		id:          runID,
		owner:       owner,
		requestID:   requestID,
		payloadHash: payloadHash,
		taskType:    taskType,
		traceID:     traceID,
		createdAt:   now,
		expiresAt:   now.Add(m.cfg.maxQueueWait + m.cfg.resultTTL),
		status:      dialtestingDebugStatusPending,
		execute:     execute,
		notify:      make(chan struct{}),
	}
	m.runs[runID] = run
	m.idempotency[key] = runID
	m.queue = append(m.queue, run)
	m.queueCond.Signal()
	dialtestingDebugRunsTotal.WithLabelValues(run.taskType, run.status).Inc()
	dialtestingDebugQueueSize.Set(float64(len(m.queue)))
	l.Infof("dialtesting debug run created: run_id=%s request_id=%s type=%s status=%s duration=0s trace_id=%s",
		run.id, run.requestID, run.taskType, run.status, run.traceID)
	response := snapshotDialtestingDebugRun(run)
	m.completePreparationLocked(key, preparation, response, nil)
	return response, nil
}

func runDialtestingDebugPreparation(
	prepare dialtestingDebugPrepare,
) (execute func() (*dialtestingDebugResponse, error), err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			execute = nil
			err = fmt.Errorf("dialtesting debug preparation panic: %v", recovered)
		}
	}()
	return prepare(context.Background())
}

func (m *dialtestingDebugManager) completePreparationLocked(
	key string,
	preparation *dialtestingDebugPreparation,
	response *dialtestingDebugRunResponse,
	err error,
) {
	if m.preparing[key] != preparation {
		return
	}
	preparation.response = response
	preparation.err = err
	delete(m.preparing, key)
	dialtestingDebugPreparingSize.Set(float64(len(m.preparing)))
	close(preparation.done)
}

func (m *dialtestingDebugManager) get(ctx context.Context, owner, runID string, wait time.Duration) (*dialtestingDebugRunResponse, error) {
	now := m.now()
	m.mu.Lock()
	run := m.runs[runID]
	if run != nil && run.owner == owner {
		m.cleanupRunLocked(run, now)
		run = m.runs[runID]
	}
	if run == nil || run.owner != owner {
		m.mu.Unlock()
		return nil, uhttp.Error(ErrDialtestingRunNotFound, "run not found")
	}
	response := snapshotDialtestingDebugRun(run)
	if wait <= 0 || isDialtestingDebugTerminal(run.status) {
		m.mu.Unlock()
		return response, nil
	}
	notify := run.notify
	m.mu.Unlock()

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
	case <-notify:
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	run = m.runs[runID]
	if run != nil && run.owner == owner {
		m.cleanupRunLocked(run, m.now())
		run = m.runs[runID]
	}
	if run == nil || run.owner != owner {
		return nil, uhttp.Error(ErrDialtestingRunNotFound, "run not found")
	}
	return snapshotDialtestingDebugRun(run), nil
}

func (m *dialtestingDebugManager) worker() {
	for {
		m.mu.Lock()
		for len(m.queue) == 0 && !m.stopped {
			m.queueCond.Wait()
		}
		if m.stopped {
			m.mu.Unlock()
			return
		}
		run := m.queue[0]
		m.queue[0] = nil
		m.queue = m.queue[1:]
		dialtestingDebugQueueSize.Set(float64(len(m.queue)))
		m.mu.Unlock()
		m.executeRun(run)
	}
}

func (m *dialtestingDebugManager) executeRun(run *dialtestingDebugRun) {
	now := m.now()
	m.mu.Lock()
	if run.status != dialtestingDebugStatusPending {
		m.mu.Unlock()
		return
	}
	if m.stopped {
		m.finishLocked(run, dialtestingDebugStatusFailed, nil, errDialtestingDebugManagerStopped, now)
		m.mu.Unlock()
		return
	}
	if now.Sub(run.createdAt) > m.cfg.maxQueueWait {
		m.finishLocked(run, dialtestingDebugStatusTimedOut, nil, nil, now)
		m.mu.Unlock()
		logDialtestingDebugTimedOutRun(run, now)
		return
	}
	run.status = dialtestingDebugStatusRunning
	run.startedAt = now
	dialtestingDebugRunsTotal.WithLabelValues(run.taskType, run.status).Inc()
	dialtestingDebugRunsActive.WithLabelValues(run.taskType).Inc()
	m.notifyLocked(run)
	m.mu.Unlock()

	start := time.Now()
	var result *dialtestingDebugResponse
	var runErr error
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				runErr = fmt.Errorf("dialtesting debug executor panic: %v", recovered)
			}
		}()
		result, runErr = run.execute()
	}()

	m.mu.Lock()
	if runErr != nil {
		m.finishLocked(run, dialtestingDebugStatusFailed, nil, runErr, m.now())
	} else {
		m.finishLocked(run, dialtestingDebugStatusCompleted, result, nil, m.now())
	}
	status := run.status
	m.mu.Unlock()
	dialtestingDebugRunsActive.WithLabelValues(run.taskType).Dec()
	dialtestingDebugDurationSeconds.WithLabelValues(run.taskType).Observe(time.Since(start).Seconds())
	l.Infof("dialtesting debug run finished: run_id=%s request_id=%s type=%s status=%s duration=%s trace_id=%s",
		run.id, run.requestID, run.taskType, status, time.Since(start), run.traceID)
}

func (m *dialtestingDebugManager) finishLocked(
	run *dialtestingDebugRun, status string, result *dialtestingDebugResponse, runErr error, now time.Time,
) {
	run.status = status
	run.result = result
	run.err = runErr
	run.completedAt = now
	run.expiresAt = now.Add(m.cfg.resultTTL)
	run.execute = nil
	dialtestingDebugRunsTotal.WithLabelValues(run.taskType, run.status).Inc()
	m.notifyLocked(run)
	m.terminal = append(m.terminal, run)
	m.terminalCount++
	m.enforceTerminalLimitLocked()
}

func (m *dialtestingDebugManager) notifyLocked(run *dialtestingDebugRun) {
	close(run.notify)
	run.notify = make(chan struct{})
}

func (m *dialtestingDebugManager) cleaner() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.mu.Lock()
			m.cleanupForSubmitLocked(m.now())
			m.mu.Unlock()
		}
	}
}

func (m *dialtestingDebugManager) cleanupForSubmitLocked(now time.Time) {
	m.expireQueuedLocked(now)
	m.cleanupTerminalLocked(now)
}

func (m *dialtestingDebugManager) cleanupRunLocked(run *dialtestingDebugRun, now time.Time) {
	if run.status == dialtestingDebugStatusPending && now.Sub(run.createdAt) > m.cfg.maxQueueWait {
		m.removeQueuedRunLocked(run)
		m.finishLocked(run, dialtestingDebugStatusTimedOut, nil, nil, now)
		logDialtestingDebugTimedOutRun(run, now)
	}
	if isDialtestingDebugTerminal(run.status) && !now.Before(run.expiresAt) {
		m.removeRunLocked(run.id, run)
		dialtestingDebugResultExpired.Inc()
	}
}

func (m *dialtestingDebugManager) expireQueuedLocked(now time.Time) {
	kept := m.queue[:0]
	for _, run := range m.queue {
		if run.status == dialtestingDebugStatusPending && now.Sub(run.createdAt) > m.cfg.maxQueueWait {
			m.finishLocked(run, dialtestingDebugStatusTimedOut, nil, nil, now)
			logDialtestingDebugTimedOutRun(run, now)
			continue
		}
		kept = append(kept, run)
	}
	for i := len(kept); i < len(m.queue); i++ {
		m.queue[i] = nil
	}
	m.queue = kept
	dialtestingDebugQueueSize.Set(float64(len(m.queue)))
}

func (m *dialtestingDebugManager) removeQueuedRunLocked(target *dialtestingDebugRun) {
	for i, run := range m.queue {
		if run != target {
			continue
		}
		copy(m.queue[i:], m.queue[i+1:])
		m.queue[len(m.queue)-1] = nil
		m.queue = m.queue[:len(m.queue)-1]
		dialtestingDebugQueueSize.Set(float64(len(m.queue)))
		return
	}
}

func (m *dialtestingDebugManager) cleanupTerminalLocked(now time.Time) {
	for len(m.terminal) > 0 {
		run := m.terminal[0]
		if m.runs[run.id] != run {
			m.terminal = m.terminal[1:]
			continue
		}
		if now.Before(run.expiresAt) {
			break
		}
		m.removeRunLocked(run.id, run)
		m.terminal = m.terminal[1:]
		dialtestingDebugResultExpired.Inc()
	}
}

func (m *dialtestingDebugManager) enforceTerminalLimitLocked() {
	for m.terminalCount > m.cfg.maxRetainedTerminalRuns && len(m.terminal) > 0 {
		run := m.terminal[0]
		m.terminal = m.terminal[1:]
		if m.runs[run.id] == run {
			m.removeRunLocked(run.id, run)
		}
	}
}

func logDialtestingDebugTimedOutRun(run *dialtestingDebugRun, now time.Time) {
	l.Infof("dialtesting debug run finished: run_id=%s request_id=%s type=%s status=%s duration=%s trace_id=%s",
		run.id, run.requestID, run.taskType, run.status, now.Sub(run.createdAt), run.traceID)
}

func (m *dialtestingDebugManager) removeRunLocked(runID string, run *dialtestingDebugRun) {
	if m.runs[runID] != run {
		return
	}
	delete(m.runs, runID)
	if isDialtestingDebugTerminal(run.status) {
		m.terminalCount--
	}
	key := run.owner + "\x00" + run.requestID
	if m.idempotency[key] == runID {
		delete(m.idempotency, key)
	}
}

func isDialtestingDebugTerminal(status string) bool {
	return status == dialtestingDebugStatusCompleted || status == dialtestingDebugStatusTimedOut || status == dialtestingDebugStatusFailed
}

func snapshotDialtestingDebugRun(run *dialtestingDebugRun) *dialtestingDebugRunResponse {
	response := &dialtestingDebugRunResponse{
		RunID:     run.id,
		Status:    run.status,
		Result:    run.result,
		ExpiresAt: run.expiresAt.Unix(),
	}
	if run.status == dialtestingDebugStatusRunning {
		response.Phase = "probing"
	}
	if run.err != nil {
		if isDialtestingDebugError(run.err, ErrDialtestingInternalNetworkDenied.ErrCode) {
			response.Error = "InternalNetworkDenied"
		} else if errors.Is(run.err, errDialtestingDebugManagerStopped) {
			response.Error = "ServerUnavailable"
		} else {
			response.Error = "ExecutionFailed"
		}
	}
	return response
}

func normalizedDialtestingDebugPayloadHash(req *dialtestingDebugRequest) (string, error) {
	payload := struct {
		Task      interface{} `json:"task"`
		TaskType  string      `json:"task_type"`
		Regions   []string    `json:"regions,omitempty"`
		Variables interface{} `json:"variables"`
	}{
		Task: req.Task, TaskType: strings.ToUpper(req.TaskType), Regions: req.Regions, Variables: req.Variables,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeDialtestingDebugRequestID(requestID string) (string, error) {
	id, err := uuid.Parse(requestID)
	if err != nil || id.Version() != 4 {
		return "", uhttp.Error(ErrDialtestingInvalidTask, "request_id must be a UUID v4")
	}
	return id.String(), nil
}

func apiCreateDialtestingDebugRun(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	owner := strings.TrimSpace(req.Header.Get(dialtestingDebugOwnerHeader))
	if owner == "" {
		return nil, uhttp.Error(ErrDialtestingInvalidTask, "missing X-Dialing-Debug-Owner")
	}
	reqDebug, err := getAPIDebugDialtestingRequest(req)
	if err != nil {
		return nil, uhttp.Error(ErrDialtestingInvalidTask, err.Error())
	}
	reqDebug.RequestID, err = normalizeDialtestingDebugRequestID(reqDebug.RequestID)
	if err != nil {
		return nil, err
	}
	payloadHash, err := normalizedDialtestingDebugPayloadHash(reqDebug)
	if err != nil {
		return nil, uhttp.Error(ErrDialtestingInvalidTask, err.Error())
	}
	manager := defaultDialtestingDebugManager
	if manager == nil {
		return nil, uhttp.Error(ErrDialtestingExecutionFailed, "dialtesting debug manager is not initialized")
	}
	tid := req.Header.Get(uhttp.XTraceID)
	response, err := manager.submit(req.Context(), owner, reqDebug.RequestID, payloadHash, strings.ToUpper(reqDebug.TaskType), tid,
		func(prepareCtx context.Context) (func() (*dialtestingDebugResponse, error), error) {
			prepared, err := prepareDialtestingDebug(prepareCtx, reqDebug, tid)
			if err != nil {
				return nil, uhttp.Error(ErrDialtestingInvalidTask, err.Error())
			}
			return func() (*dialtestingDebugResponse, error) {
				result, err := executeDialtestingDebug(prepared, tid)
				if err != nil {
					if errors.Is(err, errDialtestingInternalNetworkDenied) {
						return nil, uhttp.Error(
							ErrDialtestingInternalNetworkDenied,
							dialtestingInternalNetworkDeniedMessage,
						)
					}
					return nil, uhttp.Error(ErrDialtestingExecutionFailed, err.Error())
				}
				return result, nil
			}, nil
		})
	if errors.Is(err, errDialtestingDebugQueueFull) {
		return nil, uhttp.Error(ErrDialtestingQueueFull, "dialtesting debug queue is full")
	}
	if errors.Is(err, errDialtestingDebugManagerStopped) {
		return nil, uhttp.Error(ErrDialtestingUnavailable, "dialtesting debug manager is stopped")
	}
	if err != nil {
		if isDialtestingDebugHTTPError(err) {
			return nil, err
		}
		l.Errorf("[%s] create dialtesting debug run failed: %s", tid, err.Error())
		return nil, uhttp.Error(ErrDialtestingExecutionFailed, "create dialtesting debug run failed")
	}
	return response, nil
}

func isDialtestingDebugHTTPError(err error) bool {
	switch err.(type) {
	case *uhttp.HttpError, *uhttp.MsgError:
		return true
	default:
		return false
	}
}

func apiGetDialtestingDebugRun(w http.ResponseWriter, req *http.Request, whatever ...interface{}) (interface{}, error) {
	owner := strings.TrimSpace(req.Header.Get(dialtestingDebugOwnerHeader))
	if owner == "" {
		return nil, uhttp.Error(ErrDialtestingRunNotFound, "run not found")
	}
	runID, err := normalizeDialtestingDebugRunID(req.URL.Query().Get("run_id"))
	if err != nil {
		return nil, err
	}
	wait, err := parseDialtestingDebugWait(req)
	if err != nil {
		return nil, err
	}
	manager := defaultDialtestingDebugManager
	if manager == nil {
		return nil, uhttp.Error(ErrDialtestingRunNotFound, "run not found")
	}
	response, err := manager.get(req.Context(), owner, runID, wait)
	if err != nil {
		return nil, err
	}
	if response.Status == dialtestingDebugStatusFailed {
		if response.Error == "InternalNetworkDenied" {
			return nil, uhttp.Error(
				ErrDialtestingInternalNetworkDenied,
				dialtestingInternalNetworkDeniedMessage,
			)
		}
		if response.Error == "ServerUnavailable" {
			return nil, uhttp.Error(ErrDialtestingUnavailable, "dialtesting debug manager is stopped")
		}
		return nil, uhttp.Error(ErrDialtestingExecutionFailed, "dialtesting debug execution failed")
	}
	return response, nil
}

func isDialtestingDebugError(err error, errorCode string) bool {
	var msgErr *uhttp.MsgError
	return errors.As(err, &msgErr) && msgErr.ErrCode == errorCode
}

func normalizeDialtestingDebugRunID(runID string) (string, error) {
	id, err := uuid.Parse(strings.TrimSpace(runID))
	if err != nil || id.Version() != 4 {
		return "", uhttp.Error(ErrDialtestingInvalidTask, "run_id must be a UUID v4")
	}
	return id.String(), nil
}

func parseDialtestingDebugWait(req *http.Request) (time.Duration, error) {
	maxWait := dialtestingDebugAbsoluteMaxLongPollWait
	if defaultDialtestingDebugManager != nil && defaultDialtestingDebugManager.cfg.maxLongPollWait < maxWait {
		maxWait = defaultDialtestingDebugManager.cfg.maxLongPollWait
	}

	raw := req.URL.Query().Get("wait")
	wait := time.Duration(dialtestingDebugDefaultWaitSeconds) * time.Second
	if raw == "" {
		// Apply the configured cap to the default just like an explicit value.
	} else {
		seconds, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 {
			return 0, uhttp.Error(ErrDialtestingInvalidTask, "wait must be a finite non-negative number of seconds")
		}
		if seconds > maxWait.Seconds() {
			wait = maxWait
		} else {
			wait = time.Duration(seconds * float64(time.Second))
		}
	}
	if wait > maxWait {
		wait = maxWait
	}
	return wait, nil
}
