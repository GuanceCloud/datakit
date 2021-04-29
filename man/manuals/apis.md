{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}

# Datakit web apis

## /v1/write/metric

### example:

```
http://{host}:{port}/v1/write/metric?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

### Body 数据 **(行协议)**:

_Tags_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

_Fields_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

## /v1/write/object

### example:

```
http://{host}:{port}/v1/write/object?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

### Body 数据 **(行协议)**:

_Tags_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

_Fields_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

## /v1/write/rum

### example:

```
http://{host}:{port}/v1/write/rum?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

### Body 数据 **(行协议)**:

_Tags_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

_Fields_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

## /v1/write/logging

### example:

```
http://{host}:{port}/v1/write/logging?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

### Measurement:

使用采集器配置文件中的 source 字段，如果此字段为空，则使用默认值 default。

### Body 数据 **(行协议)**:

_Tags_

> | 名称     | 描述                                                                      |
> | -------- | ------------------------------------------------------------------------- |
> | filename | 日志文件名（不带绝对路径文件名）                                          |
> | service  | 使用采集器配置文件中的 service 字段，如果此字段为空，则默认跟 source 相同 |

_Fields_

> | 名称    | 描述                                     |
> | ------- | ---------------------------------------- | ------ |
> | status  | 日志状态，分 9 级，默认为 info，详情见下 | string |
> | message | 基础日志数据，存放有效的单行或多行数据   | string |

_status_

> | status 有效字段值     | 结果值   |
> | --------------------- | -------- |
> | f/emerg               | emerg    |
> | a/alert               | alert    |
> | c/critical            | critical |
> | e/error               | error    |
> | w/warning             | warning  |
> | n/notice              | notice   |
> | i/info                | info     |
> | d/debug/trace/verbose | debug    |
> | o/s/OK                | OK       |

### Time:

默认使用此条日志采集到的时间。如果使用 pipeline 对日志文本进行切割，且切割后的 time 字段符合转换规则，会将 time 字段转换为标准时间应用在此。

## /v1/write/security

### example:

```
http://{host}:{port}/v1/write/security?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

### Body 数据 **(行协议)**:

_Tags_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

_Fields_

> | 名称 | 描述 |
> | ---- | ---- |
> |      |      |
> |      |      |

## /v1/write/tracing

### example:

```
http://{host}:{port}/v1/write/tracing?token={token}
```

### Method: _POST_

### URL 参数:

token={token_string}

### Measurement:

根据 Tracing 数据来源决定，目前支持 jeager/zipkin/skywalking/ddtrace 这几种。

### Body 数据 **(行协议)**:

_Tags_

> | 名称             | 描述            | 可选值                  |
> | ---------------- | --------------- | ----------------------- |
> | project          | 项目名称        |                         |
> | operation        | span 名         |                         |
> | service          | 服务名          |                         |
> | parent_id        | 父 span 编号    |                         |
> | trace_id         | 连路编号        |                         |
> | span_id          | span 编号       |                         |
> | version          | 版本号          |                         |
> | http_method      | HTTP 请求方法   |                         |
> | http_status_code | HTTP 请求状态码 |                         |
> | type             | span 请求类型   | app/db/web/cache/custom |
> | status           | span 状态       | ok/error                |
> | span_type        | span 类型       | entry/local/exit        |
> | container_host   | 容器内部主机名  |                         |

_Fields_

> | 名称     | 描述          | 单位     |
> | -------- | ------------- | -------- |
> | duration | span 持续时间 | unix sec |
> | start    | span 开始时间 | unix sec |
> | message  | span 原始数据 |          |
> | resource | 资源名        |          |

### Time:

使用 trace 原始数据中时间戳。
