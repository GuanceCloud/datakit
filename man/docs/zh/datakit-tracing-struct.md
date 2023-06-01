
# Datakit Tracing 数据结构

## 简述 {#intro}

此文用于解释主流 Telemetry 平台数据结构以及与 Datakit 平台数据结构的映射关系
目前支持数据结构：DataDog/Jaeger/OpenTelemetry/SkyWalking/Zipkin/PinPoint

数据转换步骤：

1. 外部 Tracing 数据结构接入
1. Datakit Span 转换
1. span 数据运算
1. Line Protocol 转换

---

## Datakit Point Protocol 数据结构 {#point-proto}

Line Protocol 数据结构是由 Name, Tags, Fields, Timestamp 四部分和分隔符 (英文逗号，空格) 组成的字符串，形如：

``` not-set
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

> 以下简称 DKProto

| Section | Name             | Unit | Description                                                                                         |
| ---     | ---              | ---  | ---                                                                                                 |
| Tag     | container_host   |      | host name of container                                                                              |
| Tag     | endpoint         |      | end point of resource                                                                               |
| Tag     | env              |      | environment arguments                                                                               |
| Tag     | http_host        |      | HTTP host                                                                                           |
| Tag     | http_method      |      | HTTP method                                                                                         |
| Tag     | http_route       |      | HTTP route                                                                                          |
| Tag     | http_status_code |      | HTTP status code                                                                                    |
| Tag     | http_url         |      | HTTP URL                                                                                            |
| Tag     | operation        |      | operation of resource                                                                               |
| Tag     | pid              |      | process id                                                                                          |
| Tag     | project          |      | project name                                                                                        |
| Tag     | service          |      | service name                                                                                        |
| Tag     | source_type      |      | source types [app, framework, cache, message_queue, custom, db, web]                                |
| Tag     | status           |      | span status [ok, info, warning, error, critical]                                                    |
| Tag     | span_type        |      | span types [entry, local, exit, unknown]                                                            |
| Field   | duration         | 微秒 | span duration                                                                                       |
| Field   | message          |      | raw data content                                                                                    |
| Field   | parent_id        |      | parent ID of span                                                                                   |
| Field   | priority         |      | priority rules (PRIORITY_USER_REJECT, PRIORITY_AUTO_REJECT, PRIORITY_AUTO_KEEP, PRIORITY_USER_KEEP) |
| Field   | resource         |      | resource of service                                                                                 |
| Field   | sample_rate      |      | global sampling ratio (0.1 means roughly 10 percent will send to data center)                       |
| Field   | span_id          |      | span ID                                                                                             |
| Field   | start            | 微秒 | span start timestamp                                                                                |
| Field   | trace_id         |      | trace ID                                                                                            |

Span Type 为当前 span 在 trace 中的相对位置，其取值说明如下：

- entry：当前 api 为入口即链路进入进入服务后的第一个调用
- local: 当前 api 为入口后出口前的 api
- exit: 当前 api 为链路在服务上最后一个调用
- unknown: 当前 api 的相对位置状态不明确

Priority Rules 为客户端采样优先级规则

- `PRIORITY_USER_REJECT = -1` 用户选择拒绝上报
- `PRIORITY_AUTO_REJECT = 0` 客户端采样器选择拒绝上报
- `PRIORITY_AUTO_KEEP = 1` 客户端采样器选择上报
- `PRIORITY_USER_KEEP = 2` 用户选择上报

### Datakit Tracing Span 数据结构 {#span-struct}

``` golang
TraceID    string                 `json:"trace_id"`
ParentID   string                 `json:"parent_id"`
SpanID     string                 `json:"span_id"`
Service    string                 `json:"service"`     // service name
Resource   string                 `json:"resource"`    // resource or api under service
Operation  string                 `json:"operation"`   // api name
Source     string                 `json:"source"`      // client tracer name
SpanType   string                 `json:"span_type"`   // relative span position in tracing: entry, local, exit or unknow
SourceType string                 `json:"source_type"` // service type
Tags       map[string]string      `json:"tags"`
Metrics    map[string]interface{} `json:"metrics"`
Start      int64                  `json:"start"`    // unit: nano sec
Duration   int64                  `json:"duration"` // unit: nano sec
Status     string                 `json:"status"`   // span status like error, ok, info etc.
Content    string                 `json:"content"`  // raw tracing data in json
```

Datakit Span 是 Datakit 内部使用的数据结构。第三方 Tracing Agent 数据结构会转换成 Datakit Span 结构后发送到数据中心。

> 以下简称 DKSpan

| Field Name | Data Type                  | Unit | Description                                   | Correspond To              |
| ---------- | ------------------------   | ---- | -------------------------------------------   | ------------------------   |
| TraceID    | `string`                   |      | Trace ID                                      | `dkproto.fields.trace_id`  |
| ParentID   | `string`                   |      | Parent Span ID                                | `dkproto.fields.parent_id` |
| SpanID     | `string`                   |      | Span ID                                       | `dkproto.fields.span_id`   |
| Service    | `string`                   |      | Service Name                                  | `dkproto.tags.service`     |
| Resource   | `string`                   |      | Resource Name(.e.g `/get/data/from/some/api`) | `dkproto.fields.resource`  |
| Operation  | `string`                   |      | 生产此条 Span 的方法名                        | `dkproto.tags.operation`   |
| Source     | `string`                   |      | Span 接入源(.e.g `ddtrace`)                   | `dkproto.name`             |
| SpanType   | `string`                   |      | Span Type(.e.g `entry`)                       | `dkproto.tags.span_type`   |
| SourceType | `string`                   |      | Span Source Type(.e.g `web`)                  | `dkproto.tags.type`        |
| Tags       | `map[string, string]`      |      | Span Tags                                     | `dkproto.tags`             |
| Metrics    | `map[string, interface{}]` |      | Span Metrics(计算用)                          | `dkproto.fields`           |
| Start      | `int64`                    | 纳秒 | Span 起始时间                                 | `dkproto.fields.start`     |
| Duration   | `int64`                    | 纳秒 | 耗时                                          | `dkproto.fields.duration`  |
| Status     | `string`                   |      | Span 状态字段                                 | `dkproto.tags.status`      |
| Content    | `string`                   |      | Span 原始数据                                 | `dkproto.fields.message`   |

---

## DDTrace Trace&Span 数据结构 {#ddtrace-trace-span-struct}

### DDTrace Trace 数据结构 {#ddtrace-trace-struct}

DataDog Trace Structure

> Trace: []*span

DataDog Traces Structure

> Traces: []Trace

### DDTrace Span 数据结构 {#ddtrace-span-struct}

| Field Name | Data Type              | Unit | Description                                        | Correspond To                                                                                                |
| ---        | ---                    | ---  | ---                                                | ---                                                                                                          |
| TraceID    | `uint64`               |      | Trace ID                                           | `dkspan.TraceID`                                                                                             |
| ParentID   | `uint64`               |      | Parent Span ID                                     | `dkspan.ParentID`                                                                                            |
| SpanID     | `uint64`               |      | Span ID                                            | `dkspan.SpanID`                                                                                              |
| Service    | `string`               |      | 服务名                                             | `dkspan.Service`                                                                                             |
| Resource   | `string`               |      | 资源名                                             | `dkspan.Resource`                                                                                            |
| Name       | `string`               |      | 生产此条 Span 的方法名                             | `dkspan.Operation`                                                                                           |
| Start      | `int64`                | 纳秒 | Span 起始时间                                      | `dkspan.Start`                                                                                               |
| Duration   | `int64`                | 纳秒 | 耗时                                               | `dkspan.Duration`                                                                                            |
| Error      | `int32`                |      | Span 状态字段 0：无报错 1：出错                      | `dkspan.Status`                                                                                              |
| Meta       | `map[string, string]`  |      | Span 过程元数据，环境相关和服务相关 field 从此获得 | `dkspan.Project, dkspan.Env, dkspan.Version, dkspan.ContainerHost, dkspan.HTTPMethod, dkspan.HTTPStatusCode` |
| Metrics    | `map[string, float64]` |      | Span 采样，运算相关数据                            | 不直接对应 DKSpan                                                                                            |
| Type       | `string`               |      | Span Type                                          | `dkspan.SourceType`                                                                                          |

---

## OpenTelemetry Tracing 数据结构 {#otel-trace-struct}

Datakit 采集从 OpenTelemetry Exporter(OTLP) 中发送上来的数据时，简略的原始数据通过 JSON 序列化之后，如下所示：

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

OpenTelemetry 中的 `resource_spans` 和 DKSpan 的对应关系如下：

| Field Name           | Data Type           | Unit | Description    | Correspond To                                                                                                                                                       |
| ---                  | ---                 | ---  | ---            | ---                                                                                                                                                                 |
| trace_id             | `[16]byte`          |      | Trace ID       | `dkspan.TraceID`                                                                                                                                                    |
| span_id              | `[8]byte`           |      | Span ID        | `dkspan.SpanID`                                                                                                                                                     |
| parent_span_id       | `[8]byte`           |      | Parent Span ID | `dkspan.ParentID`                                                                                                                                                   |
| name                 | `string`            |      | Span Name      | `dkspan.Operation`                                                                                                                                                  |
| kind                 | `string`            |      | Span Type      | `dkspan.SpanType`                                                                                                                                                   |
| start_time_unix_nano | `int64`             | 纳秒 | Span 起始时间  | `dkspan.Start`                                                                                                                                                      |
| end_time_unix_nano   | `int64`             | 纳秒 | Span 终止时间  | `dkspan.Duration = end - start`                                                                                                                                     |
| status               | `string`            |      | Span Status    | `dkspan.Status`                                                                                                                                                     |
| name                 | `string`            |      | resource Name  | `dkspan.Resource`                                                                                                                                                   |
| resource.attributes  | `map[string]string` |      | resource 标签  | `dkspan.tags.service, dkspan.tags.project, dkspan.tags.env, dkspan.tags.version, dkspan.tags.container_host, dkspan.tags.http_method, dkspan.tags.http_status_code` |
| span.attributes      | `map[string]string` |      | Span 标签      | `dkspan.tags`                                                                                                                                                       |

OpenTelemetry 有些独有字段， 但 DKSpan 没有字段与之对应，所以就放在了标签中，只有这些值非 0 时才会显示，如：

| Field                         | Date Type | Uint | Description             | Correspond                             |
| :---                          | :---      | :--- | :---                    | :---                                   |
| span.dropped_attributes_count | `int`     |      | Span 被删除的标签数量   | `dkspan.tags.dropped_attributes_count` |
| span.dropped_events_count     | `int`     |      | Span 被删除的事件数量   | `dkspan.tags.dropped_events_count`     |
| span.dropped_links_count      | `int`     |      | Span 被删除的连接数量   | `dkspan.tags.dropped_links_count`      |
| span.events_count             | `int`     |      | Span 关联事件数量       | `dkspan.tags.events_count`             |
| span.links_count              | `int`     |      | Span 所关联的 span 数量 | `dkspan.tags.links_count`              |

---

## Jaeger Tracing 数据结构 {#jaeger-trace-struct}

### Jaeger Thrift Protocol Batch 数据结构 {#jaeger-thrift-batch-struct}

| Field Name | Data Type        | Unit | Description      | Correspond to       |
| ---------- | --------------   | ---- | ---------------- | ------------------- |
| Process    | `struct pointer` |      | 进程相关数据结构 | `dkspan.Service`    |
| SeqNo      | `int64 pointer`  |      | 序列号           | 不接对应关系 DKSpan |
| Spans      | `array`          |      | Span 数组结构    | 见下表              |
| Stats      | `struct pointer` |      | 客户端统计结构   | 不直接对应 DKSpan   |

### Jaeger Thrift Protocol Span 数据结构 {#jaeger-thrift-span-struct}

| Field Name    | Data Type | Unit | Description                               | Correspond To      |
| ------------- | --------- | ---- | ----------------------------------------- | -----------------  |
| TraceIdHigh   | `int64`   |      | Trace ID 高位与 TraceIdLow 组成 Trace ID  | `dkspan.TraceID`   |
| TraceIdLow    | `int64`   |      | Trace ID 低位与 TraceIdHigh 组成 Trace ID | `dkspan.TraceID`   |
| ParentSpanId  | `int64`   |      | Parent Span ID                            | `dkspan.ParentID`  |
| SpanId        | `int64`   |      | Span ID                                   | `dkspan.SpanID`    |
| OperationName | `string`  |      | 生产此条 Span 的方法名                    | `dkspan.Operation` |
| Flags         | `int32`   |      | Span Flags                                | 不直接对应 DKSpan  |
| Logs          | `array`   |      | Span Logs                                 | 不直接对应 DKSpan  |
| References    | `array`   |      | Span References                           | 不直接对应 DKSpan  |
| StartTime     | `int64`   | 纳秒 | Span 起始时间                             | `dkspan.Start`     |
| Duration      | `int64`   | 纳秒 | 耗时                                      | `dkspan.Duration`  |
| Tags          | `array`   |      | Span Tags 目前只取 Span 状态字段          | `dkspan.Status`    |

---

## SkyWalking Tracing Data 数据结构 {#sw-trace-struct}

<!-- markdownlint-disable MD013 -->
### Segment Object Generated By Protobuf Protocol V3 {#sw-v3-pb-struct}
<!-- markdownlint-enable -->

| Field Name      | Data Type | Unit | Description                                     | Correspond To        |
| --------------- | --------- | ---- | ----------------------------------------------- | ------------------   |
| TraceId         | `string`  |      | Trace ID                                        | `dkspan.TraceID`     |
| TraceSegmentId  | `string`  |      | Segment ID 与 Span ID 一起使用唯一标志一个 Span | `dkspan.SpanID` 高位 |
| Service         | `string`  |      | 服务名                                          | `dkspan.Service`     |
| ServiceInstance | `string`  |      | 节点逻辑关系名                                  | 未使用字段           |
| Spans           | `array`   |      | Tracing Span 数组                               | 见下表               |
| IsSizeLimited   | `bool`    |      | 是否包含连路上所有 Span                         | 未使用字段           |

### SkyWalking Span Object 数据结构 in Segment Object {#sw-span-struct}

| Field Name    | Data Type | Unit | Description                                                   | Correspond To          |
| ------------- | --------- | ---- | ------------------------------------------------------------- | --------------------   |
| ComponentId   | `int32`   |      | 第三方框架数值化定义                                          | 未使用字段             |
| Refs          | `array`   |      | 跨线程跨进程情况下存储 Parent Segment                         | `dkspan.ParentID` 高位 |
| ParentSpanId  | `int32`   |      | Parent Span ID 与 Segment ID 一起使用唯一标志一个 Parent Span | `dkspan.ParentID` 低位 |
| SpanId        | `int32`   |      | Span ID 与 Segment ID 一起使用唯一标志一个 Span               | `dkspan.SpanID` 低位   |
| OperationName | `string`  |      | Span Operation Name                                           | `dkspan.Operation`     |
| Peer          | `string`  |      | 通讯对端                                                      | `dkspan.Endpoint`      |
| IsError       | `bool`    |      | Span 状态字段                                                 | `dkspan.Status`        |
| SpanType      | `int32`   |      | Span Type 数值化定义                                          | `dkspan.SpanType`      |
| StartTime     | `int64`   | 毫秒 | Span 起始时间                                                 | `dkspan.Start`         |
| EndTime       | `int64`   | 毫秒 | Span 结束时间与 StartTime 相减代表耗时                        | `dkspan.Duration`      |
| Logs          | `array`   |      | Span Logs                                                     | 未使用字段             |
| SkipAnalysis  | `bool`    |      | 跳过后端分析                                                  | 未使用字段             |
| SpanLayer     | `int32`   |      | Span 技术栈数值化定义                                         | 未使用字段             |
| Tags          | `array`   |      | Span Tags                                                     | 未使用字段             |

---

## Zipkin Tracing Data 数据结构 {#zk-trace-struct}

### Zipkin Thrift Protocol Span 数据结构 V1 {#zk-thrift-v1-span-struct}

| Field Name        | Data Type | Unit | Description         | Correspond To      |
| ----------------- | --------- | ---- | ------------------- | ----------------   |
| TraceIDHigh       | `uint64`  |      | Trace ID 高位       | 无直接对应关系     |
| TraceID           | `uint64`  |      | Trace ID            | `dkspan.TraceID`   |
| ID                | `uint64`  |      | Span ID             | `dkspan.SpanID`    |
| ParentID          | `uint64`  |      | Parent Span ID      | `dkspan.ParentID`  |
| Annotations       | `array`   |      | 获取 Service Name   | `dkspan.Service`   |
| Name              | `string`  |      | Span Operation Name | `dkspan.Operation` |
| BinaryAnnotations | `array`   |      | 获取 Span 状态字段  | `dkspan.Status`    |
| Timestamp         | `uint64`  | 微秒 | Span 起始时间       | `dkspan.Start`     |
| Duration          | `uint64`  | 微秒 | Span 耗时           | `dkspan.Duration`  |
| Debug             | `bool`    |      | Debug 状态字段      | 未使用字段         |

### Zipkin Span 数据结构 V2 {#zk-thrift-v2-span-struct}

| Field Name     | Data Type | Unit | Description                      | Correspond To      |
| -------------- | --------- | ---- | -------------------------------- | -----------------  |
| TraceID        | `struct`  |      | Trace ID                         | `dkspan.TraceID`   |
| ID             | `uint64`  |      | Span ID                          | `dkspan.SpanID`    |
| ParentID       | `uint64`  |      | Parent Span ID                   | `dkspan.ParentID`  |
| Name           | `string`  |      | Span Operation Name              | `dkspan.Operation` |
| Debug          | `bool`    |      | Debug 状态                       | 未使用字段         |
| Sampled        | `bool`    |      | 采样状态字段                     | 未使用字段         |
| Err            | `string`  |      | Error Message                    | 不直接对应 DKSpan  |
| Kind           | `string`  |      | Span Type                        | `dkspan.SpanType`  |
| Timestamp      | `struct`  | 微秒 | 微秒级时间结构表示 Span 起始时间 | `dkspan.Start`     |
| Duration       | `int64`   | 微秒 | Span 耗时                        | `dkspan.Duration`  |
| Shared         | `bool`    |      | 共享状态                         | 未使用字段         |
| LocalEndpoint  | `struct`  |      | 用于获取 Service Name            | `dkspan.Service`   |
| RemoteEndpoint | `struct`  |      | 通讯对端                         | `dkspan.Endpoint`  |
| Annotations    | `array`   |      | 用于解释延迟相关的事件           | 未使用字段         |
| Tags           | `map`     |      | 用于获取 Span 状态               | `dkspan.Status`    |
