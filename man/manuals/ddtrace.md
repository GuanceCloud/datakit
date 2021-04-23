{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

接收符合 ddtrace 协议格式的链路数据，并把数据经过统一转换成 DataFlux 的链路数据后上报到 DataFlux 中。

## 前置条件

- 部署对应语言的 [Datadog Tracing](https://docs.datadoghq.com/tracing/)工具软件，并配置数据上报地址为 DataKit 提供的地址

## 配置

ddtrace 的数据发送到 DataKit一共需要三步：

第一步：在 DataKit 中开启链路数据接收服务

第二步：配置和获取链路数据监听的地址和端口

第三步：在应用项目中配置数据上报地址为 DataKit 的链路数据接收地址并开启应用

### 开启链路数据接收服务

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

### 配置和链路数据监听的地址和端口

打开 DataKit 安装目录下的 `datakit.conf`，找到 `http_listen` 配置项，配置数据接收的监听地址和端口，默认监听地址是 `0.0.0.0`，端口为 `9529`
完成以上配置后，即可获取：ddtrace 链路数据的接收地址`HTTP协议//:绑定地址:链路端口/v0.4/traces`。例如 DataKit 的地址是 `1.2.3.4`，配置的监听地址和端口是 `0.0.0.0:9529`，ddtrace 链路数据的接收地址是： `http://1.2.3.4:9529/v0.4/traces`，如果是本机也可以使用 `localhost` 或 `127.0.0.1`，如果是内网也可以使用内网地址。

**注意：**

- ddtrace 请求的默认路径是 `/v0.4/traces`
- 需保证数据采集端能访问该地址

### 配置并开启应用

通过 ddtarce 采集数据需要根据当前项目开发语言参考对应帮助文档 [Datadog Tracing](https://github.com/DataDog)。

这里以 Python 应用作为示范

第一步，安装相关依赖包

```shell
pip install ddtrace
```

第二步，在应用初始化时设置上报地址

```python
import os
from ddtrace import tracer

#通过环境变量设置服务名
os.environ["DD_SERVICE"] = "your_service_name"

#通过环境变量设置项目名，环境名，版本号
os.environ["DD_TAGS"] = "project:your_project_name,env=test,version=v1"

#设置链路数据datakit接收地址，
tracer.configure(
    # datakit IP 地址
    hostname="127.0.0.1",

    # datakit http 服务端口号
    port="9529",
)
``` 

第三步，开启应用

```shell
ddtrace-run python your_app.py
``` 

若需要链路数据和容器对象关联，需按照如下方式开启应用

```shell
DD_TAGS=container_host:$HOSTNAME,other_tag:other_tag_val ddtrace-run python your_app.py
``` 

若通过 `gunicorn` 运行，需要在应用初始化时进行如下配置，否则会产生相同的 `traceID`

```python
patch(gevent=True)
```


其他语言应用与此类似，配置成功后约 1-2 分钟即可在 DataFlux Studio 的 「链路追踪」中查看相关的链路数据。

除了在应用初始化时设置项目名，环境名以及版本号外，还可通过如下两种方式设置：

- 通过环境变量设置

```shell
export DD_TAGS="project:your_project_name,env=test,version=v1"
```

- 通过采集器自定义标签设置

```toml
[inputs.ddtrace]
       path = "/v0.4/traces"                   # ddtrace 链路数据接收路径，默认与ddtrace官方定义的路径相同
       [inputs.ddtrace.tags]                   # 自定义标签组
               project = "your_project_name"   # 设置项目名
               env  = "your_env_name"          # 设置环境名
               version  = "your_version"       # 设置版本信息
```

## 示例代码

以 Python 语言作为示例代码，其它编程语言与此类似，示例中`SERVICE_A`提供 HTTP 服务，并且调用`SERVICE_B` HTTP服务。

### SERVICE_A

```python
from flask import Flask
import requests, os
from ddtrace import tracer

os.environ["DD_SERVICE"] = "SERVICE_A"

tracer.configure(
    hostname="127.0.0.1",
    port="12345",
)


app = Flask(__name__)

@app.route('/a',  methods=['GET'])
def index():
    requests.get('http://127.0.0.1:10002/b')
    return 'OK', 200


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=10001, debug=True)
```

### SERVICE_B

```python
import os, time, requests
from flask import Flask
from ddtrace import tracer

os.environ["DD_SERVICE"] = "SERVICE_B"

tracer.configure(
    hostname="127.0.0.1",
    port="12345",
)

app = Flask(__name__)

@app.route('/b',  methods=['GET'])
def index():
    time.sleep(1)
    return 'OK', 200


if __name__ == '__main__':
    app.run(host="0.0.0.0", port=10002, debug=True)
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



