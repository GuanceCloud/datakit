# Flameshot 总文档

Flameshot 是 DataKit 的 Profiling Sidecar 工具，用于在业务进程出现 CPU 高水位、内存压力、定时采集或手动触发时，采集现场 Profile 数据并上传到 DataKit。

当前支持：

- Java：通过 `async-profiler` 采集 `.jfr`，并支持 OOM 后 `.hprof` 摘要恢复、主动 Heap Dump 和 hprof 对象存储上传。
- Go：通过业务进程暴露的 `net/http/pprof` HTTP 端口拉取 `.pprof` 并上传。
- Python：通过官方 `py-spy` attach 目标进程，采集 raw collapsed profile 并上传。

## 文档索引

- [Java 采集文档](./java.md)
- [Go 采集文档](./go.md)
- [Python 采集文档](./python.md)

## 工作模型

Flameshot 通常作为业务 Pod 内的 Sidecar 运行：

1. 通过进程命令行正则匹配目标进程。
1. 周期性读取目标进程的 CPU、RSS、cgroup 或 Pod 限额信息。
1. 命中阈值、定时任务或 HTTP 手动触发后，按目标语言执行对应采集流程。
1. 将采集结果以 Profiling multipart 请求上传到 DataKit。
1. Java 场景下，OOM 或高内存风险时可额外保留 `.hprof` 现场。

推荐部署条件：

- 业务容器和 Flameshot 容器运行在同一个 Pod。
- Pod 开启 `shareProcessNamespace: true`，让 Sidecar 能看到业务进程。
- Java 场景共享可写目录，用于存放 `async-profiler`、JFR、Heap Dump 和 OOM 文件。
- Go 场景要求 pprof HTTP 地址能从 Sidecar 内访问。
- Python 场景要求 Sidecar 内存在 `py-spy`，并具备 attach 目标进程所需权限，通常需要 `SYS_PTRACE`。

推荐共享路径：

- Profiling 输出目录：`/flameshot-data`
- Java heap dump 目录：`/flameshot-data/dumps`

## 通用环境变量

| 变量 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `FLAMESHOT_DATAKIT_ADDR` | 是 | - | DataKit Profiling 上传地址，例如 `http://datakit-service.datakit:9529/profiling/v1/input` |
| `FLAMESHOT_PROFILING_PATH` | Java/Python 建议配置 | `/data` | 共享可写目录。Java 用于 profiler 工具、JFR 和 OOM 文件；Python 相对 `pyspy_output_path` 会写入该目录；Go 采集不依赖本地落盘 |
| `FLAMESHOT_MONITOR_INTERVAL` | 否 | `1s` | 进程资源轮询间隔 |
| `FLAMESHOT_LOG_LEVEL` | 否 | `info` | 日志级别，支持 `debug` |
| `FLAMESHOT_LOG_PATH` | 否 | `/var/log/flameshot/log` | 日志路径配置 |
| `FLAMESHOT_HTTP_LOCAL_IP` | 是 | - | Flameshot HTTP 监听地址 |
| `FLAMESHOT_HTTP_LOCAL_PORT` | 是 | `8089` | Flameshot HTTP 监听端口 |
| `FLAMESHOT_PROFILING_ENABLED` | 否 | `true` | 是否开启 Profiling。设置为 `false` 后会关闭定时、阈值、cgroup 高水位和 HTTP 手动 Profiling，但保留 OOM 检测、hprof 上传和主动 Heap Dump |
| `FLAMESHOT_AUTO_PROFILING` | 否 | - | 定时自动采集间隔，开启时最小 1 分钟；配置 `0` 表示关闭 |
| `FLAMESHOT_AUTO_PROFILING_DURATION` | 否 | `30s` | 定时自动采集时长 |
| `FLAMESHOT_OOM_HPROF_ENABLED` | 否 | `false` | 是否开启 Java OOM `.hprof` 摘要恢复 |
| `FLAMESHOT_OOM_HPROF_MATCH_WINDOW` | 否 | `2m` | OOM 事件和 hprof 文件的匹配窗口 |
| `FLAMESHOT_HPROF_UPLOAD_ENABLED` | 否 | `false` | 是否将匹配到或主动生成的 `.hprof` 上传到对象存储 |
| `FLAMESHOT_HPROF_UPLOAD_PROVIDER` | 否 | - | 对象存储类型，支持 `oss` 或 `s3` |
| `FLAMESHOT_HPROF_UPLOAD_ENDPOINT` | 否 | - | OSS/S3 endpoint |
| `FLAMESHOT_HPROF_UPLOAD_REGION` | 否 | S3 默认为 `us-east-1` | S3 region |
| `FLAMESHOT_HPROF_UPLOAD_BUCKET` | 否 | - | 目标 bucket |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID` | 否 | - | 对象存储 AK |
| `FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET` | 否 | - | 对象存储 SK |
| `FLAMESHOT_HPROF_UPLOAD_PATH_TEMPLATE` | 否 | `{service}/{pod_name}/{timestamp}/{filename}` | 对象路径模板 |
| `FLAMESHOT_HPROF_DOWNLOAD_URL_TEMPLATE` | 否 | - | 可选下载链接模板；配置后事件中会上报渲染后的下载链接 |
| `FLAMESHOT_HPROF_UPLOAD_TIMEOUT` | 否 | `5m` | hprof 上传超时时间 |
| `FLAMESHOT_HPROF_UPLOAD_S3_PATH_STYLE` | 否 | `true` | S3 上传是否使用 path-style endpoint |
| `FLAMESHOT_HEAP_DUMP_ENABLED` | 否 | `false` | 内存紧急阈值命中时是否主动执行 Java Heap Dump |
| `FLAMESHOT_HEAP_DUMP_PATH_TEMPLATE` | 否 | `{profiling_path}/dumps/{service}_{pod_name}_{pid}_{timestamp}.hprof` | 本地 Heap Dump 输出路径模板 |
| `FLAMESHOT_HEAP_DUMP_JMAP_PATH` | 否 | `jmap` | `jmap` 可执行文件路径。官方 Sidecar 镜像默认不内置 JVM/JDK，开启主动 Heap Dump 时需显式提供可用 `jmap` |
| `FLAMESHOT_HEAP_DUMP_TIMEOUT` | 否 | `120s` | Heap Dump 命令超时时间 |
| `FLAMESHOT_HEAP_DUMP_COOLDOWN` | 否 | `10m` | 单进程主动 Heap Dump 冷却时间 |
| `FLAMESHOT_POD_MEM_LIMIT` | 否 | - | Pod 内存 limit，单位 Mi。配置后内存百分比优先按 Pod limit 计算 |
| `FLAMESHOT_POD_CPU_LIMIT` | 否 | - | Pod CPU limit，单位 millicore |
| `FLAMESHOT_SERVICE` | 否 | - | 覆盖所有进程规则中的 `service` |
| `FLAMESHOT_TAGS` | 否 | - | 全局标签，例如 `host:node-a,pod_name:demo,pod_namespace:prod` |

## 进程规则

`FLAMESHOT_PROCESSES` 是 JSON 数组字符串。每个元素定义一条进程匹配和采集规则。

| 字段 | 说明 |
| --- | --- |
| `service` | 上传到 DataKit 的服务名 |
| `command` | 目标进程命令行正则 |
| `language` | 目标语言，当前支持 `java`、`go`、`golang`、`python` |
| `duration` | 普通采集时长 |
| `emergency_duration` | 内存紧急触发时的短采集时长 |
| `events` | Java async-profiler 事件，例如 `cpu`、`alloc`、`lock`、`nativemem`、`all`；Go 未配置 `pprof_types` 时也可用作 pprof 类型列表 |
| `pprof_url` | Go pprof HTTP 基地址，例如 `http://127.0.0.1:6060` |
| `pprof_types` | Go pprof 类型，支持 `cpu`、`goroutine`、`heap`、`mutex`、`block` |
| `pprof_timeout` | Go pprof 请求超时时间，应大于 CPU profile 的 `duration` |
| `pyspy_path` | Python `py-spy` 可执行文件路径，默认 `py-spy` |
| `pyspy_output_path` | Python raw 输出路径；未配置时自动生成本地临时文件，相对路径会写入 `FLAMESHOT_PROFILING_PATH` 或默认输出目录 |
| `pyspy_rate` | Python 采样频率，默认 `100` |
| `pyspy_subprocesses` | Python 是否对子进程一起采样，默认 `false` |
| `pyspy_idle` | Python 是否采集 idle 线程，默认 `false` |
| `cpu_usage_percent` | CPU 使用率阈值 |
| `mem_usage_percent` | 最近 5 个点平均内存百分比阈值 |
| `mem_usage_mb` | 最近 5 个点平均 RSS 阈值，单位 MB |
| `mem_usage_percent_emergency` | 单点即时内存百分比紧急阈值 |
| `mem_usage_mb_emergency` | 单点即时 RSS 紧急阈值，单位 MB |
| `heap_dump_on_memory_emergency` | 当前规则是否允许在内存紧急阈值命中时主动 Heap Dump。未配置时，在 `FLAMESHOT_HEAP_DUMP_ENABLED=true` 的情况下默认允许 |
| `tags` | 当前规则的自定义标签 |

示例：

```json
[
  {
    "service": "demo-service",
    "language": "java",
    "command": "^java\\b.*app\\.jar$",
    "events": "cpu,alloc",
    "duration": "30s",
    "emergency_duration": "10s",
    "cpu_usage_percent": 80,
    "mem_usage_percent": 80,
    "mem_usage_percent_emergency": 92,
    "mem_usage_mb_emergency": 1900,
    "heap_dump_on_memory_emergency": true,
    "tags": ["env:prod", "version:v1"]
  }
]
```

Python 示例：

```json
[
  {
    "service": "python-api",
    "language": "python",
    "command": "^python\\b.*app\\.py$",
    "duration": "30s",
    "pyspy_rate": 100,
    "cpu_usage_percent": 80,
    "mem_usage_percent": 80,
    "tags": ["env:prod", "version:v1"]
  }
]
```

也可以使用带索引的环境变量：

```bash
FLAMESHOT_PROCESSES_0_SERVICE=demo-service
FLAMESHOT_PROCESSES_0_LANGUAGE=java
FLAMESHOT_PROCESSES_0_COMMAND='^java\b.*app\.jar$'
FLAMESHOT_PROCESSES_0_EVENTS=cpu,alloc
FLAMESHOT_PROCESSES_0_DURATION=30s
FLAMESHOT_PROCESSES_0_TAGS='["env:prod","version:v1"]'
```

Python 规则可用的索引环境变量还包括：

```bash
FLAMESHOT_PROCESSES_0_PYSPY_PATH=/usr/local/bin/py-spy
FLAMESHOT_PROCESSES_0_PYSPY_RATE=100
FLAMESHOT_PROCESSES_0_PYSPY_SUBPROCESSES=false
FLAMESHOT_PROCESSES_0_PYSPY_IDLE=false
```

## 触发方式

Flameshot 支持以下触发方式：

- CPU 阈值：`cpu_usage_percent`。
- 平均内存阈值：`mem_usage_percent`、`mem_usage_mb`，基于最近 5 个采样点。
- 紧急内存阈值：`mem_usage_percent_emergency`、`mem_usage_mb_emergency`，单点命中立即触发。
- cgroup 内存高水位：按进程规则中的 `mem_usage_percent_emergency` 判断。
- 定时采集：`FLAMESHOT_AUTO_PROFILING` 和 `FLAMESHOT_AUTO_PROFILING_DURATION`。
- HTTP 手动触发：`GET /v1/profile`。
- Java OOM 后处理：开启后尝试匹配并上传 `.hprof` 摘要日志。

HTTP 手动触发示例：

```bash
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s&events=cpu,alloc"
curl "http://127.0.0.1:8089/v1/profile?command=^java\\b.*app\\.jar$&duration=30s"
```

Go 进程手动触发时，Flameshot 会优先使用进程规则里的 `pprof_types`。如果没有配置 `pprof_types`，可以通过 `events=cpu,goroutine` 传入 pprof 类型。

Python 进程手动触发时，Flameshot 会执行 `py-spy record --format raw`，上传到 DataKit 的 event 使用 `format=collapse`、`profiler=pyspy`，附件名固定为 `prof`。

## Kubernetes 基础模板

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: profiled-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: profiled-app
  template:
    metadata:
      labels:
        app: profiled-app
    spec:
      shareProcessNamespace: true
      volumes:
        - name: flameshot-data
          emptyDir: {}
      containers:
        - name: app
          image: my-app:latest
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
        - name: flameshot
          image: pubrepo.jiagouyun.com/datakit/flameshot:latest
          env:
            - name: FLAMESHOT_DATAKIT_ADDR
              value: "http://datakit-service.datakit:9529/profiling/v1/input"
            - name: FLAMESHOT_PROFILING_PATH
              value: "/flameshot-data"
            - name: FLAMESHOT_MONITOR_INTERVAL
              value: "1s"
            - name: FLAMESHOT_HTTP_LOCAL_IP
              value: "0.0.0.0"
            - name: FLAMESHOT_HTTP_LOCAL_PORT
              value: "8089"
            - name: FLAMESHOT_AUTO_PROFILING
              value: "10m"
            - name: FLAMESHOT_AUTO_PROFILING_DURATION
              value: "15s"
            - name: FLAMESHOT_OOM_HPROF_ENABLED
              value: "true"
            - name: FLAMESHOT_OOM_HPROF_MATCH_WINDOW
              value: "3m"
            - name: FLAMESHOT_HPROF_UPLOAD_ENABLED
              value: "true"
            - name: FLAMESHOT_HPROF_UPLOAD_PROVIDER
              value: "s3"
            - name: FLAMESHOT_HPROF_UPLOAD_ENDPOINT
              value: "https://s3.example.com"
            - name: FLAMESHOT_HPROF_UPLOAD_REGION
              value: "us-east-1"
            - name: FLAMESHOT_HPROF_UPLOAD_BUCKET
              value: "heap-dumps"
            - name: FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: flameshot-hprof-upload
                  key: access-key-id
            - name: FLAMESHOT_HPROF_UPLOAD_ACCESS_KEY_SECRET
              valueFrom:
                secretKeyRef:
                  name: flameshot-hprof-upload
                  key: access-key-secret
            - name: FLAMESHOT_HEAP_DUMP_ENABLED
              value: "true"
            - name: FLAMESHOT_POD_MEM_LIMIT
              value: "2048"
            - name: FLAMESHOT_POD_CPU_LIMIT
              value: "1000"
            - name: FLAMESHOT_TAGS
              value: "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
            - name: FLAMESHOT_PROCESSES
              value: |
                [
                  {
                    "service": "demo-service",
                    "language": "java",
                    "command": "^java\\b.*app\\.jar$",
                    "events": "cpu,alloc",
                    "duration": "30s",
                    "emergency_duration": "10s",
                    "cpu_usage_percent": 80,
                    "mem_usage_percent": 80,
                    "mem_usage_percent_emergency": 92,
                    "mem_usage_mb_emergency": 1900,
                    "heap_dump_on_memory_emergency": true,
                    "tags": ["env:prod", "version:v1"]
                  }
                ]
          securityContext:
            capabilities:
              add: ["SYS_PTRACE"]
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
```

Java 和 Go 的完整示例见对应语言文档。

## Java OOM 与高内存现场

如果希望 Flameshot 在 OOM 后恢复 `.hprof` 摘要，目标 JVM 需要包含：

```bash
java \
  -XX:+HeapDumpOnOutOfMemoryError \
  -XX:HeapDumpPath=/flameshot-data/dumps/app.hprof \
  -jar app.jar
```

说明：

- `HeapDumpPath` 应位于共享卷内。
- Flameshot 会直接从目标 Java 进程启动参数中解析 `HeapDumpPath`。
- 如果 Pod 被过快杀死，`.hprof` 仍可能来不及生成。

内存紧急阈值命中后，Flameshot 会按配置保留现场：

1. 如果 `FLAMESHOT_HEAP_DUMP_ENABLED=true`，执行 `jmap` 主动生成 `.hprof`。
1. 如果 `FLAMESHOT_PROFILING_ENABLED=true`，触发一次使用 `emergency_duration` 的短 profiling。
1. 如果后续观察到 `oom_kill` 增量，尝试定位匹配的 `.hprof` 并上传 OOM 摘要日志。
1. 如果开启 hprof 上传，主动生成或匹配到的 `.hprof` 会上传至 OSS/S3，并在事件中带上对象路径、下载链接和上传状态。

写入 `FLAMESHOT_PROFILING_PATH` 的原始文件包括：

- `profiler_<timestamp>.jfr`
- JVM 或 `jmap` 生成的 `.hprof` 文件

## 排障入口

通用检查命令：

```bash
ps -ef
env | grep FLAMESHOT_
ls -lah /opt/async-profiler
ls -lah /flameshot-data
cat /var/log/flameshot.log
```

常见问题：

- 未触发采集：检查进程命令行是否匹配 `command` 正则，以及阈值是否实际命中。
- 上传失败：检查 `FLAMESHOT_DATAKIT_ADDR` 是否为 Profiling 上传地址。
- Sidecar 看不到进程：检查 Pod 是否开启 `shareProcessNamespace: true`。
- 标签缺失或服务名不符合预期：检查 `service`、`FLAMESHOT_SERVICE`、`FLAMESHOT_TAGS` 和规则内 `tags`。
- 主动 Heap Dump 没有生成：检查 Sidecar 内是否存在与目标 JVM 兼容的 `jmap`，`FLAMESHOT_HEAP_DUMP_JMAP_PATH` 是否指向该可执行文件，以及 Sidecar 是否有权限 attach 目标 Java 进程。
- hprof 上传失败：检查本地日志中的 `hprof upload failed` 记录。

语言相关排障：

- Java：见 [Java 采集文档](./java.md)。
- Go：见 [Go 采集文档](./go.md)。
- Python：见 [Python 采集文档](./python.md)。
