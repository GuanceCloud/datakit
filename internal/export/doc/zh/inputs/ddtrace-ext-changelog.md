---
title: 'DDTrace Java 扩展更新日志'
skip: 'not-searchable-on-index-page'
---

<!-- cspell:ignore Tyrus -->

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

## v1.65.7-ext (2026/9/17) {#cl-1.65.7-ext}

### 更新 {#cl-1.65.7-ext-update}

- 新增 `async_entry=true` Span Tag，用于筛选异步线程框架创建的入口 Span。
- 新增 ddtrace 与 SkyWalking 的 K6 压测报告。

## v1.65.6-ext (2026/9/14) {#cl-1.65.6-ext}

### 修复 {#cl-1.65.6-ext-fix}

- 修正 Netty SSE 首个非空正文块的耗时和分片统计。

## v1.65.5-ext (2026/9/10) {#cl-1.65.5-ext}

### 修复 {#cl-1.65.5-ext-fix}

- 修复 Redis 集群 span 使用配置域名命名时回退到实际节点 IP 的问题。

## v1.65.4-ext (2026/9/4) {#cl-1.65.4-ext}

### 更新 {#cl-1.65.4-ext-update}

- 新增 Tyrus 2.x WebSocket 客户端握手链路支持。
- 修复关闭 WebSocket 消息独立链路后，出站 send 和主动 close span 的链路归属问题。

## v1.65.3-ext (2026/8/28) {#cl-1.65.3-ext}

### 修复 {#cl-1.65.3-ext-fix}

- 修复 128 位 Trace ID 在 Datadog HTTP 传播格式中的兼容性问题。

## v1.65.2-ext (2026/8/28) {#cl-1.65.2-ext}

### 更新 {#cl-1.65.2-ext-update}

- 支持 Spring RabbitMQ 批量消息消费和日志关联。
- 新增可选的 CXF Invoker 兜底链路。

## v1.65.0-ext (2026/8/7) {#cl-1.65.0-ext}

### 更新 {#cl-1.65.0-ext-update}

- 合并 DataDog `dd-trace-java` v1.65.0，并加入 GuanceCloud 定制功能。
- 新增 HTTP Header/Body 采集、GWT RPC、JVM 指标、Netty SSE、BES 和 Nacos 等支持。

## v1.63.7-ext (2026/7/1) {#cl-1.63.7-ext}

### 更新 {#cl-1.63.7-ext-fix}

- 支持采集 GWT RPC 请求体标签。
- 新增 JVM 线程状态和原始 GC StatsD 指标。

## v1.63.6-ext (2026/6/29) {#cl-1.63.6-ext}

### 修复 {#cl-1.63.6-ext-fix}

- 修复 BES 探针 root span 丢失问题。

## v1.63.5-ext (2026/6/24) {#cl-1.63.5-ext}

### 更新 {#cl-1.63.5-ext-fix}

- 增强 Spring RabbitMQ 消费链路与日志关联。
- 新增 BES 11.0 应用服务器探针支持。

## v1.63.4-ext (2026/6/10) {#cl-1.63.4-ext}

### 新增 {#cl-1.63.4-ext-fix}

- 新增 Netty Client SSE 响应流 span 及首包延迟、分片数量指标。

## v1.63.3-ext (2026/6/11) {#cl-1.63.3-ext}

### 新增 {#cl-1.63.3-ext-fix}

- 新增从 Redis 客户端连接配置获取 `peer.hostname` 的可选配置。

## v1.63.2-ext (2026/6/8) {#cl-1.63.2-ext}

### 修复 {#cl-1.63.2-ext-fix}

- 修复 JMXFetch 对 `17-ea` 等 Java 版本字符串的识别问题。

## v1.63.1-ext (2026/6/4) {#cl-1.63.1-ext}

### 新增 {#cl-1.63.1-ext-fix}

- 新增 `DD_SERVICE_MAPPING_FILE`，支持从外部文件读取并合并 service mapping。

## v1.63.0-ext (2026/6/3) {#cl-1.63.0-ext}

### 新增 {#cl-1.63.0-ext-fix}

- 合并最新代码。

## v1.60.4-ext (2026/4/27) {#cl-1.60.4-ext}

### 修复 {#cl-1.60.4-ext-fix}

- 修复 Redis 显示 `service_name` 问题。

## v1.60.3-ext (2026/4/24) {#cl-1.60.3-ext}

### 更新 {#cl-1.60.3-ext-fix}

- 优化 JDBC 对 Oracle 的支持。

## v1.55.11-ext (2026/3/17) {#cl-1.55.11-ext}

### 更新 {#cl-1.55.11-ext-fix}

- 将 RocketMQ 探针最低版本从 4.8.0 降至 4.5.0。

## v1.55.10-ext (2026/2/1) {#cl-1.55.10-ext}

### 新增 {#cl-1.55.10-ext-fix}

- 增加 `java-websocket` 探针支持。

## v1.55.7-ext (2025/12/30) {#cl-1.55.7-ext}

### 修复 {#cl-1.55.7-ext-fix}

- 修复 Redis 参数未正确显示的问题。

## v1.55.6-ext (2025/12/22) {#cl-1.55.6-ext}

### 修复 {#cl-1.55.6-ext-fix}

- 修复 RocketMQ scope limit 问题。
- 新增 Response Body 白名单[配置](ddtrace-ext-java.md#response_body)。

## v1.55.5-ext (2025/12/15) {#cl-1.55.5-ext}

### 更新 {#cl-1.55.5-ext-update}

- 合并 DataDog DDTrace v1.55.0，并更新指标集及 Dubbo、RocketMQ、Kingbase 支持。

## v1.55.2-ext (2025/11/28) {#cl-1.55.2-ext}

### 修复 {#cl-1.55.2-ext-fix}

- 修复 RocketMQ scope limit 问题，并合并 v1.55.0。

## v1.53.7-ext (2025/11/28) {#cl-1.53.7-ext}

### 更新 {#cl-1.53.7-ext-fix}

- 支持 Apache Dubbo 3.2 流式调用及 `rocketmq.consume.ignore`。

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
