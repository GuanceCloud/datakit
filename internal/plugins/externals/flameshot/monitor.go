// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package flameshot  is a monitor for process.
package flameshot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type monitor struct {
	config             *Config
	cs                 []*processM
	csChan             chan *processM
	statsChan          chan *triggerStats
	exitChan           chan *processExitEvent
	oomChan            chan *OOMEvent
	oomWorkerSem       chan struct{}
	heapDumpChan       chan *heapDumpTask
	heapDumpWorkerSem  chan struct{}
	watchers           map[string]*cgroupWatcher
	watcherKeyByPID    map[int32]string
	procRoot           string
	cgroupRoot         string
	cgroupPollInterval time.Duration
	processedHProf     *processedHProfStore
	goPProfDeltas      *goPProfDeltaStore
}

type cgroupWatcher struct {
	key     string
	version string
	dir     string
	cancel  context.CancelFunc

	mu      sync.RWMutex
	members map[int32]*processM
}

func NewMonitor(config *Config) *monitor {
	return &monitor{
		config:             config,
		cs:                 []*processM{},
		csChan:             make(chan *processM, 10),
		statsChan:          make(chan *triggerStats, 5),
		exitChan:           make(chan *processExitEvent, 5),
		oomChan:            make(chan *OOMEvent, 5),
		oomWorkerSem:       make(chan struct{}, 1),
		heapDumpChan:       make(chan *heapDumpTask, 5),
		heapDumpWorkerSem:  make(chan struct{}, 1),
		watchers:           make(map[string]*cgroupWatcher),
		watcherKeyByPID:    make(map[int32]string),
		procRoot:           defaultProcRoot,
		cgroupRoot:         defaultCgroupRoot,
		cgroupPollInterval: defaultCgroupPollInterval,
		processedHProf:     newProcessedHProfStore(),
		goPProfDeltas:      newGoPProfDeltaStore(),
	}
}

func newCgroupWatcher(key, version, dir string, cancel context.CancelFunc) *cgroupWatcher {
	return &cgroupWatcher{
		key:     key,
		version: version,
		dir:     dir,
		cancel:  cancel,
		members: make(map[int32]*processM),
	}
}

func (w *cgroupWatcher) addMember(pm *processM) {
	if w == nil || pm == nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.members[pm.Pid] = pm
}

func (w *cgroupWatcher) removeMember(pid int32) int {
	if w == nil {
		return 0
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.members, pid)
	return len(w.members)
}

func (w *cgroupWatcher) snapshotMembers() []*processM {
	if w == nil {
		return nil
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	members := make([]*processM, 0, len(w.members))
	for _, pm := range w.members {
		if pm != nil {
			members = append(members, pm)
		}
	}
	return members
}

// 监控单个命令的资源使用情况.
func (m *monitor) MonitorCommand(p *processM) {
	if !p.isAlive() {
		m.handleProcessGone(p, fmt.Errorf("process pid=%d is no longer running", p.Pid))
		return
	}

	if err := p.updateProcessStats(); err != nil {
		if !p.isAlive() {
			m.handleProcessGone(p, err)
			return
		}
		log.Debugf("update process stats failed but process is still running, pid=%d err=%v", p.Pid, err)
		return
	}
	trigger, tags, emergency := p.triggerDecision()
	tags = append(tags, m.config.Tags...)
	tags = append(tags, p.configProcess.Tags...)
	// 将service作为tag，方便在中心展示。
	tags = append(tags, fmt.Sprintf("%s:%s", "service", p.configProcess.Service))
	if trigger {
		now := time.Now()
		if emergency || hasMemoryPressureTag(tags) {
			p.markMemoryPressure(now)
		}
		if emergency {
			m.enqueueHeapDumpIfNeeded(p, "process_memory_emergency", tags, now)
		}
		if !m.config.profilingEnabled() {
			log.Debugf("profiling disabled, skip profiling trigger for pid=%d service=%s", p.Pid, p.configProcess.Service)
			return
		}
		duration := p.configProcess.Duration
		if emergency {
			p.markEmergencyProfileTriggered(now)
			duration = getEmergencyProfileDuration(p.configProcess)
			tags = append(tags, "trigger:memory_emergency")
		} else {
			p.markProfileTriggered(now)
		}
		stats := m.newTriggerStatsForProcess(p, duration, tags)
		stats.CommandName = p.Name
		stats.PID = p.Pid
		stats.Triggered = true
		stats.Service = p.configProcess.Service
		m.statsChan <- stats
	}
}

func removePID(cs []*processM, command *processM) []*processM {
	for i, c := range cs {
		if c.Pid == command.Pid {
			return append(cs[:i], cs[i+1:]...)
		}
	}
	return cs
}

func (m *monitor) handleProcessGone(pm *processM, reason error) {
	if pm == nil {
		return
	}

	service := ""
	if pm.configProcess != nil {
		service = pm.configProcess.Service
	}

	log.Infof("process exit detected, pid=%d service=%s err=%v", pm.Pid, service, reason)
	m.stopWatcher(pm.Pid)
	m.cs = removePID(m.cs, pm)
	if event := m.buildProcessExitEvent(pm, time.Now(), reason); event != nil {
		m.enqueueProcessExitEvent(event)
	}
}

func (m *monitor) enqueueProcessExitEvent(event *processExitEvent) {
	if m == nil || event == nil {
		return
	}

	select {
	case m.exitChan <- event:
	default:
		log.Warnf("process exit event queue full, upload asynchronously, pid=%d service=%s", event.PID, event.Service)
		go m.handleProcessExit(event)
	}
}

func (m *monitor) runProcessExitWorker(ctx context.Context) {
	for {
		select {
		case event := <-m.exitChan:
			m.handleProcessExit(event)
		case <-ctx.Done():
			return
		}
	}
}

func (m *monitor) findProcessByPID(pid int32) *processM {
	for _, pm := range m.cs {
		if pm != nil && pm.Pid == pid {
			return pm
		}
	}
	return nil
}

func hasMemoryPressureTag(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "mem_") || strings.Contains(tag, "cgroup_mem_percent") {
			return true
		}
	}
	return false
}

func (m *monitor) Start(osSignal chan os.Signal) {
	filterProcessesTicker := time.NewTicker(time.Minute)

	if m.config.MonitorInterval == "" {
		m.config.MonitorInterval = "1s"
	}
	t, err := time.ParseDuration(m.config.MonitorInterval)
	if err != nil {
		t = time.Second
	}
	log.Infof("start monitor, interval: %s", t.String())
	monitorCommandTicker := time.NewTicker(t)

	if m.config.HTTPConfig != nil {
		go m.startHTTPServer()
	}

	exitWorkerCtx, exitWorkerCancel := context.WithCancel(context.Background())
	go m.runProcessExitWorker(exitWorkerCtx)

	autoDuration, autoEnabled := m.getAutoProfilingDuration()
	autoTicker := time.NewTicker(autoDuration)
	defer func() {
		exitWorkerCancel()
		autoTicker.Stop()
		filterProcessesTicker.Stop()
		monitorCommandTicker.Stop()
	}()

	m.findProcess()

	for {
		select {
		case <-filterProcessesTicker.C:
			m.findProcess()
		case pm := <-m.csChan:
			var has bool
			for _, c := range m.cs {
				if c.Pid == pm.Pid {
					has = true
				}
			}
			if !has {
				// 添加到监控列表
				log.Infof("match: PID=%d, name=%s or cmd=%s", pm.Pid, pm.Name, pm.Cmdline)
				pidCount.WithLabelValues(pm.configProcess.Language)
				pm.podMEMLimit = m.config.PodMEMLimit
				pm.podCPULimit = m.config.PodCPULimit
				m.cs = append(m.cs, pm)
				m.startWatcher(pm)
			}
		case <-monitorCommandTicker.C:
			commands := append([]*processM(nil), m.cs...)
			for _, c := range commands {
				m.MonitorCommand(c)
			}
		case <-autoTicker.C:
			// 对所有的监控列表 顺序进行 profiling 采集
			if autoEnabled && m.config.profilingEnabled() {
				for _, c := range m.cs {
					m.autoProfilingProcess(c)
				}
			}
		case stats := <-m.statsChan:
			// trigger 执行命令不可以超过5分钟，理论上1分钟就可以结束。
			log.Infof("start run profiling, pid: %d, command: %s,", stats.PID, stats.CommandName)
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			err = runProfiling(ctx, stats)
			cancel()
			if err != nil {
				log.Errorf("run profiling err: %v", err)
			} else {
				// upload
				err = uploadFileToDataKit(stats, m.config.DataKitAddr)
				if err != nil {
					uploadToDK.WithLabelValues(stats.Service, err.Error())
					log.Errorf("upload to DataKit err: %v", err)
				} else if pm := m.findProcessByPID(stats.PID); pm != nil {
					pm.markProfileArtifact(&profileArtifactSummary{
						OutputPath:  stats.artifactOutputPath(),
						UploadedAt:  time.Now(),
						StartTime:   stats.startTime,
						EndTime:     stats.endTime,
						Event:       stats.Event,
						DurationSec: stats.Duration,
					})
				}
				deleteFile(stats)
			}
		case oomEvent := <-m.oomChan:
			if !m.config.OOMHProfEnabled {
				continue
			}
			m.oomWorkerSem <- struct{}{}
			go func(evt *OOMEvent) {
				defer func() { <-m.oomWorkerSem }()
				m.handleOOMEvent(evt)
			}(oomEvent)
		case task := <-m.heapDumpChan:
			go m.runHeapDumpTask(task)

		case <-osSignal:
			m.stopAllWatchers()
			log.Infof("monitor stop")
			return
		}
	}
}

func (m *monitor) runHeapDumpTask(task *heapDumpTask) {
	if m == nil || task == nil {
		return
	}
	m.heapDumpWorkerSem <- struct{}{}
	defer func() { <-m.heapDumpWorkerSem }()
	m.handleHeapDumpTask(task)
}

// 抽取配置解析逻辑，保持主流程清晰.
func (m *monitor) getAutoProfilingDuration() (time.Duration, bool) {
	if m.config.AutoProfiling == "" {
		// 如果没配，返回一个较长的时间防止频繁唤醒 CPU，布尔值标为 false
		return time.Hour * 24, false
	}
	d, err := time.ParseDuration(m.config.AutoProfiling)
	if err != nil {
		log.Warnf("invalid AutoProfiling format, fallback to 5m: %v", err)
		return time.Minute * 5, true
	}
	if d <= 0 {
		return time.Hour * 24, false
	}
	return d, true
}

func (m *monitor) findProcess() {
	// 先检测一遍进程，匹配到进程之后添加到监控列表中
	for _, p := range m.config.Processes {
		pms := filterProcessesByRegex(p)
		for _, pm := range pms {
			m.csChan <- pm
		}
	}
}

func (m *monitor) autoProfilingProcess(p *processM) {
	tags := make([]string, 0, len(m.config.Tags)+len(p.configProcess.Tags))
	tags = append(tags, m.config.Tags...)
	tags = append(tags, p.configProcess.Tags...)
	tags = append(tags, fmt.Sprintf("%s:%s", "service", p.configProcess.Service))
	stats := newTriggerStats(p.configProcess.Events, m.getAutoProfileSampleDuration(), tags)
	m.applyProcessConfigToStats(stats, p.configProcess)
	stats.PID = p.Pid
	stats.Triggered = true
	stats.CommandName = p.Name
	stats.Service = p.configProcess.Service

	m.statsChan <- stats
}

func (m *monitor) newTriggerStatsForProcess(p *processM, duration string, tags []string) *triggerStats {
	events := ""
	if p != nil && p.configProcess != nil {
		events = p.configProcess.Events
	}
	stats := newTriggerStats(events, duration, tags)
	if p != nil {
		m.applyProcessConfigToStats(stats, p.configProcess)
	}
	return stats
}

func (m *monitor) applyProcessConfigToStats(stats *triggerStats, p *Process) {
	if stats == nil || p == nil {
		return
	}

	stats.Language = p.Language
	stats.PProfURL = p.PProfURL
	stats.PProfTypes = append([]string(nil), p.PProfTypes...)
	stats.PProfTimeout = p.PProfTimeout
	stats.PySpyPath = p.PySpyPath
	stats.PySpyOutput = p.PySpyOutputPath
	stats.PySpyRate = p.PySpyRate
	stats.PySpySubproc = p.PySpySubprocesses
	stats.PySpyIdle = p.PySpyIdle
	if isGoLanguage(p.Language) && len(p.PProfTypes) == 0 && strings.TrimSpace(p.Events) == "" {
		stats.PProfTypes = []string{defaultGoPProfType}
	}
	if m != nil {
		stats.goPProfDeltas = m.goPProfDeltas
	}
}

func getEmergencyProfileDuration(p *Process) string {
	if p == nil || p.EmergencyDuration == "" {
		return "15s"
	}
	return p.EmergencyDuration
}

func (m *monitor) getAutoProfileSampleDuration() string {
	if m == nil || m.config == nil || m.config.AutoProfileDuration == "" {
		return "30s"
	}
	return m.config.AutoProfileDuration
}

func (m *monitor) startWatcher(pm *processM) {
	if pm == nil {
		return
	}
	if _, ok := m.watcherKeyByPID[pm.Pid]; ok {
		return
	}

	key, dir, version, err := resolveCgroupWatcherTarget(m.procRoot, m.cgroupRoot, pm.Pid)
	if err != nil {
		log.Debugf("resolve cgroup watcher target failed for pid=%d: %v", pm.Pid, err)
		return
	}

	m.watcherKeyByPID[pm.Pid] = key
	if watcher, ok := m.watchers[key]; ok {
		watcher.addMember(pm)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	watcher := newCgroupWatcher(key, version, dir, cancel)
	watcher.addMember(pm)
	m.watchers[key] = watcher
	go m.watchCgroupMemory(ctx, watcher)
}

func (m *monitor) stopWatcher(pid int32) {
	key, ok := m.watcherKeyByPID[pid]
	if !ok {
		return
	}
	delete(m.watcherKeyByPID, pid)

	watcher, ok := m.watchers[key]
	if !ok {
		return
	}
	if watcher.removeMember(pid) > 0 {
		return
	}

	watcher.cancel()
	delete(m.watchers, key)
}

func (m *monitor) stopAllWatchers() {
	for key, watcher := range m.watchers {
		watcher.cancel()
		delete(m.watchers, key)
	}
	for pid := range m.watcherKeyByPID {
		delete(m.watcherKeyByPID, pid)
	}
}

// filterProcessesByRegex 使用正则表达式过滤进程，并通过channel发送匹配的进程.
func filterProcessesByRegex(p *Process) []*processM {
	// 编译正则表达式
	re, err := regexp.Compile(p.Command) //nolint
	if err != nil {
		log.Errorf("compile [%s] err: %v", p.Command, err)
		return nil
	}

	// 获取所有进程ID
	pids, err := process.Pids()
	if err != nil {
		log.Errorf("get all process: %v", err)
		return nil
	}

	log.Debugf("system has %d process ,start match of %s ...", len(pids), p.Command)

	matchedCount := 0
	pms := make([]*processM, 0)
	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue // 进程可能已退出，跳过
		}

		name, err := proc.Name()
		if err != nil {
			log.Errorf("process get name err: %v", err)
			continue
		}
		cmd, err := proc.Cmdline()
		if err != nil {
			log.Errorf("process get cmdline err: %v", err)
			continue
		}
		// 使用正则表达式匹配进程名 或者启动命令
		if re.MatchString(name) || re.MatchString(cmd) || name == p.Language {
			if shouldSkipMatchedProcess(name, cmd) {
				log.Debugf("skip helper process match: PID=%d, name=%s, cmd=%s", pid, name, cmd)
				continue
			}
			matchedCount++
			processInfo := newProcessM(name, cmd, pid, p)
			pms = append(pms, processInfo)
		}
	}

	log.Debugf("filter matched command count: %d", matchedCount)
	return pms
}

var ignoredJavaToolNames = map[string]struct{}{
	"jcmd":   {},
	"jhsdb":  {},
	"jinfo":  {},
	"jmap":   {},
	"jps":    {},
	"jstack": {},
	"jstat":  {},
}

func shouldSkipMatchedProcess(name, cmdline string) bool {
	if isIgnoredJavaToolName(name) {
		return true
	}

	fields := strings.Fields(cmdline)
	if len(fields) == 0 {
		return false
	}
	return isIgnoredJavaToolName(filepath.Base(fields[0]))
}

func isIgnoredJavaToolName(name string) bool {
	name = strings.ToLower(strings.TrimSpace(filepath.Base(name)))
	if name == "" {
		return false
	}
	_, ok := ignoredJavaToolNames[name]
	return ok
}
