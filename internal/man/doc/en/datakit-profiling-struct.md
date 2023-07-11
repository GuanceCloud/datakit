# Profiling Data Struct
---

This page will introduce about profiling data structure used in Datakit.

## Datakit Line Protocol Structure {#line-protocol}

- Line Protocol is the final format storing on disk
- Line Protocol data structure consists of four parts: "Name", "Tags", "Fields" and "Timestamp", which is separated by comma, for example：

```line protocol
source_name,key1=value1,key2=value2 field1=value1,field2=value2 ts
```

## Datakit line protocol tags and fields used in profiling {#tags-fields}

| <span style="color:green">**Section**</span> | <span style="color:green">**Name**</span> | <span style="color:green">**Unit**</span> | <span style="color:green">**Description**</span>        |
|----------------------------------------------|-------------------------------------------|-------------------------------------------|---------------------------------------------------------|
| Tag                                          | host                                      |                                           | host name                                               |
| Tag                                          | endpoint                                  |                                           | end point of resource                                   |
| Tag                                          | service                                   |                                           | service name                                            |
| Tag                                          | env                                       |                                           | environment arguments                                   |
| Tag                                          | version                                   |                                           | service version                                         |
| Tag                                          | language                                  |                                           | language [Java, Python, Golang, ...]                    |
| Field                                        | runtime                                   |                                           | runtime [jvm, CPython, go, ....]                        |
| Field                                        | runtime_os                                |                                           | operating system                                        |
| Field                                        | runtime_arch                              |                                           | cpu architecture                                        |
| Field                                        | runtime_version                           |                                           | programming language version                            |
| Field                                        | runtime_compiler                          |                                           | compiler                                                |
| Field                                        | runtime_id                                |                                           | allocated unique ID once process bootstrap              |
| Field                                        | profiler                                  |                                           | profiler library name [DDTrace, py-spy, pyroscope, ...] |
| Field                                        | library_ver                               |                                           | profiler library version                                |
| Field                                        | profiler_version                          |                                           | profiler library version                                |
| Field                                        | profile_id                                |                                           | profiling unique ID                                     |
| Field                                        | datakit_ver                               |                                           | datakit version                                         |
| Field                                        | start                                     | nanosecond                                | profiling start timestamp                               |
| Field                                        | end                                       | nanosecond                                | profiling end timestamp                                 |
| Field                                        | duration                                  | nanosecond                                | profiling duration                                      |
| Field                                        | pid                                       |                                           | process id                                              |
| Field                                        | process_id                                |                                           | process id                                              |
| Field                                        | format                                    |                                           | profiling file format                                   |
| Field                                        | __file_size                               | Byte                                      | profiling file total size                               |

