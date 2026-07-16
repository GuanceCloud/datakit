---
title     : 'DDTrace Java Extensions'
summary   : 'Extended capabilities and safe configuration for the <<<custom_key.brand_name>>> DDTrace Java Agent'
__int_icon: 'icon/ddtrace'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - 'APM'
  - 'TRACING'
---

## Introduction {#intro}

This page describes capabilities added by the <<<custom_key.brand_name>>> extended DDTrace Java Agent. It is built from upstream `dd-trace-java`, but it is not the same JAR: extension settings on this page work only with the <<<custom_key.brand_name>>> Agent. Complete the base receiver, DataKit address, and port setup in [DDTrace Java](ddtrace-java.md) first.

Extension capabilities change with Agent versions. Pin the JAR version, validate it in pre-production, and use the [changelog](ddtrace-ext-changelog.md) to confirm the minimum supported version and known changes. A setting appearing in this document does not mean every historical JAR supports it.

<!-- markdownlint-disable MD046 -->
???+ warning "Assess data-exposure risk first"

    Request/response bodies, HTTP headers, Redis arguments, JDBC arguments, and custom-method parameters can contain passwords, tokens, cookies, personal data, or business secrets. Keep them disabled by default. Enable only the minimum approved low-risk endpoints or fields, and validate redaction, access control, retention, and volume before release.
<!-- markdownlint-enable MD046 -->

## Feature Overview {#feature-overview}

| Capability | Scope | Minimum extended version / note |
| --- | --- | --- |
| Java-WebSocket | Connection, message, and close tracing; message collection is off by default | `v1.55.10-ext` |
| Dubbo | Dubbo 2 (`2.7+`) and Dubbo 3 | `v1.30.4-ext` |
| RocketMQ | Apache RocketMQ `4.5+`; Alibaba Cloud RocketMQ 5.x uses a different artifact | Minimum-version note from `v1.55.11-ext` |
| Thrift | `0.9.3+` | `v0.113.0` |
| Redis arguments | Jedis `1.4+`, Lettuce, Redisson | `v1.3.2-ext` |
| MongoDB argument obfuscation and DM8 | Common MongoDB scalar values; DM8 | `v1.12.1-ext` |
| HTTP headers / request and response bodies | Servlet HTTP scenarios | Headers: `v1.25.2-ext`; body: `v1.55.6-ext` |
| Package/method instrumentation | Custom business classes and methods | Package: `v1.47.6-ext`; file-based method rules: `v1.47.4-ext` |
| 128-bit trace ID and W3C | `tracecontext` chaining with OpenTelemetry | `v1.14.0-ext` |
| Log4j2 log pattern | Log correlation with trace/span IDs | `v1.3.0-ext` |

For other extensions, including HSF, XXL-JOB, PowerJob, Pulsar, Kingbase, and MyBatis-Plus, use the version-specific [changelog](ddtrace-ext-changelog.md) and pre-production validation as the source of truth.

## Third-Party Instrumentation {#third-party}

### Java-WebSocket {#java-websocket}

The extended Agent can trace WebSocket handshakes, message send/receive, and connection closure. Message volume and contents can be large, so message tracing is disabled by default. Enable it only after assessing throughput and privacy impact:

```shell
-Ddd.trace.websocket.messages.enabled=true
```

Use sampling and capacity protection for high-frequency connections.

### Dubbo {#dubbo}

The extension supports context propagation and RPC spans for Dubbo 2 and Dubbo 3. The call path still needs compatible propagation on consumers and providers. If topology is disconnected, first check Agent versions, Dubbo versions, and [Multi-Tracing Propagation](tracing-propagator.md).

### RocketMQ {#rocketmq}

Apache RocketMQ and Alibaba Cloud RocketMQ 5.x use different client artifacts, so name alone is not enough to determine compatibility. Record client coordinates and version, then compare them with the extension changelog. Under load, validate that async-consumer spans finish, context propagates, and retries do not create unexpected duplicate spans.

### Thrift {#thrift}

The extension supports Thrift `0.9.3+`. For connection reuse, asynchronous clients, or multiplexed protocols, validate parent/child relationships end to end rather than checking only for a local span.

### HSF {#hsf}

[HSF](https://help.aliyun.com/document_detail/100087.html){:target="_blank"} is an Alibaba RPC framework. The extension documents support for `2.2.8.2--2019-06-stable`; validate other versions before production use.

### Other Job and Messaging Frameworks {#xxl-jobs}

Support for these frameworks evolves with the extension. Verify the runtime dependency version and extension-JAR version, then test a successful task, a failed task, and a retry path. A successfully loaded Agent does not by itself prove that a framework is instrumented.

## Safety Before Collecting Additional Data {#data-safety}

### Redis Command Arguments {#redis-command-args}

Redis span resources normally show only the command name. The following setting writes command arguments into the `redis.command.args` tag:

```shell
-Ddd.redis.command.args=true
# or
export DD_REDIS_COMMAND_ARGS=true
```

Arguments often contain sessions, cached payloads, or business keys. Enable this only when needed, and make sure DataKit access control, redaction, and retention policies cover the resulting data.

### JDBC Parameter Collection {#jdbc-sql-obfuscation}

Although the setting is named `dd.jdbc.sql.obfuscation`, the extension stores `PreparedStatement` placeholder values as `sql.params.index_N` tags to help SQL troubleshooting. It is **not a general-purpose data-redaction mechanism**: parameters can be plaintext sensitive values.

```shell
-Ddd.jdbc.sql.obfuscation=true
# or
export DD_JDBC_SQL_OBFUSCATION=true
```

The original SQL remains parameterized in `db.sql.origin`, while values are stored separately to avoid unreliable string replacement. Use this only for short, approved troubleshooting windows, then disable it and review access to data already collected.

### MongoDB Argument Obfuscation {#mongo-obfuscation}

Enable the MongoDB-related extension with:

```shell
-Ddd.mongo.obfuscation=true
# or
export DD_MONGO_OBFUSCATION=true
```

The feature is intended to reduce command-argument exposure, but it does not replace application data classification and verification. Supported MongoDB value types and presentation can vary by version; validate with representative, already-sanitized data before release.

### DM8 Database {#dameng-db}

The extended Agent supports DM8 tracing. Verify driver version, connection mode, and expected `db.system`, instance, and error fields in a database span.

## HTTP Data Collection {#http}

### HTTP Status Classification {#http-error}

The extended Agent can mark HTTP 4xx requests as errors with:

```shell
-Ddd.http.error.enabled=true
```

Decide the business semantics first. Marking expected 401, 404, or validation failures as errors can distort error rates and alerts. Compare error data before and after enabling it in a test environment.

### Request and Response Bodies {#response_body}

These settings are disabled by default:

```shell
-Ddd.trace.request.body.enabled=true
-Ddd.trace.response.body.enabled=true

# Environment-variable equivalents
export DD_TRACE_REQUEST_BODY_ENABLED=true
export DD_TRACE_RESPONSE_BODY_ENABLED=true
```

Reading response streams consumes memory and can affect large, streaming, or download responses. Use it only for small, low-risk APIs and restrict paths with a list:

```shell
# Blacklist: never collect response bodies on these paths
-Ddd.trace.response.body.blacklist.urls="/download,/export"

# Whitelist: collect response bodies only on these paths
-Ddd.trace.response.body.whitelist.urls="/health/detail,/api/debug/*"
```

The environment variables are `DD_TRACE_RESPONSE_BODY_BLACKLIST_URLS` and `DD_TRACE_RESPONSE_BODY_WHITELIST_URLS`. Do not rely on both lists in one environment to express policy; choose one auditable rule set and validate actual matching. Response bodies use UTF-8 by default; adjust with `dd.trace.response.body.encoding` only when required.

### HTTP Headers {#trace_header}

```shell
-Ddd.trace.headers.enabled=true
# or
export DD_TRACE_HEADERS_ENABLED=true
```

When enabled, request and response headers are written to span tags such as `servlet_request_header` and `servlet_response_header`. Do not collect `Authorization`, `Cookie`, `Set-Cookie`, or tenant/user headers unless they are removed, redacted, or limited upstream.

## Custom Business Instrumentation {#other}

### Package Instrumentation {#package}

Instrument selected business packages with:

```shell
-Ddd.trace.method.packages=com.example.api,com.example.service
# or
export DD_TRACE_METHOD_PACKAGES=com.example.api,com.example.service
```

Package instrumentation can greatly increase span volume. Start with a small number of business packages, avoid framework packages and high-frequency getters/setters, and recheck performance and span naming after upgrades.

### File-Based Method Rules {#trace-method}

Keep method rules in a file instead of a long startup option:

```shell
-Ddd.trace.method.file=/opt/ddtrace/methods.txt
# or
export DD_TRACE_METHOD_FILE=/opt/ddtrace/methods.txt
```

Each `methods.txt` line is one rule, for example:

```text
com.example.api.OrderController[*]
com.example.service.PaymentService[charge]
```

The rule syntax follows upstream [`dd.trace.methods` configuration](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"}. Version the file with the application image or Pod volume and confirm in startup logs that it was loaded.

### Parameters of Specific Methods {#dd_trace_methods}

<!-- markdownlint-disable MD033 -->
<span id="dd-trace-methods"></span>
<!-- markdownlint-enable MD033 -->

Use `dd.trace.methods` or `@Trace` to create spans for selected methods. The extended Agent can record parameter names, types, and values. Current limits include no more than five method parameters, string values capped at 1024 characters, and object representation based on `toString()`. `toString()` is not safe serialization and can expose data or be expensive, so enable it only for reviewed methods.

## Propagation and Logs {#propagation-and-logs}

### 128-Bit Trace ID {#trace_128_bit_id}

When chaining with OpenTelemetry through W3C `tracecontext`, enable 128-bit ID generation and W3C propagation in the extended Agent:

```shell
-Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext

# or
export DD_TRACE_128_BIT_TRACEID_GENERATION_ENABLED=true
export DD_TRACE_PROPAGATION_STYLE=tracecontext
```

Also enable `compatible_otel=true` in the DataKit `ddtrace` collector and keep its default `trace_128_bit_id=true`; see the [DDTrace receiver guide](ddtrace.md#trace_propagator). Verify the full 32-character Trace ID and parent/child relationships with a cross-service request.

### Log4j2 Pattern {#log-pattern}

The extension can customize the Log4j2 pattern so logs include service, trace ID, and span ID:

```shell
-Ddd.logs.pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger - %X{dd.service} %X{dd.trace_id} %X{dd.span_id} - %msg%n"
```

The environment-variable equivalent is `DD_LOGS_PATTERN`. This extension currently declares Log4j2 support only. The log collector must preserve these MDC fields for trace/log correlation.

## Default Port and Bulk Injection {#agent_port}

The common upstream Java Agent default trace port is `8126`, while some <<<custom_key.brand_name>>> extended versions have used `9529`. To avoid version-dependent routing, always set `DD_TRACE_AGENT_PORT=9529` or `-Ddd.trace.agent.port=9529` explicitly.

### Kubernetes Bulk Injection {#java-attach}

Use the [DataKit Operator](../operator-ddtrace.md) for Kubernetes bulk injection. It injects at Pod creation through a webhook; after configuration changes, recreate Pods and verify initContainer, volume mounts, startup options, and environment variables using the Operator guide. There is no separately maintained attach guide, so do not rely on stale links or manual Agent copies that are not managed by the Operator.

## Verify an Extension {#verify}

1. Pin the extended JAR version and confirm the feature's minimum version in the [changelog](ddtrace-ext-changelog.md).
1. Temporarily enable `DD_TRACE_STARTUP_LOGS=true` and confirm that the Agent loaded without compatibility warnings.
1. Enable one extension at a time, send a minimal test request, and check the expected span/tag before enabling another feature.
1. For any content or parameter collection, review the resulting data for sensitive values, acceptable span volume, and correct path-list matching.
