---
skip: 'not-searchable-on-index-page'
---

# 更新历史

---

> *作者： 刘锐、宋龙奇*

## 简介 {#intro}

原生 OTEL agent 对部分熟知的主流框架支持不够完善，我们在这个基础上，对其做了一些改进，以支持更多的主流框架和关键的数据追踪。

当前 OTEL 已增加了如下技术栈的扩展：

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: **Java**

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/opentelemetry-javaagent.jar){:target="_blank"}

</div>
<!-- markdownlint-enable -->

> V1 版本已经不再更新，V2 版本已经发布稳定版本。

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

## 2.20.0-ext (2025/9/24) {#cl-2.20.0-ext}

### 新加功能 {#cl-2.20.0-ext-new}

- 合并 open-telemetry V2 版本最新分支
- SQL 脱敏功能与主分支合并


## 1.28.0-ext (2023/7/7) {#cl-1.28.0-ext}

### 新加功能 {#cl-1.28.0-ext-new}

- 合并 open-telemetry 最新分支

---

## 1.26.2-ext (2023/6/15) {#cl-1.26.2-ext}
下载当前版本 jar 包： [v1.26.2-ext](https://static.<<<custom_key.brand_main_domain>>>/dd-image/opentelemetry-javaagent-1.26.2-ext.jar){:target="_blank"}

### 新加功能 {#cl-1.26.2-ext-new}

- 增加 DB 语句脱敏

---

## 1.26.1-ext (2023/6/9) {#cl-1.26.1-ext}

### 新加功能 {#cl-1.26.1-ext-new}

- 非侵入方式支持获取特定方法的入参信息。
- 阿里云 HSF 框架集成。

---

## 1.26.0-ext (2023/6/1) {#cl-1.26.0-ext}

### 新加功能 {#cl-1.26.0-ext-new}

- 合并最新 OpenTelemetry 分支 v1.26.0
- 支持国产达梦数据库。

---

## 1.25.0-ext (2023/5/10) {#cl-1.25.0-ext}

### 新加功能 {#cl-1.25.0-ext-new}

- 合并最新 OpenTelemetry 分支 v1.25.0
- 支持 xxl-job 2.3
- 增加支持阿里巴巴 Dubbo 及 Dubbox 框架支持。
- 支持 thrift。

---
