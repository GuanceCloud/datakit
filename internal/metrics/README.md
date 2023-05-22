# datakit 运行时指标

datakit 自身有如下运行时指标:

| 指标                             | 类型    | 说明                                                                                                                   | labels                                                    |
| ---                              | ---     | ---                                                                                                                    | ---                                                       |
| datakit_uptime                   | gauge   | Datakit uptime(second)                                                                                                 | version,build_at,branch,os_arch,docker,auto_update,cgroup |
| datakit_datakit_goroutines       | gauge   | goroutine count within Datakit                                                                                         | -                                                         |
| datakit_heap_alloc               | gauge   | Datakit memory heap bytes                                                                                              | -                                                         |
| datakit_sys_alloc                | gauge   | Datakit memory system bytes                                                                                            | -                                                         |
| datakit_cpu_usage                | gauge   | Datakit CPU usage(%)                                                                                                   | -                                                         |
| datakit_cpu_cores                | gauge   | Datakit CPU cores                                                                                                      | -                                                         |
| datakit_open_files               | gauge   | Datakit open files(only available on Linux)                                                                            | -                                                         |
| datakit_data_overuse             | gauge   | Does current workspace's data(metric/logging) usaguse(if 0 not beyond, or with a unix timestamp when overuse occurred) | -                                                         |
| datakit_gc_summary               | summary | Datakit golang GC paused(nano-second)                                                                                  | -                                                         |
| datakit_process_ctx_switch_total | count   | Datakit process context switch count(linux only)                                                                       | type                                                      |
| datakit_process_io_count_total   | count   | Datakit process IO count                                                                                               | type                                                      |
| datakit_process_io_bytes_total   | count   | Datakit process IO bytes count                                                                                         | type                                                      |

## Prometheuse 指标性能

经测试，100 个指标的性能表现如下：

``` shell
$ gtb BenchmarkP8s
goos: darwin
goarch: arm64
pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics
BenchmarkP8s
BenchmarkP8s/n-histogram-vec
BenchmarkP8s/n-histogram-vec-10                             6148        214176 ns/op      160175 B/op        2732 allocs/op
BenchmarkP8s/n-summary-vec-with-quantile
BenchmarkP8s/n-summary-vec-with-quantile-10                 9949        113302 ns/op      115364 B/op        2232 allocs/op
BenchmarkP8s/n-summary-vec
BenchmarkP8s/n-summary-vec-10                              10000        105920 ns/op       86565 B/op        1132 allocs/op
BenchmarkP8s/n-counter-vec
BenchmarkP8s/n-counter-vec-10                              12217         98239 ns/op       84167 B/op        1032 allocs/op
BenchmarkP8s/n-gauge-vec
BenchmarkP8s/n-gauge-vec-10                                12268         98022 ns/op       82564 B/op        1032 allocs/op
BenchmarkP8s/n-gauge-vec-with-long-label-vaule
BenchmarkP8s/n-gauge-vec-with-long-label-vaule-10          10000        103371 ns/op       82565 B/op        1032 allocs/op
BenchmarkP8s/n-gauge
BenchmarkP8s/n-gauge-10                                    15237         78838 ns/op       82566 B/op        1032 allocs/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics    13.659s

$ alias gtb
alias gtb='LOGGER_PATH=nul CGO_CFLAGS=-Wno-undef-prefix go test -run XXX -test.benchmem -test.v -bench'
```
