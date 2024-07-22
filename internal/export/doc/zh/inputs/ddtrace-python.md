
# Python
---

## 安装依赖 {#dependence}

安装 DDTrace SDK

```shell
pip install ddtrace
```

## 运行应用 {#instrument}

<!-- markdownlint-disable MD046 -->
=== "主机应用"

    > 注意：此处是使用 ddtrace-run 来启动 Python 应用。
    
    ```shell linenums="1"
    DD_SERVICE="<YOUR-SERVICE-Name>" \
      DD_ENV="<YOUR-ENV-NAME>" \
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
                - name: DD_LOGS_INJECTION
                  value: "true"
    ```
<!-- markdownlint-enable -->

除此以外，还有如下其它常见的几个选项可以开启。

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

> 此处需开启 [statsd 采集器](statsd.md)。

```shell linenums="1"
DD_RUNTIME_METRICS_ENABLED=true \
  ddtrace-run python my_app.py
```

## 代码示例 {#example}

```python title="service_a.py"
from flask import Flask, request
import requests, os
from ddtrace import tracer

app = Flask(__name__)

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@app.route('/a',  methods=['GET'])
def index():
    requests.get('http://127.0.0.1:54322/b')
    return 'OK', 200

@app.route('/stop',  methods=['GET'])
def stop():
    shutdown_server()
    return 'Server shutting down...\n'

# 启动 service A: HTTP 服务启动在 54321 端口上
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54321, debug=True)
```

```python title="service_b.py"
from flask import Flask, request
import os, time, requests
from ddtrace import tracer

app = Flask(__name__)

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@app.route('/b',  methods=['GET'])
def index():
    time.sleep(1)
    return 'OK', 200

@app.route('/stop',  methods=['GET'])
def stop():
    shutdown_server()
    return 'Server shutting down...\n'

# 启动 service B: HTTP 服务启动在 54322 端口上
if __name__ == '__main__':
    app.run(host="0.0.0.0", port=54322, debug=True)
```

## 运行 {#run}

这里以 Python 中常用的 Web Server Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

- 运行 `SERVICE_A`

```shell
DD_SERVICE=SERVICE_A \
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql \
DD_TAGS=project:your_project_name,env:test,version:v1 \
DD_AGENT_HOST=localhost \
DD_AGENT_PORT=9529 \
ddtrace-run python3 service_a.py &> a.log &
```

- 运行 `SERVICE_B`

```shell
DD_SERVICE=SERVICE_B \
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql \
DD_TAGS=project:your_project_name,env:test,version:v1 \
DD_AGENT_HOST=localhost \
DD_AGENT_PORT=9529 \
ddtrace-run python3 service_b.py &> b.log &
```

调用 A 服务，促使其调用 B 服务，这样就能产生对应 trace 数据（此处可多次执行触发）

```shell
curl http://localhost:54321/a
```

分别停止两个服务

```shell
curl http://localhost:54321/stop
curl http://localhost:54322/stop
```

## 环境变量支持 {#envs}

常见环境变量支持如下，完整的 python 环境变量列表参见 [DataDog 官方文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/python/){:target="_blank"}。

- `DD_ENV`: 为服务设置环境变量。
- `DD_VERSION`: APP 版本号。
- `DD_SERVICE`: 用于设置应用程序的服务名称，在为 Pylons、Flask 或 Django 等 Web 框架集成设置中间件时，会传递该值。 对于没有 Web 集成的 Tracing，建议您在代码中设置服务名称。
- `DD_SERVICE_MAPPING`: 定义服务名映射用于在 Tracing 里重命名服务。
- `DD_TAGS`: 为每个 Span 添加默认 Tags，格式为 `key:val,key:val`。
- `DD_AGENT_HOST`: Datakit 监听的地址名，默认 localhost。
- `DD_AGENT_PORT`: Datakit 监听的端口号，默认 9529。
- `DD_TRACE_SAMPLE_RATE`: 设置采样率从 0.0(0%) ~ 1.0(100%)。
