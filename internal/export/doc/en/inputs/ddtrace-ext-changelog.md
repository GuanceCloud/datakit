---
skip: 'not-searchable-on-index-page'
title: 'DDTrace Extension Changelog'
---

## Introduction {#intro}

Native DDTrace does not support some well-known mainstream frameworks perfectly. On this basis, we have made some improvements to support more mainstream frameworks and key data tracking.

Currently, DDTrace has added the following extensions to the technology stack：
<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: **Java**

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar){:target="_blank"}

</div>
<!-- markdownlint-enable -->

## changelog {#changelog}


## v1.53.1-ext (2025/9/24) {#cl-1.53.1-ext}

### fix {#cl-1.53.1-ext-fix}

- Merge DataDog DDTrace Tag: v1.53.0.


## v1.47.6-ext (2025/6/4) {#cl-1.47.6-ext}

### fix {#cl-1.47.6-ext-fix}

- How to enhance the methods in custom Package and Class by using [command to enable function](ddtrace-ext-java.md#package){:target="_blank"}


## v1.47.5-ext (2025/5/22) {#cl-1.47.5-ext}

### fix {#cl-1.47.5-ext-fix}

- Fix: Pulsar consumer bug.
- Fix: Resource catalog bug.

## v1.47.4-ext (2025/5/14) {#cl-1.47.4-ext}

### fix {#cl-1.47.4-ext-fix}

- Attach to trace methods

## v1.47.1-ext (2025/4/17) {#cl-1.47.1-ext}

### fix {#cl-1.47.1-ext-fix}

- Fix the problem that Dubbo Response does not take effect.
- Merge DDTrace Tag: v1.47.1.


## v1.42.8-ext {#cl-1.42.8-ext}

### fix {#cl-1.42.8-ext-fix}

- Fix Response Body: Add config"dd.trace.response.body.blacklist.urls".

## v1.42.7-ext {#cl-1.42.7-ext}

### fix {#cl-1.42.7-ext-fix}

- Fix Response Body ENV Bug
- Merge DDTrace tag v1.42.1 version.

## v1.36.1-ext {#cl-1.36.1-ext}

### fix {#cl-1.36.1-ext-fix}

- Merge DataDog Java Agent tag 1.36.0.
- Add `dd-ext-version` tag.
- Using `mybatis plus`, the SQL statements executed by the batch class are not recorded as span information.

## v1.34.2-ext {#cl-1.34.2-ext}

### fix {#cl-1.34.2-ext-fix}

- Due to excessive memory usage, it has been decided to remove the [add response body](ddtrace-ext-java.md#response_body) feature.

## v1.34.0-ext {#cl-1.34.0-ext}

### new {#cl-1.34.0-ext-fix}

- Merge DataDog v1.34.0

## v1.30.5-ext v1.30.6-ext {#cl-1.30.5-ext}

### fix {#cl-1.30.5-ext-fix}

- Fixed `trace_id` extraction problem under `W3C` protocol.
- Fix `Pulsar OOM` issue.
- `Lettuce5` obtains `peer_ip` in cluster mode.

## v1.30.4-ext (2024/4/25) {#cl-1.30.4-ext}

### fix {#cl-1.30.4-ext-fix}

- Solve the problem that the link cannot be interrupted due to continuous delivery of `Dubbo` service.
- Solve the problem of `Pulsar` not releasing memory.

## v1.30.2-ext (2024/4/3) {#cl-1.30.2-ext}

### fix {#cl-1.30.2-ext-fix}

- Redis SDK `Lettuce` supports viewing `Command` parameters.

## v1.30.1-ext (2024/2/6) {#cl-1.30.1-ext}

### new {#cl-1.30.1-ext-fix}

- Merge DDTrace tag  1.30.0.
- To add HTTP Response Body information in the trace data, [the command to enable it is](ddtrace-ext-java.md#response_body)

## v1.25.2-ext (2024/1/10) {#cl-1.25.2-ext}

### new {#cl-1.25.2-ext-fix}

- By using the environment `dd.trace.headers.enabled=true`, the `header` information can be placed in the `span` tag `servlet_request_header`.

## v1.21.1-ext (2023/11/1) {#cl-1.21.1-ext}

### fix {#cl-1.21.1-ext-fix}

- Add Apache Pulsar consumer batch instructions.

## v1.21.0-ext (2023/10/24) {#cl-1.21.0-ext}

### fix {#cl-1.21.0-ext-fix}

- Merge DDTrace tag v1.21.0.

## v1.20.3-ext (2023/10/13) {#cl-1.20.3-ext}

### add {#cl-1.20.3-ext-fix}

- Add `xxl-job` 2.2 version.

## v1.20.2-ext (2023/9/25) {#cl-1.20.2-ext}

### add {#cl-1.20.2-ext-fix}

- Add Apache Pulsar instructions.

## v1.20.1-ext (2023/9/8) {#cl-1.20.1-ext}

### fix {#cl-1.20.1-ext-fix}

- Merge DDTrace tag v1.20.1 and release new version.

## v1.17.4-ext (2023/7/27) {#cl-1.17.4-ext}

### fix {#cl-1.17.4-ext-fix}

- Fix RocketMQ send span bug.

## v1.17.2-ext v1.17.3-ext (2023/7/20) {#cl-1.17.3-ext}

### fix {#cl-1.17.3-ext-fix}

- Fix bug for Redis not has Spans.
- Delete Info logging of Dubbo.
- Add 4 个 JVM metric:`jvm.total_thread_count`, `jvm.peak_thread_count`, `jvm.daemon_thread_count`, `jvm.gc.code_cache.used`.

## v1.17.1-ext (2023/7/11) {#cl-1.17.1-ext}

### fix {#cl-1.17.1-ext-new}

- RocketMQ returns a value when sending an asynchronous message, which can cause an NPE.
- RocketMQ will replace the message itself cache span with local cache, and users no longer need to turn off the traceContext function.

## v1.17.0-ext (2023/7/7) {#cl-1.17.0-ext}

### fix {#cl-1.17.0-ext-new}

- Merge Datadog v1.17.0.


## v1.15.4-ext (2023/6/12) {#cl-1.15.4-ext}

### new {#cl-1.15.4-ext-new}

- Merge Datadog v1.15.3 tag
- Support PowerJob.


## v1.14.0-ext (2023/5/18) {#cl-1.14.0-ext}

### fix {#cl-1.14.0-ext-new}

- Merge Datadog v1.14.0 version.
- Support `trace-128-bit-id`.


## v1.12.1-ext (2023/5/11) {#cl-1.12.1-ext}

### fix {#cl-1.12.1-ext-new}

- Supported MongoDB obfuscation.
- Supported DM8.


## v1.12.0 (2023/4/20) {#cl-1.10.2}

### fix {#cl-1.12.0-new}

- Merge ddtrace tag:1.12.0.
- Support `DangDang Dubbox`.
- Solve the confusion between jax-rs and Dubbo traces.
- Solve the problem that the order of Dubbo trace topology map is wrong bug.
- Solve the conflict between RocketMQ and customer-defined trace data bug.
- Modify RocketMQ resource name.

## v1.10.2 (2023/4/10) {#cl-1.10.2}

### fix {#cl-1.10.2-new}

- merge ddtrace tag:1.10.
- Fix Dubbo probe does not support @ DubboReference nesting.
- Fixed the issue of failed retrieval of RocketMQ link custom context.

## v1.8.0，v1.8.1，v1.8.3(2023/2/27) {#cl-1.8.0}

### new {#cl-1.8.0-new}

- merge ddtrace 1.8.0 version.
- add `Get the input parameter information of a specific function`.

## v1.4.1(2023/2/27) {#cl-1.4.1}

### new {#cl-1.4.1-new}

- Add support for Alibaba Cloud RocketMQ 4.0 series.

## v1.4.0(2023/1/12) {#cl-1.4.0}

### new {#cl-1.4.0-new}

- Merge latest ddtrace latest branch v1.4.0.

## v1.3.2(2023/1/12) {#cl-1.3.2}

### new {#cl-1.3.2-new}

- Add redis view parameter.
- Modify `dd-java-agent` default port is 9529.
- Alibaba Cloud RocketMQ bug.

## v1.3.0(2022/12/28) {#cl-1.3.0}

### new {#cl-1.3.0-new}

- Merge latest DataDog latest branch v1.3.0.
- Add log patten support.
- Add hsf framework support.
- Added axis1.4 support.
- Add support for Alibaba Cloud RocketMQ 5.0.

## v1.0.1(2022/12/23) {#cl-1.0.1}

### new {#cl-1.0.1-new}

- Merge latest DataDog latest branch v1.0.1.
- Merge attach custom content.

## v0.113.0-attach(2022/11/16) {#cl-0.113.0}

### new {#cl-0.113.0-new}

- The desensitization function adds SQL placeholder (`?`) agent support.

## 0.113.0(2022-10-25) {#cl-0.113.0}

### Function adjustment instructions {#cl-0.113.0-new}

- Based on the 0.113.0 tag, merge the previous code
- Fix thrift `TMultipexedProtocol` model support


## 0.108.1(2022-10-14) {#cl-0.118.0}

Merge DataDog v0.108.1 version, compile while retaining 0.108.1

### Description of function adjustments. {#cl-0.118.0-new}

- add thrift instrumentation（thrift version >=0.9.3）

---

## 0.108.1(2022-09-06) {#cl-0.108.1}

Merge DataDog v0.108.1 and compile it.

### 0.108.1 {#cl-0.108.1-new}

- add `xxl_job` agent ( `xxl_job` version >= 2.3.0)

---

## 0.107.0-ext((2022-08-30)) {#cl-0.107.0}

Merge DataDog 107 version, compile.

---

## 0.105.0-ext(2022-08-23) {#cl-0.105.0}

### new {#cl-0.105.0}

- add RocketMq agent, supported version(not lower than 4.8.0)。
- add Dubbo agent, supported version(not lower than2.7.0)。
- add SQL obfuscation：After opening, add the original SQL statement to the link to facilitate troubleshooting, and add configuration parameters when starting the Agent: `-Ddd.jdbc.sql.obfuscation=true`
