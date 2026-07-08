# Flameshot Go 采集文档

Go 采集通过业务进程暴露的 `net/http/pprof` 端口完成。Flameshot 在触发采集时请求 pprof HTTP 接口，拿到 `.pprof` 数据后通过 Profiling 上传接口发送到 DataKit。

## 前置条件

- 业务 Pod 开启 `shareProcessNamespace: true`，用于进程匹配和资源监控。
- Go 应用暴露 `net/http/pprof` HTTP 端口。
- `pprof_url` 能从 Flameshot Sidecar 内访问。
- pprof 端口只监听 Pod 内可访问地址，不要通过 Service 或公网暴露。

Go 应用示例：

```go
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func main() {
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)

	go func() {
		if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
			log.Printf("pprof server exited: %v", err)
		}
	}()

	select {}
}
```

`mutex` 和 `block` 需要业务代码显式开启 `runtime.SetMutexProfileFraction` 和 `runtime.SetBlockProfileRate`，否则采集结果可能为空。

## 地址选择

Go 的 `pprof_url` 必须填写 Flameshot sidecar 实际能访问到的地址。

常见情况：

- pprof 监听 `127.0.0.1:6060` 或 `0.0.0.0:6060`：配置 `http://127.0.0.1:6060`。
- pprof 只监听 Pod IP，例如 `10.x.x.x:6060`：通过 Downward API 注入 Pod IP，并配置 `http://$(POD_IP):6060`。

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
        "pprof_timeout": "50s"
      }
    ]
```

DataKit 以 DaemonSet 方式部署时，推荐让 Flameshot 通过当前节点 IP 上传 Profile，避免普通 Service 域名将请求转发到其它节点的 DataKit：

```yaml
- name: NODE_IP
  valueFrom:
    fieldRef:
      fieldPath: status.hostIP
- name: FLAMESHOT_DATAKIT_ADDR
  value: "http://$(NODE_IP):9529/profiling/v1/input"
```

## 支持类型

`pprof_types` 支持以下类型：

| 类型 | pprof endpoint | 上传文件名 | 说明 |
| --- | --- | --- | --- |
| `cpu` | `/debug/pprof/profile?seconds=<duration>` | `cpu.pprof` | 按 `duration` 进行 CPU profile |
| `goroutine` | `/debug/pprof/goroutine?debug=0` | `goroutines.pprof` | goroutine profile |
| `heap` | `/debug/pprof/heap` | `delta-heap.pprof` | delta profile，首次采样作为基线 |
| `mutex` | `/debug/pprof/mutex` | `delta-mutex.pprof` | delta profile，首次采样作为基线 |
| `block` | `/debug/pprof/block` | `delta-block.pprof` | delta profile，首次采样作为基线 |

`pprof_types` 可配置为具体列表，也可通过 `events=all` 使用全部类型。未配置 `pprof_types` 和 `events` 时，Go 默认采集 `cpu`。

## 进程规则

Go 规则的核心字段：

| 字段 | 说明 |
| --- | --- |
| `service` | Profiling 服务名 |
| `language` | `go` 或 `golang` |
| `command` | Go 进程命令行正则 |
| `pprof_url` | pprof HTTP 基地址，例如 `http://127.0.0.1:6060` |
| `pprof_types` | pprof 类型列表 |
| `pprof_timeout` | pprof 请求超时。建议大于 `duration` 加 15 秒 |
| `duration` | CPU profile 采集时长 |
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
    "service": "go-app",
    "language": "go",
    "command": "^/app/go-app$",
    "pprof_url": "http://127.0.0.1:6060",
    "pprof_types": ["cpu", "goroutine", "heap", "mutex", "block"],
    "duration": "30s",
    "pprof_timeout": "50s",
    "cpu_usage_percent": 80,
    "mem_usage_percent": 80,
    "mem_usage_percent_emergency": 92,
    "tags": ["env:prod", "version:v1"]
  }
]
```

带索引环境变量示例：

```bash
FLAMESHOT_PROCESSES_0_SERVICE=go-app
FLAMESHOT_PROCESSES_0_LANGUAGE=go
FLAMESHOT_PROCESSES_0_COMMAND='^/app/go-app$'
FLAMESHOT_PROCESSES_0_PPROF_URL=http://127.0.0.1:6060
FLAMESHOT_PROCESSES_0_PPROF_TYPES=cpu,goroutine,heap,mutex,block
FLAMESHOT_PROCESSES_0_DURATION=30s
FLAMESHOT_PROCESSES_0_PPROF_TIMEOUT=50s
```

## Kubernetes 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: go-app
  template:
    metadata:
      labels:
        app: go-app
    spec:
      shareProcessNamespace: true
      containers:
        - name: app
          image: my-go-app:latest
          ports:
            - name: pprof
              containerPort: 6060
        - name: flameshot
          image: pubrepo.jiagouyun.com/datakit/flameshot:latest
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: NODE_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.hostIP
            - name: FLAMESHOT_DATAKIT_ADDR
              value: "http://$(NODE_IP):9529/profiling/v1/input"
            - name: FLAMESHOT_MONITOR_INTERVAL
              value: "1s"
            - name: FLAMESHOT_HTTP_LOCAL_IP
              value: "0.0.0.0"
            - name: FLAMESHOT_HTTP_LOCAL_PORT
              value: "8089"
            - name: FLAMESHOT_TAGS
              value: "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
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
                    "pprof_timeout": "50s",
                    "cpu_usage_percent": 80,
                    "mem_usage_percent": 80,
                    "mem_usage_percent_emergency": 92,
                    "tags": ["env:prod", "version:v1"]
                  }
                ]
```

Go 采集不需要 `async-profiler`，也不依赖本地 JFR 文件。但如果同一个 Flameshot 镜像同时服务 Java 和 Go，仍可保留 `FLAMESHOT_PROFILING_PATH` 和共享目录配置。

## DataKit 要求

Flameshot 上传 Go Profile 前，需要确认 DataKit 侧已经满足以下条件。

开启 `profile` 采集器：

```toml
[[inputs.profile]]
  endpoints = ["/profiling/v1/input"]
```

如果 DataKit 以 DaemonSet 方式运行，且 Flameshot 通过 `NODE_IP:9529` 上传，这个请求在 DataKit 看来不是 localhost 请求。若 DataKit 配置了 HTTP API 白名单，需要追加 `/profiling/v1/input`：

```yaml
- name: ENV_HTTP_PUBLIC_APIS
  value: /otel/v1/trace,/otel/v1/metric,/otel/v1/logs,/profiling/v1/input
```

如果该变量已有其它接口，不要直接覆盖原值，应在原列表后追加 `/profiling/v1/input`。

## 手动触发

按 PID 触发：

```bash
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s"
```

指定采集类型：

```bash
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s&events=cpu,goroutine"
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s&events=all"
```

如果进程规则里配置了 `pprof_types`，Flameshot 会优先使用规则内的类型。

## 本机测试示例

假设本机 DataKit 进程暴露 pprof 到 `127.0.0.1:6060`，DataKit Profiling 上传地址为 `http://127.0.0.1:9529/profiling/v1/input`：

```bash
FLAMESHOT_DATAKIT_ADDR=http://127.0.0.1:9529/profiling/v1/input \
FLAMESHOT_MONITOR_INTERVAL=1s \
FLAMESHOT_HTTP_LOCAL_IP=127.0.0.1 \
FLAMESHOT_HTTP_LOCAL_PORT=18089 \
FLAMESHOT_LOG_LEVEL=debug \
FLAMESHOT_TAGS=host:local,env:local,version:test \
FLAMESHOT_PROCESSES='[
  {
    "service":"datakit-local",
    "language":"go",
    "command":"^/usr/local/datakit/datakit$",
    "pprof_url":"http://127.0.0.1:6060",
    "pprof_types":["cpu","goroutine","heap","mutex","block"],
    "duration":"5s",
    "pprof_timeout":"20s",
    "tags":["test:flameshot-go"]
  }
]' \
dist/flameshot-linux-amd64/flameshot
```

触发采集：

```bash
curl "http://127.0.0.1:18089/v1/profile?command=^/usr/local/datakit/datakit$&duration=5s"
```

第一次采集 `heap`、`mutex`、`block` 会保存 baseline，不上传这些 delta 类型；第二次及之后才会上传 `delta-heap.pprof`、`delta-mutex.pprof`、`delta-block.pprof`。

## 排障

检查 pprof 是否可访问：

```bash
curl -fsS "http://127.0.0.1:6060/debug/pprof/goroutine?debug=0" >/tmp/goroutine.pprof
curl -fsS "http://127.0.0.1:6060/debug/pprof/profile?seconds=5" >/tmp/cpu.pprof
```

如果 pprof 监听在 Pod IP 上，把 `127.0.0.1` 换成实际 Pod IP，或在 Flameshot 配置中使用 `http://$(POD_IP):6060`。

检查 6060 是否真的处于监听状态：

```bash
netstat -lntp | grep ':6060'
ss -ltnp | grep ':6060'
```

只有看到 `:6060 ... LISTEN` 才表示 pprof 已监听。类似 `10.x.x.x:60602 ... ESTABLISHED` 只是临时连接端口，不表示 `6060` 已打开。

检查 Flameshot 是否匹配到进程：

```bash
ps -ef
curl "http://127.0.0.1:8089/v1/profile?command=^/app/go-app$&duration=5s"
```

常见问题：

- `go pprof_url is required`：Go 规则缺少 `pprof_url`。
- `invalid response status`：pprof endpoint 路径不可用，或请求被业务服务拦截。
- `no go pprof data collected`：所有类型都采集失败，检查 pprof 地址、超时和类型配置。
- `connect: connection refused`：对应 IP:Port 没有监听，检查业务进程是否真正开启 pprof，以及 `pprof_url` 是否应使用 `POD_IP`。
- `datakit.publicAccessDisabled`：DataKit 没有对非 localhost 放行 `/profiling/v1/input`，需要在 `ENV_HTTP_PUBLIC_APIS` 中追加该接口。
- `input "profile" is not enabled for API "/profiling/v1/input"`：DataKit 没有开启 `profile` 采集器。
- `heap/mutex/block` 首次未上传：这是 delta profile 的正常行为。
- `mutex/block` 数据为空：检查业务代码是否开启对应 runtime profile。
