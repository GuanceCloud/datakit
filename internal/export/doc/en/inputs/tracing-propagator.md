# Propagation Among Multiple Tracing Stack

This article mainly introduces the products of multiple tracing stacks, and how to realize tracing information propagation them in distributed services.

The transparent transmission protocol, also known as the propagation protocol, is implemented by adding specific header information (generally referring to HTTP headers) in service requests and responses. When a service requests another service, it carries specific request headers. When the next hop receives the request, it obtains the specific link information from the request header and inherits it, and continues to propagate backward until the end of the link. In this way, the entire call chain can be correlated.

## Common propagation protocols {#propagators}

The following is a brief introduction to the differences between these transparent transmission protocols in the HTTP header:

### Trace Context {#propagators-w3c}

Trace Context is a trace protocol standardized by [W3C](https://www.w3.org/TR/trace-context/){:target="_blank"}, which defines two HTTP header fields: `traceparent` and `tracestate`:

- `traceparent` contains basic information about the current trace, such as SpanID and ParentSpanID, etc., for example: `traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01`
- `tracestate` is used to pass metadata related to the trace. For example: `tracestate: congo=t61rcWkgMzE`

### B3/B3Multi {#propagators-b3}

B3 is a popular tracking protocol that defines several HTTP header fields to identify tracking information. The B3Multi transparent transmission protocol is an extension of the B3 protocol. The commonly used fields are: `X-B3-TraceId`, `X-B3-SpanId`, `X-B3-ParentSpanId`, `X-B3-Sampled`, `X -B3-Flags` etc.

### Jaeger {#propagators-jaeger}

Jaeger is a distributed tracing system that defines several HTTP header fields for passing trace information. Commonly used fields are: `uber-trace-id`, `jaeger-baggage`, etc.

### OpenTracing {#propagators-ot}

OpenTracing is a transparent transmission protocol of OpenTelemetry, which defines multiple HTTP header fields for transmitting link information:

- `ot-tracer-traceid`: used to pass the link ID, indicating a complete request link
- `ot-tracer-spanid`: used to pass the ID of the current span, representing a single operation or event
- `ot-tracer-sampled`: used to indicate whether to sample the request to decide whether to record the trace information of the request

### Datadog {#propagators-datadog}

Datadog is a distributed tracing system that defines several HTTP header fields for passing trace information. Commonly used fields are: `x-datadog-trace-id`, `x-datadog-parent-id`, etc.

### Baggage {#propagators-baggage}

Baggage is a concept introduced by the Jaeger tracking system, which is used to transfer business-related context information. Baggage is passed through the HTTP header field `x-b3-baggage-<key>`, where `key` is the key of the business context.

The real meaning of Baggage is to propagate key-value pairs of the `key:value` nature, which is often used to propagate AppID, Host-Name, Host-IP, etc.

<!-- markdownlint-disable MD046 -->
???+ attention

    It should be noted that the specific implementation and usage of these transparent transmission protocols may be slightly different, but they all aim to pass tracking information and context information between different services through HTTP header fields to achieve distributed tracking and continuous sex.
<!-- markdownlint-enable -->

## Link manufacturers and product introduction {#tracing-info}

Products and manufacturers:

| Products      | Manufacturers     | Supported Languages                                                            |
| :---          | :---              | :---                                                                           |
| OpenTelemetry | CNCF              | Java, Python, Go, JavaScript, .NET, Ruby, PHP, Erlang, Swift, Rust, C++, etc.  |
| DDTrace       | Datadog           | Java, Python, Go, Ruby, JavaScript, PHP, .NET, Scala, Objective-C, Swift, etc. |
| SkyWalking    | Apache SkyWalking | Java, .NET, Node.js, PHP, Python, Go, Ruby, Lua, OAP, etc.                     |
| Zipkin        | OpenZipkin        | Java, Node.js, Ruby, Go, Scala, Python, etc.                                   |
| Jaeger        | CNCF              | Java, Python, Go, C++, C#, Node.js, etc.                                       |

The open source address of the product:

- [OpenTelemetry](https://github.com/open-telemetry){:target="_blank"} is a product under CNCF. At the same time, Observation Cloud also [extended it](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}
- [Jaeger](https://github.com/jaegertracing/jaeger){:target="_blank"} also belongs to CNCF
- [Datadog](https://github.com/DataDog){:target="_blank"} is a multilingual link tool, in which Observation Cloud has [extended](https://github.com/GuanceCloud /dd-trace-java){:target="_blank"}
- [SkyWalking](https://github.com/apache?q=skywalking&type=all&language=&sort=){:target="_blank"} is an open source product under the Apache Foundation
- [Zipkin](https://github.com/OpenZipkin){:target="_blank"} There are link tools in multiple languages.

## Product transparent transmission protocol {#use-propagators}

### OpenTelemetry {#use-otel}

List of tracing transparent transmission protocols supported by OTEL:

| Propagator List | Reference                                                                                                                      |
| ---             | ---                                                                                                                            |
| `tracecontext`  | [W3C Trace Context](https://www.w3.org/TR/trace-context/){:target="_blank"}                                                    |
| `baggage`       | [W3C Baggage](https://www.w3.org/TR/baggage/){:target="_blank"}                                                                |
| `b3`            | [B3](https://github.com/openzipkin/b3-propagation#single-header){:target="_blank"}                                             |
| `b3multi`       | [B3Multi](https://github.com/openzipkin/b3-propagation#multiple-headers){:target="_blank"}                                     |
| `jaeger`        | [Jaeger](https://www.jaegertracing.io/docs/1.21/client-libraries/#propagation-format){:target="_blank"}                        |
| `xray`          | [AWS X-Ray](https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html#xray-concepts-tracingheader){:target="_blank"} |
| `opentracing`   | [OpenTracing](https://github.com/opentracing?q=basic&type=&language=){:target="_blank"}                                        |

Example format of distributed link header information in transparent transmission:

```shell
# Example of command line injection (multiple communication protocols are separated by commas)
-Dotel.propagators="tracecontext,baggage"

# Environment variable injection example (Linux)
export OTEL_PROPAGATORS="tracecontext, baggage"

# Environment variable injection example (Windows)
$env:OTEL_PROPAGATORS="tracecontext,baggage"
```

### Datadog {#use-datadog}

| Supported languages | Transparent protocol support           | Commands                                                      |
| :---                | :---                                   | :---                                                          |
| Node.js             | `datadog/b3multi/tracecontext/b3/none` | `DD_TRACE_PROPAGATION_STYLE` (default `datadog`)              |
| C++                 | `datadog/b3multi/b3/none`              | `DD_TRACE_PROPAGATION_STYLE` (default `datadog`)              |
| .NET                | `datadog/b3multi/tracecontext/none`    | `DD_TRACE_PROPAGATION_STYLE` (default `datadog`)              |
| Java                | `datadog/b3multi/tracecontext/none`    | `DD_TRACE_PROPAGATION_STYLE` (default `tracecontext,datadog`) |

> Here `none` means that tracing protocol transparent transmission is not set.

#### DD_TRACE_PROPAGATION_STYLE {#dd-pg-style}

Datadog tracing can make inbound settings on the behavior of protocol transparent transmission, that is, whether to inherit the upstream protocol and whether to transparently transmit its own protocol to the downstream. It is controlled separately by the following two environment variables:

- Inbound control: `export DD_TRACE_PROPAGATION_STYLE_EXTRACT=<XXX>`
- Outbound control: `export DD_TRACE_PROPAGATION_STYLE_INJECT=<YYY>`
- It is also possible to control both inbound and outbound via a single ENV: `export DD_TRACE_PROPAGATION_STYLE="tracecontext,datadog"`

Example:

```shell
# Inbound will inherit X-Datadog-* and X-B3-* headers (if any),
# X-Datadog-* and X-B3-* request headers will be carried when outbound
$ export DD_TRACE_PROPAGATION_STYLE="datadog,b3" ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    After version V1.7.0, the default support protocol is changed to `DD_TRACE_PROPAGATION_STYLE="tracecontext,datadog"`, B3 has been deprecated, please use B3multi.
<!-- markdownlint-enable -->

For more language examples, see [here](https://github.com/DataDog/documentation/blob/4ff75ed0bcaa1269bf98e9d185935cfda675b08c/content/en/tracing/trace_collection/trace_context_propagation/_index.md){:target="_blank"}.

### SkyWalking {#use-sw8}

SkyWalking's own [protocol (SW8)](https://skywalking.apache.org/docs/main/next/en/api/x-process-propagation-headers-v3/){:target="_blank"}

### Zipkin {#use-zipkin}

[see here](https://github.com/openzipkin/b3-propagation){:target="_blank"}

### Jaeger {#use-jaeger}

All supported protocols:

- [Jaeger Propagation Format](https://www.jaegertracing.io/docs/1.21/client-libraries/#propagation-format){:target="_blank"}
- [B3 propagation](https://github.com/openzipkin/b3-propagation){:target="_blank"}
-W3C Trace-Context

## Multi-link series {#series}

Request header and vendor support list:

|               | W3C                       | b3multi                  | Jaeger                   | OpenTracing              | Datadog                  | sw8                      |
| :----         | :---                      | :---                     | :---                     | :---                     | :---                     | :---                     |
| HTTP Header   | `tracecontext/tracestate` | `X-B3-*`                 | `uber-trace-id`          | `ot-tracer-*`            | `x-datadog-*`            | `xxx-xxx-xxx-xxx`        |
| OpenTelemetry | :heavy_check_mark:        | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_multiplication_x: |
| Datadog       | :heavy_check_mark:        | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_check_mark:       | :heavy_multiplication_x: |
| SkyWalking    | :heavy_multiplication_x:  | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_check_mark:       |
| Zipkin        | :heavy_multiplication_x:  | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: |
| Jaeger        | :heavy_check_mark:        | :heavy_check_mark:       | :heavy_check_mark:       | :heavy_multiplication_x: | :heavy_multiplication_x: | :heavy_multiplication_x: |

According to the specific manufacturer's tools used, the corresponding transparent transmission protocol can be used to realize link series connection and ensure link integrity.

### Concatenation Example {#dd-otel-example}

Here is an example to illustrate the concatenation of DDTrace and OpenTelemetry link data. It can be seen from the above table that both DDTrace and OpenTelemetry support the W3C Trace Context protocol, and link concatenation can be realized through this protocol.

- TraceID in DDTrace is a 64-bit int string, and SpanID and ParentID are also 64-bit int
- TraceID in OTEL is a 128-bit hexadecimal int string, and SpanID and ParentID are 64-bit int strings

If the two want to associate TraceID, DDTrace needs to be upgraded to 128 bits.

No matter which one is the initiator of the request, DDTrace needs to enable 128bit TraceID support (`dd.trace.128.bit.traceid.generation.enabled`):

```shell
# DDTrace start example
$ java -javaagent:/usr/local/ddtrace/dd-java-agent.jar\
  -Ddd.service.name=client \
  -Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext\
  -jar springboot-client.jar

# OTEL start example
$ java -javaagent:/usr/local/ddtrace/opentelemetry-javaagent.jar\
  -dotel.service.name=server\
  -jar springboot-server.jar
```

The client will send an HTTP request to the server, and DDTrace will pass the link information in the `tracecontext` request header to the server

However, in the "service call relationship", the data from the two tools cannot be connected. This is because the SpanIDs of both parties are not uniform. DDTrace is a decimal string of numbers, while OpenTelemetry is a hexadecimal number character. string. To do this, you need to modify the configuration in the `ddtrace` collector and release `compatible_otel` in `ddtrace.conf`:

```toml
  ## compatible otel: It is possible to compatible OTEL Trace with DDTrace trace.
  ## make span_id and parent_id to hex encoding.
  compatible_otel=true
```

After `compatible_otel=true`, all DDTrace `span_id` and `parent_id` will become hexadecimal numeric strings.

<!-- markdownlint-disable MD046 -->
???+ tip "Convert `span_id` from digital to hexadecimal"

    In the loggging, the SpanId in DDTrace is still in decimal, you need to extract `span_id` in the Pipeline for collecting logs and convert it into a hexadecimal number string (the original logging text will not be modified):

    ```python
    # convert string to int64
    fn parse_int(val: str, base: int) int64

    # convert int64 to string
    fn format_int(val: int64, base: int) str
    ```
<!-- markdownlint-enable -->

So far, DDTrace and OTEL have been connected in series on the link, and the service call relationship and logs can also be connected in series:

<!-- markdownlint-disable MD046 MD033 -->
<figure>
  <img src="https://github.com/GuanceCloud/dd-trace-java/assets/31207055/9b599678-1ebc-4f1f-9993-f863fb25280b" style="height: 600px" alt="Link Details">
  <figcaption> Link Details </figcaption>
</figure>
<!-- markdownlint-enable -->
