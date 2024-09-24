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

Integrating APM into Golang involves some level of invasiveness, **requiring modifications to existing code**, but overall, common business code does not need too many changes, simply replace the relevant import packages.

## Install Dependencies {#dependence}

Install the DDTrace Golang SDK:

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer
```

Install the profiling library

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/profiler
```

Other libraries related to components, as needed, for example:

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql
```

We can learn more about available tracing SDKs from the [Github plugin library](https://github.com/DataDog/dd-trace-go/tree/main/contrib){:target="_blank"} or [Datadog's related support documentation](https://docs.datadoghq.com/tracing/trace_collection/compatibility/go/#integrations){:target="_blank"}.

## Code Examples {#examples}

### Simple HTTP Server {#sample-http-server}

``` go hl_lines="8-10 15-16 20-38" linenums="1" title="http-server.go"
package main

import (
  "log"
  "net/http"
  "time"

  httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
  "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
  "gopkg.in/DataDog/dd-trace-go.v1/profiler"
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
    go build http-server.go -o http-server
    DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./http-server
    ```

=== "Windows"

    ```powershell
    go build http-server.go -o http-server
    $env:DD_AGENT_HOST="localhost"; $env:DD_TRACE_AGENT_PORT="9529"; .\http-server.exe
    ```
<!-- markdownlint-enable -->

### Manual Tracing {#manual-tracing}

The following code demonstrates trace data collection for a file opening operation.

In the `main()` entry code, set the basic trace parameters and start tracing:

``` go hl_lines="8-9 14-17 40-45 57-66" linenums="1" title="main.go"
package main

import (
    "io/ioutil"
    "os"
    "time"

    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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
    // Start a root span.
    span := tracer.StartSpan("get.data")
    defer span.Finish(tracer.WithError(err))

    // Create a child of it, computing the time needed to read a file.
    child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
    child.SetTag(ext.ResourceName, os.Args[0])

    // Perform an operation.
    var bts []byte
    bts, err = ioutil.ReadFile(os.Args[0])
    span.SetTag("file_len", len(bts))
    child.Finish(tracer.WithError(err))
}

func runAppWithError() {
    var err error
    // Start a root span.
    span := tracer.StartSpan("get.data")

    // Create a child of it, computing the time needed to read a file.
    child := tracer.StartSpan("read.file", tracer.ChildOf(span.Context()))
    child.SetTag(ext.ResourceName, "somefile-not-found.go")

    defer func() {
        child.Finish(tracer.WithError(err))
        span.Finish(tracer.WithError(err))
    }()

    // Perform an error operation.
    if _, err = ioutil.ReadFile("somefile-not-found.go"); err != nil {
        // error handle
    }
}
```

Compile and run

<!-- markdownlint-disable MD046 -->
=== "Linux/Mac"

    ```shell
    go build main.go -o my-app
    DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 ./my-app
    ```

=== "Windows"

    ```powershell
    go build main.go -o my-app.exe
    $env:DD_AGENT_HOST="localhost"; $env:DD_TRACE_AGENT_PORT="9529"; .\my-app.exe
    ```
<!-- markdownlint-enable -->

After running the program for a while, you can see trace data similar to the following in GuanceCloud:

<figure markdown>
  ![](https://static.guance.com/images/datakit/golang-ddtrace-example.png){  width="800"}
  <figcaption>Golang program trace data display</figcaption>
</figure>

## Supported Environment Variables {#start-options}

The following environment variables are supported to specify some configuration parameters of DDTrace when starting the program, and their basic form is:

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./my-app
```

For more environment variable support, see [DDTrace-Go Documentation](https://docs.datadoghq.com/tracing/trace_collection/library_config/go/){:target="_blank"}.

<!-- markdownlint-disable MD046 -->
???+ attention

    These environment variables will be overridden by the corresponding fields injected with `WithXXX()` in the code, so the configuration injected by the code has a higher priority. These ENVs only take effect when the corresponding fields are not specified in the code.
<!-- markdownlint-enable -->

- **`DD_VERSION`**

    Sets the application version, such as `1.2.3`, `2022.02.13`

- **`DD_SERVICE`**

    Sets the application service name

- **`DD_ENV`**

    Sets the current environment of the application, such as `prod`, `pre-prod`, etc.

- **`DD_AGENT_HOST`**

    **Default**: `localhost`

    Sets the IP address of DataKit, and the trace data generated by the application will be sent to Datakit

- **`DD_TRACE_AGENT_PORT`**

    Sets the DataKit trace data receiving port. Here you need to manually specify the [DataKit HTTP port](datakit-conf.md#config-http-server) (usually 9529)

- **`DD_DOGSTATSD_PORT`**

    Default value: `8125`
    If you want to receive StatsD data generated by DDTrace, you need to manually enable the [StatsD collector](../integrations/statsd.md) on Datakit

- **`DD_TRACE_SAMPLING_RULES`**

    **Default**: `nil`

    Here a JSON array is used to represent the sampling settings (sampling rate application is in array order), where `sample_rate` is the sampling rate, and the value range is `[0.0, 1.0]`.

    **Example 1**: Set the global sampling rate to 20%: `DD_TRACE_SAMPLING_RULES='[{"sample_rate": 0.2}]' ./my-app`

    **Example 2**: Service name wildcard `app1.*`, and the span name is `abc`, set the sampling rate to 10%, otherwise, set the sampling rate to 20%: `DD_TRACE_SAMPLING_RULES='[{"service": "app1.*", "name": "b", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app`

- **`DD_TRACE_SAMPLE_RATE`**

    **Default**: `nil`

    Enable the above sampling rate switch

- **`DD_TRACE_RATE_LIMIT`**

    Sets the number of span samples per second for each Golang process. If `DD_TRACE_SAMPLE_RATE` is already turned on, the default is 100

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
[4]: datakit-conf.md#config-http-server
[5]: ../integrations/statsd.md
<!-- markdownlint-enable -->
