---
title     : 'DDTrace Golang'
summary   : 'DDTrace Golang 集成'
tags      :
  - 'DDTRACE'
  - 'GOLANG'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---

Go 应用有两种接入方式：

- **编译时插桩（Orchestrion）**：不修改业务代码，在编译阶段为支持的库插入埋点，适合希望获得广覆盖并可统一纳入 CI/CD 的场景。
- **手动插桩**：在关键入口、外部调用和业务操作中显式创建 span，控制最精确，但需要维护代码。

本页示例使用 `dd-trace-go/v2`。上游已停止维护 v1；遗留应用如暂时不能迁移，必须继续使用与现有代码匹配的 v1 API，不要混用 v1 与 v2 import 路径。

## 编译时插桩 {#compilation-automatically}

要求：

- 使用 Go Modules 管理项目；
- Go 版本至少为 1.18，生产环境优先使用上游仍维护的 Go 版本；
- 已安装 DataKit 并开启 [DDTrace 采集器](ddtrace.md)。

<!-- markdownlint-disable MD046 -->

=== "推荐流程"

    1. 安装 Orchestrion，并确保 `$(go env GOBIN)` 或 `$(go env GOPATH)/bin` 在 `PATH` 中：

    ```shell
    go install github.com/DataDog/orchestrion@latest
    ```

    2. 在项目根目录登记 Orchestrion：

    ```shell
    orchestrion pin
    ```

    此操作会更新 `go.mod`、`go.sum` 并生成 `orchestrion.tool.go`。将这些文件与应用代码一起提交，确保开发机和 CI 使用同一套插桩依赖。

    3. 使用 Orchestrion 构建、测试和运行：

    ```shell
    orchestrion go build .
    orchestrion go test ./...
    DD_SERVICE=my-go-service DD_AGENT_HOST=localhost DD_TRACE_AGENT_PORT=9529 \
      orchestrion go run .
    ```

<!-- cspell:ignore toolexec -->
=== "CI 中使用 -toolexec"

    如构建系统不能调用 `orchestrion go`，可在对应的 `go` 命令中显式传入 `-toolexec`：

    ```shell
    go build -toolexec="orchestrion toolexec" .
    go test -toolexec="orchestrion toolexec" ./...
    ```

<!-- markdownlint-enable MD046 -->

编译时插桩只改变构建产物，不会替你配置 DataKit 目标。运行时仍需设置 `DD_SERVICE`、`DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT=9529`，并按需设置环境、版本、采样和 Profiling。

### 更多文档 {#docs}

- [Tracing Go Applications](https://docs.datadoghq.com/tracing/trace_collection/dd_libraries/go/){:target="_blank"}
- [GitHub Orchestrion](https://github.com/DataDog/orchestrion){:target="_blank"}

---

## 手动方式插桩 {#manual-dependence}

安装 v2 Trace SDK：

```shell
go get github.com/DataDog/dd-trace-go/v2/ddtrace/tracer
```

如需 Profiling，再安装 profiler；同时要在 DataKit 中开启 [Profiling 采集器](profile.md)：

```shell
go get github.com/DataDog/dd-trace-go/v2/profiler
```

其它跟组件有关的库，视情况而定，比如：

```shell
go get github.com/DataDog/dd-trace-go/contrib/gorilla/mux/v2
go get github.com/DataDog/dd-trace-go/contrib/net/http/v2
go get github.com/DataDog/dd-trace-go/contrib/database/sql/v2
```

可从 [GitHub 插件库](https://github.com/DataDog/dd-trace-go/tree/main/contrib){:target="_blank"}或 [Datadog 支持矩阵](https://docs.datadoghq.com/tracing/trace_collection/compatibility/go/#integrations){:target="_blank"}了解可用集成。自动 HTTP 集成并不能覆盖所有业务操作；对关键业务逻辑仍应补充手动 span。

## 代码示例 {#examples}

### 简单的 HTTP 服务 {#sample-http-server}

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

编译运行

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

`profiler.Start()` 是可选项。启用它时，Profiling 数据会发送到 DataKit 的 Profiling 接收端，而不是本页的 trace 接收端；请确认 `profile` 采集器已经启用。

### 手动埋点 {#manual-tracing}

以下代码演示了一个文件打开操作的 trace 数据收集。

在 `main()` 入口代码中启动 tracer，并把父 span 的 context 传给下游操作。v2 推荐通过 `StartChild` 或 `StartSpanFromContext` 建立父子关系，而不是沿用 v1 的 `ChildOf` 写法：

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

编译运行

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

程序运行一段时间后，即可在<<<custom_key.brand_name>>>看到类似如下 trace 数据：

<figure markdown>
  ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/golang-ddtrace-example.png){ width="800"}
  <figcaption>Golang 程序 trace 数据展示</figcaption>
</figure>

## 支持的环境变量 {#start-options}

以下环境变量在启动程序时配置 DDTrace，基本形式为：

```shell
DD_XXX=<env-value> DD_YYY=<env-value> ./my-app
```

完整环境变量、优先级和版本差异见 [DDTrace-Go 文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/go/){:target="_blank"}。

<!-- markdownlint-disable MD046 -->
???+ info

    避免在代码和环境变量中为同一项设置冲突的值。不同 Go SDK 版本的优先级可能不同；需要覆盖配置时，请以运行版本的 SDK 文档和启动日志为准。
<!-- markdownlint-enable MD046 -->

- **`DD_VERSION`**

    设置应用程序版本，如 `1.2.3`、`2022.02.13`

- **`DD_SERVICE`**

    设置应用服务名

- **`DD_ENV`**

    设置应用当前的环境，如 `prod`、`pre-prod` 等

- **`DD_AGENT_HOST`**

    **默认值**：`localhost`

    设置 DataKit 的主机名或 IP，应用产生的 trace 会发送到该地址。若设置 `DD_TRACE_AGENT_URL`，URL 通常优先。

- **`DD_TRACE_AGENT_PORT`**

    设置 trace 接收端口。上游默认通常为 `8126`，接入 DataKit 时需显式指定 [DataKit 的 HTTP 端口][4]（通常为 `9529`）。

- **`DD_DOGSTATSD_HOST`**、**`DD_DOGSTATSD_PORT`**

    Runtime metrics 的 DogStatsD 地址，默认端口为 `8125`。如需接收 Go SDK 产生的 DogStatsD 数据，需在 DataKit 上开启 [StatsD 采集器][5]；不要将其设置为 trace 端口 `9529`。

- **`DD_TRACE_SAMPLING_RULES`**

    JSON 规则数组按顺序匹配，`sample_rate` 取值范围为 `[0.0, 1.0]`。

    **示例一**：设置全局采样率为 20%：`DD_TRACE_SAMPLING_RULES='[{"sample_rate": 0.2}]' ./my-app`

    **示例二**：服务名匹配 `app1.*`、且 span 名称为 `abc` 的 trace 采样率为 10%，其他 trace 为 20%：`DD_TRACE_SAMPLING_RULES='[{"service": "app1.*", "name": "abc", "sample_rate": 0.1}, {"sample_rate": 0.2}]' ./my-app`

- **`DD_TRACE_SAMPLE_RATE`**

    **v2 默认值**：`1.0`

    设置全局 SDK 侧采样率。它是独立于 `DD_TRACE_SAMPLING_RULES` 的简化配置；需要按服务或操作名控制时使用规则。

- **`DD_TRACE_RATE_LIMIT`**

    设置每个 Go 进程每秒最多采样的 trace 数。当设置采样规则或采样率时，默认通常为 `100`。

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

[4]: ../datakit/datakit-conf.md#config-http-server
[5]: statsd.md
