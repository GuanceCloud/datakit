---
title     : 'Tracing Sample'
summary   : '链路采样实践指南'
tags      :
  - 'sample'
  - 'ddtrace'
  - 'otel'
  - '采样'
__int_icon: ''
---


基于 DDTrace 与 OpenTelemetry 的链路采样实践指南。

重点解决多链路串联时的采样问题。

## DDTrace 与 OTel Agent 的采样机制对比 {#sampling-agency}

### DDTrace 采样行为分析 {#ddtrace}

- **采样优先级字段逻辑**  
  在 DDTrace 中，`_sampling_priority_v1` 字段是关键标识：
    - `1`: 默认未配置采样时的标记，表示链路按系统规则保留
    - `2`: 用户显式配置采样后的保留标记（如设置 50% 采样率时命中采样的链路）
    - `-1`: 用户配置规则下需删除的链路（如低优先级业务链路）
    - `0`: 内置规则下需要删除的标记（如上游服务标记为删除，但当前没有开启采样规则）

- **采样决策传递特性**  
  当存在上游服务时（非 DDTrace Agent ），下游的采样配置会失效，决策权移交上游。此特性可能导致链路标签与预期不符

- **采样配置示例**

  ```shell
  -Ddd.trace.sample.rate=0.5
  ```

### OpenTelemetry 采样机制 {#otel}

- **W3C 协议透传特性**  
  OTEL 通过 `traceparent` 头部的 `trace-flags` 传递采样状态：
    - `00`：未采样链路，数据不会上报 DK（DataKit）
    - `01`：已采样链路，数据完整上报

- **采样配置示例**

  ```shell
  #基于父级 TraceID 的 50% 比例采样
  -Dotel.traces.sampler=parentbased_traceidratio -Dotel.traces.sampler.arg=0.5
  ```

OpenTelemetry Agent 发送到 DataKit 的链路已经是采样结束的。 DDTrace 策略是采样在服务端进行。

### 核心差异对比 {#difference}

| 特性                | DDTrace                     | OpenTelemetry          |
|---------------------|-----------------------------|------------------------|
| 数据上报策略        | 全量采集，服务端过滤        | 客户端直接过滤未采样数据 |
| 协议兼容性          | 支持 W3C 但存在字段冲突风险   | 原生支持 W3C 标准        |
| 多级服务控制能力    | 下游采样受上游限制          | 支持分布式决策协同      |

---

## 混合环境下的串联问题与解决方案 {#mixed}

### W3C 透传协议的兼容性问题 {#compatible}

- **字段映射冲突**  
  DDTrace 的 `_sampling_priority_v1` 与 OTEL 的 `trace-flags` 存在语义重叠但取值逻辑差异：
    - `trace-flags` 只有 0,1 两种情况，但 DDTrace 有四种情况。
    - DDTrace 在将 `trace-flags` 的 0,1 转成了 `_sampling_priority_v1` 的 0,1 并不是采样规则下的 -1 和 2 。

- **决策权覆盖风险**  
  当链路中存在 DDTrace 与 OTEL Agent 混用时，会出现：上游强制覆盖下游的采样配置，这是因为上游已经对链路做好了采样标记。


### 数据一致性挑战 {#consistency}

- **采样状态断流**  
  当 OTel Agent 处理过的链路（标记为 `01` ）传递至 DDTrace 服务时，可能出现：
    - DDTrace 新增 Span 的 `_sampling_priority_v1` 被重置为 `1` 这时候 DataKit 再开启采样就会误删数据。
    - DataKit OTEL 采集器开启采样后会导致本应该保留的链路被删除。


混合使用的情况会出现多种串联形式，最佳的配置方案是头部服务开启采样，这样会通过透传协议层级下传。

---

## 推荐配置实践 {#config}

**单类型配置：**
  无论是单个服务还是多个服务串联，都使用 DDTrace Agent 的时候：Agent 端和 DK 端两者只开启一端都可以。

  OTEL 在 Agent 端和 DK 端只能有一端配置采样，否则链路就不完整。



**多个 Agent 级联配置：**
  在多链路级联情况下，推荐使用 **Agent 端头部采样** 和 **DK 端关闭采样**。这样做是为了防止一种情况：当 DDTrace 从 W3C 协议头中转换到自身采样规则的时候
  会在 span 的 `_sampling_priority_v1` 重置为 `1` ，而 `1` 是可以被 DDTrace 采集器的采样策略删除的，这样 就出现了链路丢失 span 的情况。

  串联的协议使用 `tracecontext` 多链路串联的配置和说明，可以查看 [多链路串联](tracing-propagator.md){:target="_blank"}


## 参考 {#docs}

- [W3C 透传协议](https://www.w3.org/TR/trace-context/){:target="_blank"}
- [OpenTelemetry 采样配置](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration){:target="_blank"}
- [DDTrace 采样配置](https://docs.datadoghq.com/tracing/trace_pipeline/ingestion_mechanisms/?tab=java#in-tracing-libraries-user-defined-rules){:target="_blank"}
- [多链路串联](tracing-propagator.md){:target="_blank"}
