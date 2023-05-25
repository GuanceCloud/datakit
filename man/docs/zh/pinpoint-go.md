# Golang 示例

- [Pinpoint Golang Agent 代码仓库](https://github.com/pinpoint-apm/pinpoint-go-agent){:target="_blank"}
- [Pinpoint Golang 代码示例](https://github.com/pinpoint-apm/pinpoint-go-agent/tree/main/example){:target="_blank"}
- [Pinpoint Golang Agent 配置文档](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/doc/config.md){:target="_blank"}

---

## 配置 Datakit Agent {#config-datakit-agent}

参考[配置 Datakit 中的 Pinpoint Agent](pinpoint.md#agent-config)

## 配置 Pinpoint Golang Agent {#config-pinpoint-golang-agent}

Pinpoint Golang Agent 可以通过包括命令行参数，配置文件，环境变量在内的多种方式进行配置，配置优先级由高到低为：

- 命令行参数
- 环境变量
- 配置文件
- 配置函数
- 默认配置

Pinpoint Golang Agent 也支持运行期动态更改配置，所有标记了 dynamic 的配置项参数可以在运行期进行动态配置。

基本参数说明：

以下每个标题为配置文件中的配置项，并且每一个配置项描述中的列表将依次列出对应的命令行参数，环境变量，配置函数，配置值类型以及附加信息。

<!-- markdownlint-disable MD006 MD007 -->
`ConfigFile`

:   配置文件支持 [JSON](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.json){:target="_blank"}，[YAML](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.yaml){:target="_blank"}，[Properties 配置文件](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.prop){:target="_blank"}。配置文件中的配置项是大小写敏感的

   - --pinpoint-configfile
   - PINPOINT_GO_CONFIGFILE
   - WithConfigFile()
   - string
   - case-sensitive

   对于那些以 '.' 分隔的配置字段，它们在配置文件中将以缩进的方式出现。例如：

   ``` yaml
   applicationName: "MyAppName"
   collector:
     host: "collector.myhost.com"
   sampling:
     type: "percent"
     percentRate: 10
   logLevel: "error"
   ```

`ApplicationName`

:   ApplicationName 配置应用程序的名字，该项如果没有配置 Agent 将无法启动。

   - --pinpoint-applicationname
   - PINPOINT_GO_APPLICATIONNAME
   - WithAppName()
   - string
   - case-sensitive
   - max-length: 24

`ApplicationType`

:   ApplicationType 配置应用程序的类型

   - --pinpoint-applicationtype
   - PINPOINT_GO_APPLICATIONTYPE
   - WithAppType()
   - int
   - default: 1800 (ServiceTypeGoApp)

`AgentId`

:   AgentId 用于配置 id 来区分不同的 Agent。建议将 hostname 包含进来。如果没有配置或配置错误，Agent 将使用自动生成的 id。

   - --pinpoint-agentid
   - PINPOINT_GO_AGENTID
   - WithAgentId()
   - string
   - case-sensitive
   - max-length: 24

`AgentName`

:   AgentName 配置 Agent 名字。

   - --pinpoint-agentname
   - PINPOINT_GO_AGENTNAME
   - WithAgentName()
   - string
   - case-sensitive
   - max-length: 255

`Collector.Host`

:   Collector.Host 配置 Pinpoint Collector 的主机地址。

   - --pinpoint-collector-host
   - PINPOINT_GO_COLLECTOR_HOST
   - WithCollectorHost()
   - string
   - default: "localhost"
   - case-sensitive

`Collector.AgentPort`

:   Collector.AgentPort 配置 Pinpoint Collector 的 Agent 端口号。

   - --pinpoint-collector-agentport
   - PINPOINT_GO_COLLECTOR_AGENTPORT
   - WithCollectorAgentPort()
   - int
   - default: 9991 (Datakit Pinpoint Agent 中默认端口号为 9991)

`Collector.SpanPort`

:   Collector.SpanPort 配置 Pinpoint Collector 的 Span 端口号。

   - --pinpoint-collector-spanport
   - PINPOINT_GO_COLLECTOR_SPANPORT
   - WithCollectorSpanPort()
   - int
   - default: 9993 (Datakit Pinpoint Agent 中默认端口号为 9991)

`Collector.StatPort`

:   Collector.StatPort 配置 Pinpoint Collector 的 Stat 端口号。

   - --pinpoint-collector-statport
   - PINPOINT_GO_COLLECTOR_STATPORT
   - WithCollectorStatPort()
   - int
   - default: 9992 (Datakit Pinpoint Agent 中默认端口号为 9991)

`Sampling.Type`

:   Sampling.Type 配置采样器类型， "COUNTER" 或者 "PERCENT"。

   - --pinpoint-sampling-type
   - PINPOINT_GO_SAMPLING_TYPE
   - WithSamplingType()
   - string
   - default: "COUNTER"
   - case-insensitive
   - dynamic

`Sampling.CounterRate`

:   Sampling.CounterRate 配置计数采样器采样率。 采样率为 1/rate。 比如，如果 rate 配置为 1 那么采样率为 100%，那么如果 rate 配置为 100 那么采样率为 1%。

   - --pinpoint-sampling-counterrate
   - PINPOINT_GO_SAMPLING_COUNTERRATE
   - WithSamplingCounterRate()
   - int
   - default: 1
   - valid range: 0 ~ 100
   - dynamic

`Sampling.PercentRate`

:   Sampling.PercentRate 配置百分比采样器的采样率。

   - --pinpoint-sampling-percentrate
   - PINPOINT_GO_SAMPLING_PERCENTRATE
   - WithSamplingPercentRate()
   - float
   - default: 100
   - valid range: 0.01 ~ 100
   - dynamic

`Span.QueueSize`

:   Span.QueueSize 配置 Span Queue 的大小。

   - --pinpoint-span-queuesize
   - PINPOINT_GO_SPAN_QUEUESIZE
   - WithSpanQueueSize()
   - type: int
   - default: 1024

`Span.MaxCallStackDepth`

:   Span.MaxCallStackDepth 配置 Span 可探测的调用栈最大深度。

   - --pinpoint-span-maxcallstackdepth
   - PINPOINT_GO_SPAN_MAXCALLSTACKDEPTH
   - WithSpanMaxCallStackDepth()
   - type: int
   - default: 64
   - min: 2
   - ultimate: -1
   - dynamic

`Span.MaxCallStackSequence`

:   Span.MaxCallStackDepth 配置 Span 可探测的调用站最长序列。

   - --pinpoint-span-maxcallstacksequence
   - PINPOINT_GO_SPAN_MAXCALLSTACKSEQUENCE
   - WithSpanMaxCallStackSequence()
   - type: int
   - default: 5000
   - min: 4
   - ultimate: -1
   - dynamic

`Stat.CollectInterval`

:   Stat.CollectInterval 配置统计频率。

   - --pinpoint-stat-collectinterval
   - PINPOINT_GO_STAT_COLLECTINTERVAL
   - WithStatCollectInterval()
   - type: int
   - default: 5000
   - unit: milliseconds

`Stat.BatchCount`

:   Stat.BatchCount 配置批次发送统计数据的个数。

   - --pinpoint-stat-batchcount
   - PINPOINT_GO_STAT_BATCHCOUNT
   - WithStatBatchCount()
   - type: int
   - default: 6

`Log.Level`

:   Log.Level 配置 Agent 日志的层级，trace/debug/info/warn/error 必须配置。

   - --pinpoint-log-level
   - PINPOINT_GO_LOG_LEVEL
   - WithLogLevel()
   - type: string
   - default: "info"
   - case-insensitive
   - dynamic

`Log.Output`

:   Log.Output 配置日志的输出， stderr/stdout/file path。

   - --pinpoint-log-output
   - PINPOINT_GO_LOG_OUTPUT
   - WithLogOutput()
   - type: string
   - default: "stderr"
   - case-insensitive
   - dynamic

`Log.MaxSize`

:   Log.MaxSize 配置日志文件的最大容量。

   - --pinpoint-log-maxsize
   - PINPOINT_GO_LOG_MAXSIZE
   - WithLogMaxSize()
   - type: int
   - default: 10
   - dynamic

<!-- markdownlint-enable -->

## 手动检测应用程序 {#manual-instrumentation}

对于有虚拟机的程序语言例如：JAVA 可以通过直接向虚拟机中注入检测 Agent 启动自动检测，但是对于需要进行编译后独立运行的编程语言例如 Golang 需要进行手动检测。

Pinpoint Golang Agent 中可以通过两种方式进行手动检测：

- 使用 [Pinpoint Golang 插件库](https://github.com/pinpoint-apm/pinpoint-go-agent/tree/main/plugin){:target="_blank"}
- 使用 Pinpoint Agent Golang API 进行手动检测

<!-- markdownlint-disable MD006 MD007  MD038 -->
`Span`

:   Pinpoint 中 Span 代表着 service 或 application 的顶层程序操作，例如在 HTTP handler 中创建 Span：

   ```golang
   func doHandle(w http.ResponseWriter, r *http.Request) {
     tracer = pinpoint.GetAgent().NewSpanTracerWithReader("HTTP Server", r.URL.Path, r.Header)
     defer tracer.EndSpan()

     span := tracer.Span()
     span.SetEndPoint(r.Host)
   }
   ```

   你可以检测单一调用栈的应用并产生一个 Span。 Tracer.EndSpan() 必须被调用以完成这个 Span 并发送到远端的 Collector。 SpanRecorder 和 Annotation 接口可以用来将链路数据记录在 Span 中。

`SpanEvent`

:   Pinpoint 中每一个 SpanEvent 都代表着一个 Span 探测范围内的一次程序操作，例如：访问数据库，调用函数，向另一个服务发起请求等等。你可以通过 Tracer.NewSpanEvent() 上报一个 span，必须调用 Tracer.EndSpanEvent() 来完成一个 span。

   ```golang
   func doHandle(w http.ResponseWriter, r *http.Request) {
       tracer := pinpoint.GetAgent().NewSpanTracerWithReader("HTTP Server", r.URL.Path, r.Header)
       defer tracer.EndSpan()

       span := tracer.Span()
       span.SetEndPoint(r.Host)
       defer tracer.NewSpanEvent("doHandle").EndSpanEvent()

       func() {
           defer tracer.NewSpanEvent("func_1").EndSpanEvent()

           func() {
               defer tracer.NewSpanEvent("func_2").EndSpanEvent()
               time.Sleep(100 * time.Millisecond)
           }()
           time.Sleep(1 * time.Second)
       }()
   }
   ```

` 分发 Tracing Context`

:   如果一次请求来自另一个由 Pinpoint 监测的节点，那么这次数据交换中将包含一个数据交换上下文。大多数此类型的数据来自上一个节点并且被打包在请求消息体内。 Pinpoint 提供了两个函数完成对数据交换上下文的读写。

   - Tracer.Extract(reader DistributedTracingContextReader) // 提取分发的上下文。
   - Tracer.Inject(writer DistributedTracingContextWriter) // 向请求中注入上下文。

   ```golang
   func externalRequest(tracer pinpoint.Tracer) int {
    req, err := http.NewRequest("GET", "http://localhost:9000/async_wrapper", nil)
    client := &http.Client{}

    tracer.NewSpanEvent("externalRequest")
    defer tracer.EndSpanEvent()

    se := tracer.SpanEvent()
    se.SetEndPoint(req.Host)
    se.SetDestination(req.Host)
    se.SetServiceType(pinpoint.ServiceTypeGoHttpClient)
    se.Annotations().AppendString(pinpoint.AnnotationHttpUrl, req.URL.String())
    tracer.Inject(req.Header)

    resp, err := client.Do(req)
    defer resp.Body.Close()

    tracer.SpanEvent().SetError(err)
    return resp.StatusCode
   }
   ```

` 函数间调用上下文透传 `

:   在同一个服务中的不同 API 间和不同进程中进行调用上下文的透传是通过对 context.Context 的操作现的。 Pinpoint Golang Agent 通过向 Context 中注入 Tracer 实现上下文的链接。

   - NewContext() // 注入 Tracer 到 Context 中。
   - FromContext() // 导入一个 Tracer。

   ```golang
   func tableCount(w http.ResponseWriter, r *http.Request) {
    tracer := pinpoint.FromContext(r.Context())

    db, err := sql.Open("mysql-pinpoint", "root:p123@tcp(127.0.0.1:3306)/information_schema")
    defer db.Close()

    ctx := pinpoint.NewContext(context.Background(), tracer)
    row := db.QueryRowContext(ctx, "SELECT count(*) from tables")
    var count int
    row.Scan(&count)

    fmt.Println("number of tables in information_schema", count)
   }
   ```

` 检测 Goroutine`

:   Pinpoint Tracer 设计之初是用来检测单一调用栈的应用，所以在不同的线程间共享同一个 Tracer 将引起资源抢占而造成程序崩溃。你可以通过调用 Tracer.NewGoroutineTracer() 创建一个新的 Tracer 来检测 Goroutine。

   在线程间传递 Tracer 可以通过以下几种方式：

   - function parameter

      ```golang
      func outGoingRequest(ctx context.Context) {
        client := pphttp.WrapClient(nil)

        request, _ := http.NewRequest("GET", "https://github.com/pinpoint-apm/pinpoint-go-agent", nil)
        request = request.WithContext(ctx)

        resp, err := client.Do(request)
        if nil != err {
            log.Println(err.Error())
            return
        }
        defer resp.Body.Close()
        log.Println(resp.Body)
      }

      func asyncWithTracer(w http.ResponseWriter, r *http.Request) {
          tracer := pinpoint.FromContext(r.Context())
          wg := &sync.WaitGroup{}
          wg.Add(1)

          go func(asyncTracer pinpoint.Tracer) {
              defer wg.Done()

              defer asyncTracer.EndSpan() // must be called
              defer asyncTracer.NewSpanEvent("asyncWithTracer_goroutine").EndSpanEvent()

              ctx := pinpoint.NewContext(context.Background(), asyncTracer)
              outGoingRequest(w, ctx)
          }(tracer.NewGoroutineTracer())

          wg.Wait()
      }
      ```

   - channel

      ```golang
      func asyncWithChan(w http.ResponseWriter, r *http.Request) {
          tracer := pinpoint.FromContext(r.Context())
          wg := &sync.WaitGroup{}
          wg.Add(1)

          ch := make(chan pinpoint.Tracer)

          go func() {
              defer wg.Done()

              asyncTracer := <-ch
              defer asyncTracer.EndSpan() // must be called
              defer asyncTracer.NewSpanEvent("asyncWithChan_goroutine").EndSpanEvent()

              ctx := pinpoint.NewContext(context.Background(), asyncTracer)
              outGoingRequest(w, ctx)
          }()

          ch <- tracer.NewGoroutineTracer()
          wg.Wait()
      }
      ```

   - context.Context

      ```golang
      func asyncWithContext(w http.ResponseWriter, r *http.Request) {
          tracer := pinpoint.FromContext(r.Context())
          wg := &sync.WaitGroup{}
          wg.Add(1)

          go func(asyncCtx context.Context) {
              defer wg.Done()

              asyncTracer := pinpoint.FromContext(asyncCtx)
              defer asyncTracer.EndSpan() // must be called
              defer asyncTracer.NewSpanEvent("asyncWithContext_goroutine").EndSpanEvent()

              ctx := pinpoint.NewContext(context.Background(), asyncTracer)
              outGoingRequest(w, ctx)
          }(pinpoint.NewContext(context.Background(), tracer.NewGoroutineTracer()))

          wg.Wait()
      }
      ```
<!-- markdownlint-enable -->