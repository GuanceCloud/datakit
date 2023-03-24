# changelog

---

> *作者： 刘锐、宋龙奇*

## Introduction {#intro}

Native DDTrace does not support some well-known mainstream frameworks perfectly. On this basis, we have made some improvements to support more mainstream frameworks and key data tracking.

Currently DDTrace has added the following extensions to the technology stack：

<div class="grid cards" markdown>

-   :material-language-java: __Java__

    ---

    [SDK :material-download:](https://static.guance.com/ddtrace/dd-java-agent.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/dd-trace-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/dd-trace-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/dd-trace-java/releases){:target="_blank"}

</div>

## changelog {#changelog}

## v1.8.0，v1.8.1，v1.8.3(2023/2/27) {#cl-1.8.0}

### new {#cl-1.8.0-new}

- meger ddtrace 1.8.0 version.
- add [Get the input parameter information of a specific function](https://github.com/GuanceCloud/dd-trace-java/issues/26){:target="_blank"}

## v1.4.1(2023/2/27) {#cl-1.4.1}

### new {#cl-1.4.1-new}

- Add support for Alibaba Cloud RocketMQ 4.0 series.

## v1.4.0(2023/1/12) {#cl-1.4.0}

### new {#cl-1.4.0-new}

- Merge latest ddtrace latest branch v1.4.0.

## v1.3.2(2023/1/12) {#cl-1.3.2}

### new {#cl-1.3.2-new}

- Add redis view parameter [github-#19](https://github.com/GuanceCloud/dd-trace-java/issues/19){:target="_blank"})
- Modify dd-java-agent default port [github-#18](https://github.com/GuanceCloud/dd-trace-java/issues/18){:target="_blank"})
- Alibaba Cloud RocketMQ bug [github-#22](https://github.com/GuanceCloud/dd-trace-java/issues/22){:target="_blank"})

## v1.3.0(2022/12/28) {#cl-1.3.0}

### new {#cl-1.3.0-new}

- Merge latest datadog latest branch v1.3.0.
- Add log patten support.
- Add hsf framework support.
- Added axis1.4 support.
- Add support for Alibaba Cloud rocketmq 5.0.


## v1.0.1(2022/12/23) {#cl-1.0.1}

### new {#cl-1.0.1-new}

- Merge latest datadog latest branch v1.0.1.
- Merge attach custom content.

## v0.113.0-attach(2022/11/16) {#cl-0.113.0}

### new {#cl-0.113.0-new}

- The desensitization function adds SQL placeholder (`?`) agent support([#7](https://github.com/GuanceCloud/dd-trace-java/issues/7){:target="_blank"})

## 0.113.0(2022-10-25) {#cl-0.113.0}

- [github download](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.113.0-guance){:target="_blank"}

### Function adjustment instructions {#cl-0.113.0-new}
- Based on the 0.113.0 tag, merge the previous code
- Fix thrift TMultipexedProtocol model support


## 0.108.1(2022-10-14) {#cl-0.118.0}

Merge DataDog v0.108.1 version, compile while retaining 0.108.1

- [github download](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1){:target="_blank"}

### Description of function adjustments. {#cl-0.118.0-new}

- add thrift instrumentation（thrift version >=0.9.3）

---

## 0.108.1(2022-09-06) {#cl-0.108.1}

Merge DataDog v0.108.1 and compile it.

- [github download](https://github.com/GuanceCloud/dd-trace-java/releases/tag/v0.108.1)

### 0.108.1 {#cl-0.108.1-new}

- add xxl_job agent ( xxl_job version >= 2.3.0)

---

## guance-0.107.0((2022-08-30)) {#cl-0.107.0}

Merge DataDog 107 version, compile.

- [github download](https://github.com/GuanceCloud/dd-trace-java/releases/tag/guance-107)

---

## guance-0.105.0(2022-08-23) {#cl-0.105.0}

[github download](https://static.guance.com/ddtrace/dd-java-agent-guance-0.106.0-SNAPSHOT.jar)

### new {#cl-0.105.0}

- add RocketMq agent, supported version(not lower than 4.8.0)。
- add Dubbo agent, supported version(not lower than2.7.0)。
- add Sql obfuscation：After opening, add the original sql statement to the link to facilitate troubleshooting, and add configuration parameters when starting the Agent: `-Ddd.jdbc.sql.obfuscation=true`
