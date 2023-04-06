
# filter 模块设计

io 模块主要负责对 inputs 采集到的数据进行过滤,并丢弃符合条件的数据.

    inputs --> Feed --> pipeline/filter --> mem-cache...

## Prometheus Metrics

filter 模块暴露如下 metrics：

| 指标                               | 类型    | 说明                                              | labels                  |
| ---                                | ---     | ---                                               | ---                     |
| datakit_filter_point_dropped_total | count   | Dropped points of filters                         | category,filters,source |
| datakit_filter_point_total         | count   | Filter points of filters                          | category,filters,source |
| datakit_filter_update_total        | count   | Filters(remote) updated count                     | -                       |
| datakit_filter_pull_latency        | summary | Filter pull(remote) latency(ms)                   | status                  |
| datakit_filter_latency             | summary | Filter latency(us) of these filters               | category,filters,source |
| datakit_filter_point_dropped_total | count   | Dropped points of filters                         | category,filter,source  |
| datakit_filter_last_update         | gauge   | filter last update time(in unix timestamp second) | -                       |
