# Tracing Data Struct

## 简述

此文用于解释主流 Telemetry 平台数据结构以及与 Datakit 平台数据结构的映射关系。
包括: DataDog Tracing, Jaeger Tracing, Skywalking, Zipking###

---

## Datakit Point Protocol Structure for Tracing

### Datakit Point Protocol

Name | Tags | Fields | Timestamp

> 以下简称 dpp

### Datakit Tracing Structure

> 以下简称 dts

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| ContainerHost                                   | string                                         |                                            | 容器主机名                                       | dpp.tags.container_host                         |
| Content                                         | string                                         |                                            | Span 原始数据                                    | dpp.fields.message                              |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dpp.fields.duration                             |
| EndPoint                                        | string                                         |                                            | 通讯对端                                         | dpp.tags.endpoint                               |
| Env                                             | string                                         |                                            | 环境变量                                         | dpp.tags.env                                    |
| HTTPMethod                                      | string                                         |                                            | HTTP Method                                      | dpp.tags.http_method                            |
| HTTPStatusCode                                  | string                                         |                                            | HTTP Response Status Code                        | dpp.tags.http_status_code                       |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                              |
| ParentID                                        | string                                         |                                            | Parent Span ID                                   | dpp.fields.parent_id                            |
| Project                                         | string                                         |                                            | 项目名                                           | dpp.tags.project                                |
| Resource                                        | string                                         |                                            | 资源名                                           | dpp.fields.resource                             |
| ServiceName                                     | string                                         |                                            | 服务名                                           | dpp.tags.service                                |
| Source                                          | string                                         |                                            | Span 生产者                                      | dpp.name                                        |
| SpanID                                          | string                                         |                                            | Span ID                                          | dpp.fields.span_id                              |
| SpanType                                        | string                                         |                                            | Span Type                                        | dpp.tags.span_type                              |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                    | dpp.fields.start                                |
| Status                                          | string                                         |                                            | Span 状态字段                                    | dpp.tags.status                                 |
| Tags                                            | map[string, string]                            |                                            | Span Tags                                        | dpp.tags                                        |
| TraceID                                         | string                                         |                                            | Trace ID                                         | dpp.fields.trace_id                             |
| Type                                            | string                                         |                                            | Span Type                                        | dpp.tags.span_type                              |
| Version                                         | string                                         |                                            | 版本号                                           | dpp.tags.version                                |

---

## DDTrace Trace&Span Structures

### DDTrace Trace Structure

DataDog 里 Trace 代表一个 Span 的数组结构

> trace: span[]

### DDTrace Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                                                                            |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dpp.fields.duration                                                                                                        |
| Error                                           | int32                                          |                                            | Span 状态字段 0:无报错 1:出错                    | dpp.tags.status                                                                                                            |
| Meta                                            | map[string, string]                            |                                            | Span 过程元数据                                  | dpp.tags.project, dpp.tags.env, dpp.tags.version, dpp.tags.container_host, dpp.tags.http_method, dpp.tags.http_status_code |
| Metrics                                         | map[string, float64]                           |                                            | Span 过程需要参与运算数据例如采样                | 无直接对应关系                                                                                                             |
| Name                                            | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                                                                                                         |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dpp.fields.parent_id                                                                                                       |
| Resource                                        | string                                         |                                            | 资源名                                           | dpp.fields.resource                                                                                                        |
| Service                                         | string                                         |                                            | 服务名                                           | dpp.tags.service                                                                                                           |
| SpanID                                          | uint64                                         |                                            | Span ID                                          | dpp.fields.span_id                                                                                                         |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                    | dpp.fields.start                                                                                                           |
| TraceID                                         | uint64                                         |                                            | Trace ID                                         | dpp.fields.trace_id                                                                                                        |
| Type                                            | string                                         |                                            | Span Type                                        | dpp.tags.span_type                                                                                                         |

---

## Jaeger Tracing Data Structure

### Jaeger Thrift Protocol Batch Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Process                                         | struct pointer                                 |                                            | 进程相关项                                       | dpp.tags.service                                |
| SeqNo                                           | int64 pointer                                  |                                            | 序列号                                           | 无直接对应关系                                  |
| Spans                                           | array                                          |                                            | Span 数组结构                                    | 多重对应关系                                    |
| Stats                                           | struct pointer                                 |                                            | 客户端统计结构                                   | 无直接对应关系                                  |

### Jaeger Thrift Protocol Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dpp.fields.duration                             |
| Flags                                           | int32                                          |                                            | Span Flags                                       | 无直接对应关系                                  |
| Logs                                            | array                                          |                                            | Span Logs                                        | 无直接对应关系                                  |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                              |
| ParentSpanId                                    | int64                                          |                                            | Parent Span ID                                   | dpp.fields.parent_id                            |
| References                                      | array                                          |                                            | Span References                                  | 无直接对应关系                                  |
| SpanId                                          | int64                                          |                                            | Span ID                                          | dpp.fields.span_id                              |
| StartTime                                       | int64                                          | 纳秒                                       | Span 起始时间                                    | dpp.fields.start                                |
| Tags                                            | array                                          |                                            | Span Tags 目前只取 Span 状态字段                 | dpp.tags.status                                 |
| TraceIdHigh                                     | int64                                          |                                            | Trace ID 高位 TraceIdLow 组成 Trace ID           | dpp.fields.trace_id                             |
| TraceIdLow                                      | int64                                          |                                            | Trace ID 低位与 TraceIdHigh 组成 Trace ID        | dpp.fields.trace_id                             |

---

## Skywalking Tracing Data Structure

### Skywalking Segment Object Generated By Proto Buffer Protocol V3

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                        |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ---------------------------------------------------------------------- |
| IsSizeLimited                                   | bool                                           |                                            | 是否包含连路上所有 Span                          | 未使用字段                                                             |
| Service                                         | string                                         |                                            | 服务名                                           | dpp.tags.service                                                       |
| ServiceInstance                                 | string                                         |                                            | 借点逻辑关系名                                   | 未使用字段                                                             |
| Spans                                           | array                                          |                                            | Tracing Span 数组                                | 对应关系见下表                                                         |
| TraceId                                         | string                                         |                                            | Trace ID                                         | dpp.fields.trace_id                                                    |
| TraceSegmentId                                  | string                                         |                                            | Segment ID                                       | 与 Span ID 一起使用唯一标志一个 Span, 对应 dpp.fields.span_id 中的高位 |

### Skywalking Span Object Structure in Segment Object

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                                    |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ---------------------------------------------------------------------------------- |
| ComponentId                                     | int32                                          |                                            | 第三方框架数值化定义                             | 未使用字段                                                                         |
| EndTime                                         | int64                                          | 毫秒                                       | Span 结束时间                                    | EndTime 减去 StartTime 对应 dpp.fields.duration                                    |
| IsError                                         | bool                                           |                                            | Span 状态字段                                    | dpp.tags.status                                                                    |
| Logs                                            | array                                          |                                            | Span Logs                                        | 未使用字段                                                                         |
| OperationName                                   | string                                         |                                            | Span Operation Name                              | dpp.tags.operation                                                                 |
| ParentSpanId                                    | int32                                          |                                            | Parent Span ID                                   | 与 Segment ID 一起使用唯一标志一个 Parent Span, 对应 dpp.fields.parent_id 中的低位 |
| Peer                                            | string                                         |                                            | 通讯对端                                         | dpp.tags.endpoint                                                                  |
| Refs                                            | array                                          |                                            | 跨线程跨进程情况下存储 Parent Segment            | ParentTraceSegmentId 对应 dpp.fields.span_id 中的高位                              |
| SkipAnalysis                                    | bool                                           |                                            | 跳过后端分析                                     | 未使用字段                                                                         |
| SpanId                                          | int32                                          |                                            | Span ID                                          | 与 Segment ID 一起使用唯一标志一个 Span, 对应 dpp.fields.span_id 中的低位          |
| SpanLayer                                       | int32                                          |                                            | Span 技术栈数值化定义                            | 未使用字段                                                                         |
| SpanType                                        | int32                                          |                                            | Span Type 数值化定义                             | dpp.tags.span_type                                                                 |
| StartTime                                       | int64                                          | 毫秒                                       | Span 起始时间                                    | dpp.fields.start                                                                   |
| Tags                                            | array                                          |                                            | Span Tags                                        | 未使用字段                                                                         |

---

## Zipking Tracing Data Structure

### Zipkin Thrift Protocol Span Structure V1

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Annotations                                     | array                                          |                                            | 参与获取 Service Name                            | dpp.tags.service                                |
| BinaryAnnotations                               | array                                          |                                            | 参与获取 Span 状态字段                           | dpp.tags.status                                 |
| Debug                                           | bool                                           |                                            | Debug 状态字段                                   | 未使用字段                                      |
| Duration                                        | uint64                                         | 微秒                                       | Span 耗时                                        | dpp.fields.duration                             |
| ID                                              | uint64                                         |                                            | Span ID                                          | dpp.fields.span_id                              |
| Name                                            | string                                         |                                            | Span Operation Name                              | dpp.tags.operation                              |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dpp.fields.parent_id                            |
| Timestamp                                       | uint64                                         | 微秒                                       | Span 起始时间                                    | dpp.fields.start                                |
| TraceID                                         | uint64                                         |                                            | Trace ID                                         | dpp.fields.trace_id                             |
| TraceIDHigh                                     | uint64                                         |                                            | Trace ID 高位                                    | 无直接对应关系                                  |

### Zipkin Span Structure V2

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| TraceID                                         | struct                                         |                                            | Trace ID                                         | dpp.fields.trace_id                             |
| ID                                              | uint64                                         |                                            | Span ID                                          | dpp.fields.span_id                              |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dpp.fields.parent_id                            |
| Debug                                           | bool                                           |                                            | Debug 状态                                       | 未使用字段                                      |
| Sampled                                         | bool                                           |                                            | 采样状态字段                                     | 未使用字段                                      |
| Err                                             | string                                         |                                            | Error Message                                    | 无直接对应关系                                  |
| Name                                            | string                                         |                                            | Span Operation Name                              | dpp.tags.operation                              |
| Kind                                            | string                                         |                                            | Span Type                                        | dpp.tags.span_type                              |
| Timestamp                                       | struct                                         | 微秒                                       | 微秒级时间结构表示 Span 起始时间                 | dpp.fields.start                                |
| Duration                                        | int64                                          | 微秒                                       | Span 耗时                                        | dpp.fields.duration                             |
| Shared                                          | bool                                           |                                            | 共享状态                                         | 未使用字段                                      |
| LocalEndpoint                                   | struct                                         |                                            | 用于获取 Service Name                            | dpp.tags.service                                |
| RemoteEndpoint                                  | struct                                         |                                            | 通讯对端                                         | dpp.tags.endpoint                               |
| Annotations                                     | array                                          |                                            | 用于解释延迟相关的事件                           | 未使用字段                                      |
| Tags                                            | map                                            |                                            | 用于获取 Span 状态                               | dpp.tags.status                                 |
