---
title: 'DDTrace Java 扩展更新日志'
skip: 'not-searchable-on-index-page'
---

## 简介 {#intro}

本文记录 DataKit 中用于兼容 DDTrace Java 探针接入场景的 Java 扩展包更新内容。该扩展基于 `DataDog/dd-trace-java` 开发，遵循 `Apache License 2.0`。
有关法律文件、校验和及源代码变更信息，请参阅<<<custom_key.brand_name>>> [Java Tracer 扩展声明](../../application-performance-monitoring/java-tracer-extension)

本日志用于确认某个扩展能力首次出现、行为变更或修复所在的版本，不是完整的配置手册。接入时请按以下顺序使用：

1. 在部署清单中固定实际使用的扩展 JAR 版本；
1. 在本日志中确认目标能力的最低版本和后续行为变化；
1. 回到 [Java 扩展说明](ddtrace-ext-java.md) 查看配置、数据安全边界和验证步骤；
1. 在预发布环境完成兼容性测试后再升级生产环境。

历史条目保留当时的发布语义；较新的版本并不保证继续保留每个早期实验性行为。

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: **Java**

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar){:target="_blank"}

</div>
<!-- markdownlint-enable MD046 MD030 -->

## 更新历史 {#changelog}

<!--

更新历史可以参考 DataKit 的基本范式：

## 1.2.3(2022/12/12) {#cl-1.2.3}
本次发布主要有如下更新：

### 新加功能 {#cl-1.2.3-new}
### 问题修复 {#cl-1.2.3-fix}
### 功能优化 {#cl-1.2.3-opt}
### 兼容调整 {#cl-1.2.3-brk}

--->

## v1.65.0-ext (2026/8/7) {#cl-1.65.0-ext}

### 更新 {#cl-1.65.0-ext-update}

- 基于 DataDog `dd-trace-java` v1.65.0，合并上游功能更新与问题修复，并加入 GuanceCloud 定制内容。
- 新增可配置的 HTTP 请求和响应 Header、Body 采集，支持捕获 JSON 响应体以及黑白名单过滤；响应可注入 `ext_trace_id`，便于关联前后端链路。
- 支持 GWT RPC 请求体 tagging，并新增 JVM 线程状态、GC StatsD 指标以及 Netty Client SSE 响应流 span。
- 新增 BES 应用服务器、Nacos 异步线程链路透传和基于包名配置的 Trace 方法级增强；改进 xxl-job、Java-WebSocket、RabbitMQ 等探针，并修复 Redis Lettuce/Redisson 集群的 split-by-host 处理。
- 支持 Log4j2 2.7 日志 Pattern 替换；新增 Kafka 3.8+ 实验性开关 `dd.trace.experimental.kafka.enabled`；修复 Redis 和 RabbitMQ 客户端消息宿主命名。


## v1.63.7-ext (2026/7/1) {#cl-1.63.7-ext}

### 更新 {#cl-1.63.7-ext-fix}

- Spring WebMVC 请求体采集新增对 `text/x-gwt-rpc` 的支持；开启 `dd.trace.request.body.enabled` 后，POST GWT RPC 请求会写入 `request_body` span 标签。
- 请求体解码优先使用请求声明的 character encoding，未声明或无效时回退 UTF-8。
- 未启用 OTLP runtime metrics 时，通过 DogStatsD 上报 `jvm.thread.count`，并按 `jvm.thread.daemon:true|false` 与 `jvm.thread.state:<state>` 区分 daemon / 非 daemon 线程和线程状态。
- 新增原始 GC MXBean 指标 `jvm.gc.collection_count` 与 `jvm.gc.collection_time`，并保留 collector 名称为 `gc:<collector name>` 标签。
- 复用 JVM 线程状态分桶逻辑，补充 GWT RPC 请求体采集、JVM 线程状态统计和 GC StatsD 上报的测试覆盖。


## v1.63.6-ext (2026/6/29) {#cl-1.63.6-ext}

### 修复 {#cl-1.63.6-ext-fix}

- 修复 BES 探针 root span 丢失问题。
- 支持在 Spring WebMVC 过滤器中为 `text/x-gwt-rpc` 请求打上 `request_body` 标签，并补充对应的 forked coverage。


## v1.63.5-ext (2026/6/24) {#cl-1.63.5-ext}

### 更新 {#cl-1.63.5-ext-fix}

- 增强 Spring RabbitMQ 消费链路与日志串联能力。
- 修复 Spring RabbitMQ 消费消息时缺少业务消费 span 的问题，使消费者日志可在 listener 执行期间获取 `trace_id` / `span_id`。
- 保留 RabbitMQ `basic.deliver` 低层 AMQP span，同时补充 Spring listener 业务处理阶段 span。
- 新增 BES 11.0 应用服务器探针支持。


## v1.63.4-ext (2026/6/10) {#cl-1.63.4-ext}

### 新增 {#cl-1.63.4-ext-fix}

- 新增 `netty.client.stream` span，用于统计 SSE 响应体持续读取阶段。
- 保留现有 `netty.client.request` span，继续表示请求发出到响应头返回阶段。
- 新增指标 `stream.first_chunk.ms` 与 `stream.chunk_count`，用于观察首包延迟和 Netty 内容分片数量。
- 对 `text/event-stream` 响应，将 stream span 标记为 `internal`，并使用 `SSE stream ...` 资源名。


## v1.63.3-ext (2026/6/11) {#cl-1.63.3-ext}

### 新增 {#cl-1.63.3-ext-fix}

- 新增配置项 `DD_TRACE_PEER_HOSTNAME_FROM_CONFIG_ENABLED` / `trace.peer.hostname.from.config.enabled`。
- 默认关闭；开启后优先使用客户端连接配置中的 host 作为 `peer.hostname`。
- 覆盖 `Jedis`、`Lettuce 5`、`Redisson`、`Valkey`、`Vertx Redis Client`。


## v1.63.2-ext (2026/6/8) {#cl-1.63.2-ext}

### 修复 {#cl-1.63.2-ext-fix}

- 修复 JMXFetch 对 `17-ea` 等 Java 版本字符串的识别问题。
- 新增 `org.datadog.jmxfetch.util.JavaVersion` 版本解析工具。
- 同时兼容 `java.specification.version` 与 `java.version` 的多种格式。


## v1.63.1-ext (2026/6/4) {#cl-1.63.1-ext}

### 新增 {#cl-1.63.1-ext-fix}

- 新增配置项 `DD_SERVICE_MAPPING_FILE` / `dd.service.mapping.file`。
- 支持从外部文件读取 service mapping，并与 `DD_SERVICE_MAPPING` 的内联配置合并。
- 补充 `supported-configurations` 元数据以及对应单元测试。


## v1.63.0-ext (2026/6/3) {#cl-1.63.0-ext}

### 新增 {#cl-1.63.0-ext-fix}

- 合并最新代码


## v1.60.4-ext (2026/4/27) {#cl-1.60.4-ext}

### 新增 {#cl-1.60.4-ext-fix}

- 修复 Redis 显示 `service_name` 问题。


## v1.60.3-ext (2026/4/24) {#cl-1.60.3-ext}

### 新增 {#cl-1.60.3-ext-fix}

- 优化 JDBC 对于 Oracle 的支持。


## v1.55.11-ext (2026/3/17) {#cl-1.55.11-ext}

### 新增 {#cl-1.55.11-ext-fix}

- 探针 RocketMQ 最低版本从 4.8.0 到 4.5.0


## v1.55.10-ext (2026/2/1) {#cl-1.55.10-ext}

### 新增 {#cl-1.55.10-ext-fix}

- 增加 `java-websocket` 探针支持。


## v1.55.7-ext (2025/12/30) {#cl-1.55.7-ext}

### 修复 {#cl-1.55.7-ext-fix}

- 修复：Redis 参数没有正确显示问题。


## v1.55.6-ext (2025/12/22) {#cl-1.55.6-ext}

### 修复 {#cl-1.55.6-ext-fix}

- 修复：RocketMQ scope limit error.
- 添加 Response Body 白名单 [配置并开启](ddtrace-ext-java.md#response_body) 功能。

## v1.55.2-ext (2025/11/28) {#cl-1.55.2-ext}

### 修复 {#cl-1.55.2-ext-fix}

- fix:RocketMQ scope limit error.
- Merge latest branch v1.55.0

## v1.53.7-ext (2025/11/28) {#cl-1.53.7-ext}

### 修复 {#cl-1.53.7-ext-fix}

- support **Apache Dubbo** stream version 3.2

## v1.53.1-ext (2025/9/24) {#cl-1.53.1-ext}

### 修复 {#cl-1.53.1-ext-fix}

- 合并主分支： 1.53.0


## v1.47.6-ext (2025/6/4) {#cl-1.47.6-ext}

### 修复 {#cl-1.47.6-ext-fix}

- 可以针对自定义 Package 及 Class 中方法进行增强，如何通过[命令开启功能](ddtrace-ext-java.md#package){:target="_blank"}


## v1.47.5-ext (2025/5/22) {#cl-1.47.5-ext}

### 修复 {#cl-1.47.5-ext-fix}

- 修复： Pulsar 消费者链路断开问题。
- 修复： 资源目录字段缺失问题。

## v1.47.4-ext (2025/5/14) {#cl-1.47.4-ext}

### 新增 {#cl-1.47.4-ext-fix}

- 方法级的插桩，[配置并开启](ddtrace-ext-java.md#trace-method) 功能。


## v1.47.1-ext (2025/4/17) {#cl-1.47.1-ext}

### 修复 {#cl-1.47.1-ext-fix}

- 修复 Dubbo Response 不生效的问题。
- 合并最新分支 v1.47.1


## v1.42.8-ext {#cl-1.42.8-ext}

### 修复 {#cl-1.42.8-ext-fix}

- Response Body 功能添加配置： "dd.trace.response.body.blacklist.urls".

## v1.42.7-ext {#cl-1.42.7-ext}

### 修复 {#cl-1.42.7-ext-fix}

- 修复 Response Body 功能中的环境变量不生效的 Bug
- 合并最新 DDTrace tag v1.42.1 版本

## v1.36.1-ext {#cl-1.36.1-ext}

### 修复 {#cl-1.36.1-ext-fix}

- 合并最新 DataDog Java Agent 分支 1.36.0
- 增加 `dd-ext-version` tag, 方便快速定位版本。
- `mybatis-plus batch` 类执行的 `sql` 语句都没有被记录为 `span` 信息。

## v1.34.2-ext {#cl-1.34.2-ext}

### 修复 {#cl-1.34.2-ext-fix}

- 由于太占用内存，决定移除 [添加 response_body](ddtrace-ext-java.md#response_body) 功能。

## v1.34.0-ext {#cl-1.34.0-ext}

### 更新 {#cl-1.34.0-ext-fix}

- 合并最新 `v1.34.0` 代码。

## v1.30.5-ext v1.30.6-ext {#cl-1.30.5-ext}

### 更新 {#cl-1.30.5-ext-fix}

- 修复 `W3C` 协议下 `trace_id` 提取问题。
- 修复 `Pulsar OOM` 问题。
- `Lettuce5` 集群模式下获取 `peer_ip`.

## v1.30.4-ext (2024/4/25) {#cl-1.30.4-ext}

### 更新 {#cl-1.30.4-ext-fix}

- 解决 `Dubbo` 服务连续传递导致的链路无法中断问题。
- 解决 `Pulsar` 没有释放内存问题。

## v1.30.2-ext (2024/4/3) {#cl-1.30.2-ext}

### 更新 {#cl-1.30.2-ext-fix}

- Redis SDK `Lettuce` 支持查看 `Command` 参数。

## v1.30.1-ext (2024/2/6) {#cl-1.30.1-ext}

### 更新 {#cl-1.30.1-ext-fix}

- 合并最新 DataDog Java Agent 分支 1.30.0.
- 链路数据中添加 HTTP Response Body 信息，[使用命令开启](ddtrace-ext-java.md#response_body)

## v1.25.2-ext (2024/1/10) {#cl-1.25.2-ext}

### 更新 {#cl-1.25.2-ext-fix}

- 链路数据中添加 HTTP Header 信息，[使用命令开启](ddtrace-ext-java.md#trace_header)

## v1.21.1-ext (2023/11/1) {#cl-1.21.1-ext}

### 更新 {#cl-1.21.1-ext-fix}

- 增加 Apache Pulsar 批量消费支持。

## v1.21.0-ext (2023/10/24) {#cl-1.21.0-ext}

### 更新 {#cl-1.21.0-ext-fix}

- 合并最新 DDTrace 分支 v1.21.0 并发布新版本。

## v1.20.3-ext (2023/10/13) {#cl-1.20.3-ext}

### 新增 {#cl-1.20.3-ext-fix}

- 增加 xxl-job 支持 2.2 版本探针。

## v1.20.2-ext (2023/9/25) {#cl-1.20.2-ext}

### 新增 {#cl-1.20.2-ext-fix}

- 增加 Apache Pulsar 探针支持。

## v1.20.1-ext (2023/9/8) {#cl-1.20.1-ext}

### 更新 {#cl-1.20.1-ext-fix}

- 合并最新 DDTrace 分支 v1.20.1 并发布新版本。

## v1.17.4-ext (2023/7/27) {#cl-1.17.4-ext}

### 修复 {#cl-1.17.4-ext-fix}

- 修复 RocketMQ 在高并发中丢失 Span 问题。

## v1.17.2-ext v1.17.3-ext (2023/7/20) {#cl-1.17.3-ext}

### 修复 {#cl-1.17.3-ext-fix}

- 修复 Redis 没有链路信息的问题。
- 去除 Dubbo 中大量的调试日志。
- 增加 4 个 JVM 指标： `jvm.total_thread_count`, `jvm.peak_thread_count`, `jvm.daemon_thread_count`, `jvm.gc.code_cache.used`.

## v1.17.1-ext (2023/7/11) {#cl-1.17.1-ext}

### 修复 {#cl-1.17.1-ext-new}

- RocketMQ 在发送异步消息时返回值会引起 npe 异常。
- RocketMQ 将使用消息本身缓存 span 替换为本地缓存，用户不再需要关闭 traceContext 功能。

### 优化 {#cl-1.17.1-ext-opt}

- 优化日志输出

## v1.17.0-ext (2023/7/7) {#cl-1.17.0-ext}

### 修复 {#cl-1.17.0-ext-new}

- 合并最新的 Datadog v1.17.0 版本


## v1.15.4-ext (2023/6/12) {#cl-1.15.4-ext}

### 修复 {#cl-1.15.4-ext-new}

- 合并最新的 Datadog v1.15.3 版本
- 支持 PowerJob


## v1.14.0-ext (2023/5/18) {#cl-1.14.0-ext}

### 修复 {#cl-1.14.0-ext-new}

- 合并最新的 Datadog v1.14.0 版本
- 支持链路 ID 128 位。


## v1.12.1-ext (2023/5/11) {#cl-1.12.1-ext}

### 修复 {#cl-1.12.1-ext-new}

- 支持 MongoDB 脱敏。
- 支持达梦国产数据库。


## v1.12.0 (2023/4/20) {#cl-1.12.0}

### 修复 {#cl-1.12.0-new}

- 合并最新 DDTrace Tag:1.12.0
- 当当网 Dubbox 支持。
- 解决 jax-rs 与 `Dubbo` 链路产生混淆的问题。
- 解决 `Dubbo` 链路拓扑图顺序不对的问题。
- 解决 RocketMQ 与客户自定义链路数据冲突问题。
- 解决 RocketMQ Resource Name 问题。

## v1.10.2 (2023/4/10) {#cl-1.10.2}

### 修复 {#cl-1.10.2-new}

- 合并最新 DDTrace Tag: 1.10
- 修复 Dubbo 探针不支持 `@DubboReference` 嵌套
- 修复 RocketMQ 链路客户自定义 context 之后获取失败问题

## v1.8.0，v1.8.1，v1.8.3(2023/2/27) {#cl-1.8.0}

### 新加功能 {#cl-1.8.0-new}

- 合并最新 DDTrace 分支
- 增加功能 获取特定函数的入参信息。

## v1.4.1(2023/2/27) {#cl-1.4.1}

### 新加功能 {#cl-1.4.1-new}

- 增加支持阿里云 RocketMQ 4.0 系列

## v1.4.0(2023/1/12) {#cl-1.4.0}

### 新加功能 {#cl-1.4.0-new}

- 合并最新 DDTrace 最新分支 v1.4.0

## v1.3.2(2023/1/12) {#cl-1.3.2}

### 新加功能 {#cl-1.3.2-new}

- 增加 Redis 查看参数功能。
- 修改 DDTrace-Java-Agent 默认端口为 9529 。
- 阿里云 RocketMQ 解决单端为链路问题。

## v1.3.0(2022/12/28) {#cl-1.3.0}

### 新加功能 {#cl-1.3.0-new}

- 合并最新 DataDog 最新分支 v1.3.0
- 增加 Log Patten 支持
- 增加 HSF 框架支持
- 增加 Axis 1.4 支持
- 增加阿里云 RocketMQ 5.0 支持

## v1.0.1(2022/12/23) {#cl-1.0.1}

### 新加功能 {#cl-1.0.1-new}

- 合并最新 DataDog 最新分支 v1.0.1.
- 合并 attach 定制内容。

## v0.113.0-attach(2022/11/16) {#cl-0.113.0}

### 新加功能 {#cl-0.113.0-new}

- 脱敏功能增加 SQL 占位符（`?`）探针支持。

## 0.113.0(2022-10-25) {#cl-0.113.0}

### 功能调整说明 {#cl-0.113.0-new}

- 以 0.113.0 tag 为基准，合并之前的代码

- 修复 Thrift `TMultipexedProtocol` 模型支持

## 0.108.1(2022-10-14) {#cl-0.118.0}

合并 DataDog v0.108.1 版本，进行编译同时保留了 0.108.1


### 功能调整说明 {#cl-0.118.0-new}

- 新增 thrift instrumentation（thrift version >=0.9.3 以上版本）

---

## 0.108.1(2022-09-06) {#cl-0.108.1}

合并 DataDog v0.108.1 版本，进行编译。


### 功能调整说明 {#cl-0.108.1-new}

- 增加 `xxl_job` 探针(`xxl_job` 版本 >= 2.3.0)

---

## 0.107.0((2022-08-30)) {#cl-0.107.0}

合并 DataDog 107 版本，进行编译。

---

## 0.105.0(2022-08-23) {#cl-0.105.0}

### 功能调整说明 {#cl-0.105.0}

- 增加 RocketMq 探针 支持的版本(不低于 4.8.0)。
- 增加 Dubbo 探针 支持的版本(不低于 2.7.0)。
- 增加 SQL 脱敏功能：开启后将原始的 SQL 语句添加到链路中以方便排查问题，启动 Agent 时增加配置参数 `-Ddd.jdbc.sql.obfuscation=true`
