---
title     : 'DDTrace Golang'
summary   : 'Tracing Golang application with DDTrace'
tags      :
  - 'DDTRACE'
  - 'GOLANG'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/ddtrace'
---

There are two ways to instrument a Go application:

## 1. Compile-time Instrumentation {#compile-time}

- **Does not require source-code changes** and instruments supported libraries at build time.
- **Works well in CI/CD** when you want broad, centrally managed coverage.

## 2. Manual Instrumentation {#manual}

Use `dd-trace-go` in code to create spans around selected operations. This option:

- **Gives precise control** over which parts of the application are traced.
- **Requires source-code changes**.

This guide uses `dd-trace-go/v2`. Upstream no longer maintains v1. A legacy application that cannot migrate must keep its matching v1 API; never mix v1 and v2 import paths.

---

### Requirements {#requirements}

- Applications must use Go modules. Module vendoring is supported.
- Go must be **1.18+**; for production, prefer a Go version still maintained upstream.
- Install DataKit and enable the [DDTrace collector](ddtrace.md){:target="_blank"}.

<!-- markdownlint-disable MD046 -->

=== "Recommended Workflow"

    1. Install Orchestrion and ensure `$(go env GOBIN)` or `$(go env GOPATH)/bin` is on `PATH`:

    ```shell
    go install github.com/DataDog/orchestrion@latest
    ```

    2. Register Orchestrion in the project root:

    ```shell
    orchestrion pin
    ```

    This updates `go.mod`, `go.sum`, and creates `orchestrion.tool.go`. Commit those files with the application so local and CI builds use the same instrumentation dependency.

    3. Build, test, and run through Orchestrion:

    ```shell
    orchestrion go build .
    orchestrion go test ./...
    DD_SERVICE=my-go-service DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 \
      orchestrion go run .
    ```

<!-- cspell:ignore toolexec -->
=== "Use -toolexec in CI"

    If the build system cannot invoke `orchestrion go`, pass `-toolexec` explicitly to each relevant `go` command:

    ```shell
    go build -toolexec="orchestrion toolexec" .
    go test -toolexec="orchestrion toolexec" ./...
    ```
<!-- markdownlint-enable MD046 -->

### More Documentation {#docs}

- [Tracing Go Applications](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/go/){:target="_blank"}
- [GitHub Orchestrion](https://github.com/DataDog/orchestrion){:target="_blank"}

---

## Manual instrumentation {#dependence}

Install the v2 trace SDK:

```shell
go get github.com/DataDog/dd-trace-go/v2/ddtrace/tracer
```

If profiling is needed, install the profiler as well and enable the DataKit [Profiling collector](profile.md):

```shell
go get github.com/DataDog/dd-trace-go/v2/profiler
```

Other libraries related to components, as needed, for example:

```shell
go get github.com/DataDog/dd-trace-go/contrib/gorilla/mux/v2
go get github.com/DataDog/dd-trace-go/contrib/net/http/v2
go get github.com/DataDog/dd-trace-go/contrib/database/sql/v2
```

See the [GitHub integration library](https://github.com/DataDog/dd-trace-go/tree/main/contrib){:target="_blank"} or the [Datadog support matrix](https://docs.datadoghq.com/tracing/trace_collection/compatibility/go/#integrations){:target="_blank"} for available integrations. Automatic HTTP integration does not cover every business operation, so add manual spans for critical business logic.

## Code Examples {#examples}

### Simple HTTP Server {#sample-http-server}

``` go hl_lines="8-10 15-16 20-38" linenums="1" title="http-server.go"
package main

import (
  "log"
  "net/http"
  "time"

  httptrace "github.com/DataDog/dd-trace-go/contrib/net/http/v2"
  "github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
  "github.com/DataDog/dd-trace-go/v2/profiler"
)

func main() {
  tracer.Start(
    tracer.WithService("test"),
    tracer.WithEnv("test"),
  )
  defer tracer.Stop()

  err := profiler.Start(
    profiler.WithService("test"),
    profiler.WithEnv("test"),
    profiler.WithProfileTypes(
      profiler.CPUProfile,
      profiler.HeapProfile,
      // The profiles below are disabled by
      // default to keep overhead low, but
      // can be enabled as needed.
      // profiler.BlockProfile,
      // profiler.MutexProfile,
      // profiler.GoroutineProfile,
    ),
  )
  if err != nil {
    log.Fatal(err)
  }
  defer profiler.Stop()

  // Create a traced mux router
  mux := httptrace.NewServeMux()
  // Continue using the router as you normally would.
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    time.Sleep(time.Second)
    w.Write([]byte("Hello World!"))
  })
  if err := http.ListenAndServe(":18080", mux); err != nil {
    log.Fatal(err)
  }
}
```

Compile and run

<!-- markdownlint-disable MD046 -->
=== "Linux/Mac"

    ```shell
    go build -o http-server http-server.go
    DD_SERVICE=go-http-server DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./http-server
    ```

=== "Windows"

    ```powershell
    go build -o http-server.exe http-server.go
    $env:DD_SERVICE="go-http-server"; $env:DD_AGENT_HOST="localhost"; $env:DD_TRACE_AGENT_PORT="9529"; .\http-server.exe
    ```
<!-- markdownlint-enable MD046 -->

`profiler.Start()` is optional. When enabled, profiling data is delivered to the DataKit Profiling receiver, not the trace receiver described on this page. Confirm that the `profile` collector is enabled.

### Manual Tracing {#manual-tracing}

The following code demonstrates trace data collection for a file opening operation.

Start the tracer at the application entry point and pass the parent span context to downstream operations. In v2, use `StartChild` or `StartSpanFromContext` to establish parent/child relationships instead of the v1 `ChildOf` pattern:

``` go linenums="1" title="main.go"
package main

import (
    "os"
    "time"

    "github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
    "github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
)

func main() {
    tracer.Start(
        tracer.WithEnv("prod"),
        tracer.WithService("test-file-read"),
        tracer.WithServiceVersion("1.2.3"),
        tracer.WithGlobalTag("project", "add-ddtrace-in-golang-project"),
    )

    // end of app exit, make sure tracer stopped
    defer tracer.Stop()

    tick := time.NewTicker(time.Second)
    defer tick.Stop()

    // your-app-main-entry...
    for {
        runApp()
        runAppWithError()

        select {
        case <-tick.C:
        }
    }
}

func runApp() {
    var err error
    span := tracer.StartSpan("get.data")
    defer func() { span.Finish(tracer.WithError(err)) }()

    child := span.StartChild("read.file")
    child.SetTag(ext.ResourceName, os.Args[0])

    bts, err := os.ReadFile(os.Args[0])
    span.SetTag("file_len", len(bts))
    child.Finish(tracer.WithError(err))
}

func runAppWithError() {
    var err error
    span := tracer.StartSpan("get.data")
    child := span.StartChild("read.file")
    child.SetTag(ext.ResourceName, "somefile-not-found.go")

    defer func() {
        child.Finish(tracer.WithError(err))
        span.Finish(tracer.WithError(err))
    }()

    _, err = os.ReadFile("somefile-not-found.go")
}
```

Compile and run

<!-- markdownlint-disable MD046 -->
=== "Linux/Mac"

    ```shell
    go build -o my-app main.go
    DD_SERVICE=test-file-read DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./my-app
    ```

=== "Windows"

    ```powershell
    go build -o my-app.exe main.go
    $env:DD_SERVICE="test-file-read"; $env:DD_AGENT_HOST="localhost"; $env:DD_TRACE_AGENT_PORT="9529"; .\my-app.exe
    ```
<!-- markdownlint-enable MD046 -->

After running the program for a while, you can see trace data similar to the following in <<<custom_key.brand_name>>>:

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/golang-ddtrace-example.png){  width="800"}
  <figcaption>Golang program trace data display</figcaption>
</figure>

## Supported Environment Variables {#start-options}

The following environment variables configure DDTrace at process startup:

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./my-app
```

See [DDTrace-Go documentation](https://docs.datadoghq.com/tracing/trace_collection/library_config/go/){:target="_blank"} for the complete list, precedence, and version-specific behavior.

<!-- markdownlint-disable MD046 -->
???+ attention

    Avoid setting conflicting values for the same setting in code and environment variables. Precedence can differ by Go SDK version; when overriding a value, use the running SDK's documentation and startup logs as the source of truth.
<!-- markdownlint-enable MD046 -->

- **`DD_VERSION`**

    Sets the application version, such as `1.2.3`, `2022.02.13`

- **`DD_SERVICE`**

    Sets the application service name

- **`DD_ENV`**

    Sets the current environment of the application, such as `prod`, `pre-prod`, etc.

- **`DD_AGENT_HOST`**

    **Default**: `localhost`

    Sets the DataKit host name or IP to which traces are sent. `DD_TRACE_AGENT_URL`, when set, normally takes precedence.

- **`DD_TRACE_AGENT_PORT`**

    Sets the trace receiver port. The common upstream default is `8126`; explicitly specify the [DataKit HTTP port][4] (normally `9529`) for DataKit.

- **`DD_DOGSTATSD_HOST`**, **`DD_DOGSTATSD_PORT`**

    The DogStatsD destination for runtime metrics. The normal default port is `8125`. To receive DogStatsD data from the Go SDK, enable the DataKit [StatsD collector][5]; do not point it at the trace port `9529`.

- **`DD_TRACE_SAMPLING_RULES`**

    A JSON rule array evaluated in order. `sample_rate` ranges from `[0.0, 1.0]`.

    **Example 1**: Set the global sampling rate to 20%: `DD_TRACE_SAMPLING_RULES='[{"sample_rate": 0.2}]' ./my-app`

    **Example 2**: Sample traces at 10% when service matches `app1.*` and the span name is `abc`, otherwise at 20%: `DD_TRACE_SAMPLING_RULES='[{"service": "app1.*", "name": "abc", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app`

- **`DD_TRACE_SAMPLE_RATE`**

    **v2 default**: `1.0`

    Sets the global SDK-side sample rate. It is a simpler, separate setting from `DD_TRACE_SAMPLING_RULES`; use rules when sampling must vary by service or operation.

- **`DD_TRACE_RATE_LIMIT`**

    Sets the maximum number of sampled traces per second for each Go process. When sampling rules or a sample rate are set, the usual default is `100`.

- **`DD_TAGS`**

    **Default**: `[]`

    Here you can inject a set of global tags, which will appear in each span and profile data. Multiple tags can be separated by spaces and commas, such as `layer:api,team:intake`, `layer:api team:intake`

- **`DD_TRACE_STARTUP_LOGS`**

    **Default**: `true`

    Enable DDTrace-related configuration and diagnostic logs

- **`DD_TRACE_DEBUG`**

    **Default**: `false`

    Enable DDTrace-related debug logs

- **`DD_TRACE_ENABLED`**

    **Default**: `true`

    Enable trace switch. If this switch is manually turned off, no trace data will be generated

- **`DD_SERVICE_MAPPING`**

    **Default**: `null`
    Dynamically rename service names, service name mappings can be separated by spaces and commas, such as `mysql:mysql-service-name,postgres:postgres-service-name`, `mysql:mysql-service-name postgres:postgres-service-name`

---

<!-- markdownlint-disable MD053 -->
[4]: ../datakit/datakit-conf.md#config-http-server
[5]: statsd.md
<!-- markdownlint-enable MD053 -->
