# eBPF Netlog VM Pressure Test Report

Date: `2026-04-09`

## Scope

This report captures a pure `bpf-netlog` pressure run after the netlog rewrite work was integrated into the runtime path.

Unlike the earlier mixed runs, this validation intentionally disables `ebpf-net` and profiles only:

- NIC packet capture through raw sockets
- packet-to-flow handling in `internal/l4log`
- HTTP parsing on top of the shared bounded stream reassembler
- netlog log and netflow export

## VM Harness

The reusable VM harness added in this round lives under `internal/plugins/externals/ebpf/vm`:

- `session-netlog-vm-up.sh`
- `session-netlog-vm-down.sh`
- `session-netlog-profile.sh`
- `ok_server.go`
- `loadblast.go`

For this run, the harness targeted the already running guest `ebpf-stress-lab` at `192.168.122.216`.

## Environment

- Host: Ubuntu 24.04.4 LTS, `x86_64`
- Guest VM: `ebpf-stress-lab`
- Guest kernel: `6.8.0-106-generic`
- Collector binary: locally built from the current workspace and copied in as `/tmp/datakit-ebpf-netlog`

## Collector Command

```bash
sudo /tmp/datakit-ebpf-netlog run \
  --enabled bpf-netlog \
  --netlog-log \
  --netlog-metric \
  --pprof-host 127.0.0.1 \
  --pprof-port 6071 \
  --interval 60s \
  --log-level info \
  --datakit-apiserver 127.0.0.1:9529 \
  --log /tmp/datakit-ebpf-netlog.log \
  --pidfile /tmp/datakit-ebpf-netlog.pid
```

## Workload

Topology:

- `40` network namespaces: `ns1` to `ns40`
- `40` veth pairs on the guest root namespace: `veth1h` to `veth40h`
- Root server: `/tmp/ok-server -addr :18081`
- Namespace servers: `/tmp/ok-server -addr :18080`
- Load generator: `/tmp/loadblast`

Load shape:

- `80` workers total
- Host namespace to each netns target:
  `40` workers, each `20s`, each at concurrency `32`
- Each netns back to the paired root target:
  `40` workers, each `20s`, each at concurrency `16`

## Results

Observed load:

```text
LOAD_WORKERS 80
LOAD_ATTEMPTS 2082514
LOAD_SUCCESS 2080944
LOAD_FAIL 1570
LOAD_AGG_QPS 104008.24
```

Representative low-end workers:

```text
LOAD_LOWEST_QPS 192.1 ns13-to-root 16
LOAD_LOWEST_QPS 202.39 ns1-to-root 16
LOAD_LOWEST_QPS 202.79 ns25-to-root 14
LOAD_LOWEST_QPS 203.05 ns18-to-root 16
LOAD_LOWEST_QPS 206.69 ns15-to-root 16
```

Collector snapshots:

```text
BASELINE
PID     %CPU  %MEM   RSS    VSZ      ELAPSED
39828   0.9   0.4    50404  2057484  00:01

AFTER_LOAD
PID     %CPU  %MEM   RSS    VSZ      ELAPSED
39828   0.0   0.5    63256  2066516  01:36
```

Mock API posts:

```text
POST_TOTAL 12
POST_PATH /v1/write/logging?input=bpf-netlog%2Fnetlog 9 1423501
POST_PATH /v1/write/network?input=bpf-netlog%2Fnetflow 3 1097
```

Heap profile summary:

```text
Total in-use heap: ~11.6 MB
Top retained memory:
- runtime / package init allocations: dominant
- l4log.clonePayload: ~513 KB
- l4log.(*PktChunk).GetMacID: ~512 KB
```

CPU profile summary:

```text
Duration: 10s
Total samples: 0
```

## Findings

- The new VM harness successfully recreated a pure netlog workload and exported both log and netflow payloads.
- With a `40`-namespace topology, the current harness reached about `104k` aggregate successful requests/sec inside the guest.
- Collector RSS stayed around `50 MB` to `63 MB` during the run and did not show runaway growth.
- On this pure netlog workload, user-space Go CPU remained so low that `pprof/profile` did not collect a meaningful CPU sample set over `10s`.
- Heap retained memory was modest; the only clearly visible runtime allocations attributable to the new netlog path were small retained slices from out-of-order payload buffering and MAC-ID mapping.

## Interpretation

- The runtime rewrite has pushed a large part of the netlog hot path low enough that, in this VM shape, Go-side CPU is no longer the first visible bottleneck during a pure netlog run.
- The absence of CPU samples does not mean the workload is idle overall; it means the collector spends too little time in sampled Go user-space work for this VM and load shape.
- The remaining optimization work is better guided by:
  - targeted microbenchmarks for log/point shaping
  - heap/retained-allocation signals such as payload copies in reorder buffering
  - stronger future load sources if we need actionable Go CPU profiles

## Follow-Up Optimization In The Same Round

After this VM run, the netlog point-building path in `internal/l4log/data_point.go` was tightened by replacing repeated `point.KVs.Set` / `SetTag` scans with append-only KV construction for unique keys.

Focused benchmark comparison:

```text
Before
BenchmarkBuildHTTPLog   3019 ns/op   3658 B/op   64 allocs/op
BenchmarkBuildTCPLog    5343 ns/op   6457 B/op   60 allocs/op

After
BenchmarkBuildHTTPLog   1358 ns/op   3146 B/op   63 allocs/op
BenchmarkBuildTCPLog    2902 ns/op   6004 B/op   62 allocs/op
```

Interpretation:

- `buildHTTPLog` improved by about `2.2x`
- `buildTCPLog` improved by about `1.8x`
- allocation size also dropped further on both paths
- this is a better next-step optimization target than chasing a zero-sample CPU profile
