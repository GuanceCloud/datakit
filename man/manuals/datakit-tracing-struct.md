# Tracing Data Struct

## 简述

此文用于解释主流 Telemetry 平台数据结构以及与 Datakit 平台数据结构的映射关系。
目前支持数据结构：DataDog，Jaeger，OpenTelemetry，Skywalking，Zipkin

数据转换流：

> 外部 Tracing 数据结构 --> Datakit Span --> Line Protocol

---

## Datakit Point Protocol Structure for Tracing

### Datakit Line Protocol

行协议数据结构是由 Name, Tags, Fields, Timestamp 四部分和分隔符 (英文逗号，空格) 组成的字符串，形如：

```example
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

> 以下简称 dkproto

### Datakit Tracing Span Structure

Datakit Span 是 Datakit 内部使用的数据结构。第三方 Tracing Agent 数据结构会转换成 Datakit Span 结构后发送到数据中心。

> 以下简称 dkspan

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span>                                        | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------- | -------------------------------------------------- |
| TraceID                                         | string                                         |                                            | Trace ID                                                                                | dkproto.fields.trace_id                            |
| ParentID                                        | string                                         |                                            | Parent Span ID                                                                          | dkproto.fields.parent_id                           |
| SpanID                                          | string                                         |                                            | Span ID                                                                                 | dkproto.fields.span_id                             |
| Service                                         | string                                         |                                            | Service Name                                                                            | dkproto.tags.service                               |
| Resource                                        | string                                         |                                            | Resource Name(.e.g /get/data/from/some/api)                                             | dkproto.fields.resource                            |
| Operation                                       | string                                         |                                            | 生产此条 Span 的方法名                                                                  | dkproto.tags.operation                             |
| Source                                          | string                                         |                                            | Span 接入源(.e.g ddtrace)                                                               | dkproto.name                                       |
| SpanType                                        | string                                         |                                            | Span Type(.e.g Entry)                                                                   | dkproto.tags.span_type                             |
| SourceType                                      | string                                         |                                            | Span Source Type(.e.g Web)                                                              | dkproto.tags.type                                  |
| Env                                             | string                                         |                                            | Environment Variables                                                                   | dkproto.tags.env                                   |
| Project                                         | string                                         |                                            | App 项目名                                                                              | dkproto.tags.project                               |
| Version                                         | string                                         |                                            | App 版本号                                                                              | dkproto.tags.version                               |
| Tags                                            | map[string, string]                            |                                            | Span Tags                                                                               | dkproto.tags                                       |
| EndPoint                                        | string                                         |                                            | 通讯对端                                                                                | dkproto.tags.endpoint                              |
| HTTPMethod                                      | string                                         |                                            | HTTP Method                                                                             | dkproto.tags.http_method                           |
| HTTPStatusCode                                  | string                                         |                                            | HTTP Response Status Code(.e.g 200)                                                     | dkproto.tags.http_status_code                      |
| ContainerHost                                   | string                                         |                                            | 容器主机名                                                                              | dkproto.tags.container_host                        |
| PID                                             | string                                         |                                            | Process ID                                                                              | dkproto.                                           |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                                                           | dkproto.fields.start                               |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                                                                    | dkproto.fields.duration                            |
| Status                                          | string                                         |                                            | Span 状态字段                                                                           | dkproto.tags.status                                |
| Content                                         | string                                         |                                            | Span 原始数据                                                                           | dkproto.fields.message                             |
| Priority                                        | int                                            |                                            | Span 上报优先级 -1:reject 0:auto consider with sample rate 1:always send to data center | dkproto.fields.priority                            |
| SamplingRateGlobal                              | float64                                        |                                            | Global Sampling Rate                                                                    | dkproto.fields.sampling_rate_global                |

---

## DDTrace Trace&Span Structures

### DDTrace Trace Structure

DataDog Trace Struct

> Trace: []\*span

DataDog Traces Struct

> Traces: []Trace

### DDTrace Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span>   | <span style="color:green">**Correspond To**</span>                                                         |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | -------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| TraceID                                         | uint64                                         |                                            | Trace ID                                           | dkspan.TraceID                                                                                             |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                     | dkspan.ParentID                                                                                            |
| SpanID                                          | uint64                                         |                                            | Span ID                                            | dkspan.SpanID                                                                                              |
| Service                                         | string                                         |                                            | 服务名                                             | dkspan.Service                                                                                             |
| Resource                                        | string                                         |                                            | 资源名                                             | dkspan.Resource                                                                                            |
| Name                                            | string                                         |                                            | 生产此条 Span 的方法名                             | dkspan.Operation                                                                                           |
| Start                                           | int64                                          | 纳秒                                       | Span 起始时间                                      | dkspan.Start                                                                                               |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                               | dkspan.Duration                                                                                            |
| Error                                           | int32                                          |                                            | Span 状态字段 0:无报错 1:出错                      | dkspan.Status                                                                                              |
| Meta                                            | map[string, string]                            |                                            | Span 过程元数据，环境相关和服务相关 field 从此获得 | dkspan.Project, dkspan.Env, dkspan.Version, dkspan.ContainerHost, dkspan.HTTPMethod, dkspan.HTTPStatusCode |
| Metrics                                         | map[string, float64]                           |                                            | Span 采样，运算相关数据                            | 不直接对应 dkspan                                                                                          |
| Type                                            | string                                         |                                            | Span Type                                          | dkspan.SourceType                                                                                          |

---

## OpenTelemetry Tracing Data Structure

datakit 采集从 OpenTelemetry exporter:Otlp 中发送上来的数据时，简略的原始数据通过 json 序列化之后，如下所示：

```text
resource_spans:{
    resource:{
        attributes:{key:"message.type"  value:{string_value:"message-name"}}
        attributes:{key:"service.name"  value:{string_value:"test-name"}}
    }
    instrumentation_library_spans:{instrumentation_library:{name:"test-tracer"}
    spans:{
        trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"
        span_id:".\xbd\x06c\x10ɫ*"
        parent_span_id:"\xa7*\x80Z#\xbeL\xf6"
        name:"Sample-0"
        kind:SPAN_KIND_INTERNAL
        start_time_unix_nano:1644312397453313100
        end_time_unix_nano:1644312398464865900
        status:{}
    }
    spans:{
           ...
        }
}

```

otel 中的 `resource_spans` 和 dkspan 的对应关系 如下：

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span>                                                                                                                |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| trace_id                                        | [16]byte                                       |                                            | Trace ID                                         | dkspan.TraceID                                                                                                                                                    |
| span_id                                         | [8]byte                                        |                                            | Span ID                                          | dkspan.SpanID                                                                                                                                                     |
| parent_span_id                                  | [8]byte                                        |                                            | Parent Span ID                                   | dkspan.ParentID                                                                                                                                                   |
| name                                            | string                                         |                                            | Span Name                                        | dkspan.Operation                                                                                                                                                  |
| kind                                            | string                                         |                                            | Span Type                                        | dkspan.SpanType                                                                                                                                                   |
| start_time_unix_nano                            | int64                                          | 纳秒                                       | Span 起始时间                                    | dkspan.Start                                                                                                                                                      |
| end_time_unix_nano                              | int64                                          | 纳秒                                       | Span 终止时间                                    | dkspan.Duration = end - start                                                                                                                                     |
| status                                          | string                                         |                                            | Span Status                                      | dkspan.Status                                                                                                                                                     |
| name                                            | string                                         |                                            | resource Name                                    | dkspan.Resource                                                                                                                                                   |
| resource.attributes                             | map[string]string                              |                                            | resource 标签                                     | dkspan.tags.service, dkspan.tags.project, dkspan.tags.env, dkspan.tags.version, dkspan.tags.container_host, dkspan.tags.http_method, dkspan.tags.http_status_code |
| span.attributes                                 | map[string]string                              |                                            | Span 标签                                        | dkspan.tags                                                                                                                                                       |

otel 有些独有字段， 但 DKspan 没有字段与之对应，所以就放在了标签中，只有这些值非 0 时才会显示，如：

| Field | Date Type | Uint | Description |Correspond |
| :--- | :--- | :--- | :--- | :--- |
| span.dropped_attributes_count   | int   |     | Span 被删除的标签数量      | dkspan.tags.dropped_attributes_count                                                                                                                              |
| span.dropped_events_count       | int   |     | Span 被删除的事件数量      | dkspan.tags.dropped_events_count                                                                                                                                  |
| span.dropped_links_count        | int   |     | Span 被删除的连接数量      | dkspan.tags.dropped_links_count                                                                                                                                   |
| span.events_count               | int   |     | Span 关联事件数量          | dkspan.tags.events_count                                                                                                                                          |
| span.links_count                | int   |     | Span 所关联的 span 数量    | dkspan.tags.links_count                                                                                                                                           |

---

## Jaeger Tracing Data Structure

### Jaeger Thrift Protocol Batch Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| Process                                         | struct pointer                                 |                                            | 进程相关数据结构                                 | dkspan.Service                                     |
| SeqNo                                           | int64 pointer                                  |                                            | 序列号                                           | 不接对应关系 dkspan                                |
| Spans                                           | array                                          |                                            | Span 数组结构                                    | 见下表                                             |
| Stats                                           | struct pointer                                 |                                            | 客户端统计结构                                   | 不直接对应 dkspan                                  |

### Jaeger Thrift Protocol Span Structure

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| TraceIdHigh                                     | int64                                          |                                            | Trace ID 高位与 TraceIdLow 组成 Trace ID         | dkspan.TraceID                                     |
| TraceIdLow                                      | int64                                          |                                            | Trace ID 低位与 TraceIdHigh 组成 Trace ID        | dkspan.TraceID                                     |
| ParentSpanId                                    | int64                                          |                                            | Parent Span ID                                   | dkspan.ParentID                                    |
| SpanId                                          | int64                                          |                                            | Span ID                                          | dkspan.SpanID                                      |
| OperationName                                   | string                                         |                                            | 生产此条 Span 的方法名                           | dkspan.Operation                                   |
| Flags                                           | int32                                          |                                            | Span Flags                                       | 不直接对应 dkspan                                  |
| Logs                                            | array                                          |                                            | Span Logs                                        | 不直接对应 dkspan                                  |
| References                                      | array                                          |                                            | Span References                                  | 不直接对应 dkspan                                  |
| StartTime                                       | int64                                          | 纳秒                                       | Span 起始时间                                    | dkspan.Start                                       |
| Duration                                        | int64                                          | 纳秒                                       | 耗时                                             | dkspan.Duration                                    |
| Tags                                            | array                                          |                                            | Span Tags 目前只取 Span 状态字段                 | dkspan.Status                                      |

---

## Skywalking Tracing Data Structure

### Skywalking Segment Object Generated By Proto Buffer Protocol V3

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| TraceId                                         | string                                         |                                            | Trace ID                                         | dkspan.TraceID                                     |
| TraceSegmentId                                  | string                                         |                                            | Segment ID 与 Span ID 一起使用唯一标志一个 Span  | dkspan.SpanID 高位                                 |
| Service                                         | string                                         |                                            | 服务名                                           | dkspan.Service                                     |
| ServiceInstance                                 | string                                         |                                            | 节点逻辑关系名                                   | 未使用字段                                         |
| Spans                                           | array                                          |                                            | Tracing Span 数组                                | 见下表                                             |
| IsSizeLimited                                   | bool                                           |                                            | 是否包含连路上所有 Span                          | 未使用字段                                         |

### Skywalking Span Object Structure in Segment Object

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span>              | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------- | -------------------------------------------------- |
| ComponentId                                     | int32                                          |                                            | 第三方框架数值化定义                                          | 未使用字段                                         |
| Refs                                            | array                                          |                                            | 跨线程跨进程情况下存储 Parent Segment                         | dkspan.ParentID 高位                               |
| ParentSpanId                                    | int32                                          |                                            | Parent Span ID 与 Segment ID 一起使用唯一标志一个 Parent Span | dkspan.ParentID 低位                               |
| SpanId                                          | int32                                          |                                            | Span ID 与 Segment ID 一起使用唯一标志一个 Span               | dkspan.SpanID 低位                                 |
| OperationName                                   | string                                         |                                            | Span Operation Name                                           | dkspan.Operation                                   |
| Peer                                            | string                                         |                                            | 通讯对端                                                      | dkspan.Endpoint                                    |
| IsError                                         | bool                                           |                                            | Span 状态字段                                                 | dkspan.Status                                      |
| SpanType                                        | int32                                          |                                            | Span Type 数值化定义                                          | dkspan.SpanType                                    |
| StartTime                                       | int64                                          | 毫秒                                       | Span 起始时间                                                 | dkspan.Start                                       |
| EndTime                                         | int64                                          | 毫秒                                       | Span 结束时间与 StartTime 相减代表耗时                        | dkspan.Duration                                    |
| Logs                                            | array                                          |                                            | Span Logs                                                     | 未使用字段                                         |
| SkipAnalysis                                    | bool                                           |                                            | 跳过后端分析                                                  | 未使用字段                                         |
| SpanLayer                                       | int32                                          |                                            | Span 技术栈数值化定义                                         | 未使用字段                                         |
| Tags                                            | array                                          |                                            | Span Tags                                                     | 未使用字段                                         |

---

## Zipkin Tracing Data Structure

### Zipkin Thrift Protocol Span Structure V1

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| TraceIDHigh                                     | uint64                                         |                                            | Trace ID 高位                                    | 无直接对应关系                                     |
| TraceID                                         | uint64                                         |                                            | Trace ID                                         | dkspan.TraceID                                     |
| ID                                              | uint64                                         |                                            | Span ID                                          | dkspan.SpanID                                      |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dkspan.ParentID                                    |
| Annotations                                     | array                                          |                                            | 获取 Service Name                                | dkspan.Service                                     |
| Name                                            | string                                         |                                            | Span Operation Name                              | dkspan.Operation                                   |
| BinaryAnnotations                               | array                                          |                                            | 获取 Span 状态字段                               | dkspan.Status                                      |
| Timestamp                                       | uint64                                         | 微秒                                       | Span 起始时间                                    | dkspan.Start                                       |
| Duration                                        | uint64                                         | 微秒                                       | Span 耗时                                        | dkspan.Duration                                    |
| Debug                                           | bool                                           |                                            | Debug 状态字段                                   | 未使用字段                                         |

### Zipkin Span Structure V2

| <span style="color:green">**Field Name**</span> | <span style="color:green">**Data Type**</span> | <span style="color:green"> **Unit**</span> | <span style="color:green">**Description**</span> | <span style="color:green">**Correspond To**</span> |
| ----------------------------------------------- | ---------------------------------------------- | ------------------------------------------ | ------------------------------------------------ | -------------------------------------------------- |
| TraceID                                         | struct                                         |                                            | Trace ID                                         | dkspan.TraceID                                     |
| ID                                              | uint64                                         |                                            | Span ID                                          | dkspan.SpanID                                      |
| ParentID                                        | uint64                                         |                                            | Parent Span ID                                   | dkspan.ParentID                                    |
| Name                                            | string                                         |                                            | Span Operation Name                              | dkspan.Operation                                   |
| Debug                                           | bool                                           |                                            | Debug 状态                                       | 未使用字段                                         |
| Sampled                                         | bool                                           |                                            | 采样状态字段                                     | 未使用字段                                         |
| Err                                             | string                                         |                                            | Error Message                                    | 不直接对应 dkspan                                  |
| Kind                                            | string                                         |                                            | Span Type                                        | dkspan.SpanType                                    |
| Timestamp                                       | struct                                         | 微秒                                       | 微秒级时间结构表示 Span 起始时间                 | dkspan.Start                                       |
| Duration                                        | int64                                          | 微秒                                       | Span 耗时                                        | dkspan.Duration                                    |
| Shared                                          | bool                                           |                                            | 共享状态                                         | 未使用字段                                         |
| LocalEndpoint                                   | struct                                         |                                            | 用于获取 Service Name                            | dkspan.Service                                     |
| RemoteEndpoint                                  | struct                                         |                                            | 通讯对端                                         | dkspan.Endpoint                                    |
| Annotations                                     | array                                          |                                            | 用于解释延迟相关的事件                           | 未使用字段                                         |
| Tags                                            | map                                            |                                            | 用于获取 Span 状态                               | dkspan.Status                                      |
