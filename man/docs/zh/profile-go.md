
# Go profiling

---

Go 内置了性能分析 (Profiling) 工具 `pprof`，可以采集程序运行中的性能数据，可通过以下两种方式使用：

- `runtime/pprof`: 通过编程方式，自定义采集运行数据，然后保存分析
- `net/http/pprof`: 调用 `runtime/pprof`，封装成接口，通过 HTTP Server 的方式对外提供性能数据

性能数据主要包括如下：

- `goroutine`: 运行的 Goroutine 的调用栈分析
- `heap`: 活跃对象的内存分配情况
- `allocs`: 所有对象的内存分配情况
- `threadcreate`: OS 线程创建分析
- `block`: 阻塞分析
- `mutex`: 互斥锁分析

收集到的数据，可以通过官方 [`pprof`](https://github.com/google/pprof/blob/main/doc/README.md) 工具进行分析。

DataKit 可通过[主动拉取](profile-go.md#pull-mode) (pull) 或[被动推送](profile-go.md#push-mode) (push) 的方式来获取这些数据。

## push 方式 {#push-mode}

### DataKit 配置 {#push-datakit-config}

DataKit 开启 [profile](profile.md#config)  采集器，注册 profile http 服务。

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]
```

### Go 应用配置 {#push-app-config}

集成 DataDog 开源库 [dd-trace-go](https://github.com/DataDog/dd-trace-go)，采集应用性能数据并发送至 DataKit。 代码参考如下：

```go
package main

import (
    "log"
    "time"

    "gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
    err := profiler.Start(
        profiler.WithService("dd-service"),
        profiler.WithEnv("dd-env"),
        profiler.WithVersion("dd-1.0.0"),
        profiler.WithTags("k:1", "k:2"),
        profiler.WithAgentAddr("localhost:9529"), // DataKit url
        profiler.WithProfileTypes(
            profiler.CPUProfile,
            profiler.HeapProfile,
            // The profiles below are disabled by default to keep overhead
            // low, but can be enabled as needed.

            // profiler.BlockProfile,
            // profiler.MutexProfile,
            // profiler.GoroutineProfile,
        ),
    )

    if err != nil {
        log.Fatal(err)
    }
    defer profiler.Stop()

    // your code here
    demo()
}

func demo() {
    for {
        time.Sleep(100 * time.Millisecond)
        go func() {
            buf := make([]byte, 100000)
            _ = len(buf)
            time.Sleep(1 * time.Hour)
        }()
    }
}
```

运行该程序后，DDTrace 会定期（默认 1 分钟一次）将数据推送给 DataKit。

## Pull 方式 {#pull-mode}

### Go 应用开启 Profiling {#app-config}

应用中开启 Profiling 只需要引用 `pprof` 包即可，参考如下:

```go
package main

import (
  "net/http"
   _ "net/http/pprof"
)

func main() {
    http.ListenAndServe(":6060", nil)
}
```

运行代码后，可通过 `http://localhost:6060/debug/pprof/heap?debug=1` 来查看是否开启成功。

- Mutex 和 Block 性能分析

默认情况下，mutex 和 block 性能采集并未开启，如果需要开启，可添加如下代码：

```go
var rate = 1

// enable mutex profiling
runtime.SetMutexProfileFraction(rate)

// enable block profiling
runtime.SetBlockProfileRate(rate)
```

`rate` 设置采集频率，即 1 / rate 的事件被采集， 如设置为 0或小于 0 的数值，是不进行采集的。

### DataKit 配置 {#datakit-config}

[开启 Profile 采集器](profile.md)，进行如下设置 `[[inputs.profile.go]]`。

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]

  ## set true to enable election
  election = true

 ## go pprof config
[[inputs.profile.go]]
  ## pprof url
  url = "http://localhost:6060"

  ## pull interval, should be greater or equal than 10s
  interval = "10s"

  ## service name
  service = "go-demo"

  ## app env
  env = "dev"

  ## app version
  version = "0.0.0"

  ## types to pull 
  ## values: cpu, goroutine, heap, mutex, block
  enabled_types = ["cpu","goroutine","heap","mutex","block"]

[inputs.profile.go.tags]
  # tag1 = "val1"
```

> 注意： 如果不需要开启 Profile 的 HTTP 服务，可将 endpoints 字段注释掉。

### 字段说明 {#fields-info}

- `url`: 上报地址，如 `http://localhost:6060`
- `interval`: 采集间隔时间，最小 10s
- `service`： 服务名称
- `env`： 应用环境类型
- `version`: 应用的版本
- `enabled_types`: 性能类型，如 `cpu, goroutine, heap, mutex, block`

配置好 Profile 采集器，启动或重启 DataKit，一段时间后即可在观测云中心查看 Go 的性能数据。
