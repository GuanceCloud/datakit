---
title     : 'Flameshot'
summary   : '自动收集 profiling 工具'
tags      :
  - 'java'
  - 'async-profiler'
  - 'profiling'
  - 'flameshot'
---

Flameshot 是一个基于 Sidecar 模式运行的轻量级自动性能剖析（Profiling）工具。它通过监控目标进程的资源使用情况（CPU/内存），在达到预设阈值时自动触发底层 Profiler（如 `async-profiler`），从而实现无侵入的现场快照采集。

---

## 核心功能与原理 {#core-concepts}

### 运行模式 {#running-mode}

Flameshot 采用 **Sidecar 容器** 模式部署。它必须与业务主容器（Main Container）运行在同一个 Pod 中，并开启 **PID 命名空间共享**。

1. **监控 (Monitor)**：Flameshot 持续轮询主容器内目标进程的资源水位。
1. **触发 (Trigger)**：当满足阈值（如 CPU > 80%）或收到 HTTP API 请求时，触发采集任务。
1. **执行 (Execute)**：根据配置的语言类型（目前支持 Java 和 Go），调用对应的 Profiler 工具采集目标进程。
1. **收集 (Collect)**：生成的 Profile 文件（如 `.jfr` 或 `.pprof`）随后上传至数据观测中心。
1. **定时**: 配置 `FLAMESHOT_AUTO_PROFILING` 后，会定时对所有匹配到的进程采集一次 Profiling 数据，采集时长默认 30s，可通过 `FLAMESHOT_AUTO_PROFILING_DURATION` 调整。
1. **OOM 摘要**: 当检测到容器 `oom_kill` 增量时，Flameshot 会尝试从目标 Java 进程启动参数中自动解析 `-XX:+HeapDumpOnOutOfMemoryError` 与 `-XX:HeapDumpPath=...`。如果 dump 文件位于共享卷内且生成成功，Flameshot 会找到对应 `.hprof`，并上传一条摘要日志。

### 适用场景 {#use-cases}

- **生产环境兜底**：在服务因 CPU 飙高或内存泄漏即将崩溃前，自动保留现场证据。
- **性能压测分析**：配合压测平台，自动采集高负载下的性能热点。

---

## 配置详解 {#configuration}

Flameshot 的所有行为均通过环境变量进行控制。配置分为 **全局设置** 和 **采集策略** 两部分。

### 全局环境变量 {#global-env}

这些变量控制 Sidecar 容器的基础行为。

| 变量名称                         | 必填    | 默认值     | 说明                                                                        |
|:-----------------------------|:------|:--------|:--------------------------------------------------------------------------|
| `FLAMESHOT_DATAKIT_ADDR`     | **是** | -       | DataKit 的 Profiling 数据接收接口地址。                                             |
| `FLAMESHOT_PROFILING_PATH`   | **是** | `/data` | **共享目录路径**。用于存放工具库和生成的临时文件，需与主容器挂载一致。                                     |
| `FLAMESHOT_MONITOR_INTERVAL` | 否     | `1`     | 监控轮询间隔（秒）。                                                                |
| `FLAMESHOT_LOG_LEVEL`        | 否     | `info`  | 日志级别，可选：`debug`, `info`, `warn`, `error`。                                 |
| `FLAMESHOT_HTTP_LOCAL_IP`    | **是** | `-`     | Sidecar 自身 HTTP 服务监听地址。                                                   |
| `FLAMESHOT_HTTP_LOCAL_PORT`  | **是** | `8089`  | Sidecar 自身 HTTP 服务监听端口。                                                   |
| `FLAMESHOT_PROFILING_ENABLED` | 否 | `true` | 是否开启 JFR Profiling。设置为 `false` 后会关闭定时、阈值、cgroup 高水位和 HTTP 手动 Profiling，但保留 OOM 检测、hprof 上传和主动 Heap Dump。 |
| `FLAMESHOT_AUTO_PROFILING`   | 否     | -       | 定时对所有匹配到的进程采集一次 Profiling 数据。最小不得低于一分钟，如五分钟："5m" 或者一小时 "1h"         |
| `FLAMESHOT_AUTO_PROFILING_DURATION` | 否 | `30s` | 定时采集模式下的单次采样时长。 |
| `FLAMESHOT_OOM_HPROF_ENABLED` | 否    | `false` | 开启 OOM 后 `.hprof` 摘要恢复链路。仅对 Java 进程生效，且要求目标 JVM 显式开启 `-XX:+HeapDumpOnOutOfMemoryError` 并配置位于共享卷内的 `-XX:HeapDumpPath=...`。建议在发布配置中显式声明。 |
| `FLAMESHOT_OOM_HPROF_MATCH_WINDOW` | 否 | `2m` | OOM 事件与 `.hprof` 文件修改时间的匹配窗口。 |
| `FLAMESHOT_HPROF_UPLOAD_ENABLED` | 否 | `false` | 是否将匹配到或主动生成的 `.hprof` 上传到对象存储。 |
| `FLAMESHOT_HPROF_UPLOAD_PROVIDER` | 否 | - | 对象存储类型，支持 `oss` 和 `s3`。 |
| `FLAMESHOT_HPROF_UPLOAD_ENDPOINT` | 否 | - | OSS/S3 endpoint。 |
| `FLAMESHOT_HPROF_UPLOAD_REGION` | 否 | S3 默认为 `us-east-1` | S3 region。 |
| `FLAMESHOT_HPROF_UPLOAD_BUCKET` | 否 | - | 目标 bucket。 |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID` | 否 | - | 对象存储 AK。 |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET` | 否 | - | 对象存储 SK。 |
| `FLAMESHOT_HPROF_UPLOAD_PATH_TEMPLATE` | 否 | `{service}/{pod_name}/{timestamp}/{filename}` | 对象路径模板，支持 service / pod_name / pod_namespace / host / pid / timestamp / filename 等变量。 |
| `FLAMESHOT_HPROF_DOWNLOAD_URL_TEMPLATE` | 否 | - | 可选下载链接模板；配置后事件中按该模板生成 `hprof_download_url`。 |
| `FLAMESHOT_HEAP_DUMP_ENABLED` | 否 | `false` | 内存紧急阈值命中时是否主动执行 Java Heap Dump。 |
| `FLAMESHOT_HEAP_DUMP_PATH_TEMPLATE` | 否 | `{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof` | 本地 Heap Dump 输出路径模板。 |
| `FLAMESHOT_HEAP_DUMP_JMAP_PATH` | 否 | `jmap` | `jmap` 可执行文件路径。官方 Sidecar 镜像默认不内置 JVM/JDK，开启主动 Heap Dump 时需显式提供可用 `jmap`。 |
| `FLAMESHOT_HEAP_DUMP_TIMEOUT` | 否 | `120s` | Heap Dump 命令超时时间。 |
| `FLAMESHOT_HEAP_DUMP_COOLDOWN` | 否 | `10m` | 单进程主动 Heap Dump 冷却时间。 |
| `FLAMESHOT_POD_MEM_LIMIT` | 否 | - | Pod 内存 limit，单位 Mi。配置后会优先按 Pod limit 计算内存使用率。 |
| `FLAMESHOT_POD_CPU_LIMIT` | 否 | - | Pod CPU limit，单位 m。配置后会按 Pod CPU limit 计算 CPU 使用率。 |
| `FLAMESHOT_SERVICE`          | 否     | -       | 可以不用在 `FLAMESHOT_PROCESSES` 中配置 `service`, 会全部替换。                         |
| `FLAMESHOT_TAGS`             | 否     | -       | 建议配置 `host` `pod_name` `pod_namespace` 如： "host:host_name,pod_name:pod_a" |

如果 DataKit 以 DaemonSet 方式部署，并通过 `hostNetwork`/`hostPort` 暴露 `9529`，推荐让 Flameshot 直连**当前业务 Pod 所在节点**的 DataKit，而不是通过普通 Service 域名随机转发：

```yaml
- name: NODE_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: FLAMESHOT_DATAKIT_ADDR
  value: "http://$(NODE_IP):9529/profiling/v1/input"
```

这样每个业务 Pod 都会将 Profile 上传到本节点 DataKit，便于排查和保持节点本地采集语义。

### 采集策略配置 (`FLAMESHOT_PROCESSES`) {#profiling-policy}

通过环境变量 `FLAMESHOT_PROCESSES` 定义监控目标。该变量的值必须是一个标准的 **JSON 数组** 字符串。

为了在 Kubernetes YAML 中保持配置的可读性，**强烈建议**使用 YAML 的多行文本语法（`|`）来书写 JSON 配置，如下所示：

```yaml
    env:
      # ... 其他环境变量 ...
      - name: FLAMESHOT_PROCESSES
        value: |
          [
            {
              "service": "user-service",
              "language": "java",
              "command": "^java.*user-service\\.jar$",
              "duration": "60s",
              "events": "cpu,alloc",
              "cpu_usage_percent": 80,
              "mem_usage_percent": 80,
              "mem_usage_mb": 1024,
              "mem_usage_percent_emergency": 92,
              "mem_usage_mb_emergency": 1536,
              "heap_dump_on_memory_emergency": true,
              "emergency_duration": "10s",
              "tags": [
                "env:prod",
                "version:v1.2"
              ]
            }
          ]
```

**通用字段说明：**

- **`service`** (String): 上报到观测中心的服务名称。
- **`language`** (String): 目标进程语言。目前支持 `java`、`go`、`golang`。
- **`command`** (String): 匹配进程命令行的正则表达式。
- **`duration`** (String): 单次采集时长（例如 `30s`, `1m`）。**注意**：受限于执行超时，建议不超过 5 分钟。
- **`emergency_duration`** (String): 内存紧急阈值命中后的快速采集时长，建议配置为 `10s` 或 `15s`。
- **`pprof_url`** (String): Go pprof HTTP 地址，例如 `http://127.0.0.1:6060`。当 `language` 为 `go` 或 `golang` 时需要配置。
- **`pprof_types`** (List): Go pprof 类型，支持 `cpu`、`goroutine`、`heap`、`mutex`、`block`，与 profile 采集器现有 Go pull 模式保持一致。
- **`pprof_timeout`** (String): Go pprof 请求超时时间，建议大于 CPU profile 的 `duration`。
- **`tags`** (List): 自定义标签列表，建议包含 `env`, `version` 等元信息。
- **`cpu_usage_percent`** (Int): CPU 触发阈值 (0-N)。多核环境下数值可能超过 100。
- **`mem_usage_percent`** (Int): 内存使用率平均阈值 (0-100)，按最近 5 个点平均值触发。
- **`mem_usage_mb`** (Int): 内存使用量平均阈值 (MB)，按最近 5 个点平均值触发。
- **`mem_usage_percent_emergency`** (Int): 内存使用率紧急瞬时阈值 (0-100)，单点命中立即触发。
- **`mem_usage_mb_emergency`** (Int): 内存使用量紧急瞬时阈值 (MB)，单点命中立即触发。
- **`heap_dump_on_memory_emergency`** (Bool): 内存紧急阈值命中时是否允许该进程规则主动 Heap Dump。未配置时，在 `FLAMESHOT_HEAP_DUMP_ENABLED=true` 的情况下默认允许。
- `cpu_usage_percent`、`mem_usage_percent`、`mem_usage_mb` 不配置或者配置 0 都会略过该项的阈值检查。
- 配置了 `FLAMESHOT_POD_MEM_LIMIT` 后，`mem_usage_percent` 与 `mem_usage_percent_emergency` 会优先按 Pod limit 视角计算，而不是宿主机视角。
- 配置 `FLAMESHOT_HEAP_DUMP_ENABLED=true` 后，内存紧急阈值命中会投递 `jmap` Heap Dump 任务；该能力要求 Sidecar 内可执行 `FLAMESHOT_HEAP_DUMP_JMAP_PATH` 指向的 `jmap`。如果同时配置 hprof 对象存储上传，生成的 `.hprof` 会继续上传，并在事件中带上 `hprof_object_key`、`hprof_download_url` 和上传状态。


---

## 语言特定指南 {#language-specifics}

根据被监控应用的技术栈，Flameshot 会调用不同的底层工具。

<!-- markdownlint-disable MD046 -->
=== "Java"

    ### Java Profiling {#java-profiling}
    
    针对 Java 应用，Flameshot 内置了 **async-profiler** (支持 `linux-amd64` / `linux-arm64`)。
    
    **关键配置字段 (`FLAMESHOT_PROCESSES`):**
    
    - **`language`**: 必须设置为 `java`。
    - **`events`**: 支持 `cpu` (CPU cycles), `alloc` (内存分配), `lock` (锁竞争), `cache-misses`, `nativemem`。默认为 `all`。
    - **`jdk_version`**: (可选) 用于元数据展示的 JDK 版本。
    
    **注意事项：**

    - 无需依赖 JVM Safepoint，开销极低。
    - 如果希望在 OOM 后自动发现并上传 `.hprof` 摘要日志，业务 JVM 必须显式开启 `-XX:+HeapDumpOnOutOfMemoryError`，并配置 `-XX:HeapDumpPath=...`。仅设置 `FLAMESHOT_OOM_HPROF_ENABLED=true` 并不会自动修改目标 JVM 的启动参数。
    - 如果开启 `FLAMESHOT_HEAP_DUMP_ENABLED=true`，Flameshot 会在内存紧急阈值命中时执行 `jmap -dump:format=b,file=<path> <pid>` 主动生成 `.hprof`。官方 Sidecar 镜像默认不内置 JVM/JDK，需要通过自定义镜像、挂载工具或其它方式显式提供与目标 JVM 兼容的 `jmap`，并用 `FLAMESHOT_HEAP_DUMP_JMAP_PATH` 指向它。
    - `HeapDumpPath` 必须指向业务容器和 Flameshot Sidecar **共同挂载**的共享目录；建议为每个进程配置稳定且可区分的 dump 路径。否则 Flameshot 即使检测到 OOM，也无法读取 dump 文件。
    - 建议在发布配置中显式声明 `.hprof` 摘要恢复相关开关，而不要依赖隐式默认值。

=== "Go"

    ### Go Profiling {#go-profiling}

    针对 Go 应用，Flameshot 通过业务进程暴露的 `net/http/pprof` HTTP 接口拉取 `.pprof` 数据并上传到 DataKit。

    **关键配置字段 (`FLAMESHOT_PROCESSES`):**

    - **`language`**: 必须设置为 `go` 或 `golang`。
    - **`pprof_url`**: 业务进程 pprof HTTP 地址，例如 `http://127.0.0.1:6060`。
    - **`pprof_types`**: 支持 `cpu`、`goroutine`、`heap`、`mutex`、`block`。
    - **`duration`**: `cpu` profile 的采集时长，会映射到 `/debug/pprof/profile?seconds=<duration>`。
    - **`pprof_timeout`**: pprof 请求超时时间，应大于 `duration`。

    **Go 应用侧要求：**

    ```go
    import (
        "net/http"
        _ "net/http/pprof"
    )

    func main() {
        go http.ListenAndServe("127.0.0.1:6060", nil)
    }
    ```

    `pprof_url` 必须填写 Flameshot Sidecar **实际能够访问到**的地址。常见有两种情况：

    - pprof 监听 `127.0.0.1:6060` 或 `0.0.0.0:6060`：可以配置 `http://127.0.0.1:6060`。
    - pprof 只监听 Pod IP，例如 `10.x.x.x:6060`：需要通过 Downward API 注入 Pod IP，并配置 `http://$(POD_IP):6060`。

    ```yaml
    - name: POD_IP
      valueFrom:
        fieldRef:
          fieldPath: status.podIP
    - name: FLAMESHOT_PROCESSES
      value: |
        [
          {
            "service": "go-app",
            "language": "go",
            "command": "^/app/go-app$",
            "pprof_url": "http://$(POD_IP):6060",
            "pprof_types": ["cpu", "goroutine", "heap", "mutex", "block"],
            "duration": "30s",
            "pprof_timeout": "45s"
          }
        ]
    ```

    配置示例：

    ```json
    {
      "service": "go-app",
      "language": "go",
      "command": "^/app/go-app",
      "pprof_url": "http://127.0.0.1:6060",
      "pprof_types": ["cpu", "goroutine", "heap", "mutex", "block"],
      "duration": "30s",
      "pprof_timeout": "45s",
      "tags": ["env:prod", "version:v1"]
    }
    ```

    **注意事项：**

    - `heap`、`mutex`、`block` 以 delta profile 形式上传，首次采样会作为基线保存，不上传这些 delta 类型。
    - `mutex` 和 `block` 默认不会采集有效数据，业务代码需要显式开启 `runtime.SetMutexProfileFraction` 和 `runtime.SetBlockProfileRate`。
    - pprof 接口可能暴露敏感运行时信息，建议只监听 Pod 内本地地址，不要通过 Service 或公网暴露。

=== "Python (Coming Soon)"

    ### Python Profiling {#python-profiling}
    
    *计划中*：将集成 `py-spy` 等无侵入式工具。
<!-- markdownlint-enable MD046 -->

---

## 部署指南 {#deployment}

### Kubernetes Sidecar 部署 {#k8s-sidecar}

为了使 Flameshot 正常工作，Pod 配置必须满足以下三个条件：

1. **共享进程空间** (`shareProcessNamespace: true`)。
1. **共享存储卷** (EmptyDir)。
1. **系统权限** (Capabilities)。

**YAML 示例：**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: java-app-profiled
spec:
  # 1. [核心] 开启 PID 共享，让 Sidecar 能看到 Java 进程
  shareProcessNamespace: true
  
  volumes:
  - name: shared-data
    emptyDir: {}

  containers:
  # 业务容器
  - name: my-app
    image: my-app:latest
    volumeMounts:
    - name: shared-data
      mountPath: /data # 需与 Sidecar 配置一致

  # Flameshot Sidecar
  - name: flameshot
    image: pubrepo.jiagouyun.com/datakit/flameshot:latest
    env:
      - name: FLAMESHOT_PROFILING_PATH
        value: "/data"
      # ... 其他环境变量 ...
    
    # 2. [核心] 赋予 ptrace 权限
    securityContext:
      capabilities:
        add: ["SYS_PTRACE"]
    
    # 3. [核心] 挂载同一目录
    volumeMounts:
    - name: shared-data
      mountPath: /data
```

### DataKit DaemonSet 配置注意事项 {#datakit-daemonset}

当 DataKit 以 DaemonSet 方式部署时，Flameshot 推荐通过当前节点 IP 上传 Profile：

```yaml
- name: NODE_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: FLAMESHOT_DATAKIT_ADDR
  value: "http://$(NODE_IP):9529/profiling/v1/input"
```

同时需要确认 DataKit 已满足以下条件：

1. **已开启 profile 采集器**，并注册 Profiling 上传接口：

    ```toml
    [[inputs.profile]]
      endpoints = ["/profiling/v1/input"]
    ```

1. **允许非 localhost 访问 Profiling API**。如果 DataKit 开启了 HTTP API 白名单，需要将 `/profiling/v1/input` 加入 `ENV_HTTP_PUBLIC_APIS`：

    ```yaml
    - name: ENV_HTTP_PUBLIC_APIS
      value: /otel/v1/trace,/otel/v1/metric,/otel/v1/logs,/profiling/v1/input
    ```

    如果该变量已经配置了其它接口，不要直接覆盖原值，应在原列表后追加 `/profiling/v1/input`。否则会出现：

    ```text
    datakit.publicAccessDisabled: api /profiling/v1/input disabled from external IP, only loopback(localhost) allowed
    ```

1. **DataKit 监听地址能被节点 IP 访问**。DaemonSet 常见配置是 `hostNetwork: true`、`hostPort: 9529`、`ENV_HTTP_LISTEN=0.0.0.0:9529`。

### OOM HProf 摘要要求 {#oom-hprof}

如果希望在 Java 进程 OOM 后由 Flameshot 自动补抓 `.hprof` 摘要，请同时满足以下条件：

1. 在业务 JVM 启动参数中开启 `-XX:+HeapDumpOnOutOfMemoryError`。
1. 在业务 JVM 启动参数中配置 `-XX:HeapDumpPath=/data/...`，且该路径位于共享卷内。
1. 为 Flameshot 设置 `FLAMESHOT_OOM_HPROF_ENABLED=true`。
1. 建议同时显式设置 `FLAMESHOT_OOM_HPROF_MATCH_WINDOW`，使运维侧对匹配窗口有明确预期。

例如：

```bash
java \
  -XX:+HeapDumpOnOutOfMemoryError \
  -XX:HeapDumpPath=/data/dumps/app.hprof \
  -jar app.jar
```

说明：

- Flameshot 会直接从目标 Java 进程的启动参数中自动解析 `HeapDumpPath`，不再单独通过配置指定 `.hprof` 路径。
- `FLAMESHOT_OOM_HPROF_ENABLED` 只是开启 Flameshot 侧的恢复逻辑，并不会替目标 JVM 注入 HeapDump 相关参数。
- 如果目标进程没有开启 `HeapDumpOnOutOfMemoryError`，或者 `HeapDumpPath` 不在共享卷内，Flameshot 只能记录 OOM 事件，无法找到对应 `.hprof` 文件。
- 如果容器在 dump 完成前已被直接终止，`.hprof` 可能仍然无法生成。

### Docker 本地测试 {#docker-testing}

如果您需要在本地 Docker 环境中进行测试，可以使用以下命令启动 Flameshot 并监控目标容器。

**前提条件：**

- 使用 `--pid="container:<target_id>"` 或共享卷方式（视具体 Docker 版本而定）。

**测试镜像：** `pubrepo.jiagouyun.com/datakit/flameshot:1.85.1-testing_testing-iss-2876`

**启动命令示例：**

```bash
docker run -d \
  --name flameshot-debug \
  --volumes-from <YOUR_JAVA_APP_CONTAINER> \
  -e FLAMESHOT_DATAKIT_ADDR="http://datakit:9529/profiling/v1/input" \
  -e FLAMESHOT_PROCESSES='[{"service":"local-test","command":"java","language":"java","cpu_usage_percent":10}]' \
  pubrepo.jiagouyun.com/datakit/flameshot:1.85.1-testing_testing-iss-2876
```

---

## API 接口参考 {#api-reference}

Flameshot 提供了 HTTP 接口，允许用户或自动化运维脚本**主动触发**采集任务。

### 手动触发采集 {#manual-trigger}

**接口地址**：`GET /v1/profile`

> **语义说明**：该接口用于按需生成一份 Profile 数据，而非获取监控指标。

**请求参数：**

| 参数名 | 必填 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `pid` | **二选一** | 目标进程 ID。优先级高于 `command`。 | `1234` |
| `command` | **二选一** | 目标进程名正则。用于匹配目标进程。 | `^java.*app.jar$` |
| `duration` | 否 | 采集时长。默认为 `30s`。 | `30s` |
| `events` | 否 | Java 采集事件类型。默认为 `all`。Go 采集优先使用 `pprof_types`。 | `cpu,alloc` |

**使用示例：**

1. **按 PID 触发采集**：

    ```bash
    # 对 PID 为 1234 的进程采集 30 秒的 CPU 和内存分配数据
    curl "http://localhost:8089/v1/profile?pid=1234&duration=30s&events=cpu,alloc"
    ```

1. **按进程名正则触发采集**：

    ```bash
    # 对名称匹配 tmall.jar 的进程采集默认时长的数据
    curl "http://localhost:8089/v1/profile?command=^java\\b.*tmall\\.jar$"
    ```

---

## JFR 数据格式 {#jfr-format}

以下是几种核心事件类型的详细说明：

| 事件类型 (Event)   | 对应参数             | 核心原理                                 | 适用场景                                          | 备注                         |
|:---------------|:-----------------|:-------------------------------------|:----------------------------------------------|:---------------------------|
| CPU Time       | cpu              | 通过内核采样或 itimer 定期查看 CPU 正在处理哪些代码指令。  | 性能优化：寻找计算密集型的“热点方法”，优化算法逻辑。                   | 只记录线程在 CPU 上运行的时间。         |
| Wall-clock     | wall             | 无论线程状态如何（运行、睡眠、阻塞），均按固定频率采样。         | 响应耗时诊断：排查 I/O 阻塞、数据库调用慢、网络延迟等。                | 能够反映出线程在“等什么”。             |
| Allocation     | alloc            | 记录 TLAB（线程本地分配缓存）的分配情况及大对象分配。        | 内存优化：定位内存抖动、减少频繁 GC 导致的停顿。                    | 记录的是分配动作，而不是当前内存存活量。       |
| Lock           | lock             | 记录线程在 synchronized 关键字上的竞争和等待耗时。     | 并发瓶颈：排查锁竞争激烈、线程死锁或同步块执行过慢。                    | 默认通常记录超过一定阈值的阻塞事件。         |
| Cache Misses   | cache-misses     | 利用硬件性能计数器 (PMU) 统计 L1/L2/L3 缓存未命中次数。 | 底层调优：优化数据结构（如 CPU 亲和性、伪共享问题）。                 | 需要 Linux 内核支持 perf_events。 |
| Context Switch | context-switches | 记录操作系统调度线程切换的频率。                     | 资源调度优化：排查线程数是否过多、系统负荷是否超载。                    | 频繁切换会导致 CPU 时间浪费在管理开销上。    |
| Java Methods   | itimer           | 基于内核计时器的采样。                          | 兼容性模式：在无法使用 perf_events 的环境（如部分容器）下替代 CPU 采样。 | 精度略低于硬件采样，但兼容性极好。          |

`alloc` 并非当前所有内存的总和，而是当前采样期间内所分配的内存大小。

---
## 常见问题与排查 {#troubleshooting}

1. **无法采集数据？**

    - 检查 Pod 是否开启了 `shareProcessNamespace: true`。
    - 检查 Sidecar 是否拥有 `SYS_PTRACE` 权限。
    - Go 应用检查 `pprof_url` 是否能在 Flameshot Sidecar 内访问。
    - 对 Go 应用，先在 Flameshot 容器内确认 pprof endpoint 可访问：

        ```bash
        curl -v "http://<pprof-host>:6060/debug/pprof/"
        curl -v "http://<pprof-host>:6060/debug/pprof/goroutine?debug=0"
        ```

    - 如果日志中出现 `connect: connection refused`，说明对应 IP:Port 没有监听。进入 Pod 检查：

        ```bash
        netstat -lntp | grep ':6060'
        ss -ltnp | grep ':6060'
        ```

      如果没有 `LISTEN`，说明业务进程没有打开 pprof，或监听端口不是 6060。仅看到 `60602 ... ESTABLISHED` 这类连接不代表 6060 已监听。
    - 如果 pprof 监听在 Pod IP 上，`pprof_url` 应配置为 `http://$(POD_IP):6060`；如果监听在 loopback 或所有地址上，可以使用 `http://127.0.0.1:6060`。

2. **文件未上传？**

    - 检查 `FLAMESHOT_PROFILING_PATH` 是否在两个容器间正确挂载。
    - 系统会自动管理文件生命周期，采集完成后会尝试删除临时文件。
    - Go 采集直接以内存附件形式上传 `.pprof`，通常不依赖本地落盘；如果 Go 采集日志已经显示 `upload to DataKit err`，应优先检查 DataKit 返回的 HTTP 状态码和响应体。
    - 如果 DataKit 返回 `403`，并包含 `datakit.publicAccessDisabled`，说明 `/profiling/v1/input` 没有对非 localhost 放行。请在 DataKit 配置中追加：

        ```yaml
        - name: ENV_HTTP_PUBLIC_APIS
          value: /otel/v1/trace,/otel/v1/metric,/otel/v1/logs,/profiling/v1/input
        ```

    - 如果 DataKit 返回 `input "profile" is not enabled for API "/profiling/v1/input"`，说明 DataKit 没有开启 profile 采集器。请启用：

        ```toml
        [[inputs.profile]]
          endpoints = ["/profiling/v1/input"]
        ```

3. **配置正则太麻烦**

    - JAVA 应用的进程名都是 `java`, 所以配置 `"command":"java"` `"language": "java"` 即可匹配所有的 JAVA 应用。
    - 想要配置特定的应用而不是所有，正则是必须要配置的。

---

## 更新日志 (Changelog) {#changelog}


### 0.2.2 (2026-5-12) {#cl-0.2.2}

#### 问题修复 {#cl-0.2.2-fix}

- **修复**
    - 修复 Profiling 上报元数据中 `tags_profiler` 可能重复写入 `host`、`env`、`version`、`service` 等标签的问题。
- **调整**
    - 移除高水位 `jcmd` 快照能力及相关配置项。
    - 增加 cgroup 内存压力触发时的观测字段，便于排查阈值、cgroup 当前值与上限值。

### 0.2.1 (2026-2-11) {#cl-0.2.1}

#### 新增功能 {#cl-0.2.1-new}

- **优化**
    - 在容器环境中，使用资源配置的大小作为阈值计算的基础值。


### 0.2.0 (2026-2-4) {#cl-0.2.0}

#### 新增功能 {#cl-0.2.0-new}

- **增加配置**
    - 支持通过环境变量 `FLAMESHOT_AUTO_PROFILING` 配置定时执行 Profiling
- **优化功能**
    - 优化配置阈值

### 0.1.0 (2025-12-17) {#cl-0.1.0}

Flameshot 的第一个正式版本，专注于为容器环境下的 Java 应用提供自动化的性能剖析能力。

#### 新增功能 {#cl-0.1.0-new}

- **核心架构**：
    - 支持 Kubernetes **Sidecar 模式**部署，利用共享 PID 命名空间实现无侵入监控。
    - 支持 **Linux AMD64** 和 **ARM64** 多架构运行。
- **语言支持**：
    - **Java**: 深度集成 `async-profiler`，支持 CPU、Alloc、Lock 等多种事件采集。
    - 支持自动检测并适配目标容器的 JDK 环境。
- **触发机制**：
    - **阈值触发**: 支持基于 CPU 使用率 (`cpu_usage_percent`) 和内存使用率/量 (`mem_usage_percent`/`mem_usage_mb`) 的自动触发。
    - **API 触发**: 提供 HTTP 接口 `GET /v1/monitor`，支持通过 PID 或正则匹配进程名手动触发采集。
- **数据集成**：
    - 支持将生成的 `.jfr` 或火焰图数据自动上报至 **DataKit**。
    - 支持通过环境变量 `FLAMESHOT_PROCESSES` 灵活配置多进程监控策略及标签 (`tags`)。
