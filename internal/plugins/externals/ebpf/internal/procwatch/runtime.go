//go:build linux
// +build linux

package procwatch

import (
	"bytes"
	"context"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/dgraph-io/ristretto"
	"github.com/josharian/intern"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hash"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"golang.org/x/sys/unix"
)

// #include "../c/process_sched/process_sched.h"
import "C"

type (
	schedEvent     C.struct_rec_process_sched_status
	procFilterInfo C.struct_proc_filter_info
	procInjectInfo C.struct_proc_inject
)

const schedEventSize = int(unsafe.Sizeof(schedEvent{})) //nolint:gosec

const (
	procwatchAttachBatchSize     = 8
	procwatchAttachInterval      = 50 * time.Millisecond
	procwatchAttachQueueSize     = 1024
	procwatchAttachFuseBacklog   = 512
	procwatchAttachFuseCooldown  = 30 * time.Second
	procwatchDetachGracePeriod   = 30 * time.Second
	procwatchDetachSweepPeriod   = 5 * time.Second
	procwatchBinaryCacheMaxCost  = 4096
	procwatchBinaryCacheCounters = 8192
)

const (
	eventFork = 0b1 << iota
	eventExec
	eventExit
)

const (
	mapProcInject = "bmap_procinject"
	mapProcFilter = "bmap_proc_filter"
	mapTidToGoID  = "bmap_tid2goid"
)

var runtimeExecutePrograms = []string{"uprobe__go_runtime_execute"}

type binaryProbeState struct {
	key             uint64
	path            string
	uid             string
	inject          *procInjectInfo
	refCnt          int
	pendingDetachAt time.Time
}

type binaryRegistry struct {
	records map[uint64]*binaryProbeState
	paths   map[uint64]string
	mu      sync.RWMutex
}

type binaryResolveResult struct {
	symbol         symbolLocation
	inject         procInjectInfo
	skipReason     string
	symbolSource   string
	goidSource     string
	cacheable      bool
	attachable     bool
	useRegisterABI bool
}

type ProbeWatcher struct {
	Runtime *bpfutil.Runtime

	catalog   *Catalog
	binaries  *binaryRegistry
	binCache  cacheStore
	perfErrCh chan error
	attachCh  chan attachRequest
	selfPID   int

	sync.Mutex
	pendingAttach   map[int]struct{}
	attachFuseUntil time.Time
}

type attachRequest struct {
	pid       int
	startTime uint64
	procDirFD int
	urgent    bool
}

func newBinaryRegistry() *binaryRegistry {
	return &binaryRegistry{
		records: make(map[uint64]*binaryProbeState),
		paths:   make(map[uint64]string),
	}
}

type binaryIdentity struct {
	path    string
	dev     uint64
	inode   uint64
	size    int64
	mtimeNs int64
	ctimeNs int64
}

func resolveBinaryIdentity(path string) (*binaryIdentity, error) {
	if path == "" {
		return nil, fmt.Errorf("empty binary path")
	}

	var stat unix.Stat_t
	if err := unix.Stat(path, &stat); err != nil {
		return nil, err
	}

	return &binaryIdentity{
		path:    path,
		dev:     stat.Dev,
		inode:   stat.Ino,
		size:    stat.Size,
		mtimeNs: stat.Mtim.Nano(),
		ctimeNs: stat.Ctim.Nano(),
	}, nil
}

func (id *binaryIdentity) Key() uint64 {
	if id == nil {
		return 0
	}

	var scratch [8]byte
	hashKey := func(h uint64, value uint64) uint64 {
		binary.LittleEndian.PutUint64(scratch[:], value)
		return hash.Fnv1aHashAddByte(h, scratch[:])
	}

	h := hash.Fnv1aHashAdd(hash.Fnv1aNew(), id.path)
	h = hashKey(h, id.dev)
	h = hashKey(h, id.inode)
	h = hashKey(h, uint64(id.size))
	h = hashKey(h, uint64(id.mtimeNs))
	h = hashKey(h, uint64(id.ctimeNs))
	return h
}

func (id *binaryIdentity) UID() string {
	if id == nil {
		return ""
	}
	return shortID(
		id.path,
		fmt.Sprintf("%d", id.dev),
		fmt.Sprintf("%d", id.inode),
		fmt.Sprintf("%d", id.size),
		fmt.Sprintf("%d", id.mtimeNs),
		fmt.Sprintf("%d", id.ctimeNs),
	)
}

func (r *binaryRegistry) Check(path string) (*binaryProbeState, uint64, bool, error) {
	identity, err := resolveBinaryIdentity(path)
	if err != nil {
		return nil, 0, false, err
	}
	key := identity.Key()

	r.mu.RLock()
	last, ok := r.records[key]
	r.mu.RUnlock()

	if !ok {
		fresh := &binaryProbeState{
			key:  key,
			path: identity.path,
			uid:  identity.UID(),
		}
		r.mu.Lock()
		r.records[key] = fresh
		r.paths[key] = identity.path
		r.mu.Unlock()
		return fresh, key, true, nil
	}

	return last, key, false, nil
}

func (r *binaryRegistry) Inject(hash uint64, inject *procInjectInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if state, ok := r.records[hash]; ok {
		state.inject = inject
	}
}

func (r *binaryRegistry) ClearInject(hash uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if state, ok := r.records[hash]; ok {
		state.inject = nil
	}
}

func (r *binaryRegistry) AddRef(hash uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if state, ok := r.records[hash]; ok {
		state.refCnt++
		state.pendingDetachAt = time.Time{}
	}
}

func (r *binaryRegistry) Release(hash uint64, now time.Time, grace time.Duration) (*binaryProbeState, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, ok := r.records[hash]
	if !ok {
		return nil, false
	}

	if state.refCnt > 0 {
		state.refCnt--
	}
	if state.refCnt <= 0 {
		if grace <= 0 {
			delete(r.records, hash)
			delete(r.paths, hash)
			return state, true
		}
		if state.pendingDetachAt.IsZero() {
			state.pendingDetachAt = now.Add(grace)
		}
		return nil, false
	}

	return nil, false
}

func (r *binaryRegistry) CollectExpired(now time.Time) []*binaryProbeState {
	r.mu.Lock()
	defer r.mu.Unlock()

	expired := make([]*binaryProbeState, 0)
	for hash, state := range r.records {
		if state == nil || state.refCnt > 0 || state.pendingDetachAt.IsZero() || now.Before(state.pendingDetachAt) {
			continue
		}
		expired = append(expired, state)
		delete(r.records, hash)
		delete(r.paths, hash)
	}
	return expired
}

func (r *binaryRegistry) Reset() {
	r.mu.Lock()
	r.records = make(map[uint64]*binaryProbeState)
	r.paths = make(map[uint64]string)
	r.mu.Unlock()
}

func NewProbeWatcher(catalog *Catalog) (*ProbeWatcher, error) {
	watcher := &ProbeWatcher{
		catalog:       catalog,
		binaries:      newBinaryRegistry(),
		binCache:      newCache(procwatchBinaryCacheMaxCost, procwatchBinaryCacheCounters),
		perfErrCh:     make(chan error, 8),
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	if catalog == nil || !catalog.allowTrace || !catalog.traceTargetConfigured() {
		return watcher, nil
	}

	runtime := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{ID: bpfutil.HookID{Program: "tracepoint__sched_process_exec"}},
			{ID: bpfutil.HookID{Program: "tracepoint__sched_process_exit"}},
		},
		Streams: []*bpfutil.PerfStream{
			{
				Map: bpfutil.Map{Name: "process_sched_event"},
				PerfStreamOptions: bpfutil.PerfStreamOptions{
					PerfRingBufferSize: 32 * os.Getpagesize(),
					PerfErrChan:        watcher.perfErrCh,
					LostHandler: func(cpu int, count uint64, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.AddPerfLost("procwatch", stream.Name, count)
					},
					DataHandler: watcher.handleEvent,
				},
			},
		},
	}

	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{Cur: math.MaxUint64, Max: math.MaxUint64},
	}
	buf, err := dkebpf.ProcessSchedBin()
	if err != nil {
		return nil, fmt.Errorf("process_sched.o: %w", err)
	}
	if err := runtime.LoadFromReader(bytes.NewReader(buf), loadSpec); err != nil {
		return nil, fmt.Errorf("load procwatch runtime: %w", err)
	}

	watcher.Runtime = runtime
	return watcher, nil
}

func (w *ProbeWatcher) Start(ctx context.Context) error {
	w.selfPID = os.Getpid()
	if w.catalog == nil || !w.catalog.allowTrace || w.Runtime == nil {
		return nil
	}

	if err := w.Runtime.StartRuntime(); err != nil {
		return err
	}

	go w.watchPerfErrors(ctx)
	go w.attachLoop(ctx)
	go w.detachUnusedBinaries(ctx)

	filter := func() func(int) {
		mp, err := w.Runtime.LookupMap(mapProcFilter)
		if err != nil {
			return func(int) {}
		}
		return func(pid int) {
			key := uint32(pid)
			value := procFilterInfo{disable: 1}
			if err := mp.Update(&key, &value, ebpf.UpdateAny); err != nil {
				log.Info(err)
			}
		}
	}()
	w.catalog.SetKernelFilter(filter)

	processes, err := listProcessIDs()
	if err != nil {
		return nil
	}
	w.bootstrapExistingProcesses(processes)

	return nil
}

func (w *ProbeWatcher) Stop() error {
	if w.Runtime == nil {
		return nil
	}
	if w.binaries != nil {
		w.binaries.Reset()
	}
	return w.Runtime.Shutdown()
}

func (w *ProbeWatcher) SharedMaps() (map[string]*ebpf.Map, bool) {
	if w.Runtime == nil {
		return nil, false
	}

	result := make(map[string]*ebpf.Map, 2)
	tid2goid, err := w.Runtime.LookupMap(mapTidToGoID)
	if err != nil {
		return nil, false
	}
	procFilter, err := w.Runtime.LookupMap(mapProcFilter)
	if err != nil {
		return nil, false
	}

	result[mapTidToGoID] = tid2goid
	result[mapProcFilter] = procFilter
	return result, true
}

func (w *ProbeWatcher) watchPerfErrors(ctx context.Context) {
	for {
		select {
		case err, ok := <-w.perfErrCh:
			if !ok {
				return
			}
			if err != nil {
				exporter.IncPerfReadError("procwatch", "process_sched_event")
				log.Warnf("procwatch perf stream stopped: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *ProbeWatcher) handleEvent(cpu int, data []byte, _ *bpfutil.PerfStream, _ *bpfutil.Runtime) {
	if len(data) < schedEventSize {
		log.Debugf("drop short process sched event: got %d want >= %d", len(data), schedEventSize)
		return
	}

	event := (*schedEvent)(unsafe.Pointer(&data[0]))

	switch event.status {
	case eventFork:
		if w.catalog != nil && !w.catalog.allowTrace {
			w.catalog.ResolveLater(int(event.nxt_pid))
			return
		}
		w.enqueueAttach(int(event.nxt_pid))
	case eventExec:
		procName := schedEventCommString(event)
		log.Debugf("procwatch exec event pid=%d name=%s", int(event.nxt_pid), procName)
		w.releaseProcess(int(event.nxt_pid))
		if w.catalog != nil && !w.catalog.allowTrace {
			w.catalog.ResolveLater(int(event.nxt_pid))
			return
		}
		w.enqueueAttachWithPriority(int(event.nxt_pid), w.catalog.shouldPrioritizeAttachByName(procName))
	case eventExit:
		w.releaseProcess(int(event.nxt_pid))
	}
}

func schedEventCommString(event *schedEvent) string {
	if event == nil {
		return ""
	}
	var comm [16]byte
	for i := range comm {
		comm[i] = byte(event.comm[i])
	}
	return schedCommString(comm[:])
}

func schedCommString(comm []byte) string {
	if idx := bytes.IndexByte(comm, 0); idx >= 0 {
		comm = comm[:idx]
	}
	return string(comm)
}

func (w *ProbeWatcher) detachBinary(state *binaryProbeState, processName string) {
	if w.Runtime == nil || state == nil || state.uid == "" {
		return
	}
	uid := state.uid
	for _, program := range runtimeExecutePrograms {
		if err := w.Runtime.DetachHookSpec(bpfutil.HookID{UID: uid, Program: program}); err != nil {
			log.Warn(err)
		} else {
			log.Infof("DetachHook: %s, ShortID: %s, name: %s", state.path, uid, processName)
		}
	}
}

func (w *ProbeWatcher) releaseProcess(pid int) {
	if w.catalog == nil {
		return
	}

	info, ok := w.catalog.Lookup(pid)
	if !ok || info == nil {
		return
	}
	w.deleteProcInject(pid)

	if info.probeRef && info.BinHash() != 0 && w.binaries != nil {
		if state, shouldDetach := w.binaries.Release(info.BinHash(), time.Now(), procwatchDetachGracePeriod); shouldDetach {
			w.detachBinary(state, info.Name())
		}
		info.probeRef = false
		info.binHash = 0
		info.probeUID = ""
	}

	w.catalog.Delete(pid)
}

func (w *ProbeWatcher) deleteProcInject(pid int) {
	if pid <= 0 || w.Runtime == nil {
		return
	}

	mp, err := w.Runtime.LookupMap(mapProcInject)
	if err != nil {
		log.Debugf("lookup proc inject map failed for pid %d: %v", pid, err)
		return
	}

	pid32 := uint32(pid)
	if err := mp.Delete(unsafe.Pointer(&pid32)); err != nil {
		log.Debugf("delete proc inject failed for pid %d: %v", pid, err)
	}
}

func (w *ProbeWatcher) attachProcess(req attachRequest) error {
	if req.pid <= 0 {
		return fmt.Errorf("pid <= 0")
	}

	if req.startTime != 0 {
		var (
			currentStartTime uint64
			err              error
		)
		if req.procDirFD >= 0 {
			currentStartTime, err = readProcessStartTimeFromProcFD(req.procDirFD, req.pid)
		} else {
			currentStartTime, err = readProcessStartTime(req.pid)
		}
		if err != nil {
			if shouldLogResolveError(err) {
				log.Debugf("skip stale procwatch attach for pid %d: %v", req.pid, err)
			}
			return nil
		}
		if currentStartTime != req.startTime {
			log.Debugf("skip stale procwatch attach for pid %d: queued start=%d current=%d",
				req.pid, req.startTime, currentStartTime)
			return nil
		}
	}

	var (
		binPath string
		info    *ProcessInfo
		err     error
	)
	if req.procDirFD >= 0 {
		binPath, info, err = w.catalog.ResolveWithProcFD(req.pid, req.procDirFD)
	} else {
		binPath, info, err = w.catalog.Resolve(req.pid)
	}
	if err != nil {
		return err
	}
	if info == nil {
		log.Debugf("skip procwatch attach for pid %d: resolved no process info", req.pid)
		return nil
	}
	if info.deleted {
		log.Debugf("skip procwatch attach for pid %d (%s): process already marked deleted", req.pid, info.Name())
		return nil
	}
	if w.Runtime == nil {
		log.Debugf("skip procwatch attach for pid %d (%s): runtime unavailable", req.pid, info.Name())
		return nil
	}
	if !info.Traceable() || binPath == "" {
		log.Debugf("skip procwatch attach for pid %d (%s): traceable=%v bin=%q collectable=%v",
			req.pid, info.Name(), info.Traceable(), binPath, info.Collectable())
		return nil
	}

	state, binKey, refresh, err := w.binaries.Check(binPath)
	if err != nil {
		return err
	}
	if !refresh && state.inject == nil {
		refresh = true
	}
	if !refresh {
		if state.inject != nil {
			mp, err := w.Runtime.LookupMap(mapProcInject)
			if err != nil {
				return fmt.Errorf("get bpf map %s failed: %w", mapProcInject, err)
			}
			pid32 := uint32(req.pid)
			if err := mp.Update(unsafe.Pointer(&pid32), unsafe.Pointer(&state.inject), ebpf.UpdateAny); err != nil {
				return err
			}
			if !info.probeRef {
				w.binaries.AddRef(binKey)
				info.probeRef = true
				info.binHash = binKey
				info.probeUID = state.uid
			}
		}
		return nil
	}

	resolved, err := w.resolveBinaryAttachInfo(binKey, binPath)
	if err != nil {
		return err
	}
	if !resolved.attachable {
		if resolved.skipReason != "" {
			log.Debugf("skip procwatch uprobe attach for %s: %s", binPath, resolved.skipReason)
		}
		return nil
	}
	if resolved.goidSource == "fallback" {
		log.Warnf("procwatch using fallback runtime.g.goid offset for %s; binary DWARF unavailable", binPath)
	}
	if resolved.symbolSource != "" && resolved.symbolSource != "gopclntab" {
		log.Infof("procwatch using %s to resolve runtime.execute for %s", resolved.symbolSource, binPath)
	}

	uid := state.uid
	attached := false
	attachedPrograms := make([]string, 0, len(runtimeExecutePrograms))
	for _, program := range runtimeExecutePrograms {
		err := w.Runtime.AttachHook(bpfutil.HookSpec{
			ID:           bpfutil.HookID{UID: uid, Program: program},
			UprobeOffset: resolved.symbol.Start,
			BinaryPath:   intern.String(binPath),
		})
		if err != nil {
			log.Warn(err)
			continue
		}
		attached = true
		attachedPrograms = append(attachedPrograms, program)
		log.Infof("AddHooK: %s, ShortID: %s, name: %s, pid: %d", binPath, uid, info.Name(), req.pid)
	}
	if !attached {
		return nil
	}
	if attached && !info.probeRef {
		mp, err := w.Runtime.LookupMap(mapProcInject)
		if err != nil {
			for _, program := range attachedPrograms {
				if err := w.Runtime.DetachHookSpec(bpfutil.HookID{UID: uid, Program: program}); err != nil {
					log.Warn(err)
				}
			}
			return fmt.Errorf("get bpf map %s failed: %w", mapProcInject, err)
		}
		pid32 := uint32(req.pid)
		if err := mp.Update(unsafe.Pointer(&pid32), unsafe.Pointer(&resolved.inject), ebpf.UpdateAny); err != nil {
			for _, program := range attachedPrograms {
				if err := w.Runtime.DetachHookSpec(bpfutil.HookID{UID: uid, Program: program}); err != nil {
					log.Warn(err)
				}
			}
			w.binaries.ClearInject(binKey)
			state.inject = nil
			return err
		}
		inject := resolved.inject
		w.binaries.Inject(binKey, &inject)
		w.binaries.AddRef(binKey)
		info.binHash = binKey
		info.probeUID = uid
		info.probeRef = true
	}

	return nil
}

func (w *ProbeWatcher) resolveBinaryAttachInfo(binKey uint64, binPath string) (*binaryResolveResult, error) {
	if w != nil && w.binCache != nil {
		if cached, ok := w.binCache.Get(binKey); ok {
			if result, ok := cached.(*binaryResolveResult); ok && result != nil {
				return result, nil
			}
		}
	}

	result, err := buildBinaryResolveResult(binPath)
	if err != nil {
		return nil, err
	}

	if result != nil && result.cacheable && w != nil && w.binCache != nil {
		w.binCache.Set(binKey, result, 1)
		if cache, ok := w.binCache.(*ristretto.Cache); ok {
			cache.Wait()
		}
	}
	return result, nil
}

func buildBinaryResolveResult(binPath string) (*binaryResolveResult, error) {
	elfFile, err := elf.Open(binPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open elf file %s: %w", binPath, err)
	}
	defer elfFile.Close() //nolint:errcheck

	result := &binaryResolveResult{cacheable: true}

	goVersion, versionKnown := resolveGoVersion(binPath, elfFile)
	useRegisterABI, err := resolveGoRuntimeRegisterABI(goVersion, versionKnown, runtime.GOARCH)
	if err != nil {
		result.skipReason = err.Error()
		return result, nil
	}

	symbol, symbolSource, err := findGoRuntimeExecuteWithSource(elfFile, "runtime.execute")
	if err != nil {
		result.skipReason = fmt.Sprintf("resolve runtime.execute failed: %v", err)
		return result, nil
	}

	goidOffset, goidOffsetSource, err := resolveGoRuntimeGoidOffset(elfFile)
	if err != nil {
		result.skipReason = fmt.Sprintf("resolve runtime.g.goid offset failed via %s: %v", goidOffsetSource, err)
		return result, nil
	}

	result.symbol = *symbol
	result.symbolSource = symbolSource
	result.goidSource = goidOffsetSource
	result.useRegisterABI = useRegisterABI
	result.attachable = true
	result.inject = procInjectInfo{
		offset_go_runtime_g_goid: C.__u64(goidOffset),
	}
	if useRegisterABI {
		result.inject.go_use_register = 1
	}

	return result, nil
}

func (w *ProbeWatcher) bootstrapExistingProcesses(pids []int) {
	for _, pid := range pids {
		w.enqueueAttach(pid)
	}
}

func (w *ProbeWatcher) enqueueAttach(pid int) {
	w.enqueueAttachWithPriority(pid, false)
}

func (w *ProbeWatcher) enqueueAttachWithPriority(pid int, urgent bool) {
	if w == nil || pid <= 0 {
		return
	}

	w.Lock()
	now := time.Now()
	if now.Before(w.attachFuseUntil) {
		w.Unlock()
		return
	}
	if len(w.pendingAttach) >= procwatchAttachFuseBacklog {
		w.tripAttachFuseLocked(now, "attach_backlog")
		w.Unlock()
		return
	}
	if _, ok := w.pendingAttach[pid]; ok {
		if !urgent {
			w.Unlock()
			return
		}
		w.Unlock()
	} else {
		w.pendingAttach[pid] = struct{}{}
		w.Unlock()
	}

	req := attachRequest{pid: pid, urgent: urgent}
	if startTime, err := readProcessStartTime(pid); err == nil {
		req.startTime = startTime
	}
	req.procDirFD = -1
	if procDirFD, err := openProcessDir(pid); err == nil {
		req.procDirFD = procDirFD
	}

	select {
	case w.attachCh <- req:
	default:
		req.close()
		w.Lock()
		delete(w.pendingAttach, pid)
		w.Unlock()
		log.Warnf("procwatch attach queue full, drop pid %d", pid)
	}
}

func (w *ProbeWatcher) attachLoop(ctx context.Context) {
	if w == nil {
		return
	}

	ticker := time.NewTicker(procwatchAttachInterval)
	defer ticker.Stop()

	batch := make([]attachRequest, 0, procwatchAttachBatchSize)
	dropBatch := func() {
		if len(batch) == 0 {
			return
		}
		w.Lock()
		for _, req := range batch {
			delete(w.pendingAttach, req.pid)
			req.close()
		}
		w.Unlock()
		batch = batch[:0]
	}
	drain := func() {
		if len(batch) == 0 {
			return
		}
		for _, req := range batch {
			if err := w.attachProcess(req); err != nil {
				log.Debug(err)
			}
			req.close()
			w.Lock()
			delete(w.pendingAttach, req.pid)
			w.Unlock()
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			drain()
			return
		case req, ok := <-w.attachCh:
			if !ok {
				drain()
				return
			}
			batch = append(batch, req)
			if req.urgent && !w.attachFuseActive() {
				drain()
			} else if len(batch) >= procwatchAttachBatchSize {
				if w.attachFuseActive() {
					dropBatch()
				} else {
					drain()
				}
			}
		case <-ticker.C:
			if w.attachFuseActive() {
				dropBatch()
			} else {
				drain()
			}
		}
	}
}

func (r *attachRequest) close() {
	if r == nil || r.procDirFD < 0 {
		return
	}
	_ = unix.Close(r.procDirFD)
	r.procDirFD = -1
}

func (w *ProbeWatcher) attachFuseActive() bool {
	if w == nil {
		return false
	}
	w.Lock()
	defer w.Unlock()
	return time.Now().Before(w.attachFuseUntil)
}

func (w *ProbeWatcher) tripAttachFuseLocked(now time.Time, reason string) {
	if w == nil {
		return
	}
	if now.Before(w.attachFuseUntil) {
		return
	}
	w.attachFuseUntil = now.Add(procwatchAttachFuseCooldown)
	log.Warnf("procwatch attach fuse tripped: reason=%s pending=%d cooldown=%s", reason, len(w.pendingAttach), procwatchAttachFuseCooldown)
}

func (w *ProbeWatcher) detachUnusedBinaries(ctx context.Context) {
	if w == nil || w.binaries == nil {
		return
	}

	ticker := time.NewTicker(procwatchDetachSweepPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, state := range w.binaries.CollectExpired(time.Now()) {
				w.detachBinary(state, "")
			}
		}
	}
}
