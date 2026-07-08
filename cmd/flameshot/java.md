# Flameshot Java 采集文档

Java 采集通过 `async-profiler` 完成。Flameshot 匹配到目标 Java 进程后，会执行 Sidecar 内的 `asprof`，生成 JFR 文件并上传到 DataKit。

## 前置条件

- 业务 Pod 开启 `shareProcessNamespace: true`。
- Flameshot Sidecar 能看到目标 JVM 进程。
- Flameshot 容器内存在 `async-profiler`，默认路径为 `/opt/async-profiler`。
- `FLAMESHOT_PROFILING_PATH` 指向业务容器和 Sidecar 共享的可写目录。
- 通常需要给 Sidecar 增加 `SYS_PTRACE` capability。

## 进程规则

Java 规则的核心字段：

| 字段 | 说明 |
| --- | --- |
| `service` | Profiling 服务名 |
| `language` | 固定为 `java` |
| `command` | JVM 命令行正则 |
| `events` | async-profiler 采集事件，支持 `cpu`、`alloc`、`lock`、`nativemem`、`all` 等 |
| `duration` | 普通采集时长 |
| `emergency_duration` | 内存紧急触发时的短采集时长 |
| `cpu_usage_percent` | CPU 使用率阈值 |
| `mem_usage_percent` | 平均内存百分比阈值 |
| `mem_usage_mb` | 平均 RSS 阈值，单位 MB |
| `mem_usage_percent_emergency` | 单点内存百分比紧急阈值 |
| `mem_usage_mb_emergency` | 单点 RSS 紧急阈值，单位 MB |
| `tags` | 当前规则标签 |

示例：

```json
[
  {
    "service": "java-app",
    "language": "java",
    "command": "^java\\b.*app\\.jar$",
    "events": "cpu,alloc",
    "duration": "30s",
    "emergency_duration": "10s",
    "cpu_usage_percent": 80,
    "mem_usage_percent": 80,
    "mem_usage_mb": 1536,
    "mem_usage_percent_emergency": 92,
    "mem_usage_mb_emergency": 1900,
    "tags": ["env:prod", "version:v1"]
  }
]
```

## Kubernetes 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: java-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: java-app
  template:
    metadata:
      labels:
        app: java-app
    spec:
      shareProcessNamespace: true
      volumes:
        - name: flameshot-data
          emptyDir: {}
      containers:
        - name: app
          image: my-java-app:latest
          command:
            - java
            - -XX:+HeapDumpOnOutOfMemoryError
            - -XX:HeapDumpPath=/flameshot-data/dumps/app.hprof
            - -jar
            - app.jar
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
            - name: FLAMESHOT_POD_MEM_LIMIT
              value: "2048"
            - name: FLAMESHOT_TAGS
              value: "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
            - name: FLAMESHOT_PROCESSES
              value: |
                [
                  {
                    "service": "java-app",
                    "language": "java",
                    "command": "^java\\b.*app\\.jar$",
                    "events": "cpu,alloc",
                    "duration": "30s",
                    "emergency_duration": "10s",
                    "cpu_usage_percent": 80,
                    "mem_usage_percent": 80,
                    "mem_usage_mb": 1536,
                    "mem_usage_percent_emergency": 92,
                    "mem_usage_mb_emergency": 1900,
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

## OOM hprof 摘要恢复

如果需要在 OOM 后恢复 `.hprof` 摘要，目标 JVM 需要配置 heap dump 参数：

```bash
java \
  -XX:+HeapDumpOnOutOfMemoryError \
  -XX:HeapDumpPath=/flameshot-data/dumps/app.hprof \
  -jar app.jar
```

注意事项：

- `HeapDumpPath` 必须位于业务容器和 Flameshot 共享的目录内。
- Flameshot 会从目标 Java 进程参数中解析 `HeapDumpPath`。
- 如果 Pod 被过快杀死，JVM 可能来不及生成 `.hprof`。
- OOM 摘要上传到 DataKit logging 写入接口，measurement 为 `flameshot_oom_hprof`。

## 采集事件

`events` 会传给 `async-profiler`：

- `all` 或 `--all`：采集 async-profiler 支持的多类事件。
- `cpu`：CPU 采样。
- `alloc`：对象分配采样。
- `lock`：锁竞争采样。
- `nativemem`：Native memory 采样。

实际可用事件依赖目标 JVM、内核、容器权限和 async-profiler 版本。

## 手动触发

按 PID 触发：

```bash
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s&events=cpu,alloc"
```

按命令行正则触发：

```bash
curl "http://127.0.0.1:8089/v1/profile?command=^java\\b.*app\\.jar$&duration=30s&events=cpu"
```

## 排障

检查 Sidecar 是否能看到目标 JVM：

```bash
ps -ef
```

检查 async-profiler：

```bash
ls -lah /opt/async-profiler
ls -lah /opt/async-profiler/bin/asprof
```

检查共享目录：

```bash
ls -lah /flameshot-data
ls -lah /flameshot-data/dumps
```

常见问题：

- `asprof` 不存在：检查 Flameshot 镜像或 `async-profiler` 复制流程。
- `Permission denied`：检查 `SYS_PTRACE`、容器用户和共享目录权限。
- 没有生成 JFR：检查 `events` 是否被目标环境支持，以及 `duration` 是否足够。
- OOM 没有 hprof：检查 JVM 参数和 `HeapDumpPath` 是否在共享目录内。
