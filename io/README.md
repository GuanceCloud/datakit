
# io 模块设计

io 模块主要负责收集各个采集器上报上来的数据,对其做基本的数据处理,然后按照一定的策略,将这些数据分门别类上传到中心.

    inputs --> Feed --> pipeline/filter --> mem-cache--.
                                                       |
       ___________________F L U S H____________________.
      |
      v     sinker?
    dw-APIs -------.     
      |            |
      |            |   Yes
      |            +---------> sinkers ---> endpoint
      |            |              |            |
      |         No + <------------'            |
      |            |                           |
      |            |                           |
      |            v                           |
      | +------------+                         | 
      `--> endpoint  |---> httpcli <-----------'
        |    ...     |       |
        +------------+       v
                          openway 


## Prometheus Metrics

io 模块暴露如下 metrics：

| 指标                                | 类型    | 说明                                                                                  | labels                |
| ---                                 | ---     | ---                                                                                   | ---                   |
| datakit_io_feed_total               | count   | Input feed total                                                                      | name,category         |
| datakit_io_feed_point_total         | count   | Input feed point total                                                                | name,category         |
| datakit_error_total                 | count   | total errors, only count on error source, not include error message                   | source,category       |
| datakit_io_input_filter_point_total | count   | Input filtered point total                                                            | name,category         |
| datakit_io_collect_latency          | summary | Input collect latency(us)                                                             | name,category         |
| datakit_io_queue_pts                | gauge   | IO module queued(cached) points                                                       | category              |
| datakit_io_last_feed                | gauge   | Input last feed time(unix timestamp in second)                                        | name,category         |
| datakit_last_err                    | gauge   | Datakit errors(when error occurred), these errors come from inputs or any sub modules | source,category,error |
| datakit_io_chan_capacity            | gauge   | IO channel capacity                                                                   | category              |
| datakit_io_chan_usage               | gauge   | IO module queued(cached) points                                                       | category              |
| datakit_io_flush_total              | count   | IO flush total                                                                        | category              |
| datakit_io_flush_failcache_total    | count   | IO flush fail-cache total                                                             | category              |
| datakit_io_flush_workers            | gauge   | IO flush workers                                                                      | category              |
