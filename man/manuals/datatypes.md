{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# DataKit 支持的数据类型

本文档是 DataKit 以及 DataFlux 所采集到的各种数据类型的描述，主要涉及如下几类：

- Metric（TODO）
- Object（TODO）
- RUM（TODO）
- Logging
- Security（TODO）
- Event（TOO）
- Tracing

## Logging 行协议文档

### measurement

使用采集器配置文件中的 `source` 字段，如果此字段为空，则使用默认值 `default`。

### 标签（tag）

| 名称       | 描述                                                                          |
| :--        | ---                                                                           |
| `filename` | 日志文件名（不带绝对路径文件名）                                              |
| `service`  | 使用采集器配置文件中的 `service` 字段，如果此字段为空，则默认跟 `source` 相同 |

### 指标列表（field）

| 指标名称 | 描述                                    | 数据类型 |
| :--      | ---                                     | ---      |
| status   | 日志状态，分9级，默认为`info`，详情见下 | String   |
| message  | 基础日志数据，存放有效的单行或多行数据  | String   |

为保证 `status` 字段内容的规范，会被部分值进行转换，最终所有的 `status` 字段值为下表的【结果值】。如果转换失败，则使用默认值`info`。

| status 有效字段值       | 结果值     |
| ---                     | ----       |
| `f/emerg`               | `emerg`    |
| `a/alert`               | `alert`    |
| `c/critical`            | `critical` |
| `e/error`               | `error`    |
| `w/warning`             | `warning`  |
| `n/notice`              | `notice`   |
| `i/info`                | `info`     |
| `d/debug/trace/verbose` | `debug`    |
| `o/s/OK`                | `OK`       |

### time

默认使用此条日志采集到的时间。

如果使用 pipeline 对日志文本进行切割，且切割后的 `time` 字段符合转换规则，会将 `time` 字段转换为标准时间应用在此。

## Tracing 行协议文档

### measurement

根据 Tracing 数据来源决定，目前支持 `jeager/zipkin/skywalking/ddtrace` 这几种。

### 标签（tag）

| 指标               | 描述            | 可选值                    |
| :--                | ---             | ---                       |
| `project`          | 项目名          | -                         |
| `operation`        | span 名         | -                         |
| `service`          | 服务名          | -                         |
| `parent_id`        | 父 span 编号    | -                         |
| `trace_id`         | 链路编号        | -                         |
| `span_id`          | span 编号       | -                         |
| `version`          | 版本号          | -                         |
| `http_method`      | HTTP 请求方法   | -                         |
| `http_status_code` | HTTP 请求状态码 | -                         |
| `type`             | span 请求类型   | `app/db/web/cache/custom` |
| `status`           | span 状态       | `ok/error`                |
| `span_type`        | span 类型       | `entry/local/exit`        |
| `container_host`   | 容器内部主机名   | -                         |


### 指标（field)

| 指标       | 描述           | 数据类型 | 单位 |
| :--        | ---            | ---      | ---  |
| `duration` | span持续时间   | int      | usec |
| `start`    | span开始时间戳 | int      | usec |
| `message`  | span原始数据   | string   | -    |
| `resource` | 资源名         | string   | -    |

### time

使用 trace 原始数据中时间戳。
