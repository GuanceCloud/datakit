---
title     : 'DDTrace Python'
summary   : 'DDTrace Python 集成'
tags      :
  - 'DDTRACE'
  - 'PYTHON'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---

## 安装依赖 {#dependence}

安装 DDTrace SDK。下方示例还使用 Flask 与 requests：

```shell
python -m pip install ddtrace flask requests
```

安装后可执行 `ddtrace-run --info` 检查 SDK 是否能读取启动时配置；该输出不会反映应用运行后通过代码修改的配置。

## 运行应用 {#instrument}

<!-- markdownlint-disable MD046 -->
=== "主机应用"

    > 使用 `ddtrace-run` 包装 Python 入口命令，并在进程启动前设置服务标识和 DataKit 目标地址。上游默认 trace 端口通常为 `8126`，接入 DataKit 时应显式使用 `9529`。
    
    ```shell linenums="1"
    DD_SERVICE="<YOUR-SERVICE-NAME>" \
      DD_ENV="<YOUR-ENV-NAME>" \
      DD_VERSION="<YOUR-APP-VERSION>" \
      DD_AGENT_HOST="<YOUR-DATAKIT-HOST>" \
      DD_TRACE_AGENT_PORT="9529" \
      DD_LOGS_INJECTION=true \
      ddtrace-run python my_app.py
    ```

=== "Kubernetes"

    ```yaml hl_lines="10-19" linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
            - name: <CONTAINER_NAME>
              image: <CONTAINER_IMAGE>/<TAG>
              env:
                - name: DD_AGENT_HOST
                  value: "datakit-service.datakit.svc"
                - name: DD_TRACE_AGENT_PORT
                  value: "9529"
                - name: DD_ENV
                  value: <YOUR-ENV-NAME>
                - name: DD_SERVICE
                  value: <YOUR-SERVICE-NAME>
                - name: DD_VERSION
                  value: <YOUR-APP-VERSION>
                - name: DD_LOGS_INJECTION
                  value: "true"
    ```
<!-- markdownlint-enable MD046 -->

应用启动后，访问一个已插桩的接口并在 DataKit monitor 中确认 trace 请求。排障时可临时增加 `DD_TRACE_DEBUG=true`，确认后应关闭，避免输出过多诊断日志。

除此以外，还有如下常见选项。

### Profiling {#instrument-profile}

```shell linenums="1"
DD_PROFILING_ENABLED=true \
  ddtrace-run python my_app.py
```

### 采样率 {#instrument-sampling}

设置 0.8 的采样率，最终只有 80% 的 trace 会保留下来。

```shell linenums="1"
DD_TRACE_SAMPLE_RATE="0.8" \
  ddtrace-run python my_app.py
```

### 开启 Python 运行时指标采集 {#instrument-py-runtime-metrics}

> 运行时指标通过 DogStatsD 而不是 trace 端口发送，因此需开启 [StatsD 采集器](statsd.md)，并让 `DD_DOGSTATSD_HOST` / `DD_DOGSTATSD_PORT` 指向 DataKit 的 StatsD 服务（通常是 UDP `8125`）。不要把它们设置为 `9529`。

```shell linenums="1"
DD_RUNTIME_METRICS_ENABLED=true \
  ddtrace-run python my_app.py
```

## 代码示例 {#example}

```python title="service_a.py"
from flask import Flask
import requests

app = Flask(__name__)

@app.route('/a',  methods=['GET'])
def index():
    requests.get('http://127.0.0.1:54322/b', timeout=3)
    return 'OK', 200

# 启动 service A: HTTP 服务启动在 54321 端口上
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54321, debug=False, use_reloader=False)
```

```python title="service_b.py"
from flask import Flask
import time

app = Flask(__name__)

@app.route('/b',  methods=['GET'])
def index():
    time.sleep(1)
    return 'OK', 200

# 启动 service B: HTTP 服务启动在 54322 端口上
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54322, debug=False, use_reloader=False)
```

## 运行 {#run}

这里以 Python 中常用的 Web Server Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

- 运行 `SERVICE_A`

```shell
DD_SERVICE=service-a \
DD_ENV=test \
DD_VERSION=v1 \
DD_TAGS=project:your_project_name \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ddtrace-run python3 service_a.py >a.log 2>&1 &
SERVICE_A_PID=$!
```

- 运行 `SERVICE_B`

```shell
DD_SERVICE=service-b \
DD_ENV=test \
DD_VERSION=v1 \
DD_TAGS=project:your_project_name \
DD_AGENT_HOST=localhost \
DD_TRACE_AGENT_PORT=9529 \
ddtrace-run python3 service_b.py >b.log 2>&1 &
SERVICE_B_PID=$!
```

调用 A 服务，促使其调用 B 服务，这样就能产生对应 trace 数据（此处可多次执行触发）

```shell
curl http://localhost:54321/a
```

分别停止两个服务：

```shell
kill "$SERVICE_A_PID" "$SERVICE_B_PID"
```

## 环境变量支持 {#envs}

常见环境变量支持如下。它们应在 Python 进程启动前设置；完整列表、优先级与版本差异参见 [DataDog Python 配置文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/python/){:target="_blank"}。

- `DD_ENV`: 为服务设置环境变量。
- `DD_VERSION`: APP 版本号。
- `DD_SERVICE`: 设置应用服务名称。对 Web 框架通常会传给框架集成；生产环境建议显式设置。
- `DD_SERVICE_MAPPING`: 定义依赖服务名映射，用于在链路中重命名依赖服务。
- `DD_TAGS`: 为每个 span 添加默认标签，格式为 `key:val,key:val`；不要加入用户标识或敏感内容。
- `DD_AGENT_HOST`: DataKit 主机名或 IP。若设置 `DD_TRACE_AGENT_URL`，该 URL 通常优先。
- `DD_TRACE_AGENT_PORT`: trace 接收端端口；上游默认通常为 `8126`，DataKit 使用 `9529`。
- `DD_TRACE_SAMPLE_RATE`: 设置 SDK 侧采样率，范围为 `0.0`（0%）到 `1.0`（100%）。
- `DD_TRACE_ENABLED`: 控制 trace 生成/发送；排障时确认没有被设置为 `false`。
