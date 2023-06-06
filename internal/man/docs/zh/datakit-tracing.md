
# Datakit Tracing 综述

目前 Datakit 支持的第三方 Tracing 数据包括：

- DDTrace
- Apache Jaeger
- OpenTelemetry
- SkyWalking
- Zipkin

---

## Datakit Tracing Frontend {#datakit-tracing-frontend}

Tracing Frontend 即接收各种不同类 Trace 数据的 API，它们一般通过 HTTP 或 gRPC 等方式接收各种 Trace SDK 发送过来的数据。DataKit 收到这些数据后，会将它们转换成[统一的 Span 结构](datakit-tracing-struct.md)。然后再发送到 [Backend](datakit-tracing.md#datakit-tracing-backend) 处理。

除了转换 Span 结构外，Tracing Frontend 还会完成对[Tracing Backend](datakit-tracing.md#datakit-tracing-backend)中过滤单元和运算单元的配置

## Tracing 数据采集通用配置 {#tracing-common-config}

配置文件中的 tracer 代指当前配置的 Tracing Agent，所有已支持的 Tracing Agent，均可以使用如下配置：

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

  ## Threads config controls how many goroutines an agent cloud start.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  ## timeout is the duration(ms) before a job can return a result.
  [inputs.tracer.threads]
    buffer = 100
    threads = 8
    timeout = 1000
```

- `customer_tags`: 默认情况下 Datakit 只拾取自己感兴趣的 Tags（即观测云链路详情里可以看到的除 message 外的字段），

  如果用户对链路上报的其他 tag 感兴趣可以在这项配置添加告知 Datakit 去拾取。此项配置优先级高于 `[inputs.tracer.tags]`。

- `keep_rare_resource`: 如果来自某个 Resource 的链路在最近一小时内没有出现过，那么系统认为此条链路为稀有链路并直接上报到 Data Center。
- `omit_err_status`: 默认情况下如果链路中存在 Error 状态的 Span 那么数据会被直接上报到 Data Center，如果用户需要忽略某些 HTTP Error Status（例如：429 too many requests） 的链路可以通过配置此项告知 Datakit 忽略。
- `[inputs.tracer.close_resource]`: 用户可以通过配置此项来关闭 [span_type](datakit-tracing-struct.md) 为 Entry 的 Resource 链路。
- `[inputs.tracer.sampler]`: 配置当前 Datakit 的全局采样率，[配置示例](datakit-tracing.md#samplers)。
- `[inputs.tracer.tags]`: 配置 Datakit Global Tags，优先级低于 `customer_tags` 。
- `[inputs.tracer.threads]`: 配置当前 Tracing Agent 的线程队列用来控制处理数据过程中能使用的 CPU 和 Memory 资源。
    - buffer: 工作队列的缓存，配置越大那么内存消耗越大同时发送到 Agent 上的请求能更大概率入队成功并快速返回否则将被丢弃并返回 429 错误。
    - threads: 工作队列的最大线程数，配置越大启动的线程越多 CPU 占用越高，一般情况下配置成 CPU 的核心数。
    - timeout: 任务超时，配置越大占用 buffer 的时间越长。

## Datakit Tracing Backend {#datakit-tracing-backend}

Datakit backend 负责按照配置来操作链路数据，目前支持的操作包括 Tracing Filters 和 Samplers。

### Datakit Filters {#filters}

- `user_rule_filter`: Datakit 默认 filter，用户行为触发。
- `omit_status_code_filter`: 当配置了 `omit_err_status = ["404"]`，那么 HTTP 服务下的链路中如果包含状态码为 404 的错误将不会被上报到 Data Center。
- `penetrate_error_filter`: Datakit 默认 filter，链路错误触发。
- `close_resource_filter`: 在 `[inputs.tracer.close_resource]` 中进行配置，服务名为服务全称或 `*`，资源名为资源的正则表达式。
    - 例一：配置如 `login_server = ["^auth\_.*\?id=[0-9]*"]`，那么 `login_server` 服务名下 `resource` 形如 `auth_name?id=123` 的链路将被关闭
    - 例二：配置如 `"*" = ["heart_beat"]`，那么当前 Datakit 下的所有服务上的 `heart_beat` 资源将被关闭。
- `keep_rare_resource_filter`: 当配置了 `keep_rare_resource = true`，那么被判定为稀有的链路将会被直接上报到 Data Center。

当前的 Datakit 版本中的 Filters (Sampler 也是一种 Filter)执行顺序是固定的：

> error status penetration --> close resource filter --> omit certain http status code list --> rare resource keeper --> sampler <br>
> 每个 Datakit Filter 都具备终止执行链路的能力，即符合终止条件的 Filter 将不会在执行后续的 Filter。

### Datakit Samplers {#samplers}

目前 Datakit 尊重客户端的采样优先级配，[DDTrace Sampling Rules](https://docs.datadoghq.com/tracing/faq/trace_sampling_and_storage){:target="_blank"}。

- 情况一

以 DDTrace 为例如果 DDTrace lib sdk 或 client 中配置了 sampling priority tags 并通过环境变量(DD_TRACE_SAMPLE_RATE)或启动参数(dd.trace.sample.rate)配置了客户端采样率为 0.3 并没有指定 Datakit 采样率(inputs.tracer.sampler) 那么上报到 Data Center 中的数据量大概为总量的 30%。

- 情况二

如果客户只配置了 Datakit 采样率(inputs.tracer.sampler)，例如：sampling_rate = 0.3，那么此 Datakit 上报到 Data Center 的数据量大概为总量的 30%。

**Note** 在多服务多 Datakit 分布式部署情况下配置 Datakit 采样率需要统一配置成同一个采样率才能达到采样效果。

- 情况三

即配置了客户端采样率为 A 又配置了 Datakit 采样率为 B，这里 A，B 大于 0 且小于 1，这种情况下上报到 Data Center 的数据量大概为总量的 A\*B%。

**Note** 在多服务多 Datakit 分布式部署情况下配置 Datakit 采样率需要统一配置成同一个采样率才能达到采样效果。

## Span 结构说明 {#about-span-structure}

关于 Datakit 如何使用[DatakitSpan](datakit-tracing-struct.md)数据结构的业务解释

- 关于 Datakit Tracing 数据结构详细说明请参考 [Datakit Tracing Structure](datakit-tracing-struct.md)。
- 多个 Datakit Span 数据被放在 Datakit Trace 组成一条 Tracing 数据上传到 Data Center 并保证所有 Span 有且只有一个 TraceID。
- 对于 DDTrace 来说同一个 TraceID 的 DDTrace 数据有可能被分批上报。
- 生产环境下(多服务，多 Datakit 部署)一条完整的 Trace 数据是被分批次上传到 Data Center 的并不是按照调用先后顺序上传到 Data Center。
- `parent_id = 0` 为 root span。
- `span_type = entry` 为 service 上的首个 resource 的调用者即当前 service 上的第一个 span。
