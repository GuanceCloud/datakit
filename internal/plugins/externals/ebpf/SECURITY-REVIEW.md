# eBPF Collector Security Review

This document records the April 6, 2026 security audit and hardening work for `internal/plugins/externals/ebpf`.

## Scope

- eBPF C programs under `internal/c`
- Go-side event readers, process enrichment, and runtime glue
- Runtime behavior under malformed, truncated, or high-rate event input

## Findings

### 1. Truncated perf-event records could crash user-space handlers

Affected readers previously dereferenced `&data[0]` or parsed variable-length payloads without validating `len(data)` first:

- `internal/bashhistory/bash_history.go`
- `internal/netflow/netflow_linux.go`
- `internal/procwatch/runtime.go`
- `internal/l7flow/net_tracer.go`

Risk:

- malformed or truncated records could trigger panic or out-of-bounds reads in Go
- this creates a denial-of-service path for the collector

Fix:

- added explicit minimum-size checks before any `unsafe.Pointer(&data[0])`
- added payload boundary checks for batched L7 events before reading headers or payload slices
- malformed records are now dropped instead of crashing the collector

### 2. Hot-path callbacks could block on bounded channels

Affected paths:

- `bashhistory` event queue
- `netflow` closed-connection queue
- `procwatch` async PID resolution queue

Risk:

- under bursty or attacker-driven event rates, the perf callback goroutine could block on channel send
- a blocked callback can stall event consumption and amplify data loss or service degradation

Fix:

- converted queue sends to non-blocking `select` with drop-on-overflow behavior
- `procwatch.ResolveLater()` now clears the pending marker if the queue is full, allowing safe retry later

### 3. `l7flow` performed synchronous `/proc` resolution in the perf callback

Affected path:

- `internal/l7flow/net_tracer.go`

Risk:

- network traffic could force expensive `/proc` and process-resolution work on the packet/event hot path
- this increases exposure to CPU and I/O exhaustion

Fix:

- changed `l7flow` process enrichment to cache-first lookup
- if process info is missing, it now schedules asynchronous resolution via `ResolveLater()` instead of resolving synchronously in the perf callback

### 4. `bash_history` eBPF probe did not short-circuit `NULL` return values

Affected path:

- `internal/c/bash_history/bash_history.c`

Risk:

- `readline()` may return `NULL`; the previous program still attempted to read from that pointer
- verifier/runtime would usually reject the access or the helper would fail, but it is still an unnecessary unsafe path in kernel-side code

Fix:

- added a `NULL` guard before reading the returned line buffer

### 5. L7 perf uploads sent the entire backing buffer instead of the used bytes

Affected paths:

- `internal/c/apiflow/l7_utils.h`
- `internal/c/apiflow/apiflow.c`

Risk:

- every flush uploaded `sizeof(network_events_t)` even when only a small prefix was populated
- this wastes CPU and bandwidth on the kernel-to-user path
- it can also carry stale payload bytes from prior batches inside the unused tail of the perf sample

Fix:

- changed L7 perf-event uploads to use `sizeof(event_rec_t) + events->rec.bytes`
- changed per-event payload copy to only copy `capture_size` bytes
- changed the batch-space check to use actual event size instead of the full fixed payload envelope

### 6. Several kernel probes lacked cleanup or pointer/bounds guards on failure paths

Affected paths:

- `internal/c/netflow/netflow.c`
- `internal/c/apiflow/l7_utils.h`

Risk:

- failed `sendfile()` calls could be cast to `size_t`, producing bogus very large byte counters
- failed `udp_recvmsg()` calls left temporary map entries behind, which could accumulate under error-heavy workloads
- `sockfd_lookup_light` kretprobe dereferenced the returned socket pointer without checking for `NULL`
- `get_socket_from_fd()` indexed `fdtable->fd` without checking `max_fds`
- `inet_bind` and `inet6_bind` read user sockaddr pointers without checking for `NULL`

Fix:

- guarded `sendfile()` return values and clean up temp state on failure
- always delete `udp_recvmsg` temp state before handling the return value
- added `NULL` guard for `sockfd_lookup_light` return socket
- added `fdtable->max_fds` bounds check before indexing file descriptors
- added `NULL` guards for bind sockaddr arguments

### 7. `/proc`-derived binary paths were accepted without normalization

Affected paths:

- `internal/procwatch/procfs_linux.go`

Risk:

- executable and shared-library paths from `/proc/<pid>/exe` and `/proc/<pid>/maps` may include deleted markers, pseudo-paths, or non-regular files
- repeatedly attempting to resolve or attach uprobes to those paths adds noise and creates avoidable failure churn in complex containerized environments

Fix:

- normalize procfs-derived paths before use
- strip the trailing ` (deleted)` marker
- reject non-absolute and pseudo paths such as bracketed entries
- require the resolved host path to be a regular file before treating it as attachable

## Validation

Unit tests added:

- `internal/bashhistory/bash_history_test.go`
- `internal/netflow/netflow_test.go`
- `internal/l7flow/net_tracer_test.go`
- `internal/procwatch/runtime_test.go`
- `internal/procwatch/catalog_test.go`

Commands run:

```bash
go test ./internal/plugins/externals/ebpf/internal/bashhistory \
  ./internal/plugins/externals/ebpf/internal/netflow \
  ./internal/plugins/externals/ebpf/internal/procwatch \
  ./internal/plugins/externals/ebpf/internal/l7flow

go test ./internal/plugins/externals/ebpf/...

make -C internal/plugins/externals/ebpf bindata

go build -o /tmp/datakit-ebpf-security ./internal/plugins/externals/ebpf/cmd/datakit-ebpf

printf '%s\n' '#!mrst-vir##!' | sudo -S timeout 8s \
  /tmp/datakit-ebpf-security run \
  --enabled ebpf-net,bpf-netlog \
  --netlog-log --netlog-metric \
  --pprof-host 127.0.0.1 --pprof-port 6108
```

Result:

- new and existing eBPF unit tests passed
- bindata regeneration succeeded
- `datakit-ebpf` rebuilt successfully
- runtime smoke test started successfully with `ebpf-net` and `bpf-netlog`

## Residual Risks To Watch

- `unsafe` is still used where Go must decode packed C structs; this is expected, but every new perf/ringbuf handler should follow the same length-check pattern added in this review
- drop-on-overflow protects liveness, but it trades some completeness for resilience; any future tuning should treat bounded loss as preferable to callback blockage
