//go:build linux && cgo
// +build linux,cgo

package procwatch

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unsafe"

	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
)

func TestNewProbeWatcherLazyLoad(t *testing.T) {
	w, err := NewProbeWatcher(&Catalog{})
	if err != nil {
		t.Fatalf("new probe watcher: %v", err)
	}
	if w.Runtime != nil {
		t.Fatal("probe runtime should be nil when tracing is disabled")
	}
}

func TestNewProbeWatcherWithAllowTraceButNoTarget(t *testing.T) {
	w, err := NewProbeWatcher(&Catalog{allowTrace: true})
	if err != nil {
		t.Fatalf("new probe watcher: %v", err)
	}
	if w.Runtime != nil {
		t.Fatal("probe runtime should be nil when no trace target configured")
	}

	if os.Geteuid() != 0 {
		t.Skip("skip privileged procwatch runtime load test on non-root")
	}

	if _, err := dkebpf.ProcessSchedBin(); err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "file does not exist") {
			t.Skipf("skip procwatch runtime load test without embedded process_sched.o: %v", err)
		}
		t.Fatalf("load embedded process_sched.o: %v", err)
	}

	w, err = NewProbeWatcher(&Catalog{allowTrace: true, traceAllProc: true})
	if err != nil {
		t.Fatalf("new probe watcher: %v", err)
	}
	if w.Runtime == nil {
		t.Fatal("probe runtime should be initialized when trace target configured")
	}
}

func TestBinaryRegistryRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app")
	if err := os.WriteFile(path, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	registry := newBinaryRegistry()
	state, binKey, ok, err := registry.Check(path)
	if err != nil {
		t.Fatalf("check failed: %v", err)
	}
	if !ok {
		t.Fatal("expected first check to create a record")
	}
	if state == nil || state.path != path || state.uid == "" {
		t.Fatalf("unexpected registry state: %+v", state)
	}

	registry.AddRef(binKey)
	registry.AddRef(binKey)

	now := time.Now()
	if detached, shouldDetach := registry.Release(binKey, now, procwatchDetachGracePeriod); shouldDetach {
		t.Fatalf("unexpected detach on first release for %+v", detached)
	}

	if detached, shouldDetach := registry.Release(binKey, now, procwatchDetachGracePeriod); shouldDetach {
		t.Fatalf("expected final release to enter grace period, detached %+v", detached)
	}

	if expired := registry.CollectExpired(now.Add(procwatchDetachGracePeriod - time.Second)); len(expired) != 0 {
		t.Fatalf("expected no expired binaries before grace timeout, got %v", expired)
	}

	expired := registry.CollectExpired(now.Add(procwatchDetachGracePeriod + time.Second))
	if len(expired) != 1 || expired[0] == nil || expired[0].path != path || expired[0].uid != state.uid {
		t.Fatalf("unexpected expired binaries: %+v, want path=%s uid=%s", expired, path, state.uid)
	}
}

func TestBinaryRegistryTracksReplacedBinarySeparately(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app")
	if err := os.WriteFile(path, []byte("bin-v1"), 0o755); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	registry := newBinaryRegistry()
	first, firstKey, created, err := registry.Check(path)
	if err != nil {
		t.Fatalf("first check failed: %v", err)
	}
	if !created {
		t.Fatal("expected first check to create a record")
	}

	registry.AddRef(firstKey)

	if err := os.WriteFile(path, []byte("bin-v2 with different contents"), 0o755); err != nil {
		t.Fatalf("rewrite temp file: %v", err)
	}

	second, secondKey, created, err := registry.Check(path)
	if err != nil {
		t.Fatalf("second check failed: %v", err)
	}
	if !created {
		t.Fatal("expected replaced binary to create a new record")
	}
	if secondKey == firstKey {
		t.Fatalf("expected different binary keys after replacement, got same key %d", secondKey)
	}
	if second == nil || first == nil || second.uid == first.uid {
		t.Fatalf("expected different binary uids after replacement, first=%+v second=%+v", first, second)
	}
	if len(registry.records) != 2 {
		t.Fatalf("expected old and new binary records to coexist, got %d", len(registry.records))
	}
}

func TestBinaryRegistryClearInject(t *testing.T) {
	registry := newBinaryRegistry()
	const binHash = uint64(321)
	inject := &procInjectInfo{}

	registry.records[binHash] = &binaryProbeState{inject: inject}
	registry.ClearInject(binHash)

	if registry.records[binHash].inject != nil {
		t.Fatal("expected inject state to be cleared")
	}
}

func TestDeleteProcInjectHandlesNilRuntime(t *testing.T) {
	watcher := &ProbeWatcher{}
	watcher.deleteProcInject(12345)
}

func TestStartNilRuntimeWithTraceCatalog(t *testing.T) {
	watcher := &ProbeWatcher{
		catalog: &Catalog{allowTrace: true},
	}
	if err := watcher.Start(context.Background()); err != nil {
		t.Fatalf("start with nil runtime should be safe: %v", err)
	}
}

func TestStopNilRuntime(t *testing.T) {
	watcher := &ProbeWatcher{}
	if err := watcher.Stop(); err != nil {
		t.Fatalf("stop with nil runtime should be safe: %v", err)
	}
}

func TestHandleEventDropsShortRecord(t *testing.T) {
	watcher := &ProbeWatcher{}

	done := make(chan struct{})
	go func() {
		watcher.handleEvent(0, nil, nil, nil)
		watcher.handleEvent(0, make([]byte, schedEventSize-1), nil, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleEvent blocked on short record")
	}
}

func TestHandleEventExecQueuesAttach(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	writeFakeProcStat(t, procRoot, 1234, "app", 1)

	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	data := make([]byte, schedEventSize)
	event := (*schedEvent)(unsafe.Pointer(&data[0]))
	event.status = eventExec
	event.nxt_pid = 1234
	copy((*[16]byte)(unsafe.Pointer(&event.comm[0]))[:], []byte("app"))

	watcher.handleEvent(0, data, nil, nil)

	if got := len(watcher.attachCh); got != 1 {
		t.Fatalf("unexpected queued attach count: got %d want 1", got)
	}
	req := <-watcher.attachCh
	if req.pid != 1234 {
		t.Fatalf("unexpected attach request: %+v", req)
	}
	req.close()
}

func TestHandleEventExecQueuesUrgentAttachForWhitelistedName(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	writeFakeProcStat(t, procRoot, 3234, "app", 1)

	watcher := &ProbeWatcher{
		catalog:       NewCatalog(context.Background(), WithTracing(true), WithNameWhitelist([]string{"app"})),
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	data := make([]byte, schedEventSize)
	event := (*schedEvent)(unsafe.Pointer(&data[0]))
	event.status = eventExec
	event.nxt_pid = 3234
	copy((*[16]byte)(unsafe.Pointer(&event.comm[0]))[:], []byte("app"))

	watcher.handleEvent(0, data, nil, nil)

	req := <-watcher.attachCh
	if !req.urgent {
		t.Fatalf("expected urgent attach request, got %+v", req)
	}
	req.close()
}

func TestHandleEventForkQueuesAttach(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	writeFakeProcStat(t, procRoot, 2234, "app", 1)

	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	data := make([]byte, schedEventSize)
	event := (*schedEvent)(unsafe.Pointer(&data[0]))
	event.status = eventFork
	event.nxt_pid = 2234

	watcher.handleEvent(0, data, nil, nil)

	if got := len(watcher.attachCh); got != 1 {
		t.Fatalf("unexpected queued attach count: got %d want 1", got)
	}
	req := <-watcher.attachCh
	if req.pid != 2234 {
		t.Fatalf("unexpected attach request: %+v", req)
	}
	req.close()
}

func TestHandleEventExitDoesNotQueueAttach(t *testing.T) {
	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	data := make([]byte, schedEventSize)
	event := (*schedEvent)(unsafe.Pointer(&data[0]))
	event.status = eventExit
	event.nxt_pid = 1234

	watcher.handleEvent(0, data, nil, nil)

	if got := len(watcher.attachCh); got != 0 {
		t.Fatalf("unexpected queued attach count for exit event: got %d want 0", got)
	}
}

func TestEnqueueAttachDeduplicatesPID(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	writeFakeProcStat(t, procRoot, 1234, "app", 1)

	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	watcher.enqueueAttach(1234)
	watcher.enqueueAttach(1234)

	if got := len(watcher.attachCh); got != 1 {
		t.Fatalf("unexpected queued attach count: got %d want 1", got)
	}
	if _, ok := watcher.pendingAttach[1234]; !ok {
		t.Fatal("expected pid to remain pending after enqueue")
	}
	req := <-watcher.attachCh
	if req.pid != 1234 || req.startTime == 0 {
		t.Fatalf("unexpected attach request: %+v", req)
	}
	if req.procDirFD < 0 {
		t.Fatalf("expected proc dir fd in attach request: %+v", req)
	}
	req.close()
}

func TestBootstrapExistingProcessesQueuesPIDs(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)
	writeFakeProcStat(t, procRoot, 101, "app", 1)
	writeFakeProcStat(t, procRoot, 102, "app", 1)

	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}

	watcher.bootstrapExistingProcesses([]int{101, 102, 101})

	if got := len(watcher.attachCh); got != 2 {
		t.Fatalf("unexpected bootstrap queue count: got %d want 2", got)
	}
}

func TestEnqueueAttachTripsFuseOnBacklog(t *testing.T) {
	watcher := &ProbeWatcher{
		attachCh:      make(chan attachRequest, procwatchAttachQueueSize),
		pendingAttach: make(map[int]struct{}),
	}
	for i := 0; i < procwatchAttachFuseBacklog; i++ {
		watcher.pendingAttach[i] = struct{}{}
	}

	watcher.enqueueAttach(9999)

	if watcher.attachFuseUntil.IsZero() {
		t.Fatal("expected attach fuse to trip")
	}
	if got := len(watcher.attachCh); got != 0 {
		t.Fatalf("expected no queued attach while fuse trips, got %d", got)
	}
}

func TestAttachProcessSkipsStaleRequestBeforeResolve(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	writeFakeProcStat(t, procRoot, pid, "app", 1)

	watcher := &ProbeWatcher{}
	if err := watcher.attachProcess(attachRequest{pid: pid, startTime: 99999}); err != nil {
		t.Fatalf("stale attach request should be ignored without error: %v", err)
	}
}

func TestAttachProcessSkipsRequestWhenProcFDIsDead(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 5321
	writeFakeProcStat(t, procRoot, pid, "app", 1)

	procDirFD, err := openProcessDir(pid)
	if err != nil {
		t.Fatalf("open process dir: %v", err)
	}

	pidDir := filepath.Join(procRoot, "5321")
	if err := os.RemoveAll(pidDir); err != nil {
		t.Fatalf("remove old pid dir: %v", err)
	}
	writeFakeProcStat(t, procRoot, pid, "app-new", 1)

	req := attachRequest{
		pid:       pid,
		startTime: 12345,
		procDirFD: procDirFD,
	}
	defer req.close()

	watcher := &ProbeWatcher{}
	if err := watcher.attachProcess(req); err != nil {
		t.Fatalf("dead procfd attach request should be ignored without error: %v", err)
	}
}

func TestAttachFuseActive(t *testing.T) {
	watcher := &ProbeWatcher{
		pendingAttach: make(map[int]struct{}),
	}
	if watcher.attachFuseActive() {
		t.Fatal("expected inactive fuse by default")
	}

	watcher.attachFuseUntil = time.Now().Add(time.Second)
	if !watcher.attachFuseActive() {
		t.Fatal("expected active fuse during cooldown")
	}
}

func TestResolveBinaryAttachInfoUsesCache(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}

	identity, err := resolveBinaryIdentity(path)
	if err != nil {
		t.Fatalf("resolve binary identity: %v", err)
	}

	watcher := &ProbeWatcher{
		binCache: newCache(procwatchBinaryCacheMaxCost, procwatchBinaryCacheCounters),
	}

	first, err := watcher.resolveBinaryAttachInfo(identity.Key(), path)
	if err != nil {
		t.Fatalf("first resolve failed: %v", err)
	}
	second, err := watcher.resolveBinaryAttachInfo(identity.Key(), path)
	if err != nil {
		t.Fatalf("second resolve failed: %v", err)
	}

	if first == nil || second == nil {
		t.Fatal("expected cached binary resolve result")
	}
	if first != second {
		t.Fatal("expected second resolve to reuse cached result pointer")
	}
}
