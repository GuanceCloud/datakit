// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultPromPodLimit         = 500
	defaultPromTaskLimit        = 1000
	defaultPromScheduleInterval = time.Second
	defaultPromRequestTimeout   = 30 * time.Second
	promTaskLimitLogRate        = 0.1
)

type promTask struct {
	runner   *promRunner
	interval time.Duration
	nextDue  time.Time
	ctx      context.Context
	busy     bool
	lastRun  uint64
}

type promPodState struct {
	revision promPodRevision
	cancel   context.CancelFunc
	tasks    []*promTask
}

type promPodRevision struct {
	raw       [sha256.Size]byte
	effective [sha256.Size]byte
}

type preparedPromPod struct {
	revision promPodRevision
	configs  []*promConfig
}

type promJob struct {
	task        *promTask
	scheduledAt time.Time
}

type promTaskManager struct {
	cfg            *Config
	workerCount    int
	requestTimeout time.Duration
	now            func() time.Time
	tickC          <-chan time.Time

	ctx    context.Context
	cancel context.CancelFunc

	pods  map[types.UID]*promPodState
	tasks []*promTask

	jobs      chan promJob
	results   chan *promTask
	snapshots chan []promPodCandidate
	running   int
	runSeq    uint64

	done chan struct{}
}

func newPromTaskManager(parent context.Context, cfg *Config) *promTaskManager {
	ctx, cancel := context.WithCancel(parent)
	workers := min(4, max(2, runtime.NumCPU()))
	return &promTaskManager{
		cfg:            cfg,
		workerCount:    workers,
		requestTimeout: defaultPromRequestTimeout,
		now:            time.Now,
		ctx:            ctx,
		cancel:         cancel,
		pods:           make(map[types.UID]*promPodState),
		jobs:           make(chan promJob, workers),
		results:        make(chan *promTask, workers),
		snapshots:      make(chan []promPodCandidate),
		done:           make(chan struct{}),
	}
}

func preparePromPod(candidate promPodCandidate, labelKeys []string) (*preparedPromPod, error) {
	configs, err := parsePromConfigs(completePromConfig(candidate.pod, candidate.rawConfig))
	if err != nil {
		return nil, err
	}

	effectiveConfigs := clonePromConfigs(configs)
	applyPromPodMetadata(candidate.pod, effectiveConfigs, labelKeys)
	effectiveHash := sha256.New()
	if err := json.NewEncoder(effectiveHash).Encode(effectiveConfigs); err != nil {
		return nil, fmt.Errorf("encode effective Prometheus config: %w", err)
	}
	var effectiveRevision [sha256.Size]byte
	copy(effectiveRevision[:], effectiveHash.Sum(nil))

	return &preparedPromPod{
		revision: promPodRevision{
			raw:       sha256.Sum256([]byte(candidate.rawConfig)),
			effective: effectiveRevision,
		},
		configs: configs,
	}, nil
}

func (m *promTaskManager) applySnapshot(candidates []promPodCandidate, now time.Time) {
	desired := make(map[types.UID]promPodCandidate, len(candidates))
	orderedUIDs := make([]types.UID, 0, len(candidates))
	for _, candidate := range candidates {
		if _, found := desired[candidate.uid]; !found {
			orderedUIDs = append(orderedUIDs, candidate.uid)
		}
		desired[candidate.uid] = candidate
	}

	type preparationResult struct {
		pod *preparedPromPod
		err error
	}
	preparations := make(map[types.UID]preparationResult, len(desired))
	prepare := func(uid types.UID) (*preparedPromPod, error) {
		if result, found := preparations[uid]; found {
			return result.pod, result.err
		}
		pod, err := preparePromPod(desired[uid], m.cfg.LabelAsTagsForMetric.Keys)
		preparations[uid] = preparationResult{pod: pod, err: err}
		return pod, err
	}
	loggedPreparationErrors := make(map[types.UID]struct{})
	logPreparationError := func(candidate promPodCandidate) {
		if _, found := loggedPreparationErrors[candidate.uid]; found {
			return
		}
		loggedPreparationErrors[candidate.uid] = struct{}{}
		klog.Warnf("failed to parse legacy %s annotation for Pod %s/%s; discarded all tasks",
			annotationPromExport, candidate.pod.Namespace, candidate.pod.Name)
	}

	activeTasks := len(m.tasks)
	for uid, state := range m.pods {
		candidate, found := desired[uid]
		if !found {
			activeTasks -= len(state.tasks)
			m.removePod(uid, state)
			continue
		}
		prepared, err := prepare(uid)
		if err != nil {
			logPreparationError(candidate)
		}
		if err != nil || state.revision != prepared.revision {
			activeTasks -= len(state.tasks)
			m.removePod(uid, state)
		}
	}

	activePods := len(m.pods)
	rejectedPods, rejectedTasks := 0, 0
	for _, uid := range orderedUIDs {
		candidate := desired[uid]
		if _, found := m.pods[uid]; found {
			continue
		}
		if activePods >= defaultPromPodLimit {
			rejectedPods++
			continue
		}
		prepared, err := prepare(uid)
		if err != nil {
			logPreparationError(candidate)
			continue
		}
		if activeTasks+len(prepared.configs) > defaultPromTaskLimit {
			rejectedPods++
			rejectedTasks += len(prepared.configs)
			continue
		}
		if len(prepared.configs) == 0 {
			continue
		}
		runners, err := newPromRunnersForPod(candidate.pod, prepared.configs, m.cfg)
		if err != nil {
			klog.Warnf("failed to build legacy %s tasks for Pod %s/%s; discarded all tasks",
				annotationPromExport, candidate.pod.Namespace, candidate.pod.Name)
			continue
		}

		podCtx, podCancel := context.WithCancel(m.ctx)
		state := &promPodState{revision: prepared.revision, cancel: podCancel}
		for _, runner := range runners {
			state.tasks = append(state.tasks, &promTask{
				runner: runner, interval: runner.conf.Interval,
				nextDue: now.Add(runner.conf.Interval), ctx: podCtx,
			})
		}
		m.pods[uid] = state
		activePods++
		activeTasks += len(state.tasks)
	}

	ordered := make([]*promTask, 0, activeTasks)
	for _, uid := range orderedUIDs {
		if state, found := m.pods[uid]; found {
			ordered = append(ordered, state.tasks...)
		}
	}
	m.tasks = ordered
	podAnnotationPromActiveTasks.Set(float64(activeTasks))
	podAnnotationPromVec.WithLabelValues("prom").Observe(float64(activeTasks))
	if rejectedPods > 0 {
		klog.RLWarnf(promTaskLimitLogRate,
			"legacy %s limits reached: pods=%d/%d tasks=%d/%d dropped_pods=%d dropped_tasks=%d",
			annotationPromExport, activePods, defaultPromPodLimit, activeTasks, defaultPromTaskLimit, rejectedPods, rejectedTasks)
	}
}

func (m *promTaskManager) removePod(uid types.UID, state *promPodState) {
	state.cancel()
	for _, task := range state.tasks {
		task.runner.close()
	}
	delete(m.pods, uid)
}

func (m *promTaskManager) removeAllPods() {
	for uid, state := range m.pods {
		m.removePod(uid, state)
	}
	m.tasks = nil
	m.runSeq = 0
	podAnnotationPromActiveTasks.Set(0)
}

func (m *promTaskManager) run() {
	ticks := m.tickC
	if ticks == nil {
		ticker := time.NewTicker(defaultPromScheduleInterval)
		defer ticker.Stop()
		ticks = ticker.C
	}

	workers := goroutine.NewGroup(goroutine.Option{
		Name:       "k8s-pod-prom-workers",
		PanicTimes: 6,
	})
	for range m.workerCount {
		workers.Go(func(_ context.Context) error {
			m.runWorker()
			return nil
		})
	}
	defer func() {
		m.removeAllPods()
		if err := workers.Wait(); err != nil {
			klog.Warnf("kubernetes Prometheus workers stopped: %s", err)
		}
		close(m.done)
	}()

	for {
		select {
		case <-m.ctx.Done():
			return
		case task := <-m.results:
			m.running--
			now := m.now()
			if task.ctx.Err() == nil {
				if !now.Before(task.nextDue) {
					m.advanceNextDue(task, now)
				}
				task.busy = false
			}
			m.schedule(now)
		case snapshot := <-m.snapshots:
			m.applySnapshot(snapshot, m.now())
		case now := <-ticks:
			m.schedule(now)
		}
	}
}

func (m *promTaskManager) schedule(now time.Time) {
	for range m.workerCount - m.running {
		var next *promTask
		for _, task := range m.tasks {
			if task.busy || now.Before(task.nextDue) {
				continue
			}
			if next == nil || task.lastRun < next.lastRun {
				next = task
			}
		}
		if next == nil {
			break
		}
		scheduledAt := m.advanceNextDue(next, now)
		m.runSeq++
		next.lastRun, next.busy = m.runSeq, true
		m.running++
		m.jobs <- promJob{task: next, scheduledAt: scheduledAt}
	}
}

func (*promTaskManager) advanceNextDue(task *promTask, now time.Time) time.Time {
	steps := int64(now.Sub(task.nextDue)/task.interval) + 1
	scheduledAt := task.nextDue.Add(time.Duration((steps - 1) * task.interval.Nanoseconds()))
	task.nextDue = task.nextDue.Add(time.Duration(steps * task.interval.Nanoseconds()))
	return scheduledAt
}

func (m *promTaskManager) runWorker() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case job := <-m.jobs:
			m.runPromJob(job)
			select {
			case m.results <- job.task:
			case <-m.ctx.Done():
				return
			}
		}
	}
}

func (m *promTaskManager) runPromJob(job promJob) {
	defer func() {
		if recover() != nil {
			klog.Errorf("panic recovered while collecting legacy %s Prometheus metrics", annotationPromExport)
		}
	}()

	if job.task.ctx.Err() != nil {
		return
	}
	err := job.task.runner.scrape(job.task.ctx, job.scheduledAt, m.requestTimeout)
	if err != nil && !errors.Is(err, context.Canceled) {
		klog.Warnf("failed to collect legacy %s Prometheus metrics (result=%s)",
			annotationPromExport, promScrapeResult(err))
	}
}

func (m *promTaskManager) publishPromPods(candidates []promPodCandidate) {
	select {
	case m.snapshots <- candidates:
	case <-m.ctx.Done():
	}
}

func (m *promTaskManager) close() {
	m.cancel()
	<-m.done
}
