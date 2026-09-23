---
title     : 'Tracing Propagator'
summary   : '多链路中的信息传播机制及使用'
tags      :
  - '链路追踪'
  - 'Propagator'
  - '透传协议'
__int_icon: ''
---

分布式追踪通过 Propagator 在跨进程请求中注入和提取 Trace ID、Span ID、采样状态及 Baggage。HTTP 场景通常使用请求头承载这些信息。只有调用链两端使用兼容的传播格式，Span 才能连接成完整链路。

传播格式只负责上下文传递，不负责把 Span 数据发送到 DataKit。应用仍需分别配置 DDTrace 或 OpenTelemetry 的导出地址、端口和协议。

## 常见传播格式 {#propagators}

| 格式 | 主要 HTTP Header | 说明 |
| --- | --- | --- |
| W3C Trace Context | `traceparent`、`tracestate` | 厂商中立的标准格式，适合 DDTrace 与 OpenTelemetry 混合调用链。 |
| W3C Baggage | `baggage` | 传播应用定义的键值对；可能包含敏感信息，不要发往不可信边界。 |
| B3 Single | `b3` | Zipkin B3 单 Header 格式。 |
| B3 Multi | `X-B3-TraceId`、`X-B3-SpanId`、`X-B3-Sampled` 等 | Zipkin B3 多 Header 格式。 |
| Jaeger | `uber-trace-id` | Jaeger 传统传播格式；新接入优先使用 W3C Trace Context。 |
| Datadog | `x-datadog-trace-id`、`x-datadog-parent-id` 等 | Datadog 原生传播格式。 |

OpenTracing 是一套已归档的 API 和规范，不是 OpenTelemetry 的传播协议。部分旧实现使用 OT Trace 格式（例如 `ot-tracer-*`），但新接入不应把 `opentracing` 当成通用的 `OTEL_PROPAGATORS` 配置值。

## OpenTelemetry 配置 {#use-otel}

OpenTelemetry 默认组合通常为 W3C Trace Context 和 W3C Baggage：

```shell
export OTEL_PROPAGATORS=tracecontext,baggage
```

Java 系统属性写法：

```shell
-Dotel.propagators=tracecontext,baggage
```

`tracecontext` 和 `baggage` 是 OpenTelemetry 核心传播器。B3、Jaeger、AWS X-Ray 等格式是否可用，取决于语言 SDK、Agent 版本及是否安装相应扩展；配置前请查阅所用发行版的文档。不要在逗号后添加空格，以免某些实现把空格识别为值的一部分。

参考：

- [OpenTelemetry Propagators API](https://opentelemetry.io/docs/specs/otel/context/api-propagators/){:target="_blank"}
- [OpenTelemetry SDK 配置](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_propagators){:target="_blank"}

## DDTrace 配置 {#use-datadog}

DDTrace 使用 `DD_TRACE_PROPAGATION_STYLE` 同时控制提取和注入格式，也可以分别配置：

```shell
# 同时控制入站提取和出站注入
export DD_TRACE_PROPAGATION_STYLE=tracecontext,datadog

# 分别控制入站和出站
export DD_TRACE_PROPAGATION_STYLE_EXTRACT=tracecontext,datadog
export DD_TRACE_PROPAGATION_STYLE_INJECT=tracecontext,datadog
```

不同语言和版本支持的格式及默认值可能不同。混用 DDTrace 与 OpenTelemetry 时，建议显式配置双方共同支持的 `tracecontext`，不要依赖默认值。完整兼容矩阵见 [Datadog Trace Context Propagation](https://docs.datadoghq.com/tracing/trace_collection/trace_context_propagation/){:target="_blank"}。

## DDTrace 与 OpenTelemetry 串联示例 {#dd-otel-example}

下面的示例使用 W3C Trace Context 串联两个 Java 服务。DDTrace 应用把 Trace 发送到 DataKit HTTP 端口 `9529`，OpenTelemetry 应用通过 OTLP/gRPC 发送到 `4317`。

DDTrace 客户端：

```shell
java -javaagent:/opt/ddtrace/dd-java-agent.jar \
  -Ddd.service=client \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext \
  -jar springboot-client.jar
```

OpenTelemetry 服务端：

```shell
java -javaagent:/opt/otel/opentelemetry-javaagent.jar \
  -Dotel.service.name=server \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://127.0.0.1:4317 \
  -Dotel.propagators=tracecontext,baggage \
  -jar springboot-server.jar
```

为避免不同 DDTrace Agent 版本的 Trace ID 默认行为不同，示例显式启用 128 位 Trace ID。W3C Trace Context 的 Trace ID 为 128 位十六进制字符串，Span ID 为 64 位十六进制字符串。

DataKit 的 `ddtrace` 配置还需启用 OpenTelemetry 兼容输出：

```toml
[[inputs.ddtrace]]
  compatible_otel = true
  trace_128_bit_id = true
```

- `compatible_otel` 将 DDTrace 的 `span_id` 和 `parent_id` 输出为十六进制字符串。
- `trace_128_bit_id` 使用 `_dd.p.tid` 中的高 64 位与载荷中的低 64 位重建 128 位 Trace ID；当前默认值为 `true`，此处显式写出便于核对。

配置文件应位于 `/usr/local/datakit/conf.d/ddtrace.conf`；示例文件位于 `/usr/local/datakit/conf.d/samples/ddtrace.conf.sample`。修改后重启 DataKit，并用一次跨服务请求确认两端 Trace ID 连续、父子 Span 关系正确。

<!-- markdownlint-disable MD046 -->
???+ tip "日志中的 DDTrace Span ID"

    DDTrace 日志注入的 Span ID 可能仍为十进制。需要与 OpenTelemetry 的十六进制 Span ID 关联时，可在日志 Pipeline 中使用 `parse_int()` 和 `format_int()` 转换提取字段；这不会修改原始日志文本。
<!-- markdownlint-enable MD046 -->
