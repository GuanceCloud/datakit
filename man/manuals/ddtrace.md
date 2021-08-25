{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

接收符合 ddtrace 协议格式的链路数据，并把数据经过统一转换成 DataFlux 的链路数据后上报到 DataFlux 中。

## 前置条件

准备对应语言的 ddtrace 配置：

- [Python](https://github.com/DataDog/dd-trace-py)
- [Golang](https://github.com/DataDog/dd-trace-go)
- [NodeJS](https://github.com/DataDog/dd-trace-js)
- [PHP](https://github.com/DataDog/dd-trace-php)
- [Ruby](https://github.com/DataDog/dd-trace-rb)
- [C#](https://github.com/DataDog/dd-trace-dotnet)
- [C++](https://github.com/opentracing/opentracing-cpp)
- Java： DataKit 安装目录 `data` 目录下，有预先准备好的 `dd-java-agent.jar`（推荐使用）。也可以直接去 [Maven 下载](https://mvnrepository.com/artifact/com.datadoghq/dd-java-agent)

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

> 注意：不要修改这里的 `endpoints` 列表。

```toml
endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
```

编辑 `conf.d/datakit.conf`，将 `listen` 改为 `0.0.0.0:9529`（此处目的是开放外网访问，端口可选）。此时 ddtrace 的访问地址就是 `http://<datakit-ip>:9529`。如果 trace 数据来源就是 DataKit 本机，可不用修改 `listen` 配置，直接使用 `http://localhost:9529` 即可。

如果有 trace 数据发送给 DataKit，那么在 DataKit 的 `gin.log` 上能看到：

```shell
tail -f /var/log/datakit/gin.log
[GIN] 2021/08/02 - 17:16:31 | 200 |     386.256µs |       127.0.0.1 | POST     "/v0.4/traces"
[GIN] 2021/08/02 - 17:17:30 | 200 |     116.109µs |       127.0.0.1 | POST     "/v0.4/traces"
[GIN] 2021/08/02 - 17:17:30 | 200 |     489.428µs |       127.0.0.1 | POST     "/v0.4/traces"
...
```

> 注意：如果没有 trace 发送过来，在 [monitor 页面](datakit-how-to#44462aae)是看不到 ddtrace 的采集信息的。

## Python Flask 完整示例

这里以 Python 中常用的 Webserver Flask 应用为例。示例中 `SERVICE_A` 提供 HTTP 服务，并且调用 `SERVICE_B` HTTP 服务。

```shell
# 先安装 flask 包
pip install flask
```

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

### ddtrace 环境变量设置

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过命令行注入环境变量

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- 在 ddtrace.conf 中直接配置自定义标签。这种方式会影响**所有**发送给 DataKit tracing 服务的数据，需慎重考虑：

```toml
## tags is ddtrace configed key value pairs
# [inputs.ddtrace.tags]
	# some_tag = "some_value"
	# more_tag = "some_other_value"
	## ...
```

<!--#### 关联 ddtrace 数据和容器对象

若需要链路数据和容器对象关联，可按照如下方式开启应用（一般情况下就是修改 Dockerfile 中的启动命令 `CMD`）。这里的 `$HOSTNAME` 环境变量会自动替换成对应容器中的主机名：

```shell
DD_TAGS="container_host:$HOSTNAME,other_tag:other_tag_val" ddtrace-run python your_app.py
```
-->

### 在代码中添加业务 tag

在应用代码中，可通过诸如 `span.SetTag(some-tag-key, some-tag-value)`（不同语言方式不同） 这样的方式来设置业务自定义 tag。对于这些业务自定义 tag，可通过配置 `customer_tags` 来识别并提取：

```toml
customer_tags = []
```

注意，这些 tag-key 中不能包含英文字符 '.'，带 `.` 的 tag-key 会忽略掉，示例：

```toml
customer_tags = [
	"order_id",
	"task_id",
	"some.invalid.key",  # 无效的 tag-key，DataKit 选择将其忽略
]
```

#### 应用代码中添加业务 tag 注意事项

- 务必在 `customer_tags` 中添加 tag-key 列表，否则 DataKit 不会进行业务 tag 的提取
- 在开启了采样的情况下，部分添加了 tag 的 span 有可能被舍弃

### 设置 trace 数据采样率

默认每次调用都会产生 trace 数据，若不加以限制，会导致采集到数据量大，占用过多的存储，网络带宽等系统资源，可以通过设置采样率解决这一问题，修改 `{{.InputName}}.conf` ：

```toml
[inputs.ddtrace.sample_config]
	## sample rate, how many will be sampled
	rate = 10
	## sample scope, the range to sample
	scope = 100
```

说明：

- 此处 `rate/scope` 即最终的采样率，示例配置即采样 10%
- 如果在 DataKit 上开启了采样率，就不要在 ddtrace 上再设置采样率，这可能导致双重采样，导致数据大面积缺失
- 对 RUM 产生的 trace，这里的采样率不生效，建议在 [RUM 中设置采样率](https://www.yuque.com/dataflux/doc/eqs7v2#16fe8486)

#### ddtrace 采样透传 tag

目前的采样会忽略带有如下 tag-value 的数据：

| key          | value |
| ------------ | ----- |
| `_dd.origin` | `rum` |

## Tracing 数据

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
