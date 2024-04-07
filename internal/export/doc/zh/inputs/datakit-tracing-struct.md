
# Datakit Tracing 数据结构

## 简述 {#intro}

此文用于解释主流 Telemetry 平台数据结构以及与 Datakit 平台数据结构的映射关系
目前支持数据结构：DataDog/Jaeger/OpenTelemetry/SkyWalking/Zipkin/PinPoint

数据转换步骤：

1. 外部 Tracing 数据接入，通过数据多种协议接受数据后进行反序列化。
2. 反序列化后的对象转 `Line Protocol` （行协议格式）。
3. span 数据运算包括：采样，过滤，添加特定标签等操作。

---

## Datakit Point Protocol 数据结构 {#point-proto}

Line Protocol 数据结构是由 Name, Tags, Fields, Timestamp 四部分和分隔符 (英文逗号，空格) 组成的字符串，形如：

``` not-set
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

> 以下简称 DKProto

| Section | Name             | Unit | Description                                                                                         |
|---------|------------------|------|-----------------------------------------------------------------------------------------------------|
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
| Field   | duration         | 微秒   | span duration                                                                                       |
| Field   | message          |      | raw data content                                                                                    |
| Field   | parent_id        |      | parent ID of span                                                                                   |
| Field   | priority         |      | priority rules (PRIORITY_USER_REJECT, PRIORITY_AUTO_REJECT, PRIORITY_AUTO_KEEP, PRIORITY_USER_KEEP) |
| Field   | resource         |      | resource of service                                                                                 |
| Field   | sample_rate      |      | global sampling ratio (0.1 means roughly 10 percent will send to data center)                       |
| Field   | span_id          |      | span ID                                                                                             |
| Field   | start            | 微妙   | span start timestamp                                                                                |
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

OpenTelemetry 中的 `resource_spans` 和 DKProto 的对应关系如下：

| Field Name           | Data Type           | Unit | Description    | Correspond To                                                                                                                                                       |
| -------------------- | ------------------- | ---- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| trace_id             | `[16]byte`          |      | Trace ID       | `DKProto.TraceID`                                                                                                                                                    |
| span_id              | `[8]byte`           |      | Span ID        | `DKProto.SpanID`                                                                                                                                                     |
| parent_span_id       | `[8]byte`           |      | Parent Span ID | `DKProto.ParentID`                                                                                                                                                   |
| name                 | `string`            |      | Span Name      | `DKProto.Operation`                                                                                                                                                  |
| kind                 | `string`            |      | Span Type      | `DKProto.SpanType`                                                                                                                                                   |
| start_time_unix_nano | `int64`             | 纳秒 | Span 起始时间  | `DKProto.Start`                                                                                                                                                      |
| end_time_unix_nano   | `int64`             | 纳秒 | Span 终止时间  | `DKProto.Duration = end - start`                                                                                                                                     |
| status               | `string`            |      | Span Status    | `DKProto.Status`                                                                                                                                                     |
| name                 | `string`            |      | resource Name  | `DKProto.Resource`                                                                                                                                                   |
| resource.attributes  | `map[string]string` |      | resource 标签  | `DKProto.tags.service, DKProto.tags.project, DKProto.tags.env, DKProto.tags.version, DKProto.tags.container_host, DKProto.tags.http_method, DKProto.tags.http_status_code` |
| span.attributes      | `map[string]string` |      | Span 标签      | `DKProto.tags`                                                                                                                                                       |

OpenTelemetry 有些独有字段， 但 DKProto 没有字段与之对应，所以就放在了标签中，只有这些值非 0 时才会显示，如：

| Field                         | Date Type | Uint | Description             | Correspond                             |
| :---------------------------- | :-------- | :--- | :---------------------- | :------------------------------------- |
| span.dropped_attributes_count | `int`     |      | Span 被删除的标签数量   | `DKProto.tags.dropped_attributes_count` |
| span.dropped_events_count     | `int`     |      | Span 被删除的事件数量   | `DKProto.tags.dropped_events_count`     |
| span.dropped_links_count      | `int`     |      | Span 被删除的连接数量   | `DKProto.tags.dropped_links_count`      |
| span.events_count             | `int`     |      | Span 关联事件数量       | `DKProto.tags.events_count`             |
| span.links_count              | `int`     |      | Span 所关联的 span 数量 | `DKProto.tags.links_count`              |

---

## Jaeger Tracing 数据结构 {#jaeger-trace-struct}

### Jaeger Thrift Protocol Batch 数据结构 {#jaeger-thrift-batch-struct}

| Field Name | Data Type        | Unit | Description      | Correspond to       |
| ---------- | ---------------- | ---- | ---------------- | ------------------- |
| Process    | `struct pointer` |      | 进程相关数据结构 | `DKProto.Service`    |
| SeqNo      | `int64 pointer`  |      | 序列号           | 不接对应关系 DKProto |
| Spans      | `array`          |      | Span 数组结构    | 见下表              |
| Stats      | `struct pointer` |      | 客户端统计结构   | 不直接对应 DKProto   |

### Jaeger Thrift Protocol Span 数据结构 {#jaeger-thrift-span-struct}

| Field Name    | Data Type | Unit | Description                               | Correspond To      |
| ------------- | --------- | ---- | ----------------------------------------- | ------------------ |
| TraceIdHigh   | `int64`   |      | Trace ID 高位与 TraceIdLow 组成 Trace ID  | `DKProto.TraceID`   |
| TraceIdLow    | `int64`   |      | Trace ID 低位与 TraceIdHigh 组成 Trace ID | `DKProto.TraceID`   |
| ParentSpanId  | `int64`   |      | Parent Span ID                            | `DKProto.ParentID`  |
| SpanId        | `int64`   |      | Span ID                                   | `DKProto.SpanID`    |
| OperationName | `string`  |      | 生产此条 Span 的方法名                    | `DKProto.Operation` |
| Flags         | `int32`   |      | Span Flags                                | 不直接对应 DKProto  |
| Logs          | `array`   |      | Span Logs                                 | 不直接对应 DKProto  |
| References    | `array`   |      | Span References                           | 不直接对应 DKProto  |
| StartTime     | `int64`   | 纳秒 | Span 起始时间                             | `DKProto.Start`     |
| Duration      | `int64`   | 纳秒 | 耗时                                      | `DKProto.Duration`  |
| Tags          | `array`   |      | Span Tags 目前只取 Span 状态字段          | `DKProto.Status`    |

---

## SkyWalking Tracing Data 数据结构 {#sw-trace-struct}

<!-- markdownlint-disable MD013 -->
### Segment Object Generated By Protobuf Protocol V3 {#sw-v3-pb-struct}
<!-- markdownlint-enable -->

| Field Name      | Data Type | Unit | Description                                     | Correspond To        |
| --------------- | --------- | ---- | ----------------------------------------------- | -------------------- |
| TraceId         | `string`  |      | Trace ID                                        | `DKProto.TraceID`     |
| TraceSegmentId  | `string`  |      | Segment ID 与 Span ID 一起使用唯一标志一个 Span | `DKProto.SpanID` 高位 |
| Service         | `string`  |      | 服务名                                          | `DKProto.Service`     |
| ServiceInstance | `string`  |      | 节点逻辑关系名                                  | 未使用字段           |
| Spans           | `array`   |      | Tracing Span 数组                               | 见下表               |
| IsSizeLimited   | `bool`    |      | 是否包含连路上所有 Span                         | 未使用字段           |

### SkyWalking Span Object 数据结构 in Segment Object {#sw-span-struct}

| Field Name    | Data Type | Unit | Description                                                   | Correspond To          |
| ------------- | --------- | ---- | ------------------------------------------------------------- | ---------------------- |
| ComponentId   | `int32`   |      | 第三方框架数值化定义                                          | 未使用字段             |
| Refs          | `array`   |      | 跨线程跨进程情况下存储 Parent Segment                         | `DKProto.ParentID` 高位 |
| ParentSpanId  | `int32`   |      | Parent Span ID 与 Segment ID 一起使用唯一标志一个 Parent Span | `DKProto.ParentID` 低位 |
| SpanId        | `int32`   |      | Span ID 与 Segment ID 一起使用唯一标志一个 Span               | `DKProto.SpanID` 低位   |
| OperationName | `string`  |      | Span Operation Name                                           | `DKProto.Operation`     |
| Peer          | `string`  |      | 通讯对端                                                      | `DKProto.Endpoint`      |
| IsError       | `bool`    |      | Span 状态字段                                                 | `DKProto.Status`        |
| SpanType      | `int32`   |      | Span Type 数值化定义                                          | `DKProto.SpanType`      |
| StartTime     | `int64`   | 毫秒 | Span 起始时间                                                 | `DKProto.Start`         |
| EndTime       | `int64`   | 毫秒 | Span 结束时间与 StartTime 相减代表耗时                        | `DKProto.Duration`      |
| Logs          | `array`   |      | Span Logs                                                     | 未使用字段             |
| SkipAnalysis  | `bool`    |      | 跳过后端分析                                                  | 未使用字段             |
| SpanLayer     | `int32`   |      | Span 技术栈数值化定义                                         | 未使用字段             |
| Tags          | `array`   |      | Span Tags                                                     | 未使用字段             |

---

## Zipkin Tracing Data 数据结构 {#zk-trace-struct}

### Zipkin Thrift Protocol Span 数据结构 V1 {#zk-thrift-v1-span-struct}

| Field Name        | Data Type | Unit | Description         | Correspond To      |
| ----------------- | --------- | ---- | ------------------- | ------------------ |
| TraceIDHigh       | `uint64`  |      | Trace ID 高位       | 无直接对应关系     |
| TraceID           | `uint64`  |      | Trace ID            | `DKProto.TraceID`   |
| ID                | `uint64`  |      | Span ID             | `DKProto.SpanID`    |
| ParentID          | `uint64`  |      | Parent Span ID      | `DKProto.ParentID`  |
| Annotations       | `array`   |      | 获取 Service Name   | `DKProto.Service`   |
| Name              | `string`  |      | Span Operation Name | `DKProto.Operation` |
| BinaryAnnotations | `array`   |      | 获取 Span 状态字段  | `DKProto.Status`    |
| Timestamp         | `uint64`  | 微秒 | Span 起始时间       | `DKProto.Start`     |
| Duration          | `uint64`  | 微秒 | Span 耗时           | `DKProto.Duration`  |
| Debug             | `bool`    |      | Debug 状态字段      | 未使用字段         |

### Zipkin Span 数据结构 V2 {#zk-thrift-v2-span-struct}

| Field Name     | Data Type | Unit | Description                      | Correspond To      |
| -------------- | --------- | ---- | -------------------------------- | ------------------ |
| TraceID        | `struct`  |      | Trace ID                         | `DKProto.TraceID`   |
| ID             | `uint64`  |      | Span ID                          | `DKProto.SpanID`    |
| ParentID       | `uint64`  |      | Parent Span ID                   | `DKProto.ParentID`  |
| Name           | `string`  |      | Span Operation Name              | `DKProto.Operation` |
| Debug          | `bool`    |      | Debug 状态                       | 未使用字段         |
| Sampled        | `bool`    |      | 采样状态字段                     | 未使用字段         |
| Err            | `string`  |      | Error Message                    | 不直接对应 DKProto  |
| Kind           | `string`  |      | Span Type                        | `DKProto.SpanType`  |
| Timestamp      | `struct`  | 微秒 | 微秒级时间结构表示 Span 起始时间 | `DKProto.Start`     |
| Duration       | `int64`   | 微秒 | Span 耗时                        | `DKProto.Duration`  |
| Shared         | `bool`    |      | 共享状态                         | 未使用字段         |
| LocalEndpoint  | `struct`  |      | 用于获取 Service Name            | `DKProto.Service`   |
| RemoteEndpoint | `struct`  |      | 通讯对端                         | `DKProto.Endpoint`  |
| Annotations    | `array`   |      | 用于解释延迟相关的事件           | 未使用字段         |
| Tags           | `map`     |      | 用于获取 Span 状态               | `DKProto.Status`    |
