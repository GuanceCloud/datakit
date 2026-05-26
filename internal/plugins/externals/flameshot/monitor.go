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
	"regexp"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type monitor struct {
	config             *Config
	cs                 []*processM
	csChan             chan *processM
	statsChan          chan *triggerStats
	oomChan            chan *OOMEvent
	oomWorkerSem       chan struct{}
	jcmdChan           chan *jcmdSnapshotRequest
	jcmdWorkerSem      chan struct{}
	watchers           map[string]*cgroupWatcher
	watcherKeyByPID    map[int32]string
	procRoot           string
	cgroupRoot         string
	cgroupPollInterval time.Duration
	processedHProf     *processedHProfStore
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
		oomChan:            make(chan *OOMEvent, 5),
		oomWorkerSem:       make(chan struct{}, 1),
		jcmdChan:           make(chan *jcmdSnapshotRequest, 5),
		jcmdWorkerSem:      make(chan struct{}, 1),
		watchers:           make(map[string]*cgroupWatcher),
		watcherKeyByPID:    make(map[int32]string),
		procRoot:           defaultProcRoot,
		cgroupRoot:         defaultCgroupRoot,
		cgroupPollInterval: defaultCgroupPollInterval,
		processedHProf:     newProcessedHProfStore(),
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
	if err := p.updateProcessStats(); err != nil {
		m.stopWatcher(p.Pid)
		m.cs = removePID(m.cs, p)
		return
	}
	trigger, tags, emergency := p.triggerDecision()
	tags = append(tags, m.config.Tags...)
	tags = append(tags, p.configProcess.Tags...)
	// 将service作为tag，方便在中心展示。
	tags = append(tags, fmt.Sprintf("%s:%s", "service", p.configProcess.Service))
	if trigger {
		p.markProfileTriggered(time.Now())
		duration := p.configProcess.Duration
		if emergency {
			duration = getEmergencyProfileDuration(p.configProcess)
			tags = append(tags, "trigger:memory_emergency")
			if m.config != nil && m.config.JCmdSnapshotEnabled && !p.inJcmdCooldown(defaultCgroupEmergencyCooldown) {
				p.markJcmdTriggered(time.Now())
				jcmdTags := make([]string, 0, len(tags))
				jcmdTags = append(jcmdTags, tags...)
				m.jcmdChan <- &jcmdSnapshotRequest{
					Service:     p.configProcess.Service,
					PID:         p.Pid,
					ProcessName: p.Name,
					DetectedAt:  time.Now(),
					MemPercent:  getListAvg(p.MemPercentHistory, 1),
					Tags:        jcmdTags,
				}
			}
		}
		stats := newTriggerStats(p.configProcess.Events, duration, tags)
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

	autoDuration, autoEnabled := m.getAutoProfilingDuration()
	autoTicker := time.NewTicker(autoDuration)
	defer func() {
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
			for _, c := range m.cs {
				m.MonitorCommand(c)
			}
		case <-autoTicker.C:
			// 对所有的监控列表 顺序进行 profiling 采集
			if autoEnabled {
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
		case snapshotReq := <-m.jcmdChan:
			if !m.config.JCmdSnapshotEnabled {
				continue
			}
			m.jcmdWorkerSem <- struct{}{}
			go func(req *jcmdSnapshotRequest) {
				defer func() { <-m.jcmdWorkerSem }()
				m.handleJcmdSnapshot(req)
			}(snapshotReq)

		case <-osSignal:
			m.stopAllWatchers()
			log.Infof("monitor stop")
			return
		}
	}
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
	stats.PID = p.Pid
	stats.Triggered = true
	stats.CommandName = p.Name
	stats.Service = p.configProcess.Service

	m.statsChan <- stats
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

	key, version, err := resolveCgroupWatcherTarget(m.procRoot, m.cgroupRoot, pm.Pid)
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
	watcher := newCgroupWatcher(key, version, key, cancel)
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
			matchedCount++
			processInfo := newProcessM(name, cmd, pid, p)
			pms = append(pms, processInfo)
		}
	}

	log.Debugf("filter matched command count: %d", matchedCount)
	return pms
}
