//go:build linux
// +build linux

package sysmonitor

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	pr "github.com/shirou/gopsutil/v3/process"
)

type ProcessFilter struct {
	allowTrace    bool
	nameBlacklist map[string]struct{}
	nameWhitelist map[string]struct{}
	envBlacklist  map[string]struct{}
	envWhitelist  map[string]struct{}
	serviceEnv    []string

	selfPid int
	asynCh  chan int

	procInfo map[int]*ProcInfo
	procDel  *lru.Cache

	kernerFilter kernelProcFilter
	sync.RWMutex
}

type ProcInfo struct {
	pid         int
	name        string
	binPath     string
	serviceName string

	keep       bool
	allowTrace bool
	createTS   int64
}

func (p *ProcInfo) ServiceName() string {
	return p.serviceName
}

func (p *ProcInfo) Name() string {
	return p.name
}

func (p *ProcInfo) Filtered() bool {
	return p.keep
}

func (p *ProcInfo) TraceFilterd() bool {
	return p.allowTrace
}

type FilterOpt func(p *ProcessFilter)

func convArrToMAP(li []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range li {
		m[v] = struct{}{}
	}
	return m
}

func WithTracing(on bool) FilterOpt {
	return func(p *ProcessFilter) {
		p.allowTrace = on
	}
}

func WithSelfPid(pid int) FilterOpt {
	return func(p *ProcessFilter) {
		p.selfPid = pid
	}
}

func WithEnvWhitelist(li []string) FilterOpt {
	return func(p *ProcessFilter) {
		p.envWhitelist = convArrToMAP(li)
	}
}

func WithEnvBlacklist(li []string) FilterOpt {
	return func(p *ProcessFilter) {
		p.envBlacklist = convArrToMAP(li)
	}
}

func WithNameWhitelist(li []string) FilterOpt {
	return func(p *ProcessFilter) {
		p.nameWhitelist = convArrToMAP(li)
	}
}

func WithNameBlacklist(li []string) FilterOpt {
	return func(p *ProcessFilter) {
		p.nameBlacklist = convArrToMAP(li)
	}
}

func WithEnvService(li []string) FilterOpt {
	return func(p *ProcessFilter) {
		var l []string
		l = append(l, li...)
		p.serviceEnv = l
	}
}

func NewProcessFilter(ctx context.Context, opts ...FilterOpt) *ProcessFilter {
	filter := &ProcessFilter{
		procInfo: map[int]*ProcInfo{},
		procDel:  lru.New(100_000),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(filter)
		}
	}
	filter.asynCh = make(chan int, 64)
	go filter.asynTryAddLoop(ctx)

	return filter
}

func (p *ProcessFilter) setKernelProcFilter(fn kernelProcFilter) {
	p.kernerFilter = fn
}

func (p *ProcessFilter) asynTryAddLoop(ctx context.Context) {
	pidSet := map[int]struct{}{}
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	for {
		select {
		case pid := <-p.asynCh:
			pidSet[pid] = struct{}{}
		case <-tk.C:
			if len(pidSet) > 0 {
				for pid := range pidSet {
					if _, err := p.TryAdd(pid); err != nil {
						log.Debug(err)
					}
				}
				pidSet = map[int]struct{}{}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *ProcessFilter) tryAdd(pid, ppid int, createTS int64, procName, binPath string, env map[string]string) *ProcInfo {
	keep := true
	allowTraceAttach := p.allowTrace

	if pid == p.selfPid || strings.HasPrefix(procName, "datakit") {
		keep = false
	}
	if p.allowTrace {
		switch {
		case !keep:
		case len(p.nameWhitelist) > 0:
			if _, ok := p.nameWhitelist[procName]; !ok {
				keep = false
			}
		case len(p.envWhitelist) > 0:
			keep = false
			for envName := range p.envWhitelist {
				if _, ok := env[envName]; ok {
					keep = true
					break
				}
			}
		case len(p.nameBlacklist) > 0:
			if _, ok := p.nameBlacklist[procName]; ok {
				keep = false
			}
		case len(p.envBlacklist) > 0:
			for envName := range p.envBlacklist {
				if _, ok := env[envName]; ok {
					keep = false
				}
			}
		}
	}

	var serviceName string
	for _, envName := range p.serviceEnv {
		if v, ok := env[envName]; ok {
			if v != "" {
				serviceName = v
				break
			}
		}
	}

	if serviceName == "" {
		serviceName = procName
	}

	if !keep && p.kernerFilter != nil {
		p.kernerFilter(pid)
	}

	if allowTraceAttach && !keep {
		allowTraceAttach = false
	}

	if binPath == "" {
		allowTraceAttach = false
	}

	p.Lock()
	defer p.Unlock()
	if !keep && ppid != 0 {
		// inherits the state of the parent process
		if v, ok := p.procInfo[ppid]; ok && v.name == procName {
			keep = v.keep
		}
	}

	inf := &ProcInfo{
		pid:         pid,
		name:        procName,
		binPath:     binPath,
		serviceName: serviceName,
		keep:        keep,
		allowTrace:  allowTraceAttach,
		createTS:    createTS,
	}
	p.procInfo[pid] = inf
	return inf
}

func (p *ProcessFilter) Delete(pid int) {
	p.Lock()
	defer p.Unlock()

	if v, ok := p.procInfo[pid]; ok {
		delete(p.procInfo, pid)
		p.procDel.Add(pid, v)
	}
}

func (p *ProcessFilter) GetProcInfo(pid int) (*ProcInfo, bool) {
	p.RLock()
	defer p.RUnlock()

	if v, ok := p.procInfo[pid]; ok && v != nil {
		return v, true
	}

	if v, ok := p.procDel.Get(pid); ok {
		if v, ok := v.(*ProcInfo); ok && v != nil {
			return v, true
		}
	}

	return nil, false
}

func (p *ProcessFilter) TryAdd(pid int) (*ProcInfo, error) {
	proc, err := pr.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	pname, err := proc.Name()
	if err != nil {
		return nil, err
	}

	var ppid int
	if v, _ := proc.Ppid(); v > 0 {
		ppid = int(v)
	}

	env, _ := proc.Environ()
	envMap := map[string]string{}
	for _, v := range env {
		s := strings.Index(v, "=")
		if s > 0 {
			envMap[v[:s]] = v[s+1:]
		}
	}

	exePath, err := proc.Exe()
	if err != nil {
		return nil, err
	}

	exeResolvePath := resolveBinPath(pid, exePath)
	if exeResolvePath == "" {
		log.Debugf("process: %s, pid: %d, path: %s, resolvepath: %s",
			pname, pid, exePath, exeResolvePath)
	} else {
		exeResolvePath = HostRoot(exeResolvePath)
	}

	ts, _ := proc.CreateTime()
	return p.tryAdd(pid, ppid, ts, pname, exeResolvePath, envMap), nil
}

func (p *ProcessFilter) AsyncTryAdd(pid int) {
	p.asynCh <- pid
}
