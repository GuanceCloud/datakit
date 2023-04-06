# datakit 运行时指标

datakit 自身有如下运行时指标:

| 指标                       | 类型    | 说明                                                                                                                   | labels                                                    |
| ---                        | ---     | ---                                                                                                                    | ---                                                       |
| datakit_uptime             | gauge   | Datakit uptime(second)                                                                                                 | version,build_at,branch,os_arch,docker,auto_update,cgroup |
| datakit_datakit_goroutines | gauge   | goroutine count within Datakit                                                                                         | -                                                         |
| datakit_heap_alloc         | gauge   | Datakit memory heap bytes                                                                                              | -                                                         |
| datakit_sys_alloc          | gauge   | Datakit memory system bytes                                                                                            | -                                                         |
| datakit_cpu_usage          | gauge   | Datakit CPU usage(%)                                                                                                   | -                                                         |
| datakit_cpu_cores          | gauge   | Datakit CPU cores                                                                                                      | -                                                         |
| datakit_open_files         | gauge   | Datakit open files(only available on Linux)                                                                            | -                                                         |
| datakit_data_overuse       | gauge   | Does current workspace's data(metric/logging) usaguse(if 0 not beyond, or with a unix timestamp when overuse occurred) | -                                                         |
| datakit_gc_summary         | summary | Datakit golang GC paused(nano-second)                                                                                  | -                                                         |
