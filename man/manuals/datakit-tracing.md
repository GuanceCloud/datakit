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
  ## keys in [inputs.tracer.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
  customer_tags = ["key1", "key2", ...]

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  omit_err_status = ["404"]

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  [inputs.tracer.close_resource]
    service1 = ["resource1", "resource2", ...]
    service2 = ["resource1", "resource2", ...]
    "*" = ["close_resource_under_all_services"]

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  [inputs.tracer.sampler]
    sampling_rate = 1.0

  [inputs.tracer.tags]
    key1 = "value1"
    key2 = "value2"
```

- customer_tags: 默认情况下 Datakit 只拾取自己感兴趣的 tags，如果用户对链路上报的 tag 感兴趣可以在这项配置中进行配置。此项配置中的优先级低于
  inputs.tracer.tags 中的配置
- keep_rare_resource: 如果来自某个 resource 的链路在最近 1 小时内没有出现过，那么系统认为此条链路为稀有链路并直接上报到 Data Center。
- omit_err_status: 默认情况下如果链路中存在 error 状态的 span 那么数据会被直接上报到 Data Center，如果用户想忽略某些 HTTP error status 的链路可以配置此项。
- \[inputs.tracer.close_resource\]: 用户可以通过配置此项来关闭某些 resource 上报链路。
- \[inputs.tracer.sampler\]: 配置当前 Datakit 的全局采样率。
- \[inputs.tracer.tags\]: 配置 Datakit 默认 tags，此项配置将会覆盖 span 中 meta，metrics，tags 中重名的 key。

## Datakit Tracing Backend

Datakit backend 负责按照配置来操作链路数据，目前支持的操作包括 Tracing Filters 和 Samplers。

### Datakit Filters

- user_rule_filter: Datakit 默认 filter，用户行为触发。
- omit_status_code_filter: 当配置了 omit_err_status = ["404"]，那么 HTTP 服务下的链路中如果包含状态码为 404 的错误将不会被上报到 Data Center。
- penetrate_error_filter: Datakit 默认 filter，链路错误触发。
- close_resource_filter: 在\[inputs.tracer.close_resource\]中进行配置，服务名为服务全称或\*，资源名为资源的正则表达式。
  - 例一: 配置如 login_server = \["^auth\_.\*\\?id=\[0-9\]\*"\]，那么 login_server 服务名下 resource 形如 auth_name?id=123 的链路将被关闭
  - 例二: 配置如 "\*" = \["heart_beat"\]，那么当前 Datakit 下的所有服务上的 heart_beat 资源将被关闭。
- keep_rare_resource_filter: 当配置了 keep_rare_resource = true，那么稀有链路将会被直接上报到 Data Center。

当前的 Datakit 版本中的 Filters (Sampler 也是一种 Filter)执行顺序是固定的：

> error status penetration --> close resource filter --> omit certain http status code list --> rare resource keeper --> sampler <br>
> 每个 Datakit Filter 都具备终止执行链路的能力，即符合终止条件的 Filter 将不会在执行后续的 Filter。

### Datakit Samplers

目前 Datakit 尊重客户端的采样优先级配置，例如 ddtrace 的 priority rule tags。

> 情况一:<br>
> 以 ddtrace 为例如果 ddtrace lib sdk 或 client 中配置了 sampling priority tags 并通过环境变量(DD_TRACE_SAMPLE_RATE)或启动参数(dd.trace.sample.rate)配置了客户端采样率为 0.3 并没有指定 Datakit 采样率(inputs.tracer.sampler) 那么上报到 Data Center 中的数据量大概为总量的 30%。

> 情况二:<br>
> 如果客户只配置了 Datakit 采样率(inputs.tracer.sampler)，例如: sampling_rate = 0.3，那么此 Datakit 上报到 Data Center 的数据量大概为总量的 30%。
>
> **Note** 在多服务多 Datakit 分布式部署情况下配置 Datakit 采样率需要统一配置成同一个采样率才能达到采样效果。

> 情况三:<br>
> 即配置了客户端采样率为 A 又配置了 Datakit 采样率为 B，这里 A，B 大于 0 且小于 1，这种情况下上报到 Data Center 的数据量大概为总量的 A\*B%。
>
> **Note** 在多服务多 Datakit 分布式部署情况下配置 Datakit 采样率需要统一配置成同一个采样率才能达到采样效果。

> 关于多服务多 Datakit 采样：<br>
> A-Service(0.3) --> B-Service(0.3) --> C-Service(0.3) 配置正确，最终采样率为 30%。<br>
> A-Service(0.1) --> B-Service(0.3) --> C-Service(0.1) 配置错误，链路不能正常工作。

## About Datakit Tracing In Production（Q&A）

- 关于 Datakit Tracing 数据结构详细说明请参考 [Datakit Tracing Structure](datakit-tracing-struct)。
- 多个 Datakit Span 数据被放在 Datakit Trace 组成一条 Tracing 数据上传到 Data Center 并保证所有 Span 有且只有一个 TraceID。
- 对于 ddtrace 来说同一个 TraceID 的 ddtrace 数据有可能被分批上报。
- 生产环境下(多服务，多 Datakit 部署)一条完整的 Trace 数据是被分批次上传到 Data Center 的并不是按照调用先后顺序上传到 Data Center。
- parent_id = 0 为 root span。
- span_type = entry 为 service 上的首个 resource 的调用者即当前 service 上的第一个 span。
- 需要通过 Pipeline 脚本操作数据详细说明请参考 [Datakit Tracing With Pipeline](datakit-tracing-pl)
