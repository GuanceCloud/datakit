---
title     : 'DDTrace Golang'
summary   : 'DDTrace Golang 集成'
tags      :
  - 'DDTRACE'
  - 'GOLANG'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---


Golang 的 APM 接入有一定的侵入性，**需要修改已有代码**，但总体而言，常见的业务代码不需要做太多变更，只需要替换相关的 import 包即可。

## 安装依赖 {#dependence}

安装 DDTrace Golang SDK：

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer
```

安装 profiling 库

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/profiler
```

其它跟组件有关的库，视情况而定，比如：

```shell
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http
go get gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql
```

我们可以从 [Github 插件库](https://github.com/DataDog/dd-trace-go/tree/main/contrib){:target="_blank"}或 [Datadog 相关支持文档](https://docs.datadoghq.com/tracing/trace_collection/compatibility/go/#integrations){:target="_blank"}了解更多可用的 tracing SDK。

## 代码示例 {#examples}

### 简单的 HTTP 服务 {#sample-http-server}

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

编译运行

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

### 手动埋点 {#manual-tracing}

以下代码演示了一个文件打开操作的 trace 数据收集。

在 `main()` 入口代码中，设置好基本的 trace 参数，并启动 trace：

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

编译运行

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

程序运行一段时间后，即可在观测云看到类似如下 trace 数据：

<figure markdown>
  ![](https://static.guance.com/images/datakit/golang-ddtrace-example.png){ width="800"}
  <figcaption>Golang 程序 trace 数据展示</figcaption>
</figure>

## 支持的环境变量 {#start-options}

以下环境变量支持在启动程序的时候指定 DDTrace 的一些配置参数，其基本形式为：

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./my-app
```

更多环境变量支持，参见 [DDTrace-Go 文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/go/){:target="_blank"}

<!-- markdownlint-disable MD046 -->
???+ attention

    这些环境变量将会被代码中用 `WithXXX()` 注入的对应字段覆盖，故代码注入的配置，优先级更高，这些 ENV 只有在代码未指定对应字段时才生效。
<!-- markdownlint-enable -->

- **`DD_VERSION`**

    设置应用程序版本，如 `1.2.3`、`2022.02.13`

- **`DD_SERVICE`**

    设置应用服务名

- **`DD_ENV`**

    设置应用当前的环境，如 `prod`、`pre-prod` 等

- **`DD_AGENT_HOST`**

    **默认值**：`localhost`

    设置 DataKit 的 IP 地址，应用产生的 trace 数据将发送给 Datakit

- **`DD_TRACE_AGENT_PORT`**

    设置 DataKit trace 数据的接收端口。这里需手动指定 [DataKit 的 HTTP 端口][4]（一般为 9529）

- **`DD_DOGSTATSD_PORT`**

    默认值：`8125`
    如果要接收 DDTrace 产生的 StatsD 数据，需在 Datakit 上手动开启 [StatsD 采集器][5]

- **`DD_TRACE_SAMPLING_RULES`**

    **默认值**：`nil`

    这里用 JSON 数组来表示采样设置（采样率应用以数组顺序为准），其中 `sample_rate` 为采样率，取值范围为 `[0.0, 1.0]`。

    **示例一**：设置全局采样率为 20%：`DD_TRACE_SAMPLING_RULES='[{"sample_rate": 0.2}]' ./my-app`

    **示例二**：服务名通配 `app1.*`、且 span 名称为 `abc` 的，将采样率设置为 10%，除此之外，采样率设置为 20%：`DD_TRACE_SAMPLING_RULES='[{"service": "app1.*", "name": "b", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app`

- **`DD_TRACE_SAMPLE_RATE`**

    **默认值**：`nil`

    开启上面的采样率开关

- **`DD_TRACE_RATE_LIMIT`**

    设置每个 Golang 进程每秒钟的 span 采样数。如果 `DD_TRACE_SAMPLE_RATE` 已经打开，则默认为 100

- **`DD_TAGS`**

    **默认值**：`[]`

    这里可注入一组全局 tag，这些 tag 会出现在每个 span 和 profile 数据中。多个 tag 之间可以用空格和英文逗号分割，例如 `layer:api,team:intake`、`layer:api team:intake`

- **`DD_TRACE_STARTUP_LOGS`**

    **默认值**：`true`

    开启 DDTrace 有关的配置和诊断日志

- **`DD_TRACE_DEBUG`**

    **默认值**：`false`

    开启 DDTrace 有关的调试日志

- **`DD_TRACE_ENABLED`**

    **默认值**：`true`

    开启 trace 开关。如果手动将该开关关闭，则不会产生任何 trace 数据

- **`DD_SERVICE_MAPPING`**

    **默认值**：`null`
    动态重命名服务名，各个服务名映射之间可用空格和英文逗号分割，如 `mysql:mysql-service-name,postgres:postgres-service-name`，`mysql:mysql-service-name postgres:postgres-service-name`

---

[4]: datakit-conf.md#config-http-server
[5]: ../integrations/statsd.md
