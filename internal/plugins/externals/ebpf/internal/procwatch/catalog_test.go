//go:build linux
// +build linux

package procwatch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/sys/unix"
)

func writeFakeProcStat(t *testing.T, procRoot string, pid int, name string, ppid int) {
	t.Helper()

	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}

	stat := "1 (" + name + ") S " +
		strconv.Itoa(ppid) +
		" 1 1 0 -1 4194560 0 0 0 0 0 0 0 0 20 0 1 0 12345 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0\n"
	if err := os.WriteFile(filepath.Join(pidDir, "stat"), []byte(stat), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFakeProcCmdline(t *testing.T, procRoot string, pid int, args ...string) {
	t.Helper()

	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	payload := make([]byte, 0, 64)
	for _, arg := range args {
		payload = append(payload, []byte(arg)...)
		payload = append(payload, 0)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "cmdline"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFakeProcEnviron(t *testing.T, procRoot string, pid int, entries ...string) {
	t.Helper()

	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	payload := make([]byte, 0, 64)
	for _, entry := range entries {
		payload = append(payload, []byte(entry)...)
		payload = append(payload, 0)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "environ"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveReturnsCachedExePath(t *testing.T) {
	catalog := &Catalog{
		active:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:      newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		retry:        make(map[int]resolveRetryState),
		allowTrace:   true,
		traceAllProc: true,
	}
	pid := os.Getpid()
	info := catalog.create(pid, 0, "python3", "/opt/app/server.py", map[string]string{})
	if info == nil {
		t.Fatal("expected process info")
	}
	if cache, ok := catalog.active.(*ristretto.Cache); ok {
		cache.Wait()
	}

	catalog.active.Set(pid, info, info.cacheCost())
	if cache, ok := catalog.active.(*ristretto.Cache); ok {
		cache.Wait()
	}

	gotPath, gotInfo, err := catalog.Resolve(pid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotInfo == nil {
		t.Fatal("expected cached process info")
	}
	if gotPath != "/opt/app/server.py" {
		t.Fatalf("expected cached exe path, got %q", gotPath)
	}
}

func TestResolveIgnoresDeletedCache(t *testing.T) {
	catalog := &Catalog{
		active:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:      newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		retry:        make(map[int]resolveRetryState),
		allowTrace:   true,
		traceAllProc: true,
	}

	pid := os.Getpid()
	stale := catalog.create(pid, 0, "old-proc", "/stale/path", map[string]string{})
	if stale == nil {
		t.Fatal("expected process info")
	}
	stale.deleted = true
	catalog.deleted.Set(pid, stale, stale.cacheCost())
	if cache, ok := catalog.deleted.(*ristretto.Cache); ok {
		cache.Wait()
	}

	gotPath, gotInfo, err := catalog.Resolve(pid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotInfo == nil {
		t.Fatal("expected live process info")
	}
	if gotInfo.deleted {
		t.Fatal("expected live cache entry")
	}
	if gotPath == "/stale/path" {
		t.Fatalf("expected fresh path instead of stale deleted cache: %q", gotPath)
	}
	if absPath, err := os.Executable(); err == nil && filepath.Clean(gotPath) != filepath.Clean(absPath) {
		t.Fatalf("expected live executable path %q, got %q", absPath, gotPath)
	}
}

func TestResolveLaterDeduplicatesPendingPID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	catalog := NewCatalog(ctx)
	catalog.markPending(1234)
	catalog.ResolveLater(1234)

	if len(catalog.asyncCh) != 0 {
		t.Fatalf("expected no duplicate pending pid, got queue length %d", len(catalog.asyncCh))
	}
}

func TestResolveLaterDropsWhenQueueFull(t *testing.T) {
	catalog := &Catalog{
		asyncCh: make(chan int, 1),
		pending: make(map[int]struct{}),
		active:  newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted: newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		selfPID: os.Getpid(),
	}
	catalog.asyncCh <- 100

	done := make(chan struct{})
	go func() {
		catalog.ResolveLater(200)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ResolveLater blocked when queue is full")
	}

	if _, ok := catalog.pending[200]; ok {
		t.Fatal("expected dropped pid to be removed from pending set")
	}
}

func TestShouldLogResolveError(t *testing.T) {
	t.Run("process gone", func(t *testing.T) {
		err := &os.PathError{Op: "open", Path: "/proc/123/stat", Err: os.ErrNotExist}
		if shouldLogResolveError(err) {
			t.Fatal("expected disappeared /proc process errors to be suppressed")
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		if shouldLogResolveError(context.Canceled) {
			t.Fatal("expected context cancellation to be suppressed")
		}
	})

	t.Run("non regular executable path", func(t *testing.T) {
		err := &nonRegularExecutablePathError{pid: 123, path: "/usr/local/bin/kube-apiserver"}
		if shouldLogResolveError(err) {
			t.Fatal("expected non-regular executable path errors to be suppressed")
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		if !shouldLogResolveError(errors.New("boom")) {
			t.Fatal("expected unexpected errors to remain visible")
		}
	})
}

func TestTraceTargetConfigured(t *testing.T) {
	catalog := &Catalog{}
	if catalog.traceTargetConfigured() {
		t.Fatal("trace target should be disabled by default")
	}

	catalog.allowTrace = true
	if catalog.traceTargetConfigured() {
		t.Fatal("trace target should require explicit scope")
	}

	catalog.traceAllProc = true
	if !catalog.traceTargetConfigured() {
		t.Fatal("trace_all_proc should enable tracing scope")
	}

	catalog.traceAllProc = false
	catalog.nameWhitelist = map[string]struct{}{"nginx": {}}
	if !catalog.traceTargetConfigured() {
		t.Fatal("name whitelist should enable tracing scope")
	}

	catalog.nameWhitelist = nil
	catalog.envWhitelist = map[string]struct{}{"DD_SERVICE": {}}
	if !catalog.traceTargetConfigured() {
		t.Fatal("env whitelist should enable tracing scope")
	}
}

func TestCreateDisablesTraceableWithoutExplicitTraceScope(t *testing.T) {
	catalog := &Catalog{
		allowTrace: true,
		selfPID:    -1,
		active:     newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:    newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
	}

	info := catalog.create(1234, 0, "app", "/usr/local/bin/app", map[string]string{})
	if info == nil {
		t.Fatal("expected process info")
	}
	if !info.collectable {
		t.Fatal("expected process to remain collectable")
	}
	if info.traceable {
		t.Fatal("expected process tracing to stay disabled without explicit trace scope")
	}

	catalog.nameWhitelist = map[string]struct{}{"app": {}}
	info = catalog.create(1235, 0, "app", "/usr/local/bin/app", map[string]string{})
	if info == nil {
		t.Fatal("expected process info")
	}
	if !info.traceable {
		t.Fatal("expected whitelist-scoped process tracing to be enabled")
	}
}

func TestCreateKeepsBlacklistAsHardDeny(t *testing.T) {
	catalog := &Catalog{
		allowTrace:    true,
		selfPID:       -1,
		nameBlacklist: map[string]struct{}{"curl": {}},
		active:        newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
	}

	parent := catalog.create(2000, 0, "python3", "/usr/bin/python3", nil)
	if parent == nil || !parent.collectable {
		t.Fatal("expected parent to be collectable")
	}
	if cache, ok := catalog.active.(*ristretto.Cache); ok {
		cache.Wait()
	}

	child := catalog.create(2001, 2000, "curl", "/usr/bin/curl", nil)
	if child == nil {
		t.Fatal("expected child process info")
	}
	if child.collectable {
		t.Fatal("expected blacklist match to stay non-collectable even with collectable parent")
	}
	if child.traceable {
		t.Fatal("expected blacklist match to stay non-traceable")
	}
}

func TestCreateParentInheritanceDoesNotTriggerKernelFilter(t *testing.T) {
	catalog := &Catalog{
		allowTrace:    true,
		selfPID:       -1,
		nameWhitelist: map[string]struct{}{"python3": {}},
		active:        newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
	}

	parent := catalog.create(3000, 0, "python3", "/usr/bin/python3", nil)
	if parent == nil || !parent.collectable {
		t.Fatal("expected parent to be collectable")
	}
	if cache, ok := catalog.active.(*ristretto.Cache); ok {
		cache.Wait()
	}

	var filtered []int
	catalog.kernelFilter = func(pid int) {
		filtered = append(filtered, pid)
	}

	child := catalog.create(3001, 3000, "worker", "", nil)
	if child == nil {
		t.Fatal("expected child process info")
	}
	if !child.collectable {
		t.Fatal("expected child to inherit collectable state from parent")
	}
	if child.traceable {
		t.Fatal("expected inherited child to remain non-traceable without direct scope match")
	}
	if len(filtered) != 0 {
		t.Fatalf("expected kernel filter to stay untouched for inherited child, got %v", filtered)
	}
}

func TestResolveSkipsExeLookupWithoutTraceScope(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	writeFakeProcStat(t, procRoot, pid, "python3", 1)

	catalog := &Catalog{
		active:  newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted: newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		selfPID: -1,
		retry:   make(map[int]resolveRetryState),
	}

	path, info, err := catalog.Resolve(pid)
	if err != nil {
		t.Fatalf("unexpected resolve error without trace scope: %v", err)
	}
	if info == nil {
		t.Fatal("expected resolved process info")
	}
	if path != "" {
		t.Fatalf("expected empty path without trace scope, got %q", path)
	}
	if info.ExePath() != "" {
		t.Fatalf("expected empty cached exe path without trace scope, got %q", info.ExePath())
	}
	if !info.Collectable() {
		t.Fatal("expected process to remain collectable")
	}
}

func TestResolveStillRequiresExeLookupWhenTraceEnabled(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4322
	writeFakeProcStat(t, procRoot, pid, "python3", 1)

	catalog := &Catalog{
		active:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:      newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		selfPID:      -1,
		retry:        make(map[int]resolveRetryState),
		allowTrace:   true,
		traceAllProc: true,
	}

	if _, _, err := catalog.Resolve(pid); err == nil {
		t.Fatal("expected missing exe to fail when trace scope is enabled")
	}
}

func TestResolveSkipsExeLookupWhenNameWhitelistMisses(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4323
	writeFakeProcStat(t, procRoot, pid, "python3", 1)

	catalog := &Catalog{
		active:        newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		selfPID:       -1,
		retry:         make(map[int]resolveRetryState),
		allowTrace:    true,
		nameWhitelist: map[string]struct{}{"java": {}},
	}

	path, info, err := catalog.Resolve(pid)
	if err != nil {
		t.Fatalf("unexpected resolve error for non-whitelisted process: %v", err)
	}
	if info == nil {
		t.Fatal("expected resolved process info")
	}
	if path != "" || info.ExePath() != "" {
		t.Fatalf("expected binary lookup to be skipped, got path=%q exe=%q", path, info.ExePath())
	}
	if info.Traceable() {
		t.Fatal("expected non-whitelisted process to stay non-traceable")
	}
}

func TestResolveSkipsExeLookupWhenEnvWhitelistMisses(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4324
	writeFakeProcStat(t, procRoot, pid, "python3", 1)
	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	if err := os.WriteFile(filepath.Join(pidDir, "environ"), []byte("OTHER=value\x00"), 0o644); err != nil {
		t.Fatal(err)
	}

	catalog := &Catalog{
		active:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:      newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		selfPID:      -1,
		retry:        make(map[int]resolveRetryState),
		allowTrace:   true,
		envWhitelist: map[string]struct{}{"DD_SERVICE": {}},
	}
	catalog.envKeys = buildEnvKeySet(catalog.serviceEnv, catalog.envWhitelist, catalog.envBlacklist)

	path, info, err := catalog.Resolve(pid)
	if err != nil {
		t.Fatalf("unexpected resolve error for env whitelist miss: %v", err)
	}
	if info == nil {
		t.Fatal("expected resolved process info")
	}
	if path != "" || info.ExePath() != "" {
		t.Fatalf("expected binary lookup to be skipped, got path=%q exe=%q", path, info.ExePath())
	}
	if info.Traceable() {
		t.Fatal("expected env whitelist miss to stay non-traceable")
	}
}

func TestResolveWithProcFDUsesStableProcContext(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 5333
	writeFakeProcStat(t, procRoot, pid, "python3", 1)
	writeFakeProcCmdline(t, procRoot, pid, "python3", "/opt/app/server.py")
	writeFakeProcEnviron(t, procRoot, pid, "DD_SERVICE=checkout")

	pidDir := filepath.Join(procRoot, strconv.Itoa(pid))
	rootfs := filepath.Join(procRoot, "rootfs")
	binPath := filepath.Join(rootfs, "usr", "bin", "python3")
	if err := os.MkdirAll(filepath.Join(pidDir, "ns"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "ns", "mnt"), []byte("mntns"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rootfs, filepath.Join(pidDir, "root")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/usr/bin/python3", filepath.Join(pidDir, "exe")); err != nil {
		t.Fatal(err)
	}

	procDirFD, err := openProcessDir(pid)
	if err != nil {
		t.Fatalf("open process dir: %v", err)
	}
	defer unix.Close(procDirFD) //nolint:errcheck

	catalog := &Catalog{
		active:       newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted:      newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		retry:        make(map[int]resolveRetryState),
		allowTrace:   true,
		traceAllProc: true,
		selfPID:      -1,
		serviceEnv:   []string{"DD_SERVICE"},
	}
	catalog.envKeys = buildEnvKeySet(catalog.serviceEnv, catalog.envWhitelist, catalog.envBlacklist)

	path, info, err := catalog.ResolveWithProcFD(pid, procDirFD)
	if err != nil {
		t.Fatalf("ResolveWithProcFD() error = %v", err)
	}
	if info == nil {
		t.Fatal("expected process info")
	}
	if path != HostRoot(binPath) {
		t.Fatalf("ResolveWithProcFD() path = %q, want %q", path, HostRoot(binPath))
	}
	if info.ServiceName() != "checkout" {
		t.Fatalf("expected env-derived service name, got %q", info.ServiceName())
	}
	if !info.Traceable() {
		t.Fatal("expected traced process")
	}
}

func TestResolveWithProcFDDoesNotSeeReusedPID(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 5444
	writeFakeProcStat(t, procRoot, pid, "python3", 1)

	procDirFD, err := openProcessDir(pid)
	if err != nil {
		t.Fatalf("open process dir: %v", err)
	}
	defer unix.Close(procDirFD) //nolint:errcheck

	if err := os.RemoveAll(filepath.Join(procRoot, strconv.Itoa(pid))); err != nil {
		t.Fatalf("remove old pid dir: %v", err)
	}
	writeFakeProcStat(t, procRoot, pid, "java", 1)

	catalog := &Catalog{
		active:  newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		deleted: newCache(processCatalogCacheMaxCost, processCatalogCacheCounters),
		retry:   make(map[int]resolveRetryState),
		selfPID: -1,
	}

	if _, _, err := catalog.ResolveWithProcFD(pid, procDirFD); err == nil {
		t.Fatal("expected old procfd to stop resolving after pid reuse")
	}
}
