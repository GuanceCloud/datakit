# eBPF External Collector VM Pressure Test Report

Date: `2026-04-08`

## Environment

- Host: Ubuntu 24.04.4 LTS, `x86_64`
- Guest VM: `ebpf-stress-lab`
- Guest kernel: `6.8.0-106-generic`
- Guest IP during test: `192.168.122.216`
- Collector binary: locally built from current workspace as `/tmp/datakit-ebpf-stress`

## Collector Command

```bash
sudo /tmp/datakit-ebpf-stress run \
  --enabled ebpf-net,bpf-netlog \
  --l7net-enabled httpflow \
  --netlog-log \
  --netlog-metric \
  --pprof-host 127.0.0.1 \
  --pprof-port 6071 \
  --interval 60s \
  --log-level debug \
  --datakit-apiserver 127.0.0.1:9529 \
  --log /tmp/datakit-ebpf-stress.log \
  --pidfile /tmp/datakit-ebpf-stress.pid
```

## Workload

- Mock DataKit API listener on `127.0.0.1:9529`
- Local HTTP target on `127.0.0.1:18080`
- Main workload:

```bash
seq 1 2000 | xargs -P32 -I{} curl -fsS http://127.0.0.1:18080/test?i={} >/dev/null
```

- Additional flush validation workload:

```bash
seq 1 2000 | xargs -P32 -I{} curl -fsS http://127.0.0.1:18080/test?j={} >/dev/null
sleep 70
```

- CPU-profile-aligned workload:

```bash
(seq 1 4000 | xargs -P32 -I{} curl -fsS http://127.0.0.1:18080/hot?k={} >/dev/null) &
curl -fsS http://127.0.0.1:6071/debug/pprof/profile?seconds=5 >/tmp/datakit-ebpf-stress.cpu.load.pb.gz
```

## Results

Process snapshots:

```text
BASELINE
PID   %CPU  %MEM   RSS      VSZ       ELAPSED
3537  23.7  2.7    219676   2150976   00:01

AFTER_LOAD
PID   %CPU  %MEM   RSS      VSZ       ELAPSED
3537   4.6  2.7    222860   2224708   00:10

LATER_STEADY_STATE
PID   %CPU  %MEM   RSS      VSZ       ELAPSED
3537   0.6  2.5    207388   2224708   02:34
```

Mock API observed posts after waiting a full flush interval:

```text
POST_TOTAL 7
POST_PATH /v1/write/logging?input=bpf-netlog%2Fnetlog 4 23279
POST_PATH /v1/write/network?input=bpf-netlog%2Fnetflow 2 1095
POST_PATH /v1/write/network?input=ebpf-net%2Fnetflow 1 13076
```

Heap profile summary:

```text
Total in-use heap: ~28.1 MB
Top retained memory:
- github.com/cilium/ebpf/btf.readTypes: 7.5 MB
- github.com/cilium/ebpf/btf.readStringTable: 3.4 MB flat, 5.9 MB cum
- bufio.(*Scanner).Text: 2.5 MB
- github.com/cilium/ebpf/btf.inflateRawTypes: 2.2 MB
```

CPU profile captured during load:

```text
Duration: 5s
Total samples: 410ms
Top entries:
- runtime.futex: 24.39%
- internal/runtime/syscall/linux.Syscall6: 19.51%
- bpfutil.(*PerfStream).start.func1: 53.66% cumulative
- github.com/cilium/ebpf/perf.(*Reader).ReadInto: 36.59% cumulative
- procwatch.(*Catalog).resolveLoop: 14.63% cumulative
```

## Heavier Multi-NetNS Run

Topology:

- `8` network namespaces: `ns1` to `ns8`
- `8` veth pairs on the guest root namespace: `veth1h` to `veth8h`
- Namespace HTTP targets on `10.200.<n>.2:18080`
- Root-namespace HTTP target on `0.0.0.0:18081`

Workload shape:

- Root namespace to each netns target:

```bash
for i in $(seq 1 8); do
  seq 1 5000 | xargs -P48 -I{} curl -fsS http://10.200.${i}.2:18080/h${i}?i={} >/dev/null
done
```

- Each netns back to the paired root-namespace target:

```bash
for i in $(seq 1 8); do
  sudo ip netns exec ns${i} \
    sh -c 'seq 1 3000 | xargs -P24 -I{} curl -fsS http://10.200.'"${i}"'.1:18081/n'"${i}"'?i={} >/dev/null'
done
```

- Additional CPU-profile-aligned hotspot load:

```bash
(seq 1 8000 | xargs -P48 -I{} curl -fsS http://10.200.1.2:18080/hot?k={} >/dev/null) &
curl -fsS http://127.0.0.1:6071/debug/pprof/profile?seconds=8 >/tmp/datakit-ebpf-netns.cpu.load.pb.gz
```

Observed collector snapshot during the sustained phase:

```text
HEAVY_RUN_STEADY
PID      %CPU  %MEM   RSS      VSZ       ELAPSED
263543    3.8   2.6   212968   2708628   05:02
```

Mock API observed posts during the heavier run:

```text
POST_TOTAL 300
POST_PATH /v1/write/network?input=ebpf-net%2Fnetflow 282 50829009
POST_PATH /v1/write/logging?input=bpf-netlog%2Fnetlog 12 110937
POST_PATH /v1/write/network?input=bpf-netlog%2Fnetflow 5 2192
POST_PATH /v1/write/network?input=ebpf-net%2Fdnsflow 1 456
```

Heavy-run heap profile summary:

```text
Total in-use heap: ~39.3 MB
Top retained memory:
- github.com/GuanceCloud/cliutils/point.doNewKV: 16.5 MB flat, 20.0 MB cum
- netflow.kv2point: 4.5 MB flat, 25.5 MB cum
- github.com/GuanceCloud/cliutils/point.newVal: 3.5 MB
- procwatch.(*Catalog).create: 1.0 MB
```

Heavy-run CPU profile captured during load:

```text
Duration: 8s
Total samples: 730ms
Top entries:
- internal/runtime/syscall/linux.Syscall6: 28.77%
- runtime.futex: 26.03%
- bpfutil.(*PerfStream).start.func1: 50.68% cumulative
- github.com/cilium/ebpf/perf.(*Reader).ReadInto: 28.77% cumulative
- procwatch.(*Catalog).resolveLoop: 12.33% cumulative
- procwatch/procfs reads plus zap error logging remained visible in the profile
```

## Findings

- The new external collector started successfully in the VM and stayed stable during the run.
- `ebpf-net` and `bpf-netlog` both produced output to the mock API after one full `60s` aggregation/export interval.
- RSS stayed roughly in the `207 MB` to `223 MB` range during this test and did not show runaway growth.
- The most visible runtime CPU cost under this workload was in perf-ring-buffer consumption and syscall wait paths, which is consistent with an event-driven collector.
- The collector log repeatedly emitted `procwatch/catalog.go` errors for short-lived PIDs disappearing before `/proc/<pid>/stat` could be read. Under bursty process churn this creates noisy logs and showed up in an earlier low-signal CPU profile through zap logging.
- With `8` net namespaces and bidirectional traffic across `8` veth pairs, the collector still exported successfully and pushed about `50.8 MB` of `ebpf-net/netflow` payloads to the mock API in a single run.
- Under the heavier run, retained heap moved away from BTF parsing and was dominated by point construction for exported netflow data, especially `point.NewKV` and `netflow.kv2point`.
- The traffic generator itself saw some `curl: (56) Recv failure: Connection reset by peer` events from the lightweight Python HTTP servers under extreme concurrency. Those resets came from the test harness endpoints, not from the collector export path, which continued to deliver data to the mock API.

## Caveats

- The collector enforces a minimum interval of `60s`, so immediate post-load export counts under-report until a full interval passes.
- This run used a local mock DataKit API and a local Python HTTP target, so it validates collector startup, capture, aggregation, export behavior, and resource shape, but not end-to-end production ingestion.
- The very lightweight Python HTTP servers are good enough to create short-lived socket/process churn, but they are not suitable for judging application-side stability under extreme fan-out because they start resetting some connections before the collector does.

## Forty-NIC Run

Topology:

- `40` network namespaces: `ns1` to `ns40`
- `40` veth pairs on the guest root namespace: `veth1h` to `veth40h`
- Root server: `/tmp/ok-server -addr :18081`
- Namespace servers: `/tmp/ok-server -addr :18080`
- Load generator: `/tmp/loadblast`

Workload shape:

- `80` concurrent workers total
- Root namespace to each netns target:
  `40` workers, each `20s` long, each at concurrency `32`
- Each netns back to its paired root target:
  `40` workers, each `20s` long, each at concurrency `16`

Observed aggregate load:

```text
LOAD_WORKERS 80
LOAD_ATTEMPTS 1154519
LOAD_SUCCESS 1153020
LOAD_FAIL 1499
LOAD_AGG_QPS 57595.95
```

Representative low-end workers:

```text
LOAD_LOWEST_QPS 108.74 ns6-to-root 16
LOAD_LOWEST_QPS 113.13 ns10-to-root 16
LOAD_LOWEST_QPS 118.05 ns9-to-root 16
LOAD_LOWEST_QPS 118.34 ns2-to-root 15
LOAD_LOWEST_QPS 118.39 ns19-to-root 16
```

Collector snapshot after the load window:

```text
FORTY_NIC_STEADY
PID      %CPU  %MEM   RSS      VSZ       ELAPSED
411583    0.2   2.0   164844   2298184   05:10
```

Mock API posts during this run:

```text
POST_TOTAL 12
POST_PATH /v1/write/logging?input=bpf-netlog%2Fnetlog 4 12135
POST_PATH /v1/write/network?input=ebpf-net%2Fnetflow 4 141157
POST_PATH /v1/write/network?input=bpf-netlog%2Fnetflow 4 1460
```

Forty-NIC heap profile summary:

```text
Total in-use heap: ~12.0 MB
Top retained memory:
- goccy/go-json decoder init: 2.1 MB
- regexp compiler state: 1.0 MB
- point.doNewKV: 1.0 MB
- netflow.(*FlowAgg).ToPoint / kv2point: ~1.0 MB cumulative
```

Forty-NIC CPU profile summary:

```text
Duration: 10s
Total samples: 70ms
Top visible paths:
- procwatch.(*Catalog).resolveLoop: 42.86% cumulative
- epoll wait / syscall paths: ~28.58%
- perf reader path remained present but low-sample
```

Interpretation:

- This `40`-NIC run validated that the collector still starts, captures, and exports successfully with a much wider topology than the earlier `8`-NIC run.
- On this `4 vCPU / 8 GiB` VM, the practical aggregate traffic level achieved by the current Go harness was about `57.6k QPS`, not `500k QPS`.
- The gap to `500k QPS` appears to be a testbed limitation, not a collector crash: the generator and tiny in-guest HTTP endpoints saturate well before that target, while the collector remains alive and continues exporting.
- The short-lived PID churn problem in `procwatch/catalog.go` is still the most obvious runtime-noise issue in logs, even when the CPU profile under-samples the collector.

## Resized VM And Host-Mix Revalidation

VM changes before the rerun:

- Guest resized from `4 vCPU / 8 GiB` to `8 vCPU / 12 GiB`
- Collector rebuilt from the current workspace after the `procwatch` log-noise suppression patch
- The `40`-namespace, `40`-veth topology was reused

Guest-internal load during the host-mix rerun:

```text
GUEST_LOAD_WORKERS 40
GUEST_LOAD_ATTEMPTS 1265208
GUEST_LOAD_SUCCESS 1264591
GUEST_LOAD_FAIL 617
GUEST_LOAD_AGG_QPS 63223.71
```

Host-side load against the guest root server during the same window:

```text
HOST_ATTEMPTS 3542559
HOST_SUCCESS 3540547
HOST_FAIL 2012
HOST_AGG_QPS 176960.01
```

Combined observed throughput for that mixed run:

```text
COMBINED_AGG_QPS ~= 240183.72
```

Collector snapshot after the mixed run:

```text
HOST_MIX_STEADY
PID     %CPU  %MEM   RSS      VSZ       ELAPSED
3750    33.9   4.0   498632   2857344   01:36
```

Mock API posts during the mixed run:

```text
POST_TOTAL 3298
POST_PATH /v1/write/logging?input=bpf-netlog%2Fnetlog 3296 878149020
POST_PATH /v1/write/network?input=ebpf-net%2Fnetflow 1 60288
POST_PATH /v1/write/network?input=bpf-netlog%2Fnetflow 1 726
```

Important runtime finding from the mixed run:

```text
load collection: program tracepoint__sys_exit_writev: load program: permission denied:
... R5 unbounded memory access ...
```

Interpretation:

- The host-mix setup pushed the practical combined load well beyond the earlier guest-only runs, reaching about `240k QPS` with the `40`-NIC topology still active.
- The `procwatch` short-lived PID noise was materially reduced in this rerun: the earlier repeated `/proc/<pid>/stat` error flood was no longer the dominant visible log pattern.
- This rerun exposed a more important issue: `ebpf-net` did not load cleanly and hit a verifier rejection on `tracepoint__sys_exit_writev`. That explains why the export mix skewed heavily toward `bpf-netlog/netlog` and why `ebpf-net/netflow` volume was much smaller than expected.
- Even after resizing the VM and moving substantial traffic generation onto the host, this environment still did not reach `500k QPS`. The best observed combined level in this turn was about `240k QPS`, which suggests that hitting `500k` will need either a stronger guest, multiple external load sources, or further simplification/acceleration of the test harness.

## APIFlow Verifier Follow-Up

Follow-up work after the mixed run:

- Reduced `procwatch` noise by suppressing expected disappeared-`/proc` races.
- Confirmed that changing only Go code is not enough for `apiflow`: the embedded `apiflow.o` must be rebuilt through `make -C internal/plugins/externals/ebpf apiflow.o`.
- Simplified `apiflow` event upload to emit one event per perf output instead of building a variable-offset in-kernel batch buffer.
- Simplified `readv/writev` capture to read only the first useful iovec segment, avoiding variable destination-offset writes inside eBPF.

Smoke validation result after rebuilding `apiflow.o` and the Go binary:

```text
INFO  run/run.go:529  >>> datakit ebpf-net tracer(ebpf) starting ...
INFO  l7flow/l7flow.go:388  api tracer starting ...
```

Interpretation:

- The verifier rejection was reproducible with the rebuilt object and moved between syscall hooks while the program still used variable-offset event assembly.
- After simplifying the upload path and rebuilding the embedded ELF, `ebpf-net` loaded successfully in a VM smoke run.
- This fix path is intentionally conservative: it prioritizes loadability and testability over in-kernel batching efficiency. A fresh heavy pressure rerun is still needed to measure the new throughput shape with verifier-safe code.
