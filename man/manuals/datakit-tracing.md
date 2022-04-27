{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# Datakit Tracing Data Flow

> Third Party Tracing Agent --> Datakit Frontend --> Datakit Backend --> Data Center

## Third Party Tracing Agent

目前 Datakit 支持的第三方 Tracing 数据包括：

- DDTrace
- Apache Jaeger
- OpenTelemetry
- Skywalking
- Zipkin

## Datakit Tracing Frontend

Datakit Frontend 即 Datakit Tracing Agent 负责接收并转换第三方 Tracing Agent 数据结构，Datakit 内部使用 [DatakitSpan](datakit-tracing-struct) 数据结构。

[Datakit Tracing Frontend](#Datakit-Tracing-Frontend) 会解析接收到的 Tracing Span 数据并转换成 [DatakitSpan](datakit-tracing-struct) 后发送到 [Datakit Tracing Backend](#datakit-tracing-backend)。[Datakit Tracing Frontend](#Datakit-Tracing-Frontend) 可以完成对[Datakit Tracing Backend](#datakit-tracing-backend)中过滤单元和运算单元的配置，请参考[Datakit Tracing Common Configuration](#Datakit-Tracing-Common-Configuration)。

## Datakit Tracing Common Configuration

```toml
## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
## that want to send to data center. Those keys set by client code will take precedence over
## keys in [tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
customer_tags = ["key1", "key2", ...]

## Keep rare tracing resources list switch.
## If some resources are rare enough(not presend in 1 hour), those resource will always send
## to data center and do not consider samplers and filters.
keep_rare_resource = false

## Ignore tracing resources map like service:[resources...].
## The service name is the full service name in current application.
## The resource list is regular expressions uses to block resource names.
[close_resource]
  service1 = ["resource1", "resource2", ...]
  service2 = ["resource1", "resource2", ...]
  # ...

## Sampler config uses to set global sampling strategy.
## priority uses to set tracing data propagation level, the valid values are -1, 0, 1
##   -1: always reject any tracing data send to datakit
##    0: accept tracing data and calculate with sampling_rate
##    1: always send to data center and do not consider sampling_rate
## sampling_rate used to set global sampling rate
[sampler]
  priority = 0
  sampling_rate = 1.0

## Piplines use to manipulate message and meta data. If this item configured right then
## the current input procedure will run the scripts wrote in pipline config file against the data
## present in span message.
## The string on the left side of the equal sign must be identical to the service name that
## you try to handle.
 [inputs.ddtrace.pipelines]
  service1 = "service1.p"
  service2 = "service2.p"
  # ...
```

## Datakit Tracing Backend

Datakit Tracing Backend 包括几个部分 Tracing <!--Statistics,--> Filters 和 Samplers：

<!-- - Tracing Statistics: 统计 Tracing 链路上的业务状态，例如：访问耗时，错误率等。 -->

- Filters:
  - keep_rare_resource: 当系统监测到某些链路在一小时之内没有发送任何 Tracing 数据那么将被认定为稀有并被透穿到 Data Center。
  - close_resource: 按照正则规则关闭某些 Service 下的一个或多个 Resource。
- Samplers: 基于概率的 Tracing 数据采样。多服务环境下采样率必须配置一致才能达到采样效果，
  - 例一：A-Service(0.3) --> B-Service(0.3) --> C-Service(0.3) 配置正确，最终采样率为 30%。
  - 例二：A-Service(0.1) --> B-Service(0.3) --> C-Service(0.1) 配置错误，链路不能正常工作。
- Piplines: [Pipeline](pipeline)为 Datakit Tracing 提供通过自定义脚本进行数据操纵的能力。配置请参考安装目录下当前开启的 Tracer xxx.conf 文件，例如 ddtrace.conf：

```toml
  ## Piplines use to manipulate message and meta data. If this item configured right then
  ## the current input procedure will run the scripts wrote in pipline config file against the data
  ## present in span message.
  ## The string on the left side of the equal sign must be identical to the service name that
  ## you try to handle.
  # [inputs.ddtrace.pipelines]
    # service1 = "service1.p"
    # service2 = "service2.p"
    # ...
```

通过 Pipeline 脚本操作数据详细说明请参考[Datakit Tracing With Pipeline](datakit-tracing-pl)

### The Order of Executing Filters

当前的 Datakit 版本中的 Filters (Sampler 也是一种 Filter)的执行顺序是固定的：

> error status penetration --> close resource filter --> omit certain http status code list --> rare resource keeper --> sampler --> piplines

每个 Filter 都具备终止执行链路的能力，即符合终止条件的 Filter 将不会在执行后续的 Filter。

## About Datakit Span Struct In Production

关于 Datakit 如何使用[DatakitSpan](datakit-tracing-struct)数据结构的业务解释

- 多个 Datakit Span 数据被放在 Datakit Trace 组成一条 Tracing 数据上传到数据中心并保证所有 Span 有且只有一个 TraceID。
- 生产环境下(多服务，多 Datakit 部署)一条完整的 Trace 数据是被分批次上传到数据中心的并不是按照调用先后顺序上传到数据中心。
- parent_id = 0 为 root span。
- span_type = entry 为 service 上的首个 resource 的调用者即当前 service 上的第一个 span。
