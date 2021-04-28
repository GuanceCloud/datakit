{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

接收符合 ddtrace 协议格式的链路数据，并把数据经过统一转换成 DataFlux 的链路数据后上报到 DataFlux 中。

## 前置条件

- 部署对应语言的 [Datadog Tracing](https://docs.datadoghq.com/tracing/)工具软件，并配置数据上报地址为 DataKit 提供的地址

## 配置

1. 进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

<!-- 第三步：在应用项目中配置数据上报地址为 DataKit 的链路数据接收地址并开启应用 -->

2. 编辑 `conf.d/datakit.conf`，将 `http_listen` 改为 `0.0.0.0:9529`（此处目的是开放外网访问，端口可选）。此时 ddtrace 的访问地址就是 `http://<datakit-ip>:9529/v0.4/traces`。如果 trace 数据来源就是 DataKit 本机，可不用修改 `http_listen` 配置，直接使用 `http://localhost:9529/v0.4/traces` 即可。

## 配置并开启应用

通过 ddtarce 采集数据需要根据当前项目开发语言参考对应帮助文档 [Datadog Tracing](https://github.com/DataDog)。

### Python 快速入门

第一步，安装相关依赖包

```shell
$ pip install ddtrace
```

第二步，在应用初始化时设置上报地址

```python
# ---------- app.py ----------

import os
from ddtrace import tracer

# 通过环境变量设置服务名
os.environ["DD_SERVICE"] = "your_service_name"

# 通过环境变量设置项目名，环境名，版本号
os.environ["DD_TAGS"] = "project:your_project_name,env=test,version=v1"

# 设置链路数据 datakit 接收地址，
tracer.configure(
    hostname="127.0.0.1",  # datakit IP 地址
    port="9529",           # datakit http 服务端口号
)
``` 

第三步，开启应用

```shell
$ ddtrace-run python app.py
``` 

若通过 `gunicorn` 运行，需要在应用初始化时进行如下配置，否则会产生相同的 `traceID`

```python
patch(gevent=True)
```

其他语言应用与此类似，配置成功后约 1~2 分钟即可在 DataFlux Studio 「应用性能监测」中查看数据。

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过命令行注入环境变量

```shell
$ DD_TAGS="project:your_project_name,env=test,version=v1 ddtrace-run python app.py"
```

> Tips：若需要链路数据和容器对象关联，可按照如下方式开启应用（一般情况下就是修改容器中的启动命令 `CMD`）。这里的 `$HOSTNAME` 环境变量会自动替换成对应容器中的 hostname：

```shell
$ DD_TAGS="container_host:$HOSTNAME,other_tag:other_tag_val" ddtrace-run python your_app.py
``` 

- 通过采集器自定义标签设置

```toml
[inputs.ddtrace]
	path = "/v0.4/traces"             # ddtrace 链路数据接收路径，默认与ddtrace官方定义的路径相同
	[inputs.ddtrace.tags]             # 自定义标签组
		project = "your_project_name"   # 设置项目名
		env     = "your_env_name"       # 设置环境名
		version = "your_version"        # 设置版本信息
```

## Python Flask 完整示例

这里以 Python 中常用的 Webserver Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

```shell
# 先确保安装 flask 包
$ pip install flask
```

```python
# -*- encoding: utf8 -*-
#--------- service_a.py ----------

from flask import Flask, request
import requests, os
from ddtrace import tracer

# 设置服务名
os.environ["DD_SERVICE"] = "SERVICE_A"

# 配置 DataKit trace API 服务地址
tracer.configure(
    hostname = "localhost",  # 视具体地址而定
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

os.environ["DD_SERVICE"] = "SERVICE_B"

tracer.configure(
    hostname="localhost",
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
$ (ddtrace-run python3 service_a.py &> a.log &)
$ (ddtrace-run python3 service_b.py &> b.log &)

# 调用 A 服务，促使其调用 B 服务，这样就能产生一条 trace 数据（此处可多次尝试）
$ curl http://localhost:54321/a

# 分别停止两个服务
$ curl http://localhost:54321/stop
$ curl http://localhost:54322/stop
```

也可以通过 [DQL](http://doc/to/dql) 查找数据：

```python
T::xxxx {service="SERVICE_A"} order by time desc limit 1

# 贴出 DQL 返回的数据
```

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
