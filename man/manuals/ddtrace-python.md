{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

## Python Flask 完整示例

这里以 Python 中常用的 Webserver Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

```shell
# 安装 flask 包
pip install flask

# 安装 ddtrace python 包以及 ddtrace-run 工具
pip install ddtrace
```

以下是示例代码。

```python
# -*- encoding: utf8 -*-
#--------- service_a.py ----------

from flask import Flask, request
import requests, os
from ddtrace import tracer

# 设置服务名
os.environ["DD_SERVICE"] = "SERVICE_A"

# 设置服务名映射关系
os.environ["DD_SERVICE_MAPPING"] = "postgres:postgresql,defaultdb:postgresql"

# 通过环境变量设置项目名，环境名，版本号
os.environ["DD_TAGS"] = "project:your_project_name,env=test,version=v1"

# 配置 DataKit trace API 服务地址
tracer.configure(
    hostname = "localhost",  # 视具体 DataKit 地址而定
    port     = "9529",
)

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

```python
# -*- encoding: utf8 -*-

#--------- service_b.py ----------

from flask import Flask, request
import os, time, requests
from ddtrace import tracer

# 设置服务名
os.environ["DD_SERVICE"] = "SERVICE_B"

# 设置服务名映射关系
os.environ["DD_SERVICE_MAPPING"] = "postgres:postgresql,defaultdb:postgresql"

# 通过环境变量设置项目名，环境名，版本号
os.environ["DD_TAGS"] = "project:your_project_name,env=test,version=v1"

tracer.configure(
    hostname = "localhost",  # 视具体 DataKit 地址而定
    port="9529",
)

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

执行以下命令来验证：

```shell
# 分别后台启动两个服务：
(ddtrace-run python3 service_a.py &> a.log &)
(ddtrace-run python3 service_b.py &> b.log &)

# 调用 A 服务，促使其调用 B 服务，这样就能产生对应 trace 数据（此处可多次执行触发）
curl http://localhost:54321/a

# 分别停止两个服务
curl http://localhost:54321/stop
curl http://localhost:54322/stop
```

可以通过 [DQL](https://www.yuque.com/dataflux/doc/fsnd2r) 验证上报的数据：

```python

> T::SERVICE_A limit
-----------------[ 1.SERVICE_A ]-----------------
parent_id '14606556292855197324'
 resource 'flask.process_response'
 trace_id '3967842463447887098'
     time 2021-04-28 15:24:11 +0800 CST
span_type 'exit'
     type 'custom'
 duration 35
     host 'testing.server'
  service 'SERVICE_A'
   source 'ddtrace'
  span_id '11450815711192661499'
    start 1619594651033484
   status 'ok'
  __docid 'T_c24gr8edtv6gq5cghnvg'
  message '{"name":"flask.process_response","service":"SERVICE_A","resource":"flask.process_response","type":"","start":1619594651033484000,"duration":35000,"span_id":11450815711192661499,"trace_id":3967842463447887098,"parent_id":14606556292855197324,"error":0}'
operation 'flask.process_response'
---------
1 rows, cost 3ms
```
