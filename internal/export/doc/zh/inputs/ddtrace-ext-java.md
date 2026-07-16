---
title     : 'DDTrace Java 扩展'
summary   : '<<<custom_key.brand_name>>> DDTrace Java Agent 的扩展能力与安全配置'
__int_icon: 'icon/ddtrace'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - '链路追踪'
---

## 简介 {#intro}

本文说明<<<custom_key.brand_name>>>扩展版 DDTrace Java Agent 的附加能力。它基于上游 `dd-trace-java`，但并不等同于上游 JAR：本页中的扩展配置只有在使用<<<custom_key.brand_name>>>版 Agent 时才生效。基础接入、DataKit 地址与端口配置请先完成 [DDTrace Java](ddtrace-java.md)。

扩展能力会随 Agent 版本变化。部署前应固定 JAR 版本，在预发布环境验证，并用[更新日志](ddtrace-ext-changelog.md)确认最低支持版本与已知变更。不要将“当前文档出现某配置”理解为任意历史 JAR 都支持它。

<!-- markdownlint-disable MD046 -->
???+ warning "先评估数据暴露风险"

    请求/响应 Body、HTTP Header、Redis 参数、JDBC 参数和自定义方法入参都可能包含密码、令牌、Cookie、个人信息或业务机密。它们默认应保持关闭，仅对经过审批的低风险接口或字段最小化开启，并在上线前验证脱敏、访问权限、保留周期和数据量。
<!-- markdownlint-enable MD046 -->

## 功能概览 {#feature-overview}

| 能力 | 适用范围 | 最低扩展版本/说明 |
| --- | --- | --- |
| Java-WebSocket | 连接、消息和关闭事件的链路；消息采集默认关闭 | `v1.55.10-ext` |
| Dubbo | Dubbo 2（`2.7+`）和 Dubbo 3 | `v1.30.4-ext` |
| RocketMQ | Apache RocketMQ `4.5+`；阿里云 RocketMQ 5.x 使用不同 artifact | `v1.55.11-ext` 起的最低版本说明 |
| Thrift | `0.9.3+` | `v0.113.0` |
| Redis 参数 | Jedis `1.4+`、Lettuce、Redisson | `v1.3.2-ext` |
| MongoDB 参数脱敏、达梦 DM8 | MongoDB 常见标量参数；DM8 | `v1.12.1-ext` |
| HTTP Header / 请求与响应 Body | Servlet HTTP 场景 | Header：`v1.25.2-ext`；Body：`v1.55.6-ext` |
| 包/方法级插桩 | 自定义业务类与方法 | 包：`v1.47.6-ext`；文件化方法规则：`v1.47.4-ext` |
| 128 位 Trace ID 与 W3C | 与 OpenTelemetry 的 `tracecontext` 串联 | `v1.14.0-ext` |
| Log4j2 日志 Pattern | 日志与 trace/span ID 关联 | `v1.3.0-ext` |

其他扩展（如 HSF、XXL-JOB、PowerJob、Pulsar、Kingbase、MyBatis-Plus 等）请以对应版本的[更新日志](ddtrace-ext-changelog.md)和预发布验证结果为准。

## 三方库插桩 {#third-party-agent}

### Java-WebSocket {#java-websocket}

扩展版可为 WebSocket 握手、消息收发和连接关闭创建链路信息。消息内容和消息量可能很大，因此默认不采集消息链路；确认吞吐与隐私影响后再开启：

```shell
-Ddd.trace.websocket.messages.enabled=true
```

仅在需要分析消息收发链路时开启，并为高频连接配置采样和容量保护。

### Dubbo {#dubbo}

扩展版支持 Dubbo 2 和 Dubbo 3 的上下文透传与 RPC span。调用链仍要求消费端、提供端使用兼容的透传协议；若拓扑断开，请先检查服务端和客户端 Agent 版本、Dubbo 版本与 [多链路串联](tracing-propagator.md) 配置。

### RocketMQ {#rocketmq}

Apache RocketMQ 与阿里云 RocketMQ 5.x 使用不同客户端 artifact，不能仅按名称判断兼容性。请记录客户端坐标和版本，并与扩展版更新日志逐项核对。异步消费链路应在压测中检查 span 是否闭合、上下文是否正确传递以及失败重试是否产生重复 span。

### Thrift {#thrift}

Thrift `0.9.3+` 可使用扩展插桩。对复用连接、异步客户端或多路复用协议，应通过端到端测试确认父子关系，而不是仅检查单服务内是否有 span。

### HSF {#hsf}

[HSF](https://help.aliyun.com/document_detail/100087.html){:target="_blank"} 是阿里巴巴 RPC 框架。扩展版对文档记录的 `2.2.8.2--2019-06-stable` 版本提供支持；其他版本必须先验证。

### XXL-JOB、PowerJob、Pulsar 与其他框架 {#xxl-jobs}

这些框架的支持随扩展版本演进。启用前请确认运行时依赖版本和扩展 JAR 版本，并通过一次成功任务、失败任务和重试任务验证链路是否连续。不要因 Agent 加载成功就假设某个框架已经被插桩。

## 采集额外数据前的安全边界 {#data-safety}

### Redis 命令参数 {#redis-command-args}

Redis span 的 Resource 默认只显示命令名。以下开关会将命令参数写入 `redis.command.args` 标签：

```shell
-Ddd.redis.command.args=true
# 或
export DD_REDIS_COMMAND_ARGS=true
```

参数常包含 session、缓存内容或业务主键。启用后应在 DataKit 权限、脱敏和保留策略中覆盖这些数据；如只需命令耗时和错误，不要开启。

### JDBC 参数采集 {#jdbc-sql-obfuscation}

配置名为 `dd.jdbc.sql.obfuscation`，但扩展行为是把 `PreparedStatement` 占位参数以 `sql.params.index_N` 写入 span，便于排查 SQL。它**不是通用的数据脱敏机制**：参数可能是明文敏感信息。

```shell
-Ddd.jdbc.sql.obfuscation=true
# 或
export DD_JDBC_SQL_OBFUSCATION=true
```

原 SQL 仍以占位符形式保留在 `db.sql.origin`，参数独立存储，避免不可靠的字符串替换。仅在短期排障、经过审批的环境中开启；排障结束后关闭，并检查已有数据的访问范围。

### MongoDB 参数脱敏 {#mongo-obfuscation}

使用下列开关启用 MongoDB 相关扩展：

```shell
-Ddd.mongo.obfuscation=true
# 或
export DD_MONGO_OBFUSCATION=true
```

该能力的目标是降低命令参数暴露风险，但不能替代对应用数据的分类与验证。支持的 MongoDB 类型和展示效果会随版本变化；上线前必须用真实但脱敏的样本确认结果。

### 达梦数据库 {#dameng-db}

扩展版支持 DM8 的数据库链路信息。请同时验证驱动版本、连接方式和数据库 span 中的 `db.system`、实例名、错误字段是否符合预期。

## HTTP 数据采集 {#http}

### HTTP 状态标记 {#http-error}

扩展版可通过以下参数将 HTTP 4xx 请求标记为错误：

```shell
-Ddd.http.error.enabled=true
```

启用前先明确业务语义：大量预期的 401、404 或参数校验错误被标为错误后，错误率和告警可能失真。应在测试环境比较开启前后的错误数据。

### 请求与响应 Body {#response_body}

以下开关默认关闭：

```shell
-Ddd.trace.request.body.enabled=true
-Ddd.trace.response.body.enabled=true

# 对应环境变量
export DD_TRACE_REQUEST_BODY_ENABLED=true
export DD_TRACE_RESPONSE_BODY_ENABLED=true
```

读取响应流会增加内存占用，并可能影响大响应、流式响应或下载接口。仅对低风险、小体积 API 使用，并通过名单限制路径：

```shell
# 黑名单：这些路径不采集响应 Body
-Ddd.trace.response.body.blacklist.urls="/download,/export"

# 白名单：仅允许指定路径采集响应 Body
-Ddd.trace.response.body.whitelist.urls="/health/detail,/api/debug/*"
```

环境变量分别为 `DD_TRACE_RESPONSE_BODY_BLACKLIST_URLS` 和 `DD_TRACE_RESPONSE_BODY_WHITELIST_URLS`。同一环境不要同时依赖白名单与黑名单来表达策略；选择一种可审计的规则并验证实际匹配结果。响应 Body 默认按 UTF-8 处理，必要时可通过 `dd.trace.response.body.encoding` 调整编码。

### HTTP Header {#trace_header}

```shell
-Ddd.trace.headers.enabled=true
# 或
export DD_TRACE_HEADERS_ENABLED=true
```

开启后，请求和响应 Header 会写入 `servlet_request_header`、`servlet_response_header` 等 span 标签。`Authorization`、`Cookie`、`Set-Cookie` 和租户/用户 Header 通常不应采集；在启用前先通过网关或应用移除、脱敏或限制这些值。

## 自定义业务插桩 {#others}

### 包级插桩 {#package}

可按包名增强业务方法：

```shell
-Ddd.trace.method.packages=com.example.api,com.example.service
# 或
export DD_TRACE_METHOD_PACKAGES=com.example.api,com.example.service
```

包级插桩会显著增加 span 数量。应从少量业务包开始，避免框架包和高频 getter/setter；升级后重新检查性能与 span 命名。

### 文件化方法规则 {#trace-method}

用文件维护方法规则可避免把长规则写入启动命令：

```shell
-Ddd.trace.method.file=/opt/ddtrace/methods.txt
# 或
export DD_TRACE_METHOD_FILE=/opt/ddtrace/methods.txt
```

`methods.txt` 每行一条规则，例如：

```text
com.example.api.OrderController[*]
com.example.service.PaymentService[charge]
```

规则语法与上游 [`dd.trace.methods` 配置](https://docs.datadoghq.com/tracing/trace_collection/library_config/java/){:target="_blank"}保持一致。配置文件必须随应用镜像或 Pod 卷一起版本化，并在启动日志中确认已加载。

### 特定方法的入参 {#dd-trace-methods}

<!-- markdownlint-disable MD033 -->
<span id="dd_trace_methods"></span>
<!-- markdownlint-enable MD033 -->

可通过 `dd.trace.methods` 或 `@Trace` 标注生成特定方法的 span。扩展版可能记录入参名称、类型和值；当前限制包括最多 5 个方法入参、字符串值最多 1024 个字符，并基于 `toString()` 表示对象。`toString()` 的输出不等于安全序列化，可能泄露敏感字段或产生昂贵计算，因此只对经过审查的方法开启。

## 透传与日志 {#propagation-and-logs}

### 128 位 Trace ID {#trace_128_bit_id}

与 OpenTelemetry 使用 W3C `tracecontext` 串联时，可在扩展版 Agent 中开启 128 位 ID 生成与 W3C 透传：

```shell
-Ddd.trace.128.bit.traceid.generation.enabled=true \
  -Ddd.trace.propagation.style=tracecontext

# 或
export DD_TRACE_128_BIT_TRACEID_GENERATION_ENABLED=true
export DD_TRACE_PROPAGATION_STYLE=tracecontext
```

同时在 DataKit `ddtrace` 采集器中启用 `compatible_otel=true`，并保留默认的 `trace_128_bit_id=true`，详见 [DDTrace 接收端](ddtrace.md#trace_propagator)。配置前后应以跨服务请求验证完整 32 位 Trace ID 和父子关系。

### Log4j2 Pattern {#log-pattern}

扩展版可通过 `dd.logs.pattern` 调整 Log4j2 Pattern，使日志包含服务、trace ID 和 span ID：

```shell
-Ddd.logs.pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%thread] %-5level %logger - %X{dd.service} %X{dd.trace_id} %X{dd.span_id} - %msg%n"
```

对应环境变量为 `DD_LOGS_PATTERN`。该扩展当前仅声明支持 Log4j2；日志采集器还必须保留这些 MDC 字段，才能实现日志与链路关联。

## 默认端口与批量注入 {#agent-port}

上游 Java Agent 常见 trace 端口默认值为 `8126`，而某些<<<custom_key.brand_name>>>扩展版本曾将默认值调整为 `9529`。为避免版本差异导致数据发错位置，始终显式设置 `DD_TRACE_AGENT_PORT=9529` 或 `-Ddd.trace.agent.port=9529`。

### Kubernetes 批量注入 {#java-attach}

Kubernetes 批量注入请使用 [DataKit Operator](../operator-ddtrace.md)。它基于 Pod 创建时的 webhook 注入；修改配置后应重新创建 Pod，并按 Operator 文档验证 initContainer、卷挂载、启动参数与环境变量。当前没有独立维护的“attach”文档，因此不要依赖失效链接或手工复制未由 Operator 管理的 Agent 文件。

## 验证扩展是否生效 {#verify}

1. 固定扩展 JAR 版本，并从 [更新日志](ddtrace-ext-changelog.md)确认所用功能的最低版本。
1. 临时开启 `DD_TRACE_STARTUP_LOGS=true`，确认 Agent 被加载且没有兼容性告警。
1. 只启用一个扩展能力，发送最小测试请求，并在 trace 详情中检查预期 span/tag；随后再逐项开启其他能力。
1. 对任何会采集内容或参数的功能，复核数据是否包含敏感信息、span 数量是否可接受、是否命中路径名单。
