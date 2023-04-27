# dataway 模块设计

dataway 模块由如下三个部分组成：

- endpoint: 表示一个具体的 remote-host，这里通常指 dataway server
- sinker：dataway 上的分流模块
- httpcli：具体的 HTTP 发送模块

它们之间的关系如下：

            sinker?
    dw APIs -------.
      |            |
      |            |    Y
      |            +---------> sinkers ---> endpoint
      |            |              |            |
      |          N + <------------'            |
      |            |                           |
      |            |                           |
      |            v                           |
      | +------------+                         | 
      `--> endpoint  |---> httpcli <-----------'
        |    ...     |       |
        +------------+       v
                          openway 
      
- io 模块数据要发送的时候，直接将数据丢给 dataway 对象
- dataway 判断，当前是否挂有 sinker 要处理，如果有，将数据丢 sinker 处理
- sinker(s) 将符合条件的数据，过滤出来交给它自己的 endpoint 去处理，剩余的数据还给 dataway 继续处理
- dataway 将剩余的数据（没有 sinker 时为全量数据），交给自己的 endpoint 处理，目前 dataway 支持挂多个 endpoint
- 所有 endpoint 将数据通过 httpcli 发送给 openway。目前的 httpcli 支持重试机制（retryablehttp）
- 其它 API 都是通过 dataway endpoint 发送出去的（对于有多个 endpoint 的情况，只将请求发送到第一个 endpoint）

## Prometheus Metrics

Dataway 模块暴露如下 metrics：

| 指标                                    | 类型    | 说明                                                                                     | labels          |
| ---                                     | ---     | ---                                                                                      | ---             |
| datakit_io_dataway_api_request_total    | count   | dataway HTTP request processed, partitioned by status code and HTTP API(url path)        | api,status      |
| datakit_io_dataway_point_total          | count   | dataway uploaded points, partitioned by category and send status(HTTP status)            | category,status |
| datakit_io_dataway_point_bytes_total    | count   | dataway uploaded points bytes, partitioned by category and pint send status(HTTP status) | category,status |
| datakit_io_dataway_sink_total           | count   | dataway sink count, partitioned by category.                                             | category        |
| datakit_io_dataway_sink_point_total     | count   | dataway sink points, partitioned by category and point send status(HTTP status)          | category,status |
| datakit_io_dataway_not_sink_point_total | count   | dataway not-sinked points(condition or category not match)                               | category        |
| datakit_io_dataway_api_latency          | summary | dataway HTTP request latency(ms) partitioned by HTTP API(url path) and HTTP status       | api,status      |
| datakit_io_flush_failcache_bytes        | summary | IO flush fail-cache bytes(in gzip) summary                                               | category        |
