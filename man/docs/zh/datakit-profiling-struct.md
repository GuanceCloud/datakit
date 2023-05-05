
# Profiling Data Structure
---

## 简述 {#intro}

本文介绍 Datakit 中 profiling 相关数据结构定义

数据转换流：外部 Profiling 数据结构 --> Datakit Profiling  --> Line Protocol

## Datakit Point Protocol Structure for Profiling {#point-struct}

### Datakit Line Protocol {#line-proto}

- Line Protocol 为数据流最后落盘数据
- Line Protocol 数据结构是由 Name, Tags, Fields, Timestamp 四部分和分隔符 (英文逗号，空格) 组成的字符串，形如：

```line protocol
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

> Line Protocol 以下简称 DKProto。

| Section | Name         | Unit | Description                          |
| ---     | ---          | ---  | ---                                  |
| Tag     | host         |      | host name                            |
| Tag     | endpoint     |      | end point of resource                |
| Tag     | service      |      | service name                         |
| Tag     | env          |      | environment arguments                |
| Tag     | version      |      | service version                      |
| Tag     | language     |      | language [Java, Python, Golang, ...] |
| Tag     | Runtime      |      | runtime [jvm, CPython, go, ....]     |
| Tag     | runtime_os   |      | os                                   |
| Tag     | runtime_arch |      | cpu architecture                     |
| Field   | profile_id   |      | profiling ID                         |
| Field   | agent_ver    |      | 客户端 agent 版本                    |
| Field   | start        | 微秒 | profiling start timestamp            |
| Field   | end          | 微秒 | profiling end timestamp              |
| Field   | duration     | 微秒 | profiling duration                   |
| Field   | pid          |      | process id                           |
| Field   | format       |      | binary profiling file format         |
| Field   | __file_size  | Byte | profiling 文件大小                   |

### Datakit Profiling Structure {#profiling-struct}

Datakit Profiling 是 Datakit 使用的用于表示 profiling 的结构。

| Field Name  | Data Type               | Unit | Description                                                          | Correspond To                 |
| ---         | ---                     | ---  | ---                                                                  | ----                          |
| ProfileId   | `string`                |      | profiling 唯一 ID                                                    | `dkproto.fields.profile_id`   |
| AgentVer    | `string`                |      | profiling agent 库版本                                               | `dkproto.fields.agent_ver`    |
| Endpoint    | `string`                |      | 通信端                                                               | `dkproto.fields.endpoint`     |
| Service     | `string`                |      | Service Name                                                         | `dkproto.tags.service`        |
| Env         | `string`                |      | Environment Variables                                                | `dkproto.tags.env`            |
| Version     | `string`                |      | App 版本号                                                           | `dkproto.tags.version`        |
| Start       | `int64`                 | 纳秒 | profiling 采样开始时间                                               | `dkproto.fields.start`        |
| End         | `int64`                 | 纳秒 | profiling 采样结束时间                                               | `dkproto.fields.end`          |
| Duration    | `int64`                 | 纳秒 | 本次采样持续时间，通常为 1min                                        | `dkproto.fields.duration`     |
| Host        | `string`                |      | 主机/容器 hostname                                                   | `dkproto.tags.host`           |
| PID         | `string`                |      | Process ID                                                           | `dkproto.fields.pid`          |
| Language    | `string`                |      | 程序语言                                                             | `dkproto.fields.language`     |
| LanguageVer | `string`                |      | 程序语言版本                                                         | `dkproto.fields.language_ver` |
| Runtime     | `string`                |      | 运行时环境，`jvm/cpython/...`                                        | `dkproto.fields.runtime`      |
| RuntimeOs   | `string`                |      | 操作系统                                                             | `dkproto.tags.runtime_os`     |
| RuntimeArch | `string`                |      | CPU 架构：amd64, arm64...                                            | `dkproto.tags.runtime_arch`   |
| Format      | `string`                |      | profiling 二进制文件采用的格式，jfr, pprof...                        | `dkproto.fields.format`       |
| Tags        | `map[string]string`     |      | profiling Tags                                                       | `dkproto.tags`                |
| OssPath     | `[]string`              |      | 原始 profiling 二进制文件存储在 OSS 的路径, 用于后续解析和供用户下载 | `dkproto.fields.oss_path`     |
| Metrics     | `map[string]string`     |      | 从 profiling 二进制文件中解析出的 相关指标                           | `dkproto.fields.metrics`      |
| Samples     | `map[EventType]*Sample` |      | profiling 相关采样性能指标                                           | `dkproto.fields.samples`      |

### Sample Structure {#sample-struct}

| Field Name    | Data Type         | Description               |
| ---           | ---               | ---                       |
| Values        | []SampleValue     | 指标值                    |
| SpanId        | string            | 相关联的 trace span id    |
| RootSpanId    | string            | 关联的 trace root span id |
| TraceEndpoint | string            | trace resource            |
| Labels        | map[string]string | tag                       |
| StackTrace    | []TraceFunc       | 调用栈                    |

### SampleValue Structure {#sample-value-struct}

| Field Name | Data Type | Description                          |
| ---        | ---       | ---                                  |
| Type       | string    | 值的类别，`cpu/wall/inuse_space/...` |
| Value      | int64     | 具体数值                             |
| Unit       | string    | 值得单位                             |

### TraceFunc Structure {#trace-func-struct}

| Field Name | Data Type | Description |
| ---        | ---       | ---         |
| Name       | string    | 方法名      |
| File       | string    | 代码源文件  |
| Line       | int       | 行号        |

### FlameGraph Structure {#framegraph-struct}

> profiling 详情查看页火焰图数据结构

| Field Name         | Data Type | Description              |
| ---                | ---       | ---                      |
| Unit               | string    | 火焰图数值单位, 如：毫秒 |
| AvailableDimension | []string  | 支持的展示维度           |
| Dimension          | string    | 当前所选的展示维度       |
| RootFrame          | []Frame   | 火焰图层                 |

### Frame Structure {#frame-struct}

> 注意：根据语言不同，某些属性可能没有。

| Field Name | Data Type | Description  |
| ---        | ---       | ---          |
| Value      | int64     | 值           |
| Method     | string    | 方法名       |
| Line       | int       | 代码行号     |
| SourceFile | string    | 代码源文件名 |
| Thread     | string    | 所属线程     |
| Modifier   | string    | 方法修饰符   |
| Library    | string    | 代码库       |
| Package    | string    | 包名         |
| Class      | string    | 类名         |
| SubFrame   | []Frame   | 下级 frame   |

### Flame Graph 响应 JSON 示例 {#flamegraph-json-example}

```json
{
    "unit": "毫秒",
    "available_dimension": [
        "method",
        "package",
        "line",
        "thread"
    ],
    "dimension": "method",
    "root_frame": {
        "value": 10000,
        "method": "List TaskProcessor.processTasks(List, int, TimeUnit, TaskMethod)",
        "line": 239,
        "sub_frame": [
            {
                "value": 3000,
                "method": "Unsafe.park(boolean, long)",
                "line": 319,
                "sub_frame": [],
                "thread": "",
                "modifier": "",
                "library": "Standard Library",
                "package": "com.sun.management.internal",
                "class": "PlainSocketImpl"
            },
            {
                "value": 7000,
                "method": "SocketInputStream.socketRead0(FileDescriptor, byte[], int, int, int)",
                "line": 403,
                "sub_frame": [],
                "thread": "",
                "modifier": "",
                "library": "Standard Library",
                "package": "com.sun.management.internal",
                "class": "SocketInputStream"
            }
        ],
        "thread": "",
        "modifier": "public",
        "library": "Standard Library",
        "package": "com.sun.management.internal",
        "class": "PlainSocketImpl"
    }
}
```
