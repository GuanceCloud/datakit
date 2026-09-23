---
title     : 'Tracing Propagator'
summary   : 'Propagation between multiple tracing agents'
tags      :
  - 'APM'
  - 'TRACING'
__int_icon: ''
---

Distributed tracing uses propagators to inject and extract Trace IDs, Span IDs, sampling decisions, and baggage across process boundaries. HTTP instrumentation usually carries this information in request headers. Spans form one trace only when both ends of a call use compatible propagation formats.

Propagation carries context only; it does not export span data to DataKit. Configure the DDTrace or OpenTelemetry destination, port, and transport separately in each application.

## Common Propagation Formats {#propagators}

| Format | Primary HTTP headers | Notes |
| --- | --- | --- |
| W3C Trace Context | `traceparent`, `tracestate` | Vendor-neutral standard; recommended for mixed DDTrace and OpenTelemetry traces. |
| W3C Baggage | `baggage` | Carries application-defined key-value pairs. Do not propagate sensitive data across untrusted boundaries. |
| B3 Single | `b3` | Zipkin B3 single-header format. |
| B3 Multi | `X-B3-TraceId`, `X-B3-SpanId`, `X-B3-Sampled`, and others | Zipkin B3 multi-header format. |
| Jaeger | `uber-trace-id` | Legacy Jaeger format; prefer W3C Trace Context for new integrations. |
| Datadog | `x-datadog-trace-id`, `x-datadog-parent-id`, and others | Datadog native format. |

OpenTracing is an archived API and specification, not an OpenTelemetry propagation protocol. Some legacy implementations use the OT Trace format, such as `ot-tracer-*` headers, but new integrations should not treat `opentracing` as a portable `OTEL_PROPAGATORS` value.

## OpenTelemetry Configuration {#use-otel}

The usual OpenTelemetry default is a composite of W3C Trace Context and W3C Baggage:

```shell
export OTEL_PROPAGATORS=tracecontext,baggage
```

Java system-property form:

```shell
-Dotel.propagators=tracecontext,baggage
```

`tracecontext` and `baggage` are core OpenTelemetry propagators. Availability of B3, Jaeger, AWS X-Ray, and other formats depends on the language SDK, Agent version, and installed extensions. Check the documentation for the distribution you use. Do not add spaces after commas because some implementations may treat the space as part of the value.

References:

- [OpenTelemetry Propagators API](https://opentelemetry.io/docs/specs/otel/context/api-propagators/){:target="_blank"}
- [OpenTelemetry SDK configuration](https://opentelemetry.io/docs/languages/sdk-configuration/general/#otel_propagators){:target="_blank"}

## DDTrace Configuration {#use-datadog}

DDTrace uses `DD_TRACE_PROPAGATION_STYLE` for both extraction and injection. You can also configure them separately:

```shell
# Configure inbound extraction and outbound injection together
export DD_TRACE_PROPAGATION_STYLE=tracecontext,datadog

# Configure inbound and outbound behavior separately
export DD_TRACE_PROPAGATION_STYLE_EXTRACT=tracecontext,datadog
export DD_TRACE_PROPAGATION_STYLE_INJECT=tracecontext,datadog
```

Supported formats and defaults vary by language and version. For a mixed DDTrace and OpenTelemetry trace, explicitly select the mutually supported `tracecontext` format instead of relying on defaults. See [Datadog Trace Context Propagation](https://docs.datadoghq.com/tracing/trace_collection/trace_context_propagation/){:target="_blank"} for the current language matrix.

## DDTrace-to-OpenTelemetry Example {#dd-otel-example}

The following Java example uses W3C Trace Context across two services. The DDTrace application exports traces to DataKit's HTTP port `9529`; the OpenTelemetry application exports through OTLP/gRPC on port `4317`.

DDTrace client:

```shell
java -javaagent:/opt/ddtrace/dd-java-agent.jar \
  -Ddd.service=client \
  -Ddd.agent.host=127.0.0.1 \
  -Ddd.trace.agent.port=9529 \
  -Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext \
  -jar springboot-client.jar
```

OpenTelemetry server:

```shell
java -javaagent:/opt/otel/opentelemetry-javaagent.jar \
  -Dotel.service.name=server \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://127.0.0.1:4317 \
  -Dotel.propagators=tracecontext,baggage \
  -jar springboot-server.jar
```

The example explicitly enables 128-bit Trace ID generation to avoid version-dependent DDTrace Agent defaults. W3C Trace Context uses a 128-bit hexadecimal Trace ID and a 64-bit hexadecimal Span ID.

Also enable OpenTelemetry-compatible output in the DataKit `ddtrace` configuration:

```toml
[[inputs.ddtrace]]
  compatible_otel = true
  trace_128_bit_id = true
```

- `compatible_otel` outputs DDTrace `span_id` and `parent_id` values as hexadecimal strings.
- `trace_128_bit_id` reconstructs a 128-bit Trace ID from the high 64 bits in `_dd.p.tid` and the low 64 bits in the payload. Its current default is `true`; it is explicit here for verification.

The active configuration belongs at `/usr/local/datakit/conf.d/ddtrace.conf`; the sample is `/usr/local/datakit/conf.d/samples/ddtrace.conf.sample`. Restart DataKit after editing, then send one cross-service request and verify continuous Trace IDs and correct parent-child relationships.

<!-- markdownlint-disable MD046 -->
???+ tip "DDTrace Span IDs in logs"

    A DDTrace-injected log Span ID may remain decimal. To correlate it with a hexadecimal OpenTelemetry Span ID, convert the extracted field with `parse_int()` and `format_int()` in the log Pipeline. This does not change the original log text.
<!-- markdownlint-enable MD046 -->
