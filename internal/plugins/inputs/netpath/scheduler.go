// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"crypto/sha256"
	"hash/fnv"
	"math"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type taskContext struct {
	task                  task
	nextRun               time.Time
	runUntil              time.Time
	accountedBytes        int64
	incarnation           uint64
	dispatchGeneration    uint64
	outstandingGeneration uint64
}

type taskStoreKey [sha256.Size]byte

// scheduledTask deliberately contains no candidate-owned strings or maps, so
// an expired or updated context cannot leave its payload retained in a queue.
type scheduledTask struct {
	key                taskStoreKey
	source             string
	scheduledAt        time.Time
	incarnation        uint64
	dispatchGeneration uint64
}

type taskStore struct {
	mu                 sync.Mutex
	items              map[taskStoreKey]*taskContext
	dynamicItems       int
	dynamicLimit       int
	dynamicBytes       int64
	dynamicBytesLimit  int64
	inFlightBytes      int64
	inFlightBytesLimit int64
	nextIncarnation    uint64
}

func newTaskStore(limit int, bytesLimits ...int64) *taskStore {
	if limit <= 0 {
		limit = defaultDynamicContextsLimit
	}
	bytesLimit := defaultDynamicContextsBytesLimit
	if len(bytesLimits) > 0 && bytesLimits[0] > 0 {
		bytesLimit = bytesLimits[0]
	}
	// Reserve one worst-case context copy per worker. Store and in-flight
	// accounting then remain within the configured total even when a context is
	// updated or expires while its previous value is still running.
	inFlightBytesLimit := int64(defaultDynamicWorkers) * maxDynamicContextBytes
	if len(bytesLimits) > 1 && bytesLimits[1] > 0 {
		inFlightBytesLimit = bytesLimits[1]
	}
	if maxInFlightBytes := bytesLimit / 2; inFlightBytesLimit > maxInFlightBytes {
		inFlightBytesLimit = maxInFlightBytes
	}
	return &taskStore{
		items:              map[taskStoreKey]*taskContext{},
		dynamicLimit:       limit,
		dynamicBytesLimit:  bytesLimit - inFlightBytesLimit,
		inFlightBytesLimit: inFlightBytesLimit,
	}
}

func makeTaskStoreKey(scheduleKey string) taskStoreKey {
	return sha256.Sum256([]byte(scheduleKey))
}

// estimateDynamicTaskBytes conservatively accounts the memory retained only
// because a dynamic context stays in the store. Config tags and compiled
// filters are excluded because the scheduler config already owns them.
func estimateDynamicTaskBytes(t task) int64 {
	const (
		contextBaseBytes = int64(1024)
		tagMapBaseBytes  = int64(64)
		tagEntryBytes    = int64(64)
	)

	size := contextBaseBytes
	for _, value := range []string{
		t.ID, t.Name, t.Source, t.Origin, t.RunType, t.Target, t.Hostname,
		t.TargetIP, t.Protocol, t.Namespace, t.SourceContainerID, t.SourceHost,
		t.SourceIP, t.SourceProcess, t.SourceService, t.DstIP, t.NetNS,
		t.ScheduleKey,
	} {
		size += int64(len(value))
	}
	// Most filter strings share the task's retained backing storage. These two
	// can have different fallback precedence and are detached only when needed.
	if t.filterValues.sourceHost != t.SourceHost {
		size += int64(len(t.filterValues.sourceHost))
	}
	if t.filterValues.sourceContainer != t.SourceContainerID {
		size += int64(len(t.filterValues.sourceContainer))
	}
	for _, tags := range []map[string]string{t.RequestTags, t.Tags} {
		if len(tags) == 0 {
			continue
		}
		size += tagMapBaseBytes
		for key, value := range tags {
			size += tagEntryBytes + int64(len(key)) + int64(len(value))
		}
	}
	return size
}

func (s *taskStore) add(now time.Time, t task) (bool, string) {
	if t.ScheduleKey == "" {
		return false, "empty_key"
	}
	key := makeTaskStoreKey(t.ScheduleKey)
	if t.Interval <= 0 {
		t.Interval = defaultDynamicInterval
	}
	accountedBytes := int64(0)
	if t.Source == sourceDynamic {
		accountedBytes = estimateDynamicTaskBytes(t)
		if accountedBytes > maxDynamicContextBytes {
			return false, "context_too_large"
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	runUntil := time.Time{}
	if t.TTL > 0 {
		runUntil = now.Add(t.TTL)
	}

	if ctx, ok := s.items[key]; ok {
		if ctx.task.ScheduleKey != t.ScheduleKey {
			return false, "schedule_key_collision"
		}
		oldDynamic := ctx.task.Source == sourceDynamic
		newDynamic := t.Source == sourceDynamic
		if newDynamic && !oldDynamic && s.dynamicItems >= s.dynamicLimit {
			return false, "contexts_limit"
		}
		if accountedBytes > ctx.accountedBytes {
			increase := accountedBytes - ctx.accountedBytes
			if s.dynamicBytes >= s.dynamicBytesLimit || increase > s.dynamicBytesLimit-s.dynamicBytes {
				return false, "contexts_bytes_limit"
			}
		}
		if oldDynamic != newDynamic {
			if oldDynamic {
				s.dynamicItems--
			} else {
				s.dynamicItems++
			}
		}
		s.dynamicBytes -= ctx.accountedBytes
		s.dynamicBytes += accountedBytes
		ctx.task = t
		ctx.accountedBytes = accountedBytes
		if !runUntil.IsZero() {
			ctx.runUntil = runUntil
		}
		return true, ""
	}
	if t.Source == sourceDynamic && s.dynamicItems >= s.dynamicLimit {
		return false, "contexts_limit"
	}
	if accountedBytes > 0 && (s.dynamicBytes >= s.dynamicBytesLimit ||
		accountedBytes > s.dynamicBytesLimit-s.dynamicBytes) {
		return false, "contexts_bytes_limit"
	}

	s.nextIncarnation++
	if s.nextIncarnation == 0 {
		s.nextIncarnation++
	}
	s.items[key] = &taskContext{
		task:           t,
		nextRun:        now,
		runUntil:       runUntil,
		accountedBytes: accountedBytes,
		incarnation:    s.nextIncarnation,
	}
	if t.Source == sourceDynamic {
		s.dynamicItems++
		s.dynamicBytes += accountedBytes
	}
	return true, ""
}

func (s *taskStore) flush(now time.Time, budget int, source string) ([]scheduledTask, int) {
	if budget < 0 {
		budget = 0
	}

	out := []scheduledTask{}
	if source == sourceDynamic {
		out = make([]scheduledTask, 0, budget)
	}
	expired := 0
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, ctx := range s.items {
		if ctx.task.Source != source {
			continue
		}
		if !ctx.runUntil.IsZero() && now.After(ctx.runUntil) {
			delete(s.items, key)
			if ctx.task.Source == sourceDynamic {
				s.dynamicItems--
				s.dynamicBytes -= ctx.accountedBytes
			}
			expired++
			continue
		}
		if ctx.nextRun.After(now) {
			continue
		}
		if ctx.outstandingGeneration != 0 {
			continue
		}
		if source == sourceDynamic && budget == 0 {
			continue
		}
		ctx.dispatchGeneration++
		if ctx.dispatchGeneration == 0 {
			ctx.dispatchGeneration++
		}
		ctx.outstandingGeneration = ctx.dispatchGeneration
		out = append(out, scheduledTask{
			key:                key,
			source:             source,
			scheduledAt:        now,
			incarnation:        ctx.incarnation,
			dispatchGeneration: ctx.dispatchGeneration,
		})
		ctx.nextRun = now.Add(intervalWithJitter(ctx.task.Interval, ctx.task.ScheduleKey))
		if source == sourceDynamic {
			budget--
		}
	}
	return out, expired
}

func (s *taskStore) markDispatchFailed(run scheduledTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ctx, ok := s.items[run.key]; ok && ctx.task.Source == run.source &&
		ctx.incarnation == run.incarnation && ctx.dispatchGeneration == run.dispatchGeneration &&
		ctx.outstandingGeneration == run.dispatchGeneration {
		ctx.outstandingGeneration = 0
		ctx.nextRun = run.scheduledAt
	}
}

func (s *taskStore) acquireRun(run scheduledTask) (task, int64, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, ok := s.items[run.key]
	if !ok || ctx.task.Source != run.source || ctx.incarnation != run.incarnation ||
		ctx.dispatchGeneration != run.dispatchGeneration ||
		ctx.outstandingGeneration != run.dispatchGeneration {
		return task{}, 0, "context_unavailable"
	}
	accountedBytes := ctx.accountedBytes
	if accountedBytes > 0 && (s.inFlightBytes >= s.inFlightBytesLimit ||
		accountedBytes > s.inFlightBytesLimit-s.inFlightBytes) {
		return task{}, 0, "inflight_bytes_limit"
	}
	t := ctx.task
	t.ScheduledAt = run.scheduledAt
	s.inFlightBytes += accountedBytes
	return t, accountedBytes, ""
}

func (s *taskStore) releaseRun(run scheduledTask, accountedBytes int64) {
	s.mu.Lock()
	if ctx, ok := s.items[run.key]; ok && ctx.task.Source == run.source &&
		ctx.incarnation == run.incarnation && ctx.dispatchGeneration == run.dispatchGeneration &&
		ctx.outstandingGeneration == run.dispatchGeneration {
		ctx.outstandingGeneration = 0
	}
	if accountedBytes > 0 {
		s.inFlightBytes -= accountedBytes
	}
	s.mu.Unlock()
}

func (s *taskStore) len() int {
	n, _ := s.stats()
	return n
}

func (s *taskStore) dynamicLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dynamicItems
}

func (s *taskStore) stats() (int, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items), s.dynamicBytes + s.inFlightBytes
}

type scheduler struct {
	ipt       *Input
	cfg       DynamicConfig
	store     *taskStore
	inputCh   chan struct{}
	processCh chan scheduledTask
	localCh   chan scheduledTask
	stopCh    <-chan interface{}
	limiter   *minuteRateLimiter
	ctx       context.Context
	cancel    context.CancelFunc
}

func newScheduler(ipt *Input, cfg *DynamicConfig) *scheduler {
	c := normalizeDynamicConfig(cfg)
	s := &scheduler{
		ipt: ipt,
		cfg: c,
		store: newTaskStore(c.ContextsLimit, c.ContextsBytesLimit,
			int64(c.Workers)*maxDynamicContextBytes),
		inputCh:   make(chan struct{}, c.InputQueue),
		processCh: make(chan scheduledTask, c.ProcessQueue),
		localCh:   make(chan scheduledTask, defaultLocalProcessQueue),
		stopCh:    ipt.semStop.Wait(),
		limiter:   newMinuteRateLimiter(c.MaxPerMinute, time.Now()),
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

func (s *scheduler) start() {
	g := goGroup()
	g.Go(func(_ context.Context) error {
		select {
		case <-datakit.Exit.Wait():
		case <-s.stopCh:
		case <-s.ctx.Done():
		}
		s.cancel()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		s.runDynamicFlushLoop()
		return nil
	})
	for i := 0; i < s.cfg.Workers; i++ {
		g.Go(func(ctx context.Context) error {
			s.runWorker(s.processCh)
			return nil
		})
	}
	if len(s.ipt.Targets) == 0 {
		return
	}
	g.Go(func(ctx context.Context) error {
		s.runLocalFlushLoop()
		return nil
	})
	for i := 0; i < defaultLocalWorkers; i++ {
		g.Go(func(ctx context.Context) error {
			s.runWorker(s.localCh)
			return nil
		})
	}
}

func (s *scheduler) enqueue(t task) (bool, string) {
	if s.stopped() {
		incDrop("enqueue", "scheduler_stopped")
		return false, "scheduler_stopped"
	}
	select {
	case s.inputCh <- struct{}{}:
	default:
		incDrop("enqueue", "input_chan_full")
		observeQueues(s)
		return false, "input_chan_full"
	}
	observeQueues(s)
	ok, reason := s.store.add(time.Now(), t)
	observeStore(s.store)
	<-s.inputCh
	observeQueues(s)
	if ok && s.stopped() {
		// Do not acknowledge an admission that raced with scheduler shutdown.
		// The old store is no longer guaranteed to have a worker that can run it.
		ok = false
		reason = "scheduler_stopped"
	}
	if !ok {
		incDrop("store", reason)
	}
	return ok, reason
}

func (s *scheduler) addLocal(t task) (bool, string) {
	t.Source = sourceLocal
	ok, reason := s.store.add(time.Now(), t)
	observeStore(s.store)
	return ok, reason
}

func (s *scheduler) runDynamicFlushLoop() {
	ticker := time.NewTicker(s.cfg.FlushInterval.Duration)
	defer ticker.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			budget := s.limiter.take(now)
			tasks, expired := s.store.flush(now, budget, sourceDynamic)
			if expired > 0 {
				dropCounter.WithLabelValues("store", "expired").Add(float64(expired))
			}
			observeStore(s.store)
			s.dispatchDynamic(tasks)
		}
	}
}

func (s *scheduler) runLocalFlushLoop() {
	ticker := time.NewTicker(defaultLocalFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			tasks, _ := s.store.flush(time.Now(), 0, sourceLocal)
			observeStore(s.store)
			if !s.dispatchLocal(tasks) {
				return
			}
		}
	}
}

func (s *scheduler) dispatchDynamic(tasks []scheduledTask) {
	for _, run := range tasks {
		select {
		case s.processCh <- run:
			observeQueues(s)
		default:
			// flush advances nextRun tentatively; restore it when the worker
			// queue did not accept this run.
			s.store.markDispatchFailed(run)
			incDrop("dispatch", "process_chan_full")
			l.Warn("drop netpath target: process_chan_full")
		}
	}
}

func (s *scheduler) dispatchLocal(tasks []scheduledTask) bool {
	for _, run := range tasks {
		select {
		case <-datakit.Exit.Wait():
			return false
		case <-s.stopCh:
			return false
		case s.localCh <- run:
			observeQueues(s)
		}
	}
	return true
}

type minuteRateLimiter struct {
	maxPerMinute int
	tokens       float64
	last         time.Time
}

func newMinuteRateLimiter(maxPerMinute int, now time.Time) *minuteRateLimiter {
	if maxPerMinute <= 0 {
		maxPerMinute = defaultDynamicMaxPerMinute
	}
	return &minuteRateLimiter{maxPerMinute: maxPerMinute, last: now}
}

func (l *minuteRateLimiter) take(now time.Time) int {
	if l == nil {
		return 0
	}
	if l.last.IsZero() {
		l.last = now
		return 0
	}
	elapsed := now.Sub(l.last)
	l.last = now
	if elapsed <= 0 {
		return 0
	}
	l.tokens += elapsed.Minutes() * float64(l.maxPerMinute)
	if l.tokens > float64(l.maxPerMinute) {
		l.tokens = float64(l.maxPerMinute)
	}
	budget := int(math.Floor(l.tokens + 1e-9))
	l.tokens -= float64(budget)
	return budget
}

func intervalWithJitter(interval time.Duration, key string) time.Duration {
	if interval <= 0 {
		return interval
	}
	maxJitter := interval / 20
	if maxJitter <= 0 {
		return interval
	}
	if maxJitter > 30*time.Second {
		maxJitter = 30 * time.Second
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return interval + time.Duration(uint64(h.Sum32())%uint64(maxJitter))
}

func (s *scheduler) runWorker(queue <-chan scheduledTask) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case run := <-queue:
			observeQueues(s)
			if s.stopped() {
				return
			}
			t, accountedBytes, reason := s.store.acquireRun(run)
			if reason != "" {
				if reason == "inflight_bytes_limit" {
					s.store.markDispatchFailed(run)
					incDrop("worker", reason)
				}
				continue
			}
			observeStore(s.store)
			func() {
				defer func() {
					s.store.releaseRun(run, accountedBytes)
					observeStore(s.store)
				}()
				res := runProbeContext(s.ctx, t)
				if s.stopped() {
					return
				}
				s.ipt.enrichProbeGateway(&res)
				if s.stopped() {
					return
				}
				observeRun(res)
				s.ipt.feedResultContext(s.ctx, res)
			}()
			if s.stopped() {
				return
			}
		}
	}
}

func (s *scheduler) stopped() bool {
	if s.ctx.Err() != nil {
		return true
	}
	select {
	case <-datakit.Exit.Wait():
		return true
	case <-s.stopCh:
		return true
	default:
		return false
	}
}
