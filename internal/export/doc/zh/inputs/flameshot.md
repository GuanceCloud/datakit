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
1. **执行 (Execute)**：根据配置的语言类型（目前支持 Java），调用对应的 Profiler 工具 attach 到目标进程。
1. **收集 (Collect)**：生成的 Profile 文件（如 `.jfr`）存储于共享卷中，随后上传至数据观测中心。
1. **定时**: 配置 `FLAMESHOT_AUTO_PROFILING` 后，会定时对所有匹配到的进程采集一次 30s 的 Profiling 数据。

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
| `FLAMESHOT_AUTO_PROFILING`   | 否     | -       | 定时对所有匹配到的进程采集一次 30s 的 Profiling 数据。最小不得低于一分钟，如五分钟："5m" 或者一小时 "1h"         |
| `FLAMESHOT_SERVICE`          | 否     | -       | 可以不用在 `FLAMESHOT_PROCESSES` 中配置 `service`, 会全部替换。                         |
| `FLAMESHOT_TAGS`             | 否     | -       | 建议配置 `host` `pod_name` `pod_namespace` 如： "host:host_name,pod_name:pod_a" |

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
              "tags": [
                "env:prod",
                "version:v1.2"
              ]
            }
          ]
```

**通用字段说明：**

- **`service`** (String): 上报到观测中心的服务名称。
- **`language`** (String): 目标进程语言。目前支持 `java`。
- **`command`** (String): 匹配进程命令行的正则表达式。
- **`duration`** (String): 单次采集时长（例如 `30s`, `1m`）。**注意**：受限于执行超时，建议不超过 5 分钟。
- **`tags`** (List): 自定义标签列表，建议包含 `env`, `version` 等元信息。
- **`cpu_usage_percent`** (Int): CPU 触发阈值 (0-N)。多核环境下数值可能超过 100。
- **`mem_usage_percent`** (Int): 内存使用率触发阈值 (0-100)。
- **`mem_usage_mb`** (Int): 内存使用量绝对值触发阈值 (MB)。
- 这三个配置： **`cpu_usage_percent`** **`mem_usage_percent`** **`mem_usage_mb`** 不配置或者配置 0 都会略过该项的阈值检查。


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
    - 如果使用非标准 JDK 镜像，请确保 Sidecar 挂载了主容器的 `/tmp` 或相应的 Java 库路径。

=== "Go (Coming Soon)"

    ### Go Profiling {#go-profiling}
    
    *计划中*：将集成 `pprof` 工具链。
    
    **预期特性：**

    -   支持 Goroutine 阻塞分析。
    -   支持 Heap 内存快照。

=== "Python (Coming Soon)"

    ### Python Profiling {#python-profiling}
    
    *计划中*：将集成 `py-spy` 等无侵入式工具。
<!-- markdownlint-enable -->

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

### Docker 本地测试 {#docker-testing}

如果您需要在本地 Docker 环境中进行测试，可以使用以下命令启动 Flameshot 并监控目标容器。

**前提条件：**

- 主容器与 Flameshot 容器共享 `/opt/java/openjdk` (或实际 JDK 路径)。
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
| `events` | 否 | 采集事件类型。默认为 `all`。 | `cpu,alloc` |

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

2. **文件未上传？**

    - 检查 `FLAMESHOT_PROFILING_PATH` 是否在两个容器间正确挂载。
    - 系统会自动管理文件生命周期，采集完成后会尝试删除临时文件。

3. **配置正则太麻烦**

    - JAVA 应用的进程名都是 `java`, 所以配置 `"command":"java"` `"language": "java"` 即可匹配所有的 JAVA 应用。
    - 想要配置特定的应用而不是所有，正则是必须要配置的。


## 更新日志 (Changelog) {#changelog}

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
