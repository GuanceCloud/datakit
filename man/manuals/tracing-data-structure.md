# Tracing Data Struct

## 简述

此文用于解释主流 Telemetry 平台数据结构以及与 Datakit 平台数据结构的映射关系。
包括: DataDog Tracing, Jaeger Tracing, Skywalking, Zipking

---

## Datakit Point Protocol Structure for Tracing

### Datakit Line Protocol

Name | Tags | Fields | Timestamp

> 以下简称 dkproto

### Datakit Tracing Span Structure

> 以下简称 dtspan

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| ContainerHost                                   | string                                         |                                            | 容器主机名                                       | dkproto.tags.container_host                     |
| Content                                         | string                                         |                                            | Span 原始数据                                    | dkproto.fields.message                          |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dkproto.fields.duration                         |
| EndPoint                                        | string                                         |                                            | 通讯对端                                         | dkproto.tags.endpoint                           |
| Env                                             | string                                         |                                            | 环境变量                                         | dkproto.tags.env                                |
| HTTPMethod                                      | string                                         |                                            | HTTP Method                                      | dkproto.tags.http_method                        |
| HTTPStatusCode                                  | string                                         |                                            | HTTP Response Status Code                        | dkproto.tags.http_status_code                   |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dkproto.tags.operation                          |
| ParentID                                        | string                                         |                                            | Parent Span ID                                   | dkproto.fields.parent_id                        |
| Project                                         | string                                         |                                            | 项目名                                           | dkproto.tags.project                            |
| Resource                                        | string                                         |                                            | 资源名                                           | dkproto.fields.resource                         |
| ServiceName                                     | string                                         |                                            | 服务名                                           | dkproto.tags.service                            |
| Source                                          | string                                         |                                            | Span 生产者                                      | dkproto.name                                    |
| SpanID                                          | string                                         |                                            | Span ID                                          | dkproto.fields.span_id                          |
| SpanType                                        | string                                         |                                            | Span Type                                        | dkproto.tags.span_type                          |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                    | dkproto.fields.start                            |
| Status                                          | string                                         |                                            | Span 状态字段                                    | dkproto.tags.status                             |
| Tags                                            | map[string, string]                            |                                            | Span Tags                                        | dkproto.tags                                    |
| TraceID                                         | string                                         |                                            | Trace ID                                         | dkproto.fields.trace_id                         |
| Type                                            | string                                         |                                            | Span Type                                        | dkproto.tags.span_type                          |
| Version                                         | string                                         |                                            | 版本号                                           | dkproto.tags.version                            |

---

## DDTrace Trace&Span Structures

### DDTrace Trace Structure

DataDog 里 Trace 代表一个 Span 的数组结构

> trace: span[]

### DDTrace Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                                                                                                    |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dkproto.fields.duration                                                                                                                            |
| Error                                           | int32                                          |                                            | Span 状态字段 0:无报错 1:出错                    | dkproto.tags.status                                                                                                                                |
| Meta                                            | map[string, string]                            |                                            | Span 过程元数据                                  | dkproto.tags.project, dkproto.tags.env, dkproto.tags.version, dkproto.tags.container_host, dkproto.tags.http_method, dkproto.tags.http_status_code |
| Metrics                                         | map[string, float64]                           |                                            | Span 过程需要参与运算数据例如采样                | 无直接对应关系                                                                                                                                     |
| Name                                            | string                                         |                                            | 生产此条 Span 的方法名                           | dkproto.tags.operation                                                                                                                             |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dkproto.fields.parent_id                                                                                                                           |
| Resource                                        | string                                         |                                            | 资源名                                           | dkproto.fields.resource                                                                                                                            |
| Service                                         | string                                         |                                            | 服务名                                           | dkproto.tags.service                                                                                                                               |
| SpanID                                          | uint64                                         |                                            | Span ID                                          | dkproto.fields.span_id                                                                                                                             |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                    | dkproto.fields.start                                                                                                                               |
| TraceID                                         | uint64                                         |                                            | Trace ID                                         | dkproto.fields.trace_id                                                                                                                            |
| Type                                            | string                                         |                                            | Span Type                                        | dkproto.tags.span_type                                                                                                                             |

---

## Jaeger Tracing Data Structure

### Jaeger Thrift Protocol Batch Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Process                                         | struct pointer                                 |                                            | 进程相关项                                       | dkproto.tags.service                            |
| SeqNo                                           | int64 pointer                                  |                                            | 序列号                                           | 无直接对应关系                                  |
| Spans                                           | array                                          |                                            | Span 数组结构                                    | 多重对应关系                                    |
| Stats                                           | struct pointer                                 |                                            | 客户端统计结构                                   | 无直接对应关系                                  |

### Jaeger Thrift Protocol Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dkproto.fields.duration                         |
| Flags                                           | int32                                          |                                            | Span Flags                                       | 无直接对应关系                                  |
| Logs                                            | array                                          |                                            | Span Logs                                        | 无直接对应关系                                  |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dkproto.tags.operation                          |
| ParentSpanId                                    | int64                                          |                                            | Parent Span ID                                   | dkproto.fields.parent_id                        |
| References                                      | array                                          |                                            | Span References                                  | 无直接对应关系                                  |
| SpanId                                          | int64                                          |                                            | Span ID                                          | dkproto.fields.span_id                          |
| StartTime                                       | int64                                          | 纳秒                                       | Span 起始时间                                    | dkproto.fields.start                            |
| Tags                                            | array                                          |                                            | Span Tags 目前只取 Span 状态字段                 | dkproto.tags.status                             |
| TraceIdHigh                                     | int64                                          |                                            | Trace ID 高位 TraceIdLow 组成 Trace ID           | dkproto.fields.trace_id                         |
| TraceIdLow                                      | int64                                          |                                            | Trace ID 低位与 TraceIdHigh 组成 Trace ID        | dkproto.fields.trace_id                         |

---

## Skywalking Tracing Data Structure

### Skywalking Segment Object Generated By Proto Buffer Protocol V3

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                            |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------- |
| IsSizeLimited                                   | bool                                           |                                            | 是否包含连路上所有 Span                          | 未使用字段                                                                 |
| Service                                         | string                                         |                                            | 服务名                                           | dkproto.tags.service                                                       |
| ServiceInstance                                 | string                                         |                                            | 借点逻辑关系名                                   | 未使用字段                                                                 |
| Spans                                           | array                                          |                                            | Tracing Span 数组                                | 对应关系见下表                                                             |
| TraceId                                         | string                                         |                                            | Trace ID                                         | dkproto.fields.trace_id                                                    |
| TraceSegmentId                                  | string                                         |                                            | Segment ID                                       | 与 Span ID 一起使用唯一标志一个 Span, 对应 dkproto.fields.span_id 中的高位 |

### Skywalking Span Object Structure in Segment Object

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                                        |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------------------- |
| ComponentId                                     | int32                                          |                                            | 第三方框架数值化定义                             | 未使用字段                                                                             |
| EndTime                                         | int64                                          | 毫秒                                       | Span 结束时间                                    | EndTime 减去 StartTime 对应 dkproto.fields.duration                                    |
| IsError                                         | bool                                           |                                            | Span 状态字段                                    | dkproto.tags.status                                                                    |
| Logs                                            | array                                          |                                            | Span Logs                                        | 未使用字段                                                                             |
| OperationName                                   | string                                         |                                            | Span Operation Name                              | dkproto.tags.operation                                                                 |
| ParentSpanId                                    | int32                                          |                                            | Parent Span ID                                   | 与 Segment ID 一起使用唯一标志一个 Parent Span, 对应 dkproto.fields.parent_id 中的低位 |
| Peer                                            | string                                         |                                            | 通讯对端                                         | dkproto.tags.endpoint                                                                  |
| Refs                                            | array                                          |                                            | 跨线程跨进程情况下存储 Parent Segment            | ParentTraceSegmentId 对应 dkproto.fields.span_id 中的高位                              |
| SkipAnalysis                                    | bool                                           |                                            | 跳过后端分析                                     | 未使用字段                                                                             |
| SpanId                                          | int32                                          |                                            | Span ID                                          | 与 Segment ID 一起使用唯一标志一个 Span, 对应 dkproto.fields.span_id 中的低位          |
| SpanLayer                                       | int32                                          |                                            | Span 技术栈数值化定义                            | 未使用字段                                                                             |
| SpanType                                        | int32                                          |                                            | Span Type 数值化定义                             | dkproto.tags.span_type                                                                 |
| StartTime                                       | int64                                          | 毫秒                                       | Span 起始时间                                    | dkproto.fields.start                                                                   |
| Tags                                            | array                                          |                                            | Span Tags                                        | 未使用字段                                                                             |

---

## Zipking Tracing Data Structure

### Zipkin Thrift Protocol Span Structure V1

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Annotations                                     | array                                          |                                            | 参与获取 Service Name                            | dkproto.tags.service                            |
| BinaryAnnotations                               | array                                          |                                            | 参与获取 Span 状态字段                           | dkproto.tags.status                             |
| Debug                                           | bool                                           |                                            | Debug 状态字段                                   | 未使用字段                                      |
| Duration                                        | uint64                                         | 微秒                                       | Span 耗时                                        | dkproto.fields.duration                         |
| ID                                              | uint64                                         |                                            | Span ID                                          | dkproto.fields.span_id                          |
| Name                                            | string                                         |                                            | Span Operation Name                              | dkproto.tags.operation                          |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dkproto.fields.parent_id                        |
| Timestamp                                       | uint64                                         | 微秒                                       | Span 起始时间                                    | dkproto.fields.start                            |
| TraceID                                         | uint64                                         |                                            | Trace ID                                         | dkproto.fields.trace_id                         |
| TraceIDHigh                                     | uint64                                         |                                            | Trace ID 高位                                    | 无直接对应关系                                  |

### Zipkin Span Structure V2

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| TraceID                                         | struct                                         |                                            | Trace ID                                         | dkproto.fields.trace_id                         |
| ID                                              | uint64                                         |                                            | Span ID                                          | dkproto.fields.span_id                          |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dkproto.fields.parent_id                        |
| Debug                                           | bool                                           |                                            | Debug 状态                                       | 未使用字段                                      |
| Sampled                                         | bool                                           |                                            | 采样状态字段                                     | 未使用字段                                      |
| Err                                             | string                                         |                                            | Error Message                                    | 无直接对应关系                                  |
| Name                                            | string                                         |                                            | Span Operation Name                              | dkproto.tags.operation                          |
| Kind                                            | string                                         |                                            | Span Type                                        | dkproto.tags.span_type                          |
| Timestamp                                       | struct                                         | 微秒                                       | 微秒级时间结构表示 Span 起始时间                 | dkproto.fields.start                            |
| Duration                                        | int64                                          | 微秒                                       | Span 耗时                                        | dkproto.fields.duration                         |
| Shared                                          | bool                                           |                                            | 共享状态                                         | 未使用字段                                      |
| LocalEndpoint                                   | struct                                         |                                            | 用于获取 Service Name                            | dkproto.tags.service                            |
| RemoteEndpoint                                  | struct                                         |                                            | 通讯对端                                         | dkproto.tags.endpoint                           |
| Annotations                                     | array                                          |                                            | 用于解释延迟相关的事件                           | 未使用字段                                      |
| Tags                                            | map                                            |                                            | 用于获取 Span 状态                               | dkproto.tags.status                             |
