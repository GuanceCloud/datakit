# 多链路中的信息传播机制及使用

本篇文章的主要介绍的是多个链路厂商的产品，以及在分布式服务中多个语言或者多个产品之间如何实现 Trace 信息传播。

透传协议也被称为传播协议，是指在服务请求和响应中添加特定的头部信息（一般指 HTTP 头）来实现。当一个服务请求另一个服务时，它会携带特定请求头。当下一跳收到请求时，从请求头中获取特定的链路信息并继承之，继续向后传播，直到链路结束。这样可以将整个调用链关联起来。

## 常见的的传播协议 {#propagators}

以下是对这些透传协议在 HTTP 头部上的区别的简要介绍：

### Trace Context {#propagators-w3c}

Trace Context 是 [W3C](https://www.w3.org/TR/trace-context/){:target="_blank"} 标准化的跟踪协议，它定义了两个 HTTP 头部字段：`traceparent` 和 `tracestate`:

- `traceparent` 包含了关于当前跟踪的基本信息，如 SpanID 和 ParentSpanID 等，例如：`traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01`
- `tracestate` 用于传递与跟踪相关的元数据。例如：`tracestate: congo=t61rcWkgMzE`

### B3/B3Multi {#propagators-b3}

B3 是一种流行的跟踪协议，它定义了多个 HTTP 头部字段来标识跟踪信息。B3Multi 透传协议是对 B3 协议的扩展，常用的字段有：`X-B3-TraceId`、`X-B3-SpanId`、`X-B3-ParentSpanId`、`X-B3-Sampled`、`X-B3-Flags` 等。

### Jaeger {#propagators-jaeger}

Jaeger 是一种分布式跟踪系统，它定义了多个 HTTP 头部字段用于传递跟踪信息。常用的字段有：`uber-trace-id`、`jaeger-baggage` 等。

### OpenTracing {#propagators-ot}

OpenTracing 是 OpenTelemetry 的一种透传协议，它定义了多个 HTTP 头部字段用于传递链路信息：

- `ot-tracer-traceid`：用于传递链路 ID，表示一个完整的请求链路
- `ot-tracer-spanid`：用于传递当前 Span 的 ID，表示一个单独的操作或事件
- `ot-tracer-sampled`：用于指示是否对该请求进行采样，以决定是否记录请求的追踪信息

### Datadog {#propagators-datadog}

Datadog 是一种分布式跟踪系统，它定义了多个 HTTP 头部字段用于传递跟踪信息。常用的字段有：`x-datadog-trace-id`、`x-datadog-parent-id` 等。

### Baggage {#propagators-baggage}

Baggage 是 Jaeger 跟踪系统引入的概念，用于传递业务相关的上下文信息。Baggage 通过 HTTP 头部字段 `x-b3-baggage-<key>` 来传递，其中 `key` 是业务上下文的键。

Baggage 真正的意义是传播 `key:value` 性质的键值对，常用于传播 AppID、Host-Name、Host-IP 等。

<!-- markdownlint-disable MD046 -->
???+ attention

    需要注意的是，这些透传协议的具体实现和使用方式可能略有不同，但它们都旨在通过 HTTP 头部字段在不同的服务之间传递跟踪信息和上下文信息，以实现分布式跟踪和连续性。
<!-- markdownlint-enable -->

## 链路厂商及产品介绍 {#tracing-info}

产品及厂商：

| 产品          | 厂商              | 支持的语言                                                                  |
| :---          | :---              | :---                                                                        |
| OpenTelemetry | CNCF              | Java, Python, Go, JavaScript, .NET, Ruby, PHP, Erlang, Swift, Rust, C++ 等  |
| DDTrace       | Datadog           | Java, Python, Go, Ruby, JavaScript, PHP, .NET, Scala, Objective-C, Swift 等 |
| SkyWalking    | Apache SkyWalking | Java, .NET, Node.js, PHP, Python, Go, Ruby, Lua, OAP 等                     |
| Zipkin        | OpenZipkin        | Java, Node.js, Ruby, Go, Scala, Python 等                                   |
| Jaeger        | CNCF              | Java, Python, Go, C++, C#, Node.js 等                                       |

产品的开源地址：

- [OpenTelemetry](https://github.com/open-telemetry){:target="_blank"} 是 CNCF 下的一个产品。同时观测云也对其[做了扩展](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
- [Jaeger](https://github.com/jaegertracing/jaeger){:target="_blank"} 同属于 CNCF
- [Datadog](https://github.com/DataDog){:target="_blank"} 多语言的链路工具，其中观测云对其[做了扩展](https://github.com/GuanceCloud/dd-trace-java){:target="_blank"}
- [SkyWalking](https://github.com/apache?q=skywalking&type=all&language=&sort=){:target="_blank"} 属于 Apache 基金会下的开源产品
- [Zipkin](https://github.com/OpenZipkin){:target="_blank"} 其中有多个语言的链路工具。

## 产品的透传协议 {#use-propagators}

### OpenTelemetry {#use-otel}

OTEL 所支持的 Tracing 透传协议列表：

| Propagator 列表  | 参考                                                                                                                           |
| ---              | ---                                                                                                                            |
| `tracecontext` | [W3C Trace Context](https://www.w3.org/TR/trace-context/){:target="_blank"}                                                    |
| `baggage`      | [W3C Baggage](https://www.w3.org/TR/baggage/){:target="_blank"}                                                                |
| `b3`           | [B3](https://github.com/openzipkin/b3-propagation#single-header){:target="_blank"}                                             |
| `b3multi`      | [B3Multi](https://github.com/openzipkin/b3-propagation#multiple-headers){:target="_blank"}                                     |
| `jaeger`       | [Jaeger](https://www.jaegertracing.io/docs/1.21/client-libraries/#propagation-format){:target="_blank"}                        |
| `xray`         | [AWS X-Ray](https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html#xray-concepts-tracingheader){:target="_blank"} |
| `opentracing`  | [OpenTracing](https://github.com/opentracing?q=basic&type=&language=){:target="_blank"}                                        |

分布式链路头部信息在透传中的格式示例：

```shell
# 命令行注入示例（多个传播协议使用逗号隔开）
-Dotel.propagators="tracecontext,baggage"

# 环境变量注入示例（Linux）
export OTEL_PROPAGATORS="tracecontext,baggage"

# 环境变量注入示例（Windows）
$env:OTEL_PROPAGATORS="tracecontext,baggage"
```

### Datadog {#use-datadog}

| 支持的语言 | 透传协议支持                           | 命令                                                     |
| :---       | :---                                   | :---                                                     |
| Node.js    | `datadog/b3multi/tracecontext/b3/none` | `DD_TRACE_PROPAGATION_STYLE`(默认 `datadog`)             |
| C++        | `datadog/b3multi/b3/none`              | `DD_TRACE_PROPAGATION_STYLE`(默认 `datadog`)             |
| .NET       | `datadog/b3multi/tracecontext/none`    | `DD_TRACE_PROPAGATION_STYLE`(默认 `datadog`)             |
| Java       | `datadog/b3multi/tracecontext/none`    | `DD_TRACE_PROPAGATION_STYLE`(默认 `tracecontext,datadog`)|

> 此处 `none` 指不设置 Tracing 协议透传。

#### DD_TRACE_PROPAGATION_STYLE {#dd-pg-style}

Datadog Tracing 在协议透传行为上可以做出入站设置，即是否继承上游协议、是否将自己的协议透传给下游。通过如下两个环境变量分别控制：

- 入站控制：`export DD_TRACE_PROPAGATION_STYLE_EXTRACT=<XXX>`
- 出站控制：`export DD_TRACE_PROPAGATION_STYLE_INJECT=<YYY>`
- 也可以通过一个单独的 ENV 来同时控制出入站：`export DD_TRACE_PROPAGATION_STYLE="tracecontext,datadog"`

示例：

```shell
# 入站会继承 X-Datadog-* 和 X-B3-* 的请求头（如果有），
# 出站时会带上 X-Datadog-* 和 X-B3-* 请求头
$ export DD_TRACE_PROPAGATION_STYLE="datadog,b3" ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    在版本 V1.7.0 之后，默认的支持协议改为 `DD_TRACE_PROPAGATION_STYLE="tracecontext,datadog"`，B3 已被弃用，请使用 B3multi。
<!-- markdownlint-enable -->

更多语言示例，参见[这里](https://github.com/DataDog/documentation/blob/4ff75ed0bcaa1269bf98e9d185935cfda675b08c/content/en/tracing/trace_collection/trace_context_propagation/_index.md){:target="_blank"}。

### SkyWalking {#use-sw8}

SkyWalking 自己的[协议（SW8）](https://skywalking.apache.org/docs/main/next/en/api/x-process-propagation-headers-v3/){:target="_blank"}

### Zipkin {#use-zipkin}

[参见这里](https://github.com/openzipkin/b3-propagation){:target="_blank"}

### Jaeger {#use-jaeger}

所有支持的协议：

- [Jaeger Propagation Format](https://www.jaegertracing.io/docs/1.21/client-libraries/#propagation-format){:target="_blank"}
- [B3 propagation](https://github.com/openzipkin/b3-propagation){:target="_blank"}
- W3C Trace-Context

## 多链路串联 {#series}

请求 Header 与厂商支持列表：

|               | W3C                         | b3multi                  | Jaeger                   | OpenTracing              | Datadog                  | sw8                      |
| :---          | :---                        | :---                     | :---                     | :---                     | :---                     | :---                     |
| header        | `tracecontext`/`tracestate` | `X-B3-*`                 | `uber-trace-id`          | `ot-tracer-*`            | `x-datadog-*`            | `xxx-xxx-xxx-xxx`        |
| OpenTelemetry | :heavy_check_mark:          | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_multiplication_x: |
| Datadog       | :heavy_check_mark:          | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_check_mark:       | :heavy_multiplication_x: |
| SkyWalking    | :heavy_multiplication_x:    | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_check_mark:       |
| Zipkin        | :heavy_multiplication_x:    | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: |
| Jaeger        | :heavy_check_mark:          | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: |

可根据具体使用的厂商工具使用相应的透传协议实现链路串联，保证链路完整性。

### 串联示例 {#dd-otel-example}

这里用一个示例说明下 DDTrace 和 OpenTelemetry 链路数据串联。由上表可知：DDTrace 和 OpenTelemetry 都是支持 W3C Trace Context 协议，可以通过该协议实现链路串联。

- DDTrace 中的 TraceID 是 64 位的 int 字符串，SpanID 和 ParentID 也是 64 位的 int
- OTEL 中的 TraceID 是 128 位的 16 进制表示的 int 字符串，SpanID 和 ParentID 是 64 位的 int 类型字符串

两者要想关联 TraceID 需要将 DDTrace 提升到 128 位。

无论哪一个作为请求的发起方，DDTrace 都需要开启 128bit TraceID 支持（`dd.trace.128.bit.traceid.generation.enabled`）：

```shell
# DDTrace 启动示例
$ java -javaagent:/usr/local/ddtrace/dd-java-agent.jar \
  -Ddd.service.name=client \
  -Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext \
  -jar springboot-client.jar

# OTEL 启动示例
$ java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent.jar \
  -Dotel.service.name=server \
  -jar springboot-server.jar
```

Client 端会发送 HTTP 请求到 Server 端，DDTrace 会通过 `tracecontext` 请求头部中携带链路信息传递到服务端上

但是，在「服务调用关系」中两个工具上来的数据连接不上，这是因为双方的 SpanID 并不是统一的，DDTrace 是一个 10 进制的数字字符串，而 OpenTelemetry 是 16 进制的数字字符串。为此，需要修改 `ddtrace` 采集器中的配置，将 `ddtrace.conf` 中的 `compatible_otel` 放开：

```toml
  ## compatible otel: It is possible to compatible OTEL Trace with DDTrace trace.
  ## make span_id and parent_id to hex encoding.
  compatible_otel=true
```

将 `compatible_otel=true` 之后所有的 DDTrace 的 `span_id` 和 `parent_id` 都会变成 16 进制的数字字符串。

<!-- markdownlint-disable MD046 -->
???+ tip "日志中的 `span_id` 转 16 进制"

    在日志中 DDTrace 中的 SpanId 还是 10 进制，需要在采集日志的 Pipeline 脚本中将 `span_id` 提取之后转成 16 进制的数字字符串（不会修改原始的日志文本）：

    ```python
    # 将字符串转 int64
    fn parse_int(val: str, base: int) int64

    # 将 int64 转 string
    fn format_int(val: int64, base: int) str
    ```
<!-- markdownlint-enable -->

至此， DDTrace 和 OTEL 在链路上实现了串联，服务调用关系和日志也能串联：

<!-- markdownlint-disable MD046 MD033 -->
<figure >
  <img src="https://github.com/GuanceCloud/dd-trace-java/assets/31207055/9b599678-1ebc-4f1f-9993-f863fb25280b" style="height: 600px" alt="链路详情">
  <figcaption> 链路详情 </figcaption>
</figure>
<!-- markdownlint-enable -->
