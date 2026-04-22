# eBPF Unit Test And Performance Report

## Scope

This round focused on the eBPF collector paths that are both easy to regress and easy to overpay for at runtime:

- `netflow` flow aggregation
- `l4log` TCP/HTTP aggregation
- `netlog` event export path
- `l4log` connection map churn
- `l7flow` pooled payload objects
- `procwatch` process catalog regression coverage

## Added Tests

### Unit tests

- `internal/procwatch/catalog_test.go`
  - verifies cached process info still returns the resolved executable path
- `internal/l4log/agg_bench_test.go`
  - verifies TCP aggregation updates counters and clears consumed metric state
  - verifies HTTP aggregation merges repeated requests into one bucket
- `internal/l4log/conn_map_test.go`
  - verifies connection map shrink is triggered after sustained delete pressure
- `internal/l4log/data_point_test.go`
  - verifies `buildTCPLog()` exports both `tx_retrans` and `rx_retrans`
- `internal/l7flow/pool_test.go`
  - verifies pooled `NetwrkData` objects are fully reset before reuse
- `internal/netflow/agg_bench_test.go`
  - verifies `FlowAgg.Append()` still aggregates bytes and TCP counters correctly

### Benchmarks

- `BenchmarkFlowAggAppend`
- `BenchmarkFlowAggTCPAppend`
- `BenchmarkFlowAggHTTPAppend`
- `BenchmarkConnMapInsertDelete`
- `BenchmarkBuildCommTags`
- `BenchmarkBuildHTTPLog`
- `BenchmarkBuildTCPLog`
- `BenchmarkNetwrkDataPool`

## Optimizations

### 1. Rewrote netflow RTT aggregation to avoid per-event slice growth

File:
- `internal/netflow/agg.go`

Before this change, `FlowAgg.Append()` appended every RTT sample into `[]int64` slices and later averaged them during point conversion.

Now it stores running totals:

- `rtt`
- `rttVar`
- `count`

This preserves the exported average while removing the main aggregation-specific allocation source from the append path.

### 2. Reduced connection map shrink overhead on hot insert/delete paths

File:
- `internal/l4log/conn_map.go`

Before this change, every insert and delete checked the shrink ratio and could trigger a full map copy.

Now shrink is gated by integer thresholds:

- at least `256` historical inserts
- at least `64` deletes since the last rebuild
- live ratio below `60%`

This keeps the reclaim behavior but avoids repeated ratio work and unnecessary rebuild attempts under normal churn.

### 3. Reduced netlog export overhead and fixed retransmit field output

File:
- `internal/l4log/data_point.go`

Changes:

- fixed a real bug where `buildTCPLog()` wrote `tx_retrans` twice and dropped `rx_retrans`
- replaced high-frequency `fmt.Sprintf("%d_%d", ...)` trace-id formatting with a lighter helper
- replaced dynamic `map[string]any` JSON message payload assembly with typed structs for:
  - HTTP netlog messages
  - HTTP/2 netlog messages
  - TCP netlog messages

This keeps the exported schema stable while reducing reflection and temporary map allocation in the log export path.

### 4. Reworked eBPF constant loading so cilium/ebpf can patch offsets reliably

Files:

- `internal/c/common/load_const.h`
- `internal/c/netflow/netflow_utils.h`
- `internal/c/conntrack/conntrack.c`
- `Makefile`

Changes:

- replaced legacy inline-asm immediate constants with `const volatile` globals
- preserved `.BTF` in built eBPF objects so `RewriteConstants()` can locate `.rodata` variables
- verified generated objects now contain both:
  - `.rodata`
  - `.BTF`

This unblocked real runtime startup for `ebpf-net` and `bpf-netlog` after the runtime migration away from DataDog's manager layer.

### 5. Fixed runtime attach compatibility issues

Files:

- `internal/bpfutil/session_linux.go`

Changes:

- fixed a self-deadlock in `Runtime.Start()` by not holding the runtime mutex while starting perf readers
- fixed kprobe attachment to avoid passing `RetprobeMaxActive` to plain `kprobe`, which tracefs rejects

These were required to get stable real-host startup and runtime profiling.

### 6. Reduced netflow C-side map churn and correctness issues

Files:

- `internal/c/netflow/netflow.c`
- `internal/c/netflow/bpfmap.h`

Changes:

- fixed `kretprobe__inet_csk_accept()` to pass the correct TCP protocol flag
- ignored invalid `segs <= 0` retransmit events and skipped retransmit updates when `read_connection_info()` fails
- guarded UDP payload length subtraction against unsigned underflow in `ip_make_skb` and `ip6_make_skb`
- kept UDP bind state namespace-aware by storing `netns + port` in the temporary bind map
- avoided reading temp bind values after map deletion by copying them first
- fixed `udp_destroy_sock` cleanup so UDP bind deletion uses the correct netns even when connection extraction fails

These changes reduce bogus map updates/deletes and avoid incorrect UDP listener direction inference across namespaces.

### 7. Rewrote procwatch hot paths to avoid gopsutil process parsing overhead

Files:

- `internal/procwatch/procfs_linux.go`
- `internal/procwatch/catalog.go`
- `internal/procwatch/runtime.go`
- `internal/netflow/netflow_linux.go`
- `internal/l7flow/net_tracer.go`

Changes:

- replaced `gopsutil/process.NewProcess()` and related stat/env/exe calls on the procwatch hot path with direct `/proc/<pid>` reads
- added lightweight helpers for:
  - PID enumeration
  - `/proc/<pid>/stat` parsing
  - `/proc/<pid>/environ` parsing
  - `/proc/<pid>/exe` resolution
- in non-tracing mode, exec events are now resolved asynchronously instead of fully synchronously on the perf-reader goroutine
- `netflow` keeps the hot closed-event path cache-first and uses deferred process resolution on cache miss
- `l7flow` keeps on-demand resolution, where name/service accuracy is more valuable than raw churn throughput

This shifts short-lived process churn away from the perf-reader critical path and removes the heaviest `gopsutil` stack from runtime profiles.

## Verification Commands

### Unit tests

```bash
go test ./internal/plugins/externals/ebpf/...
```

### Focused benchmark run

```bash
go test -run '^$' -bench 'Benchmark(FlowAgg|ConnMap)' -benchmem \
  ./internal/plugins/externals/ebpf/internal/netflow \
  ./internal/plugins/externals/ebpf/internal/l4log

go test -run '^$' -bench 'Benchmark(BuildCommTags|BuildHTTPLog|BuildTCPLog)' -benchmem \
  ./internal/plugins/externals/ebpf/internal/l4log

go test -run '^$' -bench 'BenchmarkNetwrkDataPool' -benchmem \
  ./internal/plugins/externals/ebpf/internal/l7flow
```

### Runtime smoke test

```bash
go build -o /tmp/datakit-ebpf-perf ./internal/plugins/externals/ebpf/cmd/datakit-ebpf

sudo /tmp/datakit-ebpf-perf run \
  --enabled ebpf-net,bpf-netlog \
  --netlog-log \
  --netlog-metric \
  --pprof-host 127.0.0.1 \
  --pprof-port 6071 \
  --interval 10s \
  --log-level debug \
  --log /tmp/datakit-ebpf-perf.log \
  --pidfile /tmp/datakit-ebpf-perf.pid

seq 1 200 | xargs -P16 -I{} curl -ks https://example.com >/dev/null

curl -s http://127.0.0.1:6071/debug/pprof/heap >/tmp/datakit-ebpf-perf.heap.pb.gz
curl -s 'http://127.0.0.1:6071/debug/pprof/profile?seconds=5' >/tmp/datakit-ebpf-perf.cpu.pb.gz

go tool pprof -top /tmp/datakit-ebpf-perf /tmp/datakit-ebpf-perf.heap.pb.gz
go tool pprof -top /tmp/datakit-ebpf-perf /tmp/datakit-ebpf-perf.cpu.pb.gz
```

## Benchmark Results

Environment:

- Date: `2026-04-04`
- Host CPU: `12th Gen Intel(R) Core(TM) i7-12700H`
- OS: `linux/amd64`

### netflow

After the RTT aggregation rewrite:

```text
BenchmarkFlowAggAppend-20    9677028    128.7 ns/op    0 B/op    0 allocs/op
```

Earlier in the same session before the rewrite, the same benchmark was:

```text
BenchmarkFlowAggAppend-20    5764778    240.5 ns/op    111 B/op    2 allocs/op
```

Interpretation:

- the RTT slice-growth allocation is gone
- the remaining key-building allocation was also removed by switching the aggregation key to raw IP arrays and numeric netns
- this path is now zero-allocation in the benchmarked hot append loop

### l4log

```text
BenchmarkFlowAggTCPAppend-20     12835690    134.6 ns/op      0 B/op       0 allocs/op
BenchmarkFlowAggHTTPAppend-20     8441517    145.8 ns/op      0 B/op       0 allocs/op
BenchmarkConnMapInsertDelete-20      3289    358202 ns/op  682103 B/op  1062 allocs/op
BenchmarkConnMapSteadyStateChurn-20 3801554    334.5 ns/op    216 B/op      0 allocs/op
BenchmarkBuildCommTags-20          904795    1252 ns/op      1901 B/op      6 allocs/op
BenchmarkBuildHTTPLog-20           436276    3019 ns/op      3658 B/op     64 allocs/op
BenchmarkBuildTCPLog-20            203248    5343 ns/op      6457 B/op     60 allocs/op
```

Interpretation:

- TCP and HTTP aggregation paths are allocation-free in the benchmarked hot append path
- `ConnMapInsertDelete` still allocates because the benchmark intentionally creates a fresh map and new entries on every loop
- the shrink logic is now less eager and lower-overhead under churn, even though the benchmark still measures the cost of map population itself
- `ConnMapSteadyStateChurn` isolates long-lived map maintenance cost from one-shot map creation
- `buildHTTPLog` improved from `7314 ns/op, 6823 B/op, 132 allocs/op` to `3019 ns/op, 3658 B/op, 64 allocs/op`
- `buildTCPLog` improved from `7654 ns/op, 9055 B/op, 123 allocs/op` to `5343 ns/op, 6457 B/op, 60 allocs/op`
- the large `netlog` improvement came from:
  - replacing dynamic JSON maps with typed structs
  - building common `point.KVs` once per connection and cloning them per event/chunk instead of recreating tags from a map each time
- `buildTCPLog` now exports the correct `rx_retrans` field instead of overwriting `tx_retrans`

### l7flow

```text
BenchmarkNetwrkDataPool-20    73471824    16.79 ns/op    0 B/op    0 allocs/op
```

Interpretation:

- pooled payload objects are being reused correctly
- the object reset path is allocation-free

## Runtime Observation

`datakit-ebpf` was launched with:

- `ebpf-net`
- `bpf-netlog`
- local `pprof`

Observed process snapshot during the smoke test:

```text
PID      %CPU   %MEM   RSS      VSZ        ELAPSED
2588342  0.1    0.1    53120    1881144    00:32
```

Heap profile sample:

```text
Type: inuse_space
Total: 10419.43kB
```

The top heap items in this run were dominated by runtime/init and Kubernetes/client-go related startup allocations rather than the collector hot paths themselves.

CPU profile sample:

```text
Duration: 15s, Total samples = 0
```

Even after increasing concurrent traffic and profile duration, the Go-side collector work still did not accumulate enough samples to show a user-space CPU hotspot in this environment. The live process remained around `0.1%` CPU and about `54MB` RSS during the run.

## Additional Runtime Rounds

Environment:

- Date: `2026-04-06`
- Workload:
  - `ebpf-net`
  - `bpf-netlog`
  - local HTTP target on `127.0.0.1:18080`
  - `seq 1 3000 | xargs -P32 curl http://127.0.0.1:18080`
  - local `pprof`
  - `kernel.bpf_stats_enabled=1`

### Round 1: after startup/runtime bug fixes, before procwatch rewrite

Representative CPU profile:

```text
Duration: 12s, Total samples = 2.01s
```

Representative heap profile:

```text
Type: inuse_space
Total: 57.37MB
```

Observed user-space hotspots:

- `procwatch.(*ProbeWatcher).handleEvent`
- `procwatch.(*Catalog).Resolve`
- `gopsutil/process.(*Process).fillFromStatWithContext`
- `gopsutil/internal/common.ReadLinesOffsetN`

Interpretation:

- short-lived process churn was making procwatch pay heavily for `gopsutil`-based `/proc` parsing
- this cost dominated Go-side runtime work more than `netflow` aggregation itself

### Round 2: after direct `/proc` rewrite in procwatch

Representative CPU profile:

```text
Duration: 12s, Total samples = 0.94s
```

Representative heap profile:

```text
Type: inuse_space
Total: 30.86MB
```

Observed changes:

- the `gopsutil` process stack disappeared from the hot path
- procwatch resolution cost fell to lightweight local helpers:
  - `readProcessEnvironMap`
  - `readProcessStat`
  - `readProcessExePath`
- compared with Round 1:
  - CPU samples dropped from `2.01s` to `0.94s`
  - heap in-use dropped from `57.37MB` to `30.86MB`

### Kernel-side BPF runtime observation

Using `bpftool prog show` with runtime stats enabled:

- `kprobe__sockfd_lookup_light`
  - `run_cnt 191291`
  - `run_time_ns 55537610`
- `kretprobe__sockfd_lookup_light`
  - `run_cnt 191274`
  - `run_time_ns 30782525`
- `kprobe__tcp_sendmsg`
  - `run_cnt 2530`
  - `run_time_ns 4160076`
- `kprobe__tcp_close`
  - `run_cnt 1558`
  - `run_time_ns 11215019`

Interpretation:

- on this host and workload, the main kernel-side probes are active but still inexpensive relative to the process churn cost in user space
- the highest event-volume probes are the socket-lookup probes, but their aggregate runtime is still only tens of milliseconds in this sample window

## Current Assessment

The collector hot paths look healthy after this round:

- `l4log` aggregation hot paths are already zero-allocation
- `netlog` export path is now benchmarked and no longer drops `rx_retrans`
- `netlog` export path is materially cheaper after reusing common tag KVs and removing dynamic JSON map assembly
- `l7flow` pooled payload reuse is zero-allocation
- `netflow` aggregation hot append path is now zero-allocation
- `conn_map` shrink behavior is less aggressive and safer for high-churn traffic
- runtime startup now works cleanly with:
  - `.rodata` constant rewriting
  - preserved `.BTF`
  - correct kprobe/kretprobe attach behavior
- procwatch no longer depends on `gopsutil` in its hottest runtime resolution path
- under a synthetic short-lived process + local HTTP churn workload, Go-side CPU and heap are materially lower than the pre-rewrite baseline

## Remaining Optimization Ideas

If another round is needed, the next highest-value targets are:

1. reduce `buildCommTags()` allocations further, most likely by avoiding the intermediate tag map when Kubernetes enrichment is not needed
2. evaluate whether `netlog` message JSON can be partially preformatted or pooled without hurting readability and compatibility
3. if a hotter runtime profile is still needed, drive traffic through a more local/high-QPS target so user-space collector work grows enough to sample reliably
