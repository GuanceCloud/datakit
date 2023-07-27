# 更新历史

---

> *作者： 刘锐、宋龙奇*

## 简介 {#intro}

原生 OTEL agent 对部分熟知的主流框架支持不够完善，我们在这个基础上，对其做了一些改进，以支持更多的主流框架和关键的数据追踪。

当前 OTEL 已增加了如下技术栈的扩展：

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: __Java__

    ---

    [SDK :material-download:](https://static.guance.com/dd-image/opentelemetry-javaagent.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/releases){:target="_blank"}

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

## 1.28.0-guance (2023/7/7) {#cl-1.28.0-guance}

### 新加功能 {#cl-1.28.0-guance-new}

- 合并 open-telemetry 最新分支

---

## 1.26.3-guance (2023/7/7) {#cl-1.26.3-guance}

### 新加功能 {#cl-1.26.3-guance-new}

- 新增 [guance-exporter](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/17){:target="_blank"}

---

## 1.26.2-guance (2023/6/15) {#cl-1.26.2-guance}
下载当前版本 jar 包： [v1.26.2-guance](https://static.guance.com/dd-image/opentelemetry-javaagent-1.26.2-guance.jar){:target="_blank"}

### 新加功能 {#cl-1.26.2-guance-new}

- [增加 DB 语句脱敏](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/15){:target="_blank"}

---

## 1.26.1-guance (2023/6/9) {#cl-1.26.1-guance}

### 新加功能 {#cl-1.26.1-guance-new}

- 非侵入方式支持获取特定方法的入参信息 [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/12){:target="_blank"}
- 阿里云 HSF 框架集成 [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/12){:target="_blank"}

---

## 1.26.0-guance (2023/6/1) {#cl-1.26.0-guance}

### 新加功能 {#cl-1.26.0-guance-new}

- 合并最新 OpenTelemetry 分支 v1.26.0
- 支持国产达梦数据库 [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/5){:target="_blank"}

---

## 1.25.0-guance (2023/5/10) {#cl-1.25.0-guance}

### 新加功能 {#cl-1.25.0-guance-new}

- 合并最新 OpenTelemetry 分支 v1.25.0
- 支持 xxl-job 2.3 [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/1){:target="_blank"}
- 增加支持阿里巴巴 Dubbo 及 Dubbox 框架支持 [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/2){:target="_blank"}
- 支持 thrift [GitHub-Issue](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/issues/3){:target="_blank"}

---
