---
title     : 'Profiling Golang'
summary   : 'Profling Golang applications'
tags:
  - 'GOLANG'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---


Golang built-in tool `pprof` can be used to profiling go process.

- `runtime/pprof`: By programming, output profiling data to a file.
- `net/http/pprof`: Download profiling file by http request.

Types of profiles available::

- `goroutine`: Stack traces of all current goroutines
- `heap`: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
- `allocs`: A sampling of all past memory allocations
- `threadcreate`: Stack traces that led to the creation of new OS threads
- `block`: Stack traces that led to blocking on synchronization primitives
- `mutex`: Stack traces of holders of contended mutexes

You can use official tool [`pprof`](https://github.com/google/pprof/blob/main/doc/README.md){:target="_blank"} to analysis generated profile file.

DataKit can use either [Pull mode](profile-go.md#pull-mode) or [Push mode](profile-go.md#push-mode) to generate profiling file.

## push mode {#push-mode}

### Config DataKit {#push-datakit-config}

Enable [profile](profile.md#config)  inputs

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]
```

### Integrate dd-trace-go {#push-app-config}

Import [dd-trace-go](https://github.com/DataDog/dd-trace-go){:target="_blank"}, Insert code as follows to your application:

```go
package main

import (
    "log"
    "time"

    "gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
    err := profiler.Start(
        profiler.WithService("dd-service"),
        profiler.WithEnv("dd-env"),
        profiler.WithVersion("dd-1.0.0"),
        profiler.WithTags("k:1", "k:2"),
        profiler.WithAgentAddr("localhost:9529"), // DataKit url
        profiler.WithProfileTypes(
            profiler.CPUProfile,
            profiler.HeapProfile,
            // The profiles below are disabled by default to keep overhead
            // low, but can be enabled as needed.

            // profiler.BlockProfile,
            // profiler.MutexProfile,
            // profiler.GoroutineProfile,
        ),
    )

    if err != nil {
        log.Fatal(err)
    }
    defer profiler.Stop()

    // your code here
    demo()
}

func demo() {
    for {
        time.Sleep(100 * time.Millisecond)
        go func() {
            buf := make([]byte, 100000)
            _ = len(buf)
            time.Sleep(1 * time.Hour)
        }()
    }
}
```

Once your go app start, dd-trace-go will send profiling data to DataKit by interval(per 1min by default).

### Generated Metrics {#metrics}

Starting from [:octicons-tag-24: Version-1.39.0](../datakit/changelog.md#cl-1.39.0), DataKit supports extracting a set of Go runtime-related metrics from `dd-trace-go` output. These metrics are placed under the `profiling_metrics` metric set. Below are some key metrics with explanations:

| Tags & Fields | Description  |
|----------:|:------------|
| `language`<br>(`tag`) | Language of current profile |
| `host`<br>(`tag`) | Hostname of current profile |
| `service`<br>(`tag`) | Service name of current profile |
| `env`<br>(`tag`) | Env settings of current profile |
| `version`<br>(`tag`) | Version of current profile |
| `prof_go_cpu_cores` | Number of CPU cores consumed<br> *Unit: core* |
| `prof_go_cpu_cores_gc_overhead` | Number of CPU cores used for garbage collection<br> *Unit: core* |
| `prof_go_alloc_bytes_per_sec` | Memory allocation rate per second<br> *Unit: byte* |
| `prof_go_frees_per_sec` | Number of objects freed by GC per second<br> *Unit: count* |
| `prof_go_heap_growth_bytes_per_sec` | Heap memory growth rate per second<br> *Unit: byte* |
| `prof_go_allocs_per_sec` | Memory allocation operations per second<br> *Unit: count* |
| `prof_go_alloc_bytes_total` | Total memory allocated during a single profiling period (dd-trace defaults to 60-second collection cycles)<br> *Unit: byte* |
| `prof_go_blocked_time` | Total time goroutines were blocked during a single profiling period<br> *Unit: nanosecond* |
| `prof_go_mutex_delay_time` | Total time spent waiting for locks during a single profiling period<br> *Unit: nanosecond* |
| `prof_go_gcs_per_sec` | Number of GC runs per second<br> *Unit: count* |
| `prof_go_max_gc_pause_time` | Maximum single pause duration caused by GC during a profiling period<br> *Unit: nanosecond* |
| `prof_go_gc_pause_time` | Total pause time caused by GC during a profiling period<br> *Unit: nanosecond* |
| `prof_go_num_goroutine` | Current total number of goroutines<br> *Unit: count* |
| `prof_go_lifetime_heap_bytes` | Total memory size occupied by live objects in the heap<br> *Unit: byte* |
| `prof_go_lifetime_heap_objects` | Total number of live objects in the heap<br> *Unit: count* |

## Pull Mode {#pull-mode}

### Enable profiling in app {#app-config}

import `pprof` package in your code:

```go
package main

import (
  "net/http"
   _ "net/http/pprof"
)

func main() {
    http.ListenAndServe(":6060", nil)
}
```

Once start your app, you can view page `http://localhost:6060/debug/pprof/heap?debug=1` in browser to confirm running as your wish.

- Mutex and Block events

Mutex and Block events are disable by default, if you want to enable them, add below code to your app:

```go
var rate = 1

// enable mutex profiling
runtime.SetMutexProfileFraction(rate)

// enable block profiling
runtime.SetBlockProfileRate(rate)
```

Set the collection frequency, where 1/rate events are collected. Values set to 0 or less are not collected.

### Config DataKit {#datakit-config}

[Enable Profile Input](profile.md), modify `[[inputs.profile.go]]` segment as follows.

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election
  election = true

 ## go pprof config
[[inputs.profile.go]]
  ## pprof url
  url = "http://localhost:6060"

  ## pull interval, should be greater or equal than 10s
  interval = "10s"

  ## service name
  service = "go-demo"

  ## app env
  env = "dev"

  ## app version
  version = "0.0.0"

  ## types to pull 
  ## values: cpu, goroutine, heap, mutex, block
  enabled_types = ["cpu","goroutine","heap","mutex","block"]

[inputs.profile.go.tags]
  # tag1 = "val1"
```

<!-- markdownlint-disable MD046 -->
???+ note

    If there is no need to enable profile http endpoint, just comment `endpoints` item.
<!-- markdownlint-enable MD046 -->

### Field introduction {#fields-info}

- `url`: net/http/pprof listening address, such as `http://localhost:6060`
- `interval`: upload interval, minimum 10s
- `profile_duration`: CPU profile duration, 10s by default
- `http_timeout`: HTTP timeout for one profile request; it must exceed `profile_duration`
- `service`:  your service name
- `env`:  your app running env
- `version`: your app version
- `enabled_types`: available events: `cpu, goroutine, heap, mutex, block`

You should Restart DataKit after modification. After a minute or two, you can visualize your profiles on the [profile](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"}.

## Kubernetes Pod Auto-discovery {#kubernetes-discovery}

In Kubernetes, DataKit can discover Go pprof endpoints by namespace, Pod labels, and container. Each DaemonSet DataKit watches only Pods on its own node and accesses the Pod IP directly. No application Sidecar, shared PID namespace, StatefulSet, or load-balanced Service is required.

```toml
[[inputs.profile]]
  endpoints = ["/profiling/v1/input"]
  kubernetes_max_concurrency = 2
  kubernetes_delta_cache_mb = 32

[[inputs.profile.kubernetes]]
  node_local = true
  namespaces = ["guancedb"]
  selector = "app.kubernetes.io/name=guancedb,app.kubernetes.io/component=select"
  container = "select"
  port = "pprof"
  path = "/debug/pprof"

  interval = "10m"
  scheduled_types = ["heap", "goroutine"]
  trigger_types = ["cpu", "heap", "goroutine"]
  profile_duration = "30s"
  emergency_duration = "10s"
  request_timeout = "45s"

  monitor_interval = "10s"
  trigger_window = "1m"
  cooldown = "10m"
  max_concurrency = 2
  cpu_usage_base_limit = 80
  mem_usage_base_limit = 80
  cpu_emergency_base_limit = 95
  mem_emergency_base_limit = 95

[inputs.profile.kubernetes.tags]
  team = "database"

[inputs.profile.kubernetes.pod_label_as_tags]
  "app.kubernetes.io/component" = "component"
```

Percentage thresholds use the container limit as the denominator. CPU usage comes from kubelet `usageNanoCores`, and memory usage comes from `workingSetBytes`. When a container has no corresponding limit, percentage thresholds are ignored; use `cpu_usage_millicores` or `mem_usage_bytes` instead. Normal thresholds use a sliding window: a trigger fires once the number of violating samples within the most recent `trigger_window` reaches `trigger_window / monitor_interval`. Transient dips only age out of the window and do not reset it, and a brief kubelet stats outage does not discard the recorded history. Emergency thresholds trigger immediately. A target cannot trigger again during `cooldown`.

Each target is sampled at its own `monitor_interval` deadlines; intervals across rules need not divide evenly. Each monitoring round queries the local kubelet once, with a timeout equal to the smaller of the shortest monitoring interval and 10s, and cancellation on collector shutdown. Slow requests and missing data do not create synthetic samples; normal triggers are delayed when too few violating samples remain in the window.

Configure `scheduled_types` and `trigger_types` independently for each application role. For example, create a separate rule for an Insert role and omit `cpu` from both fields to disable CPU profiling for that role. If rules overlap on the same Pod endpoint, the first configured rule wins. Collections for the same endpoint are mutually exclusive. For multi-container Pods with a numeric `port`, specify `container`; when omitted, a target is accepted only if the declared TCP ports identify exactly one container. Missing or ambiguous ownership is skipped. Single-container Pods do not need to declare the port.

DataKit adds `cluster_name_k8s`, `namespace`, `pod_name`, `pod_uid`, `node_name`, `container_name`, `workload_kind`, `workload_name`, and trigger context tags. The precedence for `service`, `env`, and `version` is: explicit fields, annotation mappings, label mappings, rule `tags`, standard application labels, and derived defaults. Use `pod_label_as_tags` and `pod_annotation_as_tags` for mappings. The existing upload protocol separates tags with commas, so unrepresentable tags are dropped: names cannot contain commas, colons, or newlines; values cannot contain commas or newlines. Colons in values are allowed. Mappings cannot override Pod identity tags.

### Resource Limits and Target State {#kubernetes-limits}

- `kubernetes_max_concurrency` belongs to `inputs.profile` and defaults to 2. It limits in-flight collections across all Kubernetes rules in that input; each rule's `max_concurrency` adds a second limit. Static `go` pulls and other Profile input instances are outside this budget.
- `kubernetes_delta_cache_mb` belongs to `inputs.profile` and defaults to 32 MiB. It bounds the shared delta baseline cache, storing unparsed Protobuf bytes and accounting for a fixed overhead per entry. When full, it evicts the least recently used baselines. The next collection rebuilds an evicted baseline without emitting a delta; the following collection can produce a delta. Removing a target releases its baseline immediately.
- Discovery accepts Protobuf pprof only. HTTP responses are limited by `body_size_limit_mb`; decompressed data is limited to the smaller of that value and 8 MiB. A 100,000 field/element parsing budget is enforced before parsing, and expanded samples are also bounded by the decompressed size limit. Rejected data is not uploaded. The cache budget is not a process RSS limit: allow additional memory for concurrent parsing, upload queues, and informer caches, and measure real profile sizes and CPU overhead before increasing concurrency.
- Temporary NotReady transitions with the same Pod UID, container, and endpoint suspend collection while retaining mutual exclusion, cooldown, and backoff. Deletion or changes to UID, IP, or the matching rule cancel the old task and release its baseline. A new task for the same endpoint waits for the old task to finish.
- The `reason` label of `datakit_profile_kubernetes_skipped_total` distinguishes `delta_cache_evicted`, `delta_cache_limit`, `invalid_tag`, and `ambiguous_port`, among other skip reasons.

### Deployment and Security Requirements {#kubernetes-security}

- The pprof listener must be reachable through the Pod IP, for example by listening on `0.0.0.0`. A listener bound only to `127.0.0.1` is not reachable by DataKit.
- Discovery connects directly to Pod IPs, without environment HTTP proxies or HTTP redirects, including same-origin redirects. Set `path` to the actual pprof path.
- Do not expose pprof outside the cluster through Ingress, NodePort, or LoadBalancer. Use a NetworkPolicy that only allows DataKit Pods to reach the pprof port.
- The DataKit ServiceAccount needs `get/list/watch` access to Pods and access to the local kubelet `/stats/summary`. The standard DataKit ClusterRole already grants Pod and `nodes/stats` access.
- `node_local` currently only supports `true`. Kubernetes discovery ignores the Profile input's global election state so non-leader nodes are not skipped.
- Profile uses a separate local-node Pod informer with a server-side `spec.nodeName` filter; namespace/label filtering is local. LIST's limit of 50 is a page size, not a target limit. There are no per-Pod GETs or periodic full LISTs. The 50 QPS/burst 50 budget belongs to each Kubernetes Client, not the whole DataKit process. Large-scale startup, reconnection, and frequent Pod updates still need capacity testing. Coordination between multiple instances on one node is deferred; deploy only one DataKit responsible for a given target per node.
- For TLS endpoints, configure `tls_open`, `tls_ca`, `tls_cert`, `tls_key`, and `insecure_skip_verify`. Skipping certificate verification is not recommended in production.
