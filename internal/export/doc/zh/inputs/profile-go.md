---
title     : 'Profiling Golang'
summary   : 'Golang Profiling 集成'
tags:
  - 'GOLANG'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---


Go 内置了性能分析 (Profiling) 工具 `pprof`，可以采集程序运行中的性能数据，可通过以下两种方式使用：

- `runtime/pprof`: 通过编程方式，自定义采集运行数据，然后保存分析
- `net/http/pprof`: 调用 `runtime/pprof`，封装成接口，通过 HTTP Server 的方式对外提供性能数据

性能数据主要包括如下：

- `goroutine`: 运行的 Goroutine 的调用栈分析
- `heap`: 活跃对象的内存分配情况
- `allocs`: 所有对象的内存分配情况
- `threadcreate`: OS 线程创建分析
- `block`: 阻塞分析
- `mutex`: 互斥锁分析

收集到的数据，可以通过官方 [`pprof`](https://github.com/google/pprof/blob/main/doc/README.md){:target="_blank"} 工具进行分析。

DataKit 可通过[主动拉取](profile-go.md#pull-mode) (pull) 或[被动推送](profile-go.md#push-mode) (push) 的方式来获取这些数据。

## push 方式 {#push-mode}

### DataKit 配置 {#push-datakit-config}

DataKit 开启 [profile](profile.md#config)  采集器，注册 profile http 服务。

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]
```

### Go 应用配置 {#push-app-config}

集成 [dd-trace-go](https://github.com/DataDog/dd-trace-go){:target="_blank"}，采集应用性能数据并发送至 DataKit。 代码参考如下：

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

运行该程序后，DDTrace 会定期（默认 1 分钟一次）将数据推送给 DataKit。

### 生成性能指标 {#metrics}

DataKit 自 [:octicons-tag-24: Version-1.39.0](../datakit/changelog.md#cl-1.39.0) 开始支持从 `dd-trace-go` 的输出中抽取一组 Go 运行时的相关指标，该组指标被置于 `profiling_metrics` 指标集下，下面列举其中部分指标加以说明：

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

<!-- markdownlint-disable MD046 -->
???+ tips

    该功能默认开启，如果不需要可以通过修改采集器的配置文件 `<DATAKIT_INSTALL_DIR>/conf.d/profile/profile.conf` 把其中的配置项 `generate_metrics` 置为 false 并重启 DataKit.

    ```toml
    [[inputs.profile]]
    
    ...
    
    ## set false to stop generating apm metrics from ddtrace output.
    generate_metrics = false
    ```
<!-- markdownlint-enable MD046 -->

## Pull 方式 {#pull-mode}

### Go 应用开启 Profiling {#app-config}

应用中开启 Profiling 只需要引用 `pprof` 包即可，参考如下：

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

运行代码后，可通过 `http://localhost:6060/debug/pprof/heap?debug=1` 来查看是否开启成功。

- Mutex 和 Block 性能分析

默认情况下，mutex 和 block 性能采集并未开启，如果需要开启，可添加如下代码：

```go
var rate = 1

// enable mutex profiling
runtime.SetMutexProfileFraction(rate)

// enable block profiling
runtime.SetBlockProfileRate(rate)
```

`rate` 设置采集频率，即 1/rate 的事件被采集， 如设置为 0 或小于 0 的数值，是不进行采集的。

### DataKit 配置 {#datakit-config}

[开启 Profile 采集器](profile.md)，进行如下设置 `[[inputs.profile.go]]`。

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

    如果不需要开启 Profile 的 HTTP 服务，可将 `endpoints` 字段注释掉。
<!-- markdownlint-enable MD046 -->

### 字段说明 {#fields-info}

- `url`: 上报地址，如 `http://localhost:6060`
- `interval`: 采集间隔时间，最小 10s
- `profile_duration`: CPU Profile 的持续时间，默认 10s
- `http_timeout`: 拉取单个 Profile 的 HTTP 超时，必须大于 `profile_duration`
- `service`： 服务名称
- `env`： 应用环境类型
- `version`: 应用的版本
- `enabled_types`: 性能类型，如 `cpu, goroutine, heap, mutex, block`

配置好 Profile 采集器，启动或重启 DataKit，一段时间后即可在<<<custom_key.brand_name>>>中心查看 Go 的性能数据。

## Kubernetes Pod 自动发现 {#kubernetes-discovery}

Kubernetes 部署下，DataKit 可以按 namespace、Pod label 和容器自动发现 Go pprof endpoint。每个 DaemonSet DataKit 仅监听本节点 Pod，直接访问 Pod IP；无需为业务增加 Sidecar、共享 PID namespace、StatefulSet 或负载均衡 Service。

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

百分比阈值使用容器 limit 作为分母：CPU 使用量来自 kubelet `usageNanoCores`，内存使用量来自 `workingSetBytes`。容器没有相应 limit 时，百分比阈值不会生效，可改用 `cpu_usage_millicores`、`mem_usage_bytes` 绝对值。普通阈值采用滑动窗口判定：最近 `trigger_window` 时间内的超阈值采样点数达到 `trigger_window / monitor_interval` 时触发，瞬时回落不会重置窗口，kubelet 数据短暂不可用也不会清空已有记录；紧急阈值立即触发；同一目标在 `cooldown` 内不会重复触发。

各目标按自身 `monitor_interval` 的到期时间采样，多条规则的周期无需互相整除。一次监控轮次批量请求本节点 kubelet，超时取最短监控周期与 10s 中的较小值，并随采集器退出取消。慢请求和缺失数据不会补造采样点；窗口内有效超阈值点不足时，普通触发会推迟。

`scheduled_types` 与 `trigger_types` 可按组件分别配置。例如为 Insert 组件创建单独规则，并从这两个字段中移除 `cpu`，即可完全禁止该组件执行 CPU Profile。重叠规则命中同一个 Pod endpoint 时，优先使用配置中靠前的规则；同一 endpoint 的采集互斥。多容器 Pod 使用数字 `port` 时，应显式指定 `container`；省略时仅接受能根据容器声明的 TCP 端口唯一确定归属的目标，无法确定或存在歧义则跳过。单容器 Pod 无需声明该端口。

DataKit 会自动附加 `cluster_name_k8s`、`namespace`、`pod_name`、`pod_uid`、`node_name`、`container_name`、`workload_kind`、`workload_name` 及触发上下文标签。`service`、`env`、`version` 的优先级依次为显式字段、annotation 映射、label 映射、规则 `tags`、标准应用标签及默认推导。映射使用 `pod_label_as_tags`、`pod_annotation_as_tags`。现有上传协议以逗号分隔标签，无法安全表示的标签会被丢弃：标签名不得包含逗号、冒号或换行，标签值不得包含逗号或换行；标签值中的冒号允许保留。Pod 身份标签不能被映射覆盖。

### 资源限制与目标状态 {#kubernetes-limits}

- `kubernetes_max_concurrency` 位于 `inputs.profile`，默认 2，限制该采集器所有 Kubernetes 规则合计的在途采集数；每条规则的 `max_concurrency` 作为第二层限制。该额度不包含静态 `go` 拉取或其他 Profile 采集器实例。
- `kubernetes_delta_cache_mb` 位于 `inputs.profile`，默认 32 MiB，限制共享差量基线缓存。缓存保留未解析的 Protobuf 字节，并计入每项固定开销；空间不足时淘汰最久未使用的基线。目标下一次采集只重建已淘汰的基线，再下一次才产生差量；目标删除后立即释放对应缓存。
- 自动发现只接受 Protobuf pprof。HTTP 响应受 `body_size_limit_mb` 限制，解压后还受该限制与 8 MiB 中较小者约束；解析前另有 100,000 个字段/元素的复杂度预算，样本展开大小也受解压大小限制。超限数据不会上传。缓存预算不是进程 RSS 上限，仍需为在途解析、上传队列及 informer 缓存预留内存；增加并发前应验证实际业务 profile 的大小和 CPU 开销。
- 同一 Pod UID、容器和 endpoint 的暂时 NotReady 会暂停采集并保留互斥、冷却和退避状态；删除、UID/IP 或匹配规则变化会取消旧任务并释放基线。同一 endpoint 的新任务需等旧任务退出。
- 内部指标 `datakit_profile_kubernetes_skipped_total` 的 `reason` 标签区分 `delta_cache_evicted`、`delta_cache_limit`、`invalid_tag`、`ambiguous_port` 等跳过原因。

### 部署与安全要求 {#kubernetes-security}

- pprof 必须监听 Pod IP 可达的地址（例如 `0.0.0.0`），仅监听 `127.0.0.1` 无法被 DataKit 访问。
- 自动发现直连 Pod IP，不使用环境 HTTP 代理，也不跟随 HTTP 重定向（包括同源重定向）。请将 `path` 配置为实际 pprof 路径。
- 不要通过 Ingress、NodePort 或 LoadBalancer 将 pprof 暴露到集群外；建议使用 NetworkPolicy 仅允许 DataKit Pod 访问目标端口。
- DataKit ServiceAccount 需要对 Pod 的 `get/list/watch` 权限，并需要访问本节点 kubelet `/stats/summary`（标准 DataKit ClusterRole 已包含 Pod 与 `nodes/stats` 权限）。
- `node_local` 当前仅支持 `true`。Kubernetes 自动发现不受 Profile 的全局 election 状态影响，避免非 leader 节点漏采。
- Profile 使用独立的本节点 Pod informer，服务端按 `spec.nodeName` 过滤；namespace/label 在本地筛选，LIST 的 50 是分页大小，不是目标数量上限。没有逐 Pod GET 或周期全量 LIST。每个 Kubernetes Client 的 50 QPS/burst 50 并非 DataKit 进程总额度；大量节点启动、重连及 Pod 高频更新仍需容量验证。同节点多实例协调尚未实现，每个节点应只部署一个负责该目标的 DataKit。
- 使用 TLS 时可配置 `tls_open`、`tls_ca`、`tls_cert`、`tls_key` 和 `insecure_skip_verify`；生产环境不建议跳过证书验证。
