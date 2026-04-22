//go:build linux
// +build linux

package procwatch

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/josharian/intern"
)

type Catalog struct {
	allowTrace    bool
	traceAllProc  bool
	nameBlacklist map[string]struct{}
	nameWhitelist map[string]struct{}
	envBlacklist  map[string]struct{}
	envWhitelist  map[string]struct{}
	envKeys       map[string]struct{}
	serviceEnv    []string

	selfPID int
	asyncCh chan int

	active  cacheStore
	deleted cacheStore

	kernelFilter func(int)

	pendingMu sync.Mutex
	pending   map[int]struct{}

	retryMu sync.RWMutex
	retry   map[int]resolveRetryState
}

type cacheStore interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
	Del(key interface{})
}

type ProcessInfo struct {
	binHash  uint64
	exePath  string
	probeUID string

	name    string
	service string

	collectable bool
	traceable   bool
	probeRef    bool
	deleted     bool
}

type CatalogOption func(*Catalog)

type resolveRetryState struct {
	failures  uint8
	nextRetry time.Time
	startTime uint64
}

const (
	processCatalogCacheMaxCost  = 512 * 1024
	processCatalogCacheCounters = 20_000
	processInfoBaseCost         = int64(unsafe.Sizeof(ProcessInfo{})) //nolint:gosec
	resolveRetryBaseDelay       = 250 * time.Millisecond
	resolveRetryMaxDelay        = 5 * time.Second
)

func (p *ProcessInfo) cacheCost() int64 {
	return processInfoBaseCost + int64(len(p.exePath)+len(p.name)+len(p.service))
}
func (p *ProcessInfo) Name() string { return p.name }
func (p *ProcessInfo) ServiceName() string {
	return p.service
}
func (p *ProcessInfo) ExePath() string   { return p.exePath }
func (p *ProcessInfo) Collectable() bool { return p.collectable }
func (p *ProcessInfo) Traceable() bool   { return p.traceable }
func (p *ProcessInfo) BinHash() uint64   { return p.binHash }
func (p *ProcessInfo) ProbeUID() string  { return p.probeUID }
func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func WithSelfPID(pid int) CatalogOption {
	return func(c *Catalog) { c.selfPID = pid }
}

func WithServiceEnv(keys []string) CatalogOption {
	return func(c *Catalog) {
		c.serviceEnv = append([]string(nil), keys...)
	}
}

func WithEnvWhitelist(items []string) CatalogOption {
	return func(c *Catalog) { c.envWhitelist = toSet(items) }
}

func WithEnvBlacklist(items []string) CatalogOption {
	return func(c *Catalog) { c.envBlacklist = toSet(items) }
}

func WithNameWhitelist(items []string) CatalogOption {
	return func(c *Catalog) { c.nameWhitelist = toSet(items) }
}

func WithNameBlacklist(items []string) CatalogOption {
	return func(c *Catalog) { c.nameBlacklist = toSet(items) }
}

func WithTracing(enabled bool) CatalogOption {
	return func(c *Catalog) { c.allowTrace = enabled }
}

func WithTraceAllProc(enabled bool) CatalogOption {
	return func(c *Catalog) { c.traceAllProc = enabled }
}

func NewCatalog(ctx context.Context, opts ...CatalogOption) *Catalog {
	catalog := &Catalog{
		active:  newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted: newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		pending: make(map[int]struct{}),
		retry:   make(map[int]resolveRetryState),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(catalog)
		}
	}

	catalog.envKeys = buildEnvKeySet(catalog.serviceEnv, catalog.envWhitelist, catalog.envBlacklist)
	catalog.asyncCh = make(chan int, 64)
	go catalog.resolveLoop(ctx)
	return catalog
}

func (c *Catalog) SetKernelFilter(fn func(int)) {
	c.kernelFilter = fn
}

func (c *Catalog) resolveLoop(ctx context.Context) {
	pids := make([]int, 0, 128)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case pid := <-c.asyncCh:
			pids = append(pids, pid)
		case <-ticker.C:
			for _, pid := range pids {
				if _, _, err := c.Resolve(pid); err != nil {
					if shouldLogResolveError(err) {
						log.Errorf("resolve pid %d: %s", pid, err.Error())
					} else {
						log.Debugf("skip pid %d resolve noise: %s", pid, err.Error())
					}
				}
				c.clearPending(pid)
			}
			pids = pids[:0]
		case <-ctx.Done():
			return
		}
	}
}

func (c *Catalog) create(pid, ppid int, procName, binPath string, env map[string]string) *ProcessInfo {
	collectable := true
	hardDeny := false

	if pid == c.selfPID || strings.HasPrefix(procName, "datakit") {
		collectable = false
		hardDeny = true
	}

	if c.allowTrace {
		switch {
		case !collectable:
		case len(c.nameWhitelist) > 0:
			_, collectable = c.nameWhitelist[procName]
		case len(c.envWhitelist) > 0:
			collectable = false
			for key := range c.envWhitelist {
				if _, ok := env[key]; ok {
					collectable = true
					break
				}
			}
		case len(c.nameBlacklist) > 0:
			if _, ok := c.nameBlacklist[procName]; ok {
				collectable = false
				hardDeny = true
			}
		case len(c.envBlacklist) > 0:
			for key := range c.envBlacklist {
				if _, ok := env[key]; ok {
					collectable = false
					hardDeny = true
					break
				}
			}
		}
	}

	serviceName := detectServiceName(procName, c.serviceEnv, env, nil)
	traceable := c.allowTrace && c.traceTargetConfigured() && collectable && binPath != ""

	if !collectable && !hardDeny && ppid != 0 {
		if parent, ok := c.Lookup(ppid); ok {
			collectable = parent.collectable
		}
	}
	if !collectable && c.kernelFilter != nil {
		c.kernelFilter(pid)
	}

	info := &ProcessInfo{
		name:        intern.String(procName),
		service:     intern.String(serviceName),
		exePath:     intern.String(binPath),
		collectable: collectable,
		traceable:   traceable,
	}
	c.active.Set(pid, info, info.cacheCost())
	return info
}

func (c *Catalog) Delete(pid int) {
	if value, ok := c.active.Get(pid); ok {
		if info, ok := value.(*ProcessInfo); ok && info != nil {
			info.deleted = true
			c.deleted.Set(pid, info, info.cacheCost())
		}
		c.active.Del(pid)
	}
	c.clearRetry(pid)
}

func (c *Catalog) lookupActive(pid int) (*ProcessInfo, bool) {
	if value, ok := c.active.Get(pid); ok {
		if info, ok := value.(*ProcessInfo); ok && info != nil && !info.deleted {
			return info, true
		}
	}
	return nil, false
}

func (c *Catalog) Lookup(pid int) (*ProcessInfo, bool) {
	if info, ok := c.lookupActive(pid); ok {
		return info, true
	}
	if value, ok := c.deleted.Get(pid); ok {
		if info, ok := value.(*ProcessInfo); ok && info != nil {
			return info, true
		}
	}
	return nil, false
}

func (c *Catalog) LookupOrResolve(pid int) (*ProcessInfo, bool) {
	if info, ok := c.Lookup(pid); ok {
		return info, true
	}
	_, info, err := c.Resolve(pid)
	if err != nil || info == nil {
		return nil, false
	}
	return info, true
}

func (c *Catalog) Resolve(pid int) (string, *ProcessInfo, error) {
	if info, ok := c.lookupActive(pid); ok {
		c.clearRetry(pid)
		return info.ExePath(), info, nil
	}

	name, ppid, startTime, err := readProcessStat(pid)
	if err != nil {
		c.markResolveFailure(pid, 0)
		return "", nil, err
	}

	var envMap map[string]string
	if c.needProcessEnv() {
		envMap = readProcessEnvironMapForKeys(pid, c.envKeys)
	}

	var hostPath string
	if c.shouldResolveBinary(name, envMap) {
		exePath, err := readProcessExePath(pid)
		if err != nil {
			if isNonRegularExecutablePathError(err) {
				exePath = ""
			} else {
				log.Debugf("get exe failed for pid %d (%s): %v", pid, name, err)
				c.markResolveFailure(pid, startTime)
				return "", nil, err
			}
		}

		hostPath = resolveHostBinaryPath(pid, exePath)
		if hostPath != "" {
			hostPath = HostRoot(hostPath)
		}
	}

	info := c.create(pid, ppid, name, hostPath, envMap)
	if info != nil && info.service == name {
		var cmdline []string
		if needCmdlineDiscovery(name) {
			cmdline = readProcessCmdline(pid)
		}
		if serviceName := detectServiceName(name, c.serviceEnv, envMap, cmdline); serviceName != "" {
			info.service = intern.String(serviceName)
		}
	}
	c.deleted.Del(pid)
	c.clearRetry(pid)
	return hostPath, info, nil
}

func (c *Catalog) ResolveWithProcFD(pid int, procDirFD int) (string, *ProcessInfo, error) {
	if info, ok := c.lookupActive(pid); ok {
		c.clearRetry(pid)
		return info.ExePath(), info, nil
	}

	name, ppid, startTime, err := readProcessStatFromProcFD(procDirFD, pid)
	if err != nil {
		c.markResolveFailure(pid, 0)
		return "", nil, err
	}

	var envMap map[string]string
	if c.needProcessEnv() {
		envMap = readProcessEnvironMapForKeysFromProcFD(procDirFD, c.envKeys)
	}

	var hostPath string
	if c.shouldResolveBinary(name, envMap) {
		exePath, err := readProcessExePathFromProcFD(procDirFD, pid)
		if err != nil {
			if isNonRegularExecutablePathError(err) {
				exePath = ""
			} else {
				log.Debugf("get exe failed for pid %d (%s) via procfd: %v", pid, name, err)
				c.markResolveFailure(pid, startTime)
				return "", nil, err
			}
		}

		hostPath = resolveHostBinaryPathFromProcFD(procDirFD, pid, exePath)
		if hostPath != "" {
			hostPath = HostRoot(hostPath)
		}
	}

	info := c.create(pid, ppid, name, hostPath, envMap)
	if info != nil && info.service == name {
		var cmdline []string
		if needCmdlineDiscovery(name) {
			cmdline = readProcessCmdlineFromProcFD(procDirFD)
		}
		if serviceName := detectServiceName(name, c.serviceEnv, envMap, cmdline); serviceName != "" {
			info.service = intern.String(serviceName)
		}
	}
	c.deleted.Del(pid)
	c.clearRetry(pid)
	return hostPath, info, nil
}

func (c *Catalog) ResolveLater(pid int) {
	if pid <= 0 {
		return
	}
	if _, ok := c.lookupActive(pid); ok {
		return
	}
	if !c.allowResolveRetry(pid, time.Now()) {
		return
	}
	if !c.markPending(pid) {
		return
	}
	select {
	case c.asyncCh <- pid:
	default:
		c.clearPending(pid)
		log.Debugf("drop async pid resolve: queue full, pid=%d", pid)
	}
}

func (c *Catalog) needProcessEnv() bool {
	return len(c.envKeys) > 0
}

func (c *Catalog) traceTargetConfigured() bool {
	if c == nil || !c.allowTrace {
		return false
	}
	if c.traceAllProc {
		return true
	}
	return len(c.nameWhitelist) > 0 || len(c.envWhitelist) > 0
}

func (c *Catalog) shouldResolveBinary(procName string, env map[string]string) bool {
	if c == nil || !c.allowTrace || !c.traceTargetConfigured() {
		return false
	}
	if c.traceAllProc {
		return true
	}
	if len(c.nameWhitelist) > 0 {
		_, ok := c.nameWhitelist[procName]
		return ok
	}
	if len(c.envWhitelist) > 0 {
		for key := range c.envWhitelist {
			if _, ok := env[key]; ok {
				return true
			}
		}
	}
	return false
}

func (c *Catalog) shouldPrioritizeAttachByName(procName string) bool {
	if c == nil || !c.allowTrace || procName == "" {
		return false
	}
	if len(c.nameWhitelist) == 0 {
		return false
	}
	_, ok := c.nameWhitelist[procName]
	return ok
}

func needCmdlineDiscovery(procName string) bool {
	switch procName {
	case "java", "python", "python2", "python3", "node", "nodejs", "ruby", "perl", "php", "bash", "sh":
		return true
	default:
		return false
	}
}

func shouldLogResolveError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) &&
		!isProcessGoneError(err) && !isNonRegularExecutablePathError(err)
}

func (c *Catalog) markPending(pid int) bool {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	if _, ok := c.pending[pid]; ok {
		return false
	}
	c.pending[pid] = struct{}{}
	return true
}

func (c *Catalog) clearPending(pid int) {
	c.pendingMu.Lock()
	delete(c.pending, pid)
	c.pendingMu.Unlock()
}

func buildEnvKeySet(serviceEnv []string, envWhitelist, envBlacklist map[string]struct{}) map[string]struct{} {
	total := len(serviceEnv) + len(envWhitelist) + len(envBlacklist)
	if total == 0 {
		return nil
	}

	keys := make(map[string]struct{}, total)
	for _, key := range serviceEnv {
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	for key := range envWhitelist {
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	for key := range envBlacklist {
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	if len(keys) == 0 {
		return nil
	}
	return keys
}

func (c *Catalog) markResolveFailure(pid int, startTime uint64) {
	if c == nil || pid <= 0 {
		return
	}

	c.retryMu.Lock()
	defer c.retryMu.Unlock()

	state := c.retry[pid]
	if startTime != 0 && state.startTime != 0 && state.startTime != startTime {
		state = resolveRetryState{}
	}
	if startTime != 0 {
		state.startTime = startTime
	}
	if state.failures < ^uint8(0) {
		state.failures++
	}

	delay := resolveRetryBaseDelay
	if state.failures > 1 {
		delay <<= (state.failures - 1)
	}
	if delay > resolveRetryMaxDelay {
		delay = resolveRetryMaxDelay
	}
	state.nextRetry = time.Now().Add(delay)
	c.retry[pid] = state
}

func (c *Catalog) allowResolveRetry(pid int, now time.Time) bool {
	if c == nil || pid <= 0 {
		return false
	}

	c.retryMu.RLock()
	state, ok := c.retry[pid]
	c.retryMu.RUnlock()
	if !ok {
		return true
	}
	return !now.Before(state.nextRetry)
}

func (c *Catalog) clearRetry(pid int) {
	if c == nil || pid <= 0 {
		return
	}
	c.retryMu.Lock()
	delete(c.retry, pid)
	c.retryMu.Unlock()
}
