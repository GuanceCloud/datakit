# eBPF 数据

## eBPF Span

### eSpan 字段描述

- **Category**: `Tracing` ( or `Logging` )

- **Source**: `"ebpf"`

- **Tags**:

    | 名称          | 类型 | 描述                                        |
    | ------------- | ---- | ------------------------------------------- |
    | host          | str  | 主机名                                      |
    | app_trace_id  | str  | dd/otel/skywk 等传播的 trace id             |
    | app_parent_id | str  | dd/otel/skywk 等传播的 trace parent span id |
    | proc_trace_id | str  | app 进程内部跟踪 id                         |
    | net_trace_id  | str  | app 网络跟踪 id                             |
    | service       | str  | 服务名                                      |
    | direction     | str  | 请求方向，值 `incoming` 或 `outgoing`       |

- **Fields**:

    | 名称 | 类型 | 描述 |
    | ---- | ---- | ---- |

- **Time**

### 构建链路

1. 关联 eSpan 和 APM

1. 单独构建 eBPF 链路

## BPF Network(L4/L7) Log
