
# Go profiling

---

Golang built-in tool `pprof` can be used to profiling go process.

- `runtime/pprof`: By programming, output profiling data to a file.
- `net/http/pprof`: Download profiling file by http request.

Types of profiles available::

- `goroutine`: Stack traces of all current goroutines
- `heap`: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
- `allocs`: A sampling of all past memory allocations
- `threadcreate`: Stack traces that led to the creation of new OS threads
- `block`: Stack traces that led to blocking on synchronization primitives
- `mutex`: Stack traces of holders of contended mutexes

You can use official tool [`pprof`](https://github.com/google/pprof/blob/main/doc/README.md){:target="_blank"} to analysis generated profile file.

DataKit can use either [Pull mode](profile-go.md#pull-mode) or [Push mode](profile-go.md#push-mode) to generate profiling file.

## push mode {#push-mode}

### Config DataKit {#push-datakit-config}

Enable [profile](profile.md#config)  inputs

```toml
[[inputs.profile]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/profiling/v1/input"]
```

### Integrate dd-trace-go {#push-app-config}

Import [dd-trace-go](https://github.com/DataDog/dd-trace-go){:target="_blank"}, Insert code as follows to your application:

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

Once your go app start, dd-trace-go will send profiling data to DataKit by interval(per 1min by default).

## pull mode {#pull-mode}

### Enable profiling in app {#app-config}

import `pprof` package in your code:

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

Once start your app, you can view page `http://localhost:6060/debug/pprof/heap?debug=1` in browser to confirm running as your wish.

- Mutex and Block events

Mutex and Block events are disable by default, if you want to enable them, add below code to your app:

```go
var rate = 1

// enable mutex profiling
runtime.SetMutexProfileFraction(rate)

// enable block profiling
runtime.SetBlockProfileRate(rate)
```

Set the collection frequency, where 1/rate events are collected. Values set to 0 or less are not collected.

### Config DataKit {#datakit-config}

[Enable Profile Input](profile.md), modify `[[inputs.profile.go]]` segment as follows.

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

<!-- markdownlint-disable MD046 -->
???+ note
    If there is no need to enable profile http endpoint, just comment `endpoints` item.
<!-- markdownlint-enable -->

### Field introduction {#fields-info}

- `url`: net/http/pprof listening address, such as `http://localhost:6060`
- `interval`: upload interval, 最小 10s
- `service`:  your service name
- `env`:  your app running env
- `version`: your app version
- `enabled_types`: available events: `cpu, goroutine, heap, mutex, block`

You should Restart DataKit after modification. After a minute or two, you can visualize your profiles on the [profile](https://console.guance.com/tracing/profile){:target="_blank"}.
