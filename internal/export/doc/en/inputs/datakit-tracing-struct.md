
# Datakit Tracing Data Structure

## Brief {#intro}

This article is used to explain the data structure of the mainstream Telemetry platform and the mapping relationship with the data structure of Datakit platform.
Currently supported data structures: DataDog, Jaeger, OpenTelemetry, SkyWalking, Zipkin.

Data conversion steps:

1. External Tracing data structure access
2. Datakit Span transformation
3. span data operations
4. Line Protocol transformation

---

## Datakit Point Protocol Data Structure {#point-proto}

The Line Protocol data structure is a string composed of Name, Tags, Fields, Timestamp and delimiters (English commas, spaces), which is shaped like:

```txt
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

> Hereinafter referred to as `dkproto`

| **Section** | **Name**         | **Unit**    | **Description**                                                                                     |
| ----------- | ---------------- | ----------- | --------------------------------------------------------------------------------------------------- |
| Tag         | container_host   |             | host name of container                                                                              |
| Tag         | endpoint         |             | end point of resource                                                                               |
| Tag         | env              |             | environment arguments                                                                               |
| Tag         | http_host        |             | HTTP host                                                                                           |
| Tag         | http_method      |             | HTTP method                                                                                         |
| Tag         | http_route       |             | HTTP route                                                                                          |
| Tag         | http_status_code |             | HTTP status code                                                                                    |
| Tag         | http_url         |             | HTTP URL                                                                                            |
| Tag         | operation        |             | operation of resource                                                                               |
| Tag         | pid              |             | process id                                                                                          |
| Tag         | project          |             | project name                                                                                        |
| Tag         | service          |             | service name                                                                                        |
| Tag         | source_type      |             | source types [app, framework, cache, message_queue, custom, db, web]                                |
| Tag         | status           |             | span status [ok, info, warning, error, critical]                                                    |
| Tag         | span_type        |             | span types [entry, local, exit, unknown]                                                            |
| Tag         | version          |             | service version                                                                                     |
| Field       | duration         | Microsecond | span duration                                                                                       |
| Field       | message          |             | raw data content                                                                                    |
| Field       | parent_id        |             | parent ID of span                                                                                   |
| Field       | priority         |             | priority rules (PRIORITY_USER_REJECT, PRIORITY_AUTO_REJECT, PRIORITY_AUTO_KEEP, PRIORITY_USER_KEEP) |
| Field       | resource         |             | resource of service                                                                                 |
| Field       | sample_rate      |             | global sampling ratio (0.1 means roughly 10 percent will send to data center)                       |
| Field       | span_id          |             | span ID                                                                                             |
| Field       | start            | Microsecond | span start timestamp                                                                                |
| Field       | trace_id         |             | trace ID                                                                                            |

Span Type is the relative position of the current span in trace, and its value is described as follows:

- entry: the current api is the first call after the entry of the link into the service
- local: the current api is the api after the entrance and before the exit
- exit: the current api is the link's last call on the service
- unknown: the relative position state of the current api is not clear

Priority Rules samples priority rules for clients:

- `PRIORITY_USER_REJECT = -1` User chooses to reject reporting
- `PRIORITY_AUTO_REJECT = 0` Client sampler chooses to reject reporting
- `PRIORITY_AUTO_KEEP = 1` Client sampler select report
- `PRIORITY_USER_KEEP = 2` User chooses to report

### Datakit Tracing Span Data Structure {#span-struct}

``` golang
TraceID    string                 `json:"trace_id"`
ParentID   string                 `json:"parent_id"`
SpanID     string                 `json:"span_id"`
Service    string                 `json:"service"`     // service name
Resource   string                 `json:"resource"`    // resource or api under service
Operation  string                 `json:"operation"`   // api name
Source     string                 `json:"source"`      // client tracer name
SpanType   string                 `json:"span_type"`   // relative span position in tracing: entry, local, exit or unknown
SourceType string                 `json:"source_type"` // service type
Tags       map[string]string      `json:"tags"`
Metrics    map[string]interface{} `json:"metrics"`
Start      int64                  `json:"start"`    // unit: nano sec
Duration   int64                  `json:"duration"` // unit: nano sec
Status     string                 `json:"status"`   // span status like error, ok, info etc.
Content    string                 `json:"content"`  // raw tracing data in json
```

Datakit Span is a data structure used internally by Datakit. The third-party Tracing Agent data structure is converted into a Datakit Span structure and sent to the data center.

> Hereinafter referred to as `dkspan`

| Field Name | Data Type                | Unit       | Description                                 | Correspond To              |
| ---------- | ------------------------ | ---------- | ------------------------------------------- | -------------------------- |
| TraceID    | string                   |            | Trace ID                                    | `dkproto.fields.trace_id`  |
| ParentID   | string                   |            | Parent Span ID                              | `dkproto.fields.parent_id` |
| SpanID     | string                   |            | Span ID                                     | `dkproto.fields.span_id`   |
| Service    | string                   |            | Service Name                                | `dkproto.tags.service`     |
| Resource   | string                   |            | Resource Name(.e.g /get/data/from/some/api) | `dkproto.fields.resource`  |
| Operation  | string                   |            | The method name that produces this Span     | `dkproto.tags.operation`   |
| Source     | string                   |            | Span source(.e.g ddtrace)                   | `dkproto.name`             |
| SpanType   | string                   |            | Span Type(.e.g Entry)                       | `dkproto.tags.span_type`   |
| SourceType | string                   |            | Span Source Type(.e.g Web)                  | `dkproto.tags.type`        |
| Tags       | map[string, string]      |            | Span Tags                                   | `dkproto.tags`             |
| Metrics    | map[string, interface{}] |            | Span Metrics(for calculation)               | `dkproto.fields`           |
| Start      | int64                    | Nanosecond | Span Starting time                          | `dkproto.fields.start`     |
| Duration   | int64                    | Nanosecond | Time consuming                              | `dkproto.fields.duration`  |
| Status     | string                   |            | Span status field                           | `dkproto.tags.status`      |
| Content    | string                   |            | Span raw data                               | `dkproto.fields.message`   |

---

## DDTrace Trace&Span Data Structure {#ddtrace-trace-span-struct}

### DDTrace Trace Data Structure {#ddtrace-trace-struct}

DataDog Trace Structure

> Trace: []\*span

DataDog Traces Structure

> Traces: []Trace

### DDTrace Span Data Structure {#ddtrace-span-struct}

| Field Name | Data Type            | Unit       | Description                                                                                   | Correspond To                                                                                                          |
| ---------- | -------------------- | ---------- | --------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| TraceID    | uint64               |            | Trace ID                                                                                      | `dkspan.TraceID`                                                                                                       |
| ParentID   | uint64               |            | Parent Span ID                                                                                | `dkspan.ParentID`                                                                                                      |
| SpanID     | uint64               |            | Span ID                                                                                       | `dkspan.SpanID`                                                                                                        |
| Service    | string               |            | Server name                                                                                   | `dkspan.Service`                                                                                                       |
| Resource   | string               |            | Resource name                                                                                 | `dkspan.Resource`                                                                                                      |
| Name       | string               |            | The name of the method to produce this Span                                                   | `dkspan.Operation`                                                                                                     |
| Start      | int64                | Nanosecond | Span starting time                                                                            | `dkspan.Start`                                                                                                         |
| Duration   | int64                | Nanosecond | Time consuming                                                                                | `dkspan.Duration`                                                                                                      |
| Error      | int32                |            | Span Status field 0: No error 1: Error                                                        | `dkspan.Status`                                                                                                        |
| Meta       | map[string, string]  |            | Span process metadata, environment-related, and service-related fields are obtained from here | `dkspan.Project`, `dkspan.Env`, `dkspan.Version`, `dkspan.ContainerHost`, `dkspan.HTTPMethod`, `dkspan.HTTPStatusCode` |
| Metrics    | map[string, float64] |            | Span sampling, computing related data                                                         | Indirect correspondence to `dkspan`                                                                                    |
| Type       | string               |            | Span Type                                                                                     | `dkspan.SourceType`                                                                                                    |

---

## OpenTelemetry Tracing Data Structure {#otel-trace-struct}

When DataKit collects data sent from OpenTelemetry exporter: Otlp, the abbreviated raw data, after serialization by json, looks like this:

```golang
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

The correspondence between `resource_spans` and `dkspan` in `otel` is as follows:

| Field Name           | Data Type         | Unit       | Description        | Correspond To                                                                                                                                                                   |
| -------------------- | ----------------- | ---------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| trace_id             | [16]byte          |            | Trace ID           | `dkspan.TraceID`                                                                                                                                                                |
| span_id              | [8]byte           |            | Span ID            | `dkspan.SpanID`                                                                                                                                                                 |
| parent_span_id       | [8]byte           |            | Parent Span ID     | `dkspan.ParentID`                                                                                                                                                               |
| name                 | string            |            | Span Name          | `dkspan.Operation`                                                                                                                                                              |
| kind                 | string            |            | Span Type          | `dkspan.SpanType`                                                                                                                                                               |
| start_time_unix_nano | int64             | Nanosecond | Span starting time | `dkspan.Start`                                                                                                                                                                  |
| end_time_unix_nano   | int64             | Nanosecond | Span ending time   | `dkspan.Duration = end - start`                                                                                                                                                 |
| status               | string            |            | Span Status        | `dkspan.Status`                                                                                                                                                                 |
| name                 | string            |            | resource Name      | `dkspan.Resource`                                                                                                                                                               |
| resource.attributes  | map[string]string |            | resource tag       | `dkspan.tags.service`, `dkspan.tags.project`, `dkspan.tags.env`, `dkspan.tags.version`, `dkspan.tags.container_host`, `dkspan.tags.http_method`, `dkspan.tags.http_status_code` |
| span.attributes      | map[string]string |            | Span tag           | `dkspan.tags`                                                                                                                                                                   |

`otel` has some unique fields, but DKspan has no corresponding fields, so it is placed in the label and will only be displayed if these values are not 0, such as:

| Field                         | Date Type | Uint | Description                                | Correspond                             |
| :---------------------------- | :-------- | :--- | :----------------------------------------- | :------------------------------------- |
| span.dropped_attributes_count | int       |      | Span number of tags removed                | `dkspan.tags.dropped_attributes_count` |
| span.dropped_events_count     | int       |      | Span number of events deleted              | `dkspan.tags.dropped_events_count`     |
| span.dropped_links_count      | int       |      | Span number of connections deleted         | `dkspan.tags.dropped_links_count`      |
| span.events_count             | int       |      | Number of Span associated events           | `dkspan.tags.events_count`             |
| span.links_count              | int       |      | The number of spans associated with a span | `dkspan.tags.links_count`              |

---

## Jaeger Tracing Data Structure {#jaeger-trace-struct}

### Jaeger Thrift Protocol Batch Data Structure {#jaeger-thrift-batch-struct}

| Field Name | Data Type         | Unit | Description                    | Correspond to                          |
| ---------- | ----------------- | ---- | ------------------------------ | -------------------------------------- |
| Process    | structure pointer |      | Process-related data structure | `dkspan.Service`                       |
| SeqNo      | int64 pointer     |      | Serial number                  | Disconnected mapping relation `dkspan` |
| Spans      | array             |      | Span array structure           | See the table below                    |
| Stats      | structure pointer    |      | Client statistical structure   | not directly correspond to `dkspan`      |

### Jaeger Thrift Protocol Span Data Structure {#jaeger-thrift-span-struct}

| Field Name    | Data Type | Unit       | Description                                          | Correspond To                     |
| ------------- | --------- | ---------- | ---------------------------------------------------- | --------------------------------- |
| TraceIdHigh   | int64     |            | Trace ID High and TraceIdLow make up Trace ID        | `dkspan.TraceID`                  |
| TraceIdLow    | int64     |            | Trace ID Low and TraceIdHigh make up Trace ID        | `dkspan.TraceID`                  |
| ParentSpanId  | int64     |            | Parent Span ID                                       | `dkspan.ParentID`                 |
| SpanId        | int64     |            | Span ID                                              | `dkspan.SpanID`                   |
| OperationName | string    |            | The name of the method to produce this Span          | `dkspan.Operation`                |
| Flags         | int32     |            | Span Flags                                           | not directly correspond to `dkspan` |
| Logs          | array     |            | Span Logs                                            | not directly correspond to `dkspan` |
| References    | array     |            | Span References                                      | not directly correspond to `dkspan` |
| StartTime     | int64     | Nanosecond | Span Starting time                                   | `dkspan.Start`                    |
| Duration      | int64     | Nanosecond | Time consuming                                       | `dkspan.Duration`                 |
| Tags          | array     |            | Span Tags currently only takes the Span status field | `dkspan.Status`                   |

---
<!-- markdownlint-disable MD013 -->
## SkyWalking Tracing Data Data Structure {#sw-trace-struct}
<!-- markdownlint-enable -->
<!-- markdownlint-disable MD013 -->
### SkyWalking Segment Object Generated By Protocol Buffer Protocol V3 {#sw-v3-pb-struct}
<!-- markdownlint-enable -->
| Field Name      | Data Type | Unit | Description                                                                                   | Correspond To       |
| --------------- | --------- | ---- | --------------------------------------------------------------------------------------------- | ------------------- |
| TraceId         | string    |      | Trace ID                                                                                      | `dkspan.TraceID`    |
| TraceSegmentId  | string    |      | The Segment ID is used with the Span ID to uniquely identify a Span `dkspan.SpanID` high order. |                     |
| Service         | string    |      | service                                                                                       | `dkspan.Service`    |
| ServiceInstance | string    |      | Node logical relationship name                                                                | Fields not used     |
| Spans           | array     |      | Tracing Span Array                                                                            | See the table below |
| IsSizeLimited   | bool      |      | whether includes all Spans on the link Span                                                   | Fields not used     |
<!-- markdownlint-disable MD013 -->
### SkyWalking Span Object Data Structure in Segment Object {#sw-span-struct}
<!-- markdownlint-enable -->
| Field Name    | Data Type | Unit         | Description                                                                       | Correspond To                   |
| ------------- | --------- | ------------ | --------------------------------------------------------------------------------- | ------------------------------- |
| ComponentId   | int32     |              | Numerical definition of third-party framework                                     | Fields not used                 |
| Refs          | array     |              | Storing Parent Segment across threads and processes                               | `dkspan.ParentID` high position |
| ParentSpanId  | int32     |              | The Parent Span ID is used with the Segment ID to uniquely identify a Parent Span | `dkspan.ParentID` low position  |
| SpanId        | int32     |              | The Span ID is used with the Segment ID to uniquely identify a Span               | `dkspan.SpanID` low position    |
| OperationName | string    |              | Span Operation Name                                                               | `dkspan.Operation`              |
| Peer          | string    |              | Communication peer                                                                | `dkspan.Endpoint`               |
| IsError       | bool      |              | Span Status field                                                                 | `dkspan.Status`                 |
| SpanType      | int32     |              | Span Type Numerical definition                                                    | `dkspan.SpanType`               |
| StartTime     | int64     | Milliseconds | Span Starting time                                                                | `dkspan.Start`                  |
| EndTime       | int64     | Milliseconds | Span end time subtracted from StartTime represents elapsed time                   | `dkspan.Duration`               |
| Logs          | array     |              | Span Logs                                                                         | Fields not used                 |
| SkipAnalysis  | bool      |              | Skip back-end analysis                                                            | Fields not used                 |
| SpanLayer     | int32     |              | Span technology stack numerical definition                                        | Fields not used                 |
| Tags          | array     |              | Span Tags                                                                         | Fields not used                 |

---

## Zipkin Tracing Data Data Structure {#zk-trace-struct}

### Zipkin Thrift Protocol Span Data Structure V1 {#zk-thrift-v1-span-struct}

| Field Name        | Data Type | Unit        | Description            | Correspond To                           |
| ----------------- | --------- | ----------- | ---------------------- | --------------------------------------- |
| TraceIDHigh       | uint64    |             | Trace ID high position | There is no direct mapping relationship |
| TraceID           | uint64    |             | Trace ID               | `dkspan.TraceID`                        |
| ID                | uint64    |             | Span ID                | `dkspan.SpanID`                         |
| ParentID          | uint64    |             | Parent Span ID         | `dkspan.ParentID`                       |
| Annotations       | array     |             | get Service Name       | `dkspan.Service`                        |
| Name              | string    |             | Span Operation Name    | `dkspan.Operation`                      |
| BinaryAnnotations | array     |             | Get Span status field  | `dkspan.Status`                         |
| Timestamp         | uint64    | Microsecond | Span Starting time     | `dkspan.Start`                          |
| Duration          | uint64    | Microsecond | Span Time consuming    | `dkspan.Duration`                       |
| Debug             | bool      |             | Debug status field     | Fields not used                         |

### Zipkin Span Data Structure V2 {#zk-thrift-v2-span-struct}

| Field Name     | Data Type | Unit        | Description                                                  | Correspond To                     |
| -------------- | --------- | ----------- | ------------------------------------------------------------ | --------------------------------- |
| TraceID        | structure |             | Trace ID                                                     | `dkspan.TraceID`                  |
| ID             | uint64    |             | Span ID                                                      | `dkspan.SpanID`                   |
| ParentID       | uint64    |             | Parent Span ID                                               | `dkspan.ParentID`                 |
| Name           | string    |             | Span Operation Name                                          | `dkspan.Operation`                |
| Debug          | bool      |             | Debug status                                                 | Fields not used                   |
| Sampled        | bool      |             | Sampling status field                                        | Fields not used                   |
| Err            | string    |             | Error Message                                                | Indirect correspondence to `dkspan` |
| Kind           | string    |             | Span Type                                                    | `dkspan.SpanType`                 |
| Timestamp      | structure | Microsecond | Microsecond time structure representation span starting time | `dkspan.Start`                    |
| Duration       | int64     | Microsecond | Span Time consuming                                          | `dkspan.Duration`                 |
| Shared         | bool      |             | Shared state                                                 | Fields not used                   |
| LocalEndpoint  | structure |             | to get Service Name                                          | `dkspan.Service`                  |
| RemoteEndpoint | structure |             | Communication peer                                           | `dkspan.Endpoint`                 |
| Annotations    | array     |             | Used to explain delay-related events                         | Fields not used                   |
| Tags           | map       |             | to get Span status                                           | `dkspan.Status`                   |
