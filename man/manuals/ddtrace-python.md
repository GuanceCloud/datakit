{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Tracing Python Applications

## Install Essential Python Packages

Install Flask For Python

```shell
pip install flask
```

Install DDTrace Library for Python

```shell
pip install ddtrace
```

**Note**: This command requires pip version 18.0.0 or greater. For Ubuntu, Debian, or another package manager, update your pip version with the following command:

```shell
pip install --upgrade pip
```

## Python Code Example

service_a.py

```python
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

service_b.py

```python
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

## Run Python Code With DDTrace

这里以 Python 中常用的 Webserver Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

Run SERVICE_A

```shell
DD_SERVICE=SERVICE_A
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql
DD_TAGS=project:your_project_name,env=test,version=v1
DD_AGENT_HOST=[datakit-listening-host,default:localhost]
DD_AGENT_PORT=[datakit-listening-port,default:9529]
ddtrace-run python3 service_a.py &> a.log &
```

Run SERVICE_B

```shell
DD_SERVICE=SERVICE_B
DD_SERVICE_MAPPING=postgres:postgresql,defaultdb:postgresql
DD_TAGS=project:your_project_name,env=test,version=v1
DD_AGENT_HOST=[datakit-listening-host,default:localhost]
DD_AGENT_PORT=[datakit-listening-port,default:9529]
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
