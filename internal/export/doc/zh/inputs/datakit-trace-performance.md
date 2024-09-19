# Datakit Trace Agent 性能报告

以下测试在真实物理环境下进行，并使用测试工具发送饱和数据。

> 运行 Datakit 物理机参数

| CPU  | 内存 | 带宽    |
| ---- | ---- | ------- |
| 1 核 | 2G   | 100Mbps |

> Datakit 使用配置

安装默认

> 测试工具配置参数（发送饱和数据）

| 开启线程 | 单线程请求次数 | 单请求 Span 数量 |
| -------- | -------------- | ---------------- |
| 100      | 1000           | 10               |

> 测试工具下载地址

- [DDTrace 测试工具](https://github.com/CodapeWild/dktrace-dd-agent/releases){:target="_blank"}
- [Jaeger 测试工具](https://github.com/CodapeWild/dktrace-jaeger-agent/releases){:target="_blank"}
- [OpenTelemetry 测试工具](https://github.com/CodapeWild/dktrace-otel-agent/releases){:target="_blank"}
- [Zipkin 测试工具](https://github.com/CodapeWild/dktrace-zipkin-agent/releases){:target="_blank"}

## DDTrace 性能报告 {#ddtrace-performace}

测试 API: `/v0.4/traces`

> 不开启 `ddtrace.threads` 不开启 `ddtrace.storage`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 60.92  | 66.47   | 100.00      | 727.68           | 7.89                   | 14.96        |

> 开启 `ddtrace.threads(buffer=100 threads=8)` 开启 `ddtrace.storage(capacity=5120)`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 60.92  | 66.69   | 50.00       | 399.07           | 8.17                   | 18.98        |

## Jaeger 性能报告 {#jaeger-performace}

测试 API: `/apis/traces`

> 不开启 `jaeger.threads` 不开启 `jaeger.storage`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 80.94  | 63.24   | 100.00      | 511.17           | 5.23                   | 24.90        |

> 开启 `jaeger.threads(buffer=100 threads=8)` 开启 `jaeger.storage(capacity=5120)`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 62.95  | 60.66   | 200.00      | 912.09           | 4.67                   | 1.37         |

## OpenTelemetry 性能报告 {#opentelemetry-performace}

测试 API: `/otel/v1/trace`

> 不开启 `opentelemetry.threads` 不开启 `opentelemetry.storage`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 65.99  | 67.72   | 100.00      | 262.07           | 2.68                   | 21.98        |

> 开启 `opentelemetry.threads(buffer=100 threads=8)` 开启 `opentelemetry.storage(capacity=5120)`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 52.94  | 47.26   | 50.00       | 130.99           | 2.68                   | 3.07         |


## Zipkin 性能报告 {#zipkin-performace}

测试 API: `/api/v2/spans`

> 不开启 `zipkin.threads` 不开启 `zipkin.storage`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 70.98  | 66.26   | 100.00      | 822.16           | 8.42                   | 37.01        |

> 开启 `zipkin.threads(buffer=100 threads=8)` 开启 `zipkin.storage(capacity=5120)`

| CPU(%) | Mem(mb) | 请求次数(k) | 数据总发送量(mb) | 单次请求数据包大小(kb) | 接口延迟(ms) |
| ------ | ------- | ----------- | ---------------- | ---------------------- | ------------ |
| 59.97  | 51.88   | 50.00       | 410.51           | 8.41                   | 16.59        |
