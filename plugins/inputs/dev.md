## Logging 行协议文档

### measurement

使用采集器配置文件中的 `source` 字段，如果此字段为空，则使用默认值 `default`。

### tags

| 指标     | 描述                                                                          | 数据类型 |
| :--      | ---                                                                           | ---      |
| filename | 日志文件名，在 DataKit v1.1.3-rc4 之后，此字段由文件绝对路径改为只有文件名    | String   |
| service  | 使用采集器配置文件中的 `service` 字段，如果此字段为空，则默认跟 `source` 相同 | String   |

### fields

| 指标    | 描述                                    | 数据类型 |
| :--     | ---                                     | ---      |
| status  | 日志状态，分9级，默认为`info`，详情见下 | String   |
| message | 基础日志数据，存放有效的单行或多行数据  | String   |

为保证`status`字段内容的规范，会被部分值进行转换，最终所有的`status`字段值都将为下标的“结果值”。如果转换失败，则使用默认值`info`。

| status 有效字段值                | 结果值     |
| :---                             | --         |
| `f`, `emerg`                     | `emerg`    |
| `a`, `alert`                     | `alert`    |
| `c`, `critical`                  | `critical` |
| `e`, `error`                     | `error`    |
| `w`, `warning`                   | `warning`  |
| `n`, `notice`                    | `notice`   |
| `i`, `info`                      | `info`     |
| `d`, `debug`, `trace`, `verbose` | `debug`    |
| `o`, `s`, `OK`                   | `OK`       |

### time

默认使用此条日志采集到的时间。

如果使用pipeline对日志文本进行切割，且切割后的`time`字段符合转换规则，会将`time`字段转换为标准时间应用在此。


## Tracing 行协议文档

### measurement

根据 Tracing 数据来源决定，目前支持 `jeager`，`zipkin`，`skywalking`与`ddtrace`。


### tags

| 指标     | 描述  | 数据类型 |  可选值 |
| :--      | ---    | ---      |   ---    |
| project           | 项目名         | String   |   -    |
| operation         | span名         | String   |   -    |
| service           | 服务名         | String   |   -    |
| parent_id         | 父span编号     | String   |   -    |
| trace_id          | 链路编号       | String   |   -    |
| span_id           | span编号       | String   |   -    |
| version           | 版本号         | String   |   -    |
| http_method       | http请求方法   | String   |   -    |
| http_status_code  | http请求状态码 | String   |   -    |
| type              | span请求类型   | String   |   app，db，web，cache，custom    |
| status            | span状态       | String   |   ok，error    |
| span_type         | span类型       | String   |   entry，local，exit    |


### fields

| 指标     | 描述           | 数据类型 | 单位 |
| :--      | ---           | ---     | ---  |
| duration | span持续时间   | Int     |  毫秒 |
| start    | span开始时间戳 | Int     |  毫秒 |
| message  | span原始数据   | String  |  -   |
| resource | 资源名         | String  |  -   |

### time

使用 trace 原始数据中时间戳。