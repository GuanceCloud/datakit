---
title: 'DDTrace Java 扩展更新日志'
skip: 'not-searchable-on-index-page'
---

> *作者： 刘锐、宋龙奇*

## 简介 {#intro}

原生 DDTrace 对部分熟知的主流框架支持不够完善，我们在这个基础上，对其做了一些改进，以支持更多的主流框架和关键的数据追踪。

当前 DDTrace 已增加了如下技术栈的扩展：

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: __Java__

    ---

    [SDK :material-download:](https://static.guance.com/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/dd-trace-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/dd-trace-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/dd-trace-java/releases){:target="_blank"}

</div>
<!-- markdownlint-enable -->

## 更新历史 {#changelog}

<!--

更新历史可以参考 Datakit 的基本范式：

## 1.2.3(2022/12/12) {#cl-1.2.3}
本次发布主要有如下更新：

### 新加功能 {#cl-1.2.3-new}
### 问题修复 {#cl-1.2.3-fix}
### 功能优化 {#cl-1.2.3-opt}
### 兼容调整 {#cl-1.2.3-brk}

--->

## v1.42.7-guance {#cl-1.42.7-guance}

### 修复 {#cl-1.42.7-guance-fix}

- 修复 Response Body 功能中的环境变量不生效的 Bug
- 合并最新 DDTrace tag v1.42.1 版本

## v1.36.1-guance {#cl-1.36.1-guance}

### 修复 {#cl-1.36.1-guance-fix}

- 合并最新 DataDog Java Agent 分支 1.36.0
- 增加 `dd-guance-version` tag, 方便快速定位版本。
- `mybatis-plus batch` 类执行的 `sql` 语句都没有被记录为 `span` 信息。

## v1.34.2-guance {#cl-1.34.2-guance}

### 修复 {#cl-1.34.2-guance-fix}

- 由于太占用内存，决定移除 [添加 response_body](ddtrace-ext-java.md#response_body) 功能。

## v1.34.0-guance {#cl-1.34.0-guance}

### 更新 {#cl-1.34.0-guance-fix}

- 合并最新 `v1.34.0` 代码。

## v1.30.5-guance v1.30.6-guance {#cl-1.30.5-guance}

### 更新 {#cl-1.30.5-guance-fix}

- 修复 `W3C` 协议下 `trace_id` 提取问题。
- 修复 `Pulsar OOM` 问题。
- `Lettuce5` 集群模式下获取 `peer_ip`.

## v1.30.4-guance (2024/4/25) {#cl-1.30.4-guance}

### 更新 {#cl-1.30.4-guance-fix}

- 解决 `Dubbo` 服务连续传递导致的链路无法中断问题。
- 解决 `Pulsar` 没有释放内存问题。

## v1.30.2-guance (2024/4/3) {#cl-1.30.2-guance}

### 更新 {#cl-1.30.2-guance-fix}

- Redis SDK `Lettuce` 支持查看 `Command` 参数。

## v1.30.1-guance (2024/2/6) {#cl-1.30.1-guance}

### 更新 {#cl-1.30.1-guance-fix}

- 合并最新 DataDog Java Agent 分支 1.30.0.
- 链路数据中添加 HTTP Response Body 信息，[使用命令开启](ddtrace-ext-java.md#response_body)

## v1.25.2-guance (2024/1/10) {#cl-1.25.2-guance}

### 更新 {#cl-1.25.2-guance-fix}

- 链路数据中添加 HTTP Header 信息，[使用命令开启](ddtrace-ext-java.md#trace_header)

## v1.25.1-guance (2024/1/4) {#cl-1.25.1-guance}

### 更新 {#cl-1.25.1-guance-fix}

- 将 `Guance_trace_id` 放置在 HTTP 响应体的头部信息中。

## v1.21.1-guance (2023/11/1) {#cl-1.21.1-guance}

### 更新 {#cl-1.21.1-guance-fix}

- 增加 Apache Pulsar 批量消费支持。

## v1.21.0-guance (2023/10/24) {#cl-1.21.0-guance}

### 更新 {#cl-1.21.0-guance-fix}

- 合并最新 DDTrace 分支 v1.21.0 并发布新版本。

## v1.20.3-guance (2023/10/13) {#cl-1.20.3-guance}

### 新增 {#cl-1.20.3-guance-fix}

- 增加 xxl-job 支持 2.2 版本探针。

## v1.20.2-guance (2023/9/25) {#cl-1.20.2-guance}

### 新增 {#cl-1.20.2-guance-fix}

- 增加 Apache Pulsar 探针支持。

## v1.20.1-guance (2023/9/8) {#cl-1.20.1-guance}

### 更新 {#cl-1.20.1-guance-fix}

- 合并最新 DDTrace 分支 v1.20.1 并发布新版本。

## v1.17.4-guance (2023/7/27) {#cl-1.17.4-guance}

### 修复 {#cl-1.17.4-guance-fix}

- 修复 RocketMQ 在高并发中丢失 Span 问题。

## v1.17.2-guance v1.17.3-guance (2023/7/20) {#cl-1.17.3-guance}

### 修复 {#cl-1.17.3-guance-fix}

- 修复 Redis 没有链路信息的问题。
- 去除 Dubbo 中大量的调试日志。
- 增加 4 个 JVM 指标，详细请查看 [GitHub-Issue](https://github.com/GuanceCloud/dd-trace-java/issues/46){:target="_blank"}

## v1.17.1-guance (2023/7/11) {#cl-1.17.1-guance}

### 修复 {#cl-1.17.1-guance-new}

- RocketMQ 在发送异步消息时返回值会引起 npe 异常。
- RocketMQ 将使用消息本身缓存 span 替换为本地缓存，用户不再需要关闭 traceContext 功能。

### 优化 {#cl-1.17.1-guance-opt}

- 优化日志输出

## v1.17.0-guance (2023/7/7) {#cl-1.17.0-guance}

### 修复 {#cl-1.17.0-guance-new}

- 合并最新的 Datadog v1.17.0 版本


## v1.15.4-guance (2023/6/12) {#cl-1.15.4-guance}

### 修复 {#cl-1.15.4-guance-new}

- 合并最新的 Datadog v1.15.3 版本
- [支持 PowerJob](https://github.com/GuanceCloud/dd-trace-java/issues/42){:target="_blank"}


## v1.14.0-guance (2023/5/18) {#cl-1.14.0-guance}

### 修复 {#cl-1.14.0-guance-new}

- 合并最新的 Datadog v1.14.0 版本
- [支持链路 ID 128 位](https://github.com/GuanceCloud/dd-trace-java/issues/37){:target="_blank"}


## v1.12.1-guance (2023/5/11) {#cl-1.12.1-guance}

### 修复 {#cl-1.12.1-guance-new}

- 支持 MongoDB 脱敏， [MongoDB 脱敏问题](https://github.com/GuanceCloud/dd-trace-java/issues/38){:target="_blank"}
- [支持达梦国产数据库](https://github.com/GuanceCloud/dd-trace-java/issues/39){:target="_blank"}


## v1.12.0 (2023/4/20) {#cl-1.12.0}

### 修复 {#cl-1.12.0-new}

- 合并最新 DDTrace Tag:1.12.0
- 当当网 [Dubbox 支持](https://github.com/GuanceCloud/dd-trace-java/issues/32){:target="_blank"}
- 解决 [jax-rs 与 Dubbo 链路产生混淆的问题](https://github.com/GuanceCloud/dd-trace-java/issues/34){:target="_blank"}
- 解决 [Dubbo 链路拓扑图顺序不对的问题](https://github.com/GuanceCloud/dd-trace-java/issues/35){:target="_blank"}
- 解决 [RocketMQ 与客户自定义链路数据冲突问题](https://github.com/GuanceCloud/dd-trace-java/issues/29){:target="_blank"}
- 解决 [RocketMQ Resource Name 问题](https://github.com/GuanceCloud/dd-trace-java/issues/33){:target="_blank"}

## v1.10.2 (2023/4/10) {#cl-1.10.2}

### 修复 {#cl-1.10.2-new}

- 合并最新 DDTrace Tag: 1.10
- 修复 Dubbo 探针不支持 `@DubboReference` 嵌套
- 修复 RocketMQ 链路客户自定义 context 之后获取失败问题

## v1.8.0，v1.8.1，v1.8.3(2023/2/27) {#cl-1.8.0}

### 新加功能 {#cl-1.8.0-new}

- 合并最新 DDTrace 分支
- 增加功能 [获取特定函数的入参信息](https://github.com/GuanceCloud/dd-trace-java/issues/26){:target="_blank"}

## v1.4.1(2023/2/27) {#cl-1.4.1}

### 新加功能 {#cl-1.4.1-new}

- 增加支持阿里云 RocketMQ 4.0 系列

## v1.4.0(2023/1/12) {#cl-1.4.0}

### 新加功能 {#cl-1.4.0-new}

- 合并最新 DDTrace 最新分支 v1.4.0

## v1.3.2(2023/1/12) {#cl-1.3.2}

### 新加功能 {#cl-1.3.2-new}

- 增加 Redis [查看参数功能](https://github.com/GuanceCloud/dd-trace-java/issues/19){:target="_blank"})
- 修改 DDTrace-Java-Agent [默认端口](https://github.com/GuanceCloud/dd-trace-java/issues/18){:target="_blank"})
- 阿里云 RocketMQ 单端为链路[问题修正](https://github.com/GuanceCloud/dd-trace-java/issues/22){:target="_blank"})

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

- 脱敏功能增加 SQL 占位符（`?`）探针支持([#7](https://github.com/GuanceCloud/dd-trace-java/issues/7){:target="_blank"})

## 0.113.0(2022-10-25) {#cl-0.113.0}

- [GitHub 下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.113.0-guance){:target="_blank"}

### 功能调整说明 {#cl-0.113.0-new}

- 以 0.113.0 tag 为基准，合并之前的代码

- 修复 Thrift `TMultipexedProtocol` 模型支持

## 0.108.1(2022-10-14) {#cl-0.118.0}

合并 DataDog v0.108.1 版本，进行编译同时保留了 0.108.1

- [GitHub 下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1){:target="_blank"}

### 功能调整说明 {#cl-0.118.0-new}

- 新增 thrift instrumentation（thrift version >=0.9.3 以上版本）

---

## 0.108.1(2022-09-06) {#cl-0.108.1}

合并 DataDog v0.108.1 版本，进行编译。

- [GitHub 下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1){:target="_blank"}

### 功能调整说明 {#cl-0.108.1-new}

- 增加 xxl_job 探针( xxl_job 版本 >= 2.3.0)

---

## guance-0.107.0((2022-08-30)) {#cl-0.107.0}

合并 DataDog 107 版本，进行编译。

- [GitHub 下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/guance-107){:target="_blank"}

---

## guance-0.105.0(2022-08-23) {#cl-0.105.0}

[GitHub 下载地址](https://static.guance.com/ddtrace/dd-java-agent-guance-0.106.0-SNAPSHOT.jar){:target="_blank"}

### 功能调整说明 {#cl-0.105.0}

- 增加 RocketMq 探针 支持的版本(不低于 4.8.0)。
- 增加 Dubbo 探针 支持的版本(不低于 2.7.0)。
- 增加 SQL 脱敏功能：开启后将原始的 SQL 语句添加到链路中以方便排查问题，启动 Agent 时增加配置参数 `-Ddd.jdbc.sql.obfuscation=true`
