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
| EndPoint                                        | string                                         |                                            | 点对点通讯中的终端                               | dpp.tags.endpoint                               |
| Env                                             | string                                         |                                            | 环境变量                                         | dpp.tags.env                                    |
| HTTPMethod                                      | string                                         |                                            | HTTP Method                                      | dpp.tags.http_method                            |
| HTTPStatusCode                                  | string                                         |                                            | HTTP Response Status Code                        | dpp.tags.http_status_code                       |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                              |
| ParentID                                        | string                                         |                                            | Span Parent ID                                   | dpp.fields.parent_id                            |
| Project                                         | string                                         |                                            | 项目名                                           | dpp.tags.project                                |
| Resource                                        | string                                         |                                            | 资源名                                           | dpp.fields.resource                             |
| ServiceName                                     | string                                         |                                            | 服务名                                           | dpp.tags.service                                |
| Source                                          | string                                         |                                            | Span 生产者                                      | dpp.name                                        |
| SpanID                                          | string                                         |                                            | Span ID                                          | dpp.fields.span_id                              |
| SpanType                                        | string                                         |                                            | Span Type                                        | dpp.tags.span_type                              |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                    | dpp.fields.start                                |
| Status                                          | string                                         |                                            | 状态字段                                         | dpp.tags.status                                 |
| Tags                                            | map[string, string]                            |                                            | Span Tags                                        | dpp.tags                                        |
| TraceID                                         | string                                         |                                            | Trace ID                                         | dpp.fields.trace_id                             |
| Type                                            | string                                         |                                            | Span Type                                        | dpp.tags.span_type                              |
| Version                                         | string                                         |                                            | 版本号                                           | dpp.tags.version                                |

---

## DDTrace Trace&Span Structures

### DDTrace Trace Structure

DataDog 里 Trace 结构是一个 Span 的数组

> trace: span[]

### DDTrace Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span>                                                                            |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dpp.fields.duration                                                                                                        |
| Error                                           | int32                                          |                                            | 0:无报错 1:出错 状态字段                         | dpp.tags.status                                                                                                            |
| Meta                                            | map[string, string]                            |                                            | Span 过程元数据                                  | dpp.tags.project, dpp.tags.env, dpp.tags.version, dpp.tags.container_host, dpp.tags.http_method, dpp.tags.http_status_code |
| Metrics                                         | map[string, float64]                           |                                            | Span 过程需要参与运算数据例如采样                | 没有显式的对应关系                                                                                                         |
| Name                                            | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                                                                                                         |
| ParentID                                        | uint64                                         |                                            | Span Parent ID                                   | dpp.fields.parent_id                                                                                                       |
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
| SeqNo                                           | int64 pointer                                  |                                            | 序列号                                           | 没有显式的对应关系                              |
| Spans                                           | array                                          |                                            | Span 相关结构                                    | 多重对应关系                                    |
| Stats                                           | struct pointer                                 |                                            | 客户端统计结构                                   | 没有显式的对应关系                              |

### Jaeger Thrift Protocol Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------- |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dpp.fields.duration                             |
| Flags                                           | int32                                          |                                            | Span Flags                                       | 没有显式的对应关系                              |
| Logs                                            | array                                          |                                            | Span Logs                                        | 没有显式的对应关系                              |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dpp.tags.operation                              |
| ParentSpanId                                    | int64                                          |                                            | Span Parent ID                                   | dpp.fields.parent_id                            |
| References                                      | array                                          |                                            | Span References                                  | 没有显式的对应关系                              |
| SpanId                                          | int64                                          |                                            | Span ID                                          | dpp.fields.span_id                              |
| StartTime                                       | int64                                          | 纳秒                                       | Span 起始时间                                    | dpp.fields.start                                |
| Tags                                            | array                                          |                                            | Span Tags 目前只取 Span 状态字段                 | dpp.tags.status                                 |
| TraceIdHigh                                     | int64                                          |                                            | Trace ID 高位 TraceIdLow 组成 Trace ID           | dpp.fields.trace_id                             |
| TraceIdLow                                      | int64                                          |                                            | Trace ID 低位与 TraceIdHigh 组成 Trace ID        | dpp.fields.trace_id                             |

## Skywalking Tracing Data Structure

---

## Zipking Tracing Data Structure
