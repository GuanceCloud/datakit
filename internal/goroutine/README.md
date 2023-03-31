
# goroutine 管理模块设计

goroutine 管理模块主要用来管理 datakit 自身所消耗的 goroutine,对其进行分组/耗时/崩溃等信息的统计.

## Prometheus Metrics

本模块暴露如下 metrics：

| 指标                            | 类型  | 说明                                  | labels |
| ---                             | ---   | ---                                   | ---    |
| datakit_goroutine_groups        | gauge | goroutine group count                 | -      |
| datakit_goroutine_cost          | gauge | goroutine running time(in nanosecond) | name   |
| datakit_goroutine_stopped_total | count | stopped goroutines                    | name   |
| datakit_goroutine_alive         | gauge | alive goroutines                      | name   |
