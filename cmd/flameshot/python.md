# Flameshot Python 采集文档

Flameshot 使用官方 `py-spy` 对 Python 进程做无侵入 attach 采样。命中 CPU/内存阈值、定时任务或 HTTP 手动触发后，Flameshot 会执行 `py-spy record --format raw`，读取生成的 raw collapsed profile，并上传到 DataKit Profiling。

## 采集方式

- 本地执行：`py-spy record --format raw --output <path> --duration <seconds> --pid <pid>`。
- 上传格式：event 中 `family=python`、`format=collapse`、`profiler=pyspy`。
- 附件名：固定为 `prof`。
- 中心解析：DataKit 转发后由 pprofparser 的 collapsed/raw flamegraph parser 解析。

`py-spy` 自身负责 attach 和采样。Flameshot 的阈值只控制何时启动一次采集，不会在 attach 后按阈值持续开关采样。

## 进程规则

```json
[
  {
    "service": "python-api",
    "language": "python",
    "command": "^python\\b.*app\\.py$",
    "duration": "30s",
    "pyspy_path": "py-spy",
    "pyspy_rate": 100,
    "pyspy_subprocesses": false,
    "pyspy_idle": false,
    "cpu_usage_percent": 80,
    "mem_usage_percent": 80,
    "tags": ["env:prod", "version:v1"]
  }
]
```

字段说明：

| 字段 | 默认值 | 说明 |
| --- | --- | --- |
| `language` | - | 必须配置为 `python` |
| `command` | - | Python 目标进程命令行正则 |
| `duration` | `60s` | 单次采集时长 |
| `pyspy_path` | `py-spy` | `py-spy` 可执行文件路径 |
| `pyspy_output_path` | 自动生成 | raw profile 输出路径；未配置时自动生成本地临时文件，相对路径会写到 `FLAMESHOT_PROFILING_PATH` 或默认输出目录 |
| `pyspy_rate` | `100` | 每秒采样次数 |
| `pyspy_subprocesses` | `false` | 是否对子进程一起采样 |
| `pyspy_idle` | `false` | 是否包含 idle 线程 |

等价索引环境变量：

```bash
FLAMESHOT_PROCESSES_0_SERVICE=python-api
FLAMESHOT_PROCESSES_0_LANGUAGE=python
FLAMESHOT_PROCESSES_0_COMMAND='^python\b.*app\.py$'
FLAMESHOT_PROCESSES_0_DURATION=30s
FLAMESHOT_PROCESSES_0_PYSPY_PATH=py-spy
FLAMESHOT_PROCESSES_0_PYSPY_RATE=100
FLAMESHOT_PROCESSES_0_PYSPY_SUBPROCESSES=false
FLAMESHOT_PROCESSES_0_PYSPY_IDLE=false
```

## Kubernetes 示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: python-api
  template:
    metadata:
      labels:
        app: python-api
    spec:
      shareProcessNamespace: true
      volumes:
        - name: flameshot-data
          emptyDir: {}
      containers:
        - name: app
          image: python-api:latest
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
        - name: flameshot
          image: pubrepo.jiagouyun.com/datakit/flameshot:latest
          securityContext:
            capabilities:
              add: ["SYS_PTRACE"]
          volumeMounts:
            - name: flameshot-data
              mountPath: /flameshot-data
          env:
            - name: FLAMESHOT_DATAKIT_ADDR
              value: "http://datakit-service.datakit:9529/profiling/v1/input"
            - name: FLAMESHOT_PROFILING_PATH
              value: "/flameshot-data"
            - name: FLAMESHOT_HTTP_LOCAL_IP
              value: "0.0.0.0"
            - name: FLAMESHOT_HTTP_LOCAL_PORT
              value: "8089"
            - name: FLAMESHOT_PROCESSES
              value: |
                [
                  {
                    "service": "python-api",
                    "language": "python",
                    "command": "^python\\b.*app\\.py$",
                    "duration": "30s",
                    "pyspy_rate": 100,
                    "cpu_usage_percent": 80,
                    "tags": ["env:prod"]
                  }
                ]
```

如果运行环境启用了更严格的 seccomp 或 ptrace 限制，需要按集群安全策略额外放开 `py-spy` attach 所需权限。

## 手动触发

```bash
curl "http://127.0.0.1:8089/v1/profile?pid=1234&duration=30s"
curl "http://127.0.0.1:8089/v1/profile?command=^python\\b.*app\\.py$&duration=30s"
```

## 本机验证

```bash
py-spy record --format raw --output /tmp/prof --duration 10 --pid <pid>
```

确认 `/tmp/prof` 内容是 collapsed 文本后，可用相同命令行权限运行 Flameshot。上传成功时，Flameshot 日志会包含 `py-spy output file`，DataKit Profiling event 中应看到 `family=python`、`format=collapse`、`profiler=pyspy`。

## 通用配置

Flameshot 的进程匹配、阈值、定时采集、HTTP 手动触发等公共能力见 [总文档](./readme.md)。
