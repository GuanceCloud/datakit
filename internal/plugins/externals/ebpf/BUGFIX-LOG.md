# eBPF Bugfix Log

## 2026-04-04

### Scope

This log tracks investigation, testing, fixes, and performance work for `internal/plugins/externals/ebpf`.

### Completed Work

1. Replaced the DataDog `ebpf-manager` integration with an internal runtime built on `cilium/ebpf`.
2. Preserved Datadog ecosystem compatibility for `DD_SERVICE` and `x-datadog-*` headers.
3. Removed DataDog runtime and procfs helper dependencies from `go.mod`, `go.sum`, and `vendor/`.
4. Unified offset resolution through `BTF -> active guess -> constant patch` flow.
5. Added a local procfs implementation for mount/path resolution and shared-library discovery.
6. Added service-name discovery fallback for processes without `DD_SERVICE`-style environment variables.
7. Reduced process-watch overhead by:
   - caching mount namespace mountinfo resolution
   - shrinking process filter caches
   - deduplicating repeated async PID discovery requests
8. Replaced the old process-watcher runtime path with a new `procwatch` module:
   - `internal/plugins/externals/ebpf/internal/procwatch/catalog.go`
   - `internal/plugins/externals/ebpf/internal/procwatch/runtime.go`
   - `internal/plugins/externals/ebpf/internal/procwatch/library.go`
   - `internal/plugins/externals/ebpf/internal/procwatch/limiter.go`
9. Added a runtime repro harness for short-lived process/container teardown:
   - `internal/plugins/externals/ebpf/repro-procwatch-container-exit.sh`
   - `internal/plugins/externals/ebpf/doc/procwatch-exit-repro.md`

### Bugs Found and Fixed

#### Bug: migrated runtime could not start netflow because offset constants were not patchable

- Location:
  - `internal/c/common/load_const.h`
  - `internal/c/netflow/netflow_utils.h`
  - `internal/c/conntrack/conntrack.c`
  - `internal/plugins/externals/ebpf/Makefile`
- Symptom:
  - `ebpf-net` startup failed with:
    - `rewrite constants: some constants are missing from .rodata`
- Root cause:
  - the legacy C path still used inline-asm immediate constants instead of BPF global constants
  - build output also stripped `.BTF`, which `cilium/ebpf` needs for global constant metadata
- Fix:
  - replaced legacy inline-asm offset loads with `const volatile` BPF globals
  - preserved `.BTF` in the built objects
- Validation:
  - verified `.rodata` and `.BTF` exist in rebuilt objects
  - verified runtime progressed past constant-rewrite stage on real startup

#### Bug: plain kprobes were attached with retprobe-only maxactive settings

- Location:
  - `internal/bpfutil/session_linux.go`
- Symptom:
  - startup failed with:
    - `can only set maxactive on kretprobes`
- Root cause:
  - `attachProbe()` passed `RetprobeMaxActive` to both `kprobe` and `kretprobe`
- Fix:
  - plain `kprobe` now attaches without `RetprobeMaxActive`
  - only `kretprobe` keeps the maxactive option
- Validation:
  - verified `ebpf-net` and `bpf-netlog` can attach their kprobes on the current host

#### Bug: runtime start could deadlock while opening perf readers

- Location:
  - `internal/bpfutil/session_linux.go`
- Symptom:
  - startup could hang while `PerfStream.start()` tried to resolve maps
- Root cause:
  - `Runtime.Start()` held the runtime mutex while `PerfStream.start()` called back into `LookupMap()`
- Fix:
  - copied probe/reader slices under lock, then performed attach/start work outside the mutex
- Validation:
  - verified startup proceeds normally after the change

#### Bug: netflow C path had multiple namespace and event-accounting correctness issues

- Location:
  - `internal/c/netflow/netflow.c`
  - `internal/c/netflow/bpfmap.h`
- Symptoms:
  - UDP bind/listener direction could be wrong across namespaces
  - retransmit accounting could run on invalid connection state
  - UDP payload size handling could underflow
- Fix:
  - made UDP bind tracking namespace-aware
  - rejected invalid retransmit/update cases
  - guarded UDP length subtraction
  - fixed a wrong TCP protocol flag in the accept path
- Validation:
  - rebuilt objects
  - package tests passed
  - live runtime attached and processed traffic successfully

#### Bug: cached process lookups lost executable path

- Location:
  - legacy process-watch cache path
- Symptom:
  - `ProcessFilter.TryAdd()` returned an empty `binPath` when a PID was already cached.
  - the old process-scheduler path depended on `binPath` to decide whether a process should attach uprobes.
  - Repeated handling of the same process could therefore skip uprobe attachment incorrectly.
- Root cause:
  - An optimization added an early return path on cache hit, but did not preserve the resolved executable path.
- Fix:
  - Added `exePath` to `ProcInfo`.
  - Cached `ProcInfo` now stores and returns the resolved executable path consistently.
  - Added regression test `TestTryAddReturnsCachedExePath`.
- Validation:
  - Verified by package test and full eBPF test run.
  - The same regression coverage is now preserved in `internal/plugins/externals/ebpf/internal/procwatch/catalog_test.go`.

#### Bug: legacy process watcher reused deleted process state and unbalanced uprobe lifecycle

- Location:
  - legacy process-watcher runtime path, now replaced by `procwatch`
- Symptom:
  - the old watcher could reuse stale process metadata after process exit or `exec`.
  - dynamic Go uprobes could be detached or retained with the wrong lifetime.
  - in containerized workloads this can turn into probe unregister drift and delayed process/container teardown behavior.
- Root cause:
  - `ProcInfo.deleted` existed but was never set, so deleted cache entries were not distinguishable from live ones.
  - `ProcessFilter.TryAdd()` and `AsyncTryAdd()` treated `procDel` entries as reusable live cache.
  - `sched_process_exec` did not release the previous process image state before attaching the new executable for the same PID.
  - `PassiveFileUpdater` used a per-binary refcount for dynamic uprobes, but `AddRef()` was never called, so `Forget()` was inherently unbalanced.
- Fix:
  - `Delete()` now marks `ProcInfo.deleted = true`.
  - added a live-cache path so `TryAdd()` / `AsyncTryAdd()` only reuse active process cache entries.
  - stale deleted cache is dropped once a fresh `TryAdd()` succeeds.
  - added `SchedTracer.releaseProcess()` and reused it on both `sched_process_exec` and `sched_process_exit`.
  - dynamic uprobe refcounts are now incremented only after successful attach/inject reuse, and decremented during release.
  - `SchedTracer.Stop()` now clears the per-binary updater state during shutdown.
- Regression tests:
  - `internal/plugins/externals/ebpf/internal/procwatch/catalog_test.go`
    - `TestResolveIgnoresDeletedCache`
  - `internal/plugins/externals/ebpf/internal/procwatch/runtime_test.go`
    - `TestBinaryRegistryRelease`
- Validation:
  - `go test ./internal/plugins/externals/ebpf/internal/procwatch`
  - `go test ./internal/plugins/externals/ebpf/...`

### Testing Performed

1. Unit and package tests:

```bash
go test ./internal/plugins/externals/ebpf/internal/procwatch ./internal/plugins/externals/ebpf/...
```

2. Real runtime smoke test:

```bash
sudo /tmp/datakit-ebpf-opt run \
  --enabled ebpf-net,bpf-netlog \
  --netlog-log \
  --netlog-metric \
  --pprof-host 127.0.0.1 \
  --pprof-port 6064 \
  --interval 10s \
  --log-level debug
```

3. Runtime profiling:

```bash
curl http://127.0.0.1:6064/debug/pprof/heap
curl http://127.0.0.1:6064/debug/pprof/profile?seconds=5
go tool pprof -top /tmp/datakit-ebpf-opt /tmp/profile.pb.gz
```

4. Procwatch teardown repro harness:

```bash
cd internal/plugins/externals/ebpf
SUDO_PASSWORD='your-sudo-password' KEEP_REPRO_ARTIFACTS=1 ./repro-procwatch-container-exit.sh
```

- current observed state on this host:
  - `bpftool link show` confirms `sched_process_fork/exec/exit` tracepoints are attached
  - the dynamic uprobe path was not covered yet by the local repro workload, so the script now dumps `bpftool` and log context on failure for follow-up debugging

### Performance Notes

- Before optimization, heap `inuse_space` was approximately `56.6MB`.
- After process-watch cache/path-resolution optimization, heap `inuse_space` dropped to approximately `9.8MB`.
- CPU hotspots were concentrated in `ProcessFilter.TryAdd()` and procfs path resolution.
- Follow-up deduplication work reduced repeated async PID processing.
- Additional runtime profiling on `2026-04-06` showed a new dominant cost after the runtime migration:
  - procwatch resolving large amounts of short-lived process churn
- Rewriting procwatch's hot path from `gopsutil` to direct `/proc` access reduced a representative runtime profile from:
  - `2.01s` sampled CPU to `0.94s`
  - `57.37MB` heap in-use to `30.86MB`
