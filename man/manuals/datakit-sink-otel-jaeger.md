# OpenTelemetry and Jaeger Sink

OpenTelemetry(OTEL) 提供了多种 Export 将链路数据发送到多个采集终端中，例如：Jaeger、otlp、zipkin、prometheus。

本篇介绍如何使用 sink 将链路数据发送到 otel-collector 和 Jaeger 中。

## 1.通过配置文件指定 sink 类型


### 将链路数据发送到 otel-collector 中

1. 修改配置 datakit 配置文件

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

2. 修改 sink 配置，注意 如果从没有配置过 sink 相关，新增一个配置项即可

otel 有两种 export：http 和 grpc。只能选择其中的一种。

http配置：

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost"
    port = "8889"
    path = "/api/traces"
    categories = ["T"] # 目前仅支持 Trace 类型
    target = "otel"
```


grpc 配置：

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "grpc"
    host = "localhost"
    port = "4317"
    #使用 grpc 协议时，不需要 path
    path = ""
    # 目前仅支持 Trace 类型  
    categories = ["T"] 
    target = "otel"
```

### 将链路数据发送到 Jaeger

Sink Jaeger 支持将链路数据发送到 `jaeger.colletcor` 和 `jaeger.agent`.支持使用 `HTTP` 和 `gRPC` 两种协议。

Colletcor 端口整理

- 14267 tcp agent发送jaeger.thrift格式数据
- 14250 tcp agent发送proto格式数据（背后gRPC)
- 14268 http 直接接受客户端数据(datakit 使用 HTTP 发送到 collector)
- 14269 http 健康检查

Agent 端口整理

- 5775 UDP协议，接收兼容zipkin的协议数据
- 6831 UDP协议，接收兼容jaeger的兼容协议(datakit 使用gRPC 将链路数据发送)
- 6832 UDP协议，接收jaeger的二进制协议
- 5778 HTTP协议，数据量大不建议使用

datakit config 文件配置示例：

打开配置文件 

``` shell 
vim /usr/local/datakit/conf/datakit.conf
```

HTTP 配置

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "http"
    host = "localhost"
    port = "14268"
    path = "/api/traces"
    # 目前仅支持 Trace 类型，故使用"T"
    categories = ["T"] 
    target = "jaeger"
```

grpc 配置：

``` toml
[sinks]
  [[sinks.sink]]
    scheme = "grpc"
    host = "localhost"
    port = "6831"
    #使用 grpc 协议时，不需要 path
    path = ""
    # 目前仅支持 Trace 类型  
    categories = ["T"] 
    target = "jaeger"
```

配置完成之后重启 datakit

---

## 2.安装阶段通过环境变量形式指定 Sink 

```shell
## jaeger-collector
DK_SINK_T="jaeger://localhost?scheme=http&port=14268" \
DK_DATAWAY="https://openway.guance.com?token=<YOUR-TOKEN>" \
bash -c "$(curl -L https://static.guance.com/datakit/community/install.sh)"
```

通过环境变量安装的 Datakit，会在自动在配置文件中生成相应的配置，在之后的服务重启时会以配置文件为准。
