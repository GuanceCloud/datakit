# Datakit Profiling 相关数据结构
---

本文介绍 Datakit 中 profiling 相关数据结构定义。

## Datakit 行协议介绍 {#line-protocol}

- Line Protocol 为数据流最后落盘数据
- Line Protocol 数据结构是由 Name, Tags, Fields, Timestamp 四部分和分隔符 (英文逗号，空格) 组成的字符串，形如：

```line protocol
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

## Datakit profiling 行协议中所使用的 tags 和 fields {#tags-fields}

| Section | Name               | Unit       | Description                                                   |
|---------|--------------------|------------|---------------------------------------------------------------|
| Tag     | `host`             |            | host name                                                     |
| Tag     | `endpoint`         |            | end point of resource                                         |
| Tag     | `service`          |            | service name                                                  |
| Tag     | `env`              |            | environment arguments                                         |
| Tag     | `version`          |            | service version                                               |
| Tag     | `language`         |            | language [`Java`, `Python`, `Golang`, ...]                    |
| Field   | `runtime`          |            | runtime [`jvm`, `CPython`, `go`, ....]                        |
| Field   | `runtime_os`       |            | operating system                                              |
| Field   | `runtime_arch`     |            | cpu architecture                                              |
| Field   | `runtime_version`  |            | programming language version                                  |
| Field   | `runtime_compiler` |            | compiler                                                      |
| Field   | `runtime_id`       |            | allocated unique ID once process bootstrap                    |
| Field   | `profiler`         |            | profiler library name [`DDTrace`, `py-spy`, `Pyroscope`, ...] |
| Field   | `library_ver`      |            | profiler library version                                      |
| Field   | `profiler_version` |            | profiler library version                                      |
| Field   | `profile_id`       |            | profiling unique ID                                           |
| Field   | `datakit_ver`      |            | Datakit version                                               |
| Field   | `start`            | nanosecond | profiling start timestamp                                     |
| Field   | `end`              | nanosecond | profiling end timestamp                                       |
| Field   | `duration`         | nanosecond | profiling duration                                            |
| Field   | `pid`              |            | process id                                                    |
| Field   | `process_id`       |            | process id                                                    |
| Field   | `format`           |            | profiling file format                                         |
| Field   | `__file_size`      | Byte       | profiling file total size                                     |
