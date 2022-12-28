# 更新历史

---

> *作者： 刘锐、宋龙奇*

## 简介 {#intro}

原生 DDTrace 对部分熟知的主流框架支持不够完善，我们在这个基础上，对其做了一些改进，以支持更多的主流框架和关键的数据追踪。

当前 DDTrace 已增加了如下技术栈的扩展：

<div class="grid cards" markdown>

-   :material-language-java: __Java__

    ---

    [SDK :material-download:](https://static.guance.com/ddtrace/dd-java-agent.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/dd-trace-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/dd-trace-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/dd-trace-java/releases){:target="_blank"}

</div>

## 更新历史 {#changelog}

<!--

更新历史可以参考 datakit 的基本范式：

## 1.2.3(2022/12/12) {#cl-1.2.3}
本次发布主要有如下更新：

### 新加功能 {#cl-1.2.3-new}
### 问题修复 {#cl-1.2.3-fix}
### 功能优化 {#cl-1.2.3-opt}
### 兼容调整 {#cl-1.2.3-brk} 

--->

## v1.3.0(2022/12/28) {#cl-1.3.0}

### 新加功能 {#cl-1.3.0-new}

- 合并最新 datadog 最新分支 v1.3.0.
- 增加log patten支持.
- 增加 hsf 框架支持.
- 增加 axis1.4支持.
- 增加阿里云 rocketmq 5.0 支持.


## v1.0.1(2022/12/23) {#cl-1.0.1}

### 新加功能 {#cl-1.0.1-new}

- 合并最新 datadog 最新分支 v1.0.1.
- 合并 attach 定制内容.

## v0.113.0-attach(2022/11/16) {#cl-0.113.0}

### 新加功能 {#cl-0.113.0-new}

- 脱敏功能增加 SQL 占位符（`?`）探针支持([#7](https://github.com/GuanceCloud/dd-trace-java/issues/7){:target="_blank"})

## 0.113.0(2022-10-25) {#cl-0.113.0}

- [github下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.113.0-guance){:target="_blank"}

### 功能调整说明 {#cl-0.113.0-new}

- 以0.113.0 tag 为基准，合并之前的代码

- 修复thrift TMultipexedProtocol 模型支持


## 0.108.1(2022-10-14) {#cl-0.118.0}

合并 DataDog v0.108.1版本，进行编译同时保留了0.108.1

- [github 下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1){:target="_blank"}

### 功能调整说明 {#cl-0.118.0-new}

- 新增 thrift instrumentation（thrift version >=0.9.3 以上版本）

---

## 0.108.1(2022-09-06) {#cl-0.108.1}

合并 DataDog v0.108.1版本，进行编译。

- [github下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1)

### 功能调整说明 {#cl-0.108.1-new}

- 增加 xxl_job 探针( xxl_job 版本 >= 2.3.0)

---

## guance-0.107.0((2022-08-30)) {#cl-0.107.0}

合并 DataDog 107 版本，进行编译。

- [github下载地址](https://github.com/GuanceCloud/dd-trace-java/releases/tag/guance-107)

---

## guance-0.105.0(2022-08-23) {#cl-0.105.0}

[github下载地址](https://static.guance.com/ddtrace/dd-java-agent-guance-0.106.0-SNAPSHOT.jar)

### 功能调整说明 {#cl-0.105.0}

- 增加 RocketMq 探针 支持的版本(不低于4.8.0)。
- 增加 Dubbo 探针 支持的版本(不低于2.7.0)。
- 增加 Sql 脱敏功能：开启后将原始的 sql 语句添加到链路中以方便排查问题，启动 Agent 时增加配置参数 `-Ddd.jdbc.sql.obfuscation=true`

