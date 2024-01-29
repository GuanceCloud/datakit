# Golang example

- [Pinpoint Golang Agent code repository](https://github.com/pinpoint-apm/pinpoint-go-agent){:target="_blank"}
- [Pinpoint Golang code example](https://github.com/pinpoint-apm/pinpoint-go-agent/tree/main/example){:target="_blank"}
- [Pinpoint Golang Agent Configuration Document](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/doc/config.md){:target="_blank"}

---

## Configure Datakit Agent {#config-datakit-agent}

Refer to [Configuring Pinpoint Agent in Datakit](pinpoint.md#agent-config)

## Configure Pinpoint Golang Agent {#config-pinpoint-golang-agent}

Pinpoint Golang Agent can be configured through a variety of methods including command line parameters, configuration files, and environment variables. The configuration priority from high to low is:

- Command line parameters
- environment variables
- Configuration file
- Configuration function
- default allocation

Pinpoint Golang Agent also supports dynamic configuration changes during runtime. All configuration item parameters marked dynamic can be dynamically configured during runtime.

Basic parameter description:

Each of the following titles is a configuration item in the configuration file, and the list in the description of each configuration item will sequentially list the corresponding command line parameters, environment variables, configuration functions, configuration value types, and additional information.

<!-- markdownlint-disable MD006 MD007 -->
`ConfigFile`

:   Configuration file supports [JSON](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.json){:target="_blank"}, [YAML]( https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.yaml){:target="_blank"},[Properties configuration file](https://github.com/pinpoint-apm/pinpoint-go-agent/blob/main/example/pinpoint-config.prop){:target="_blank"}. Configuration items in the configuration file are case-sensitive

   - --pinpoint-configfile
   - PINPOINT_GO_CONFIGFILE
   - WithConfigFile()
   - string
   - case-sensitive

   Configuration fields that are separated by '.' will appear indented in the configuration file. For example:

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

:   ApplicationName configures the name of the application. If this item is not configured, the Agent will not be able to start.

   - --pinpoint-applicationname
   - PINPOINT_GO_APPLICATIONNAME
   - WithAppName()
   - string
   - case-sensitive
   - max-length: 24

`ApplicationType`

:   ApplicationType configures the type of application

   - --pinpoint-applicationtype
   - PINPOINT_GO_APPLICATIONTYPE
   - WithAppType()
   - int
   - default: 1800 (ServiceTypeGoApp)

`AgentId`

:   AgentId is used to configure ID to distinguish different Agents. It is recommended to include hostname. If not configured or misconfigured, the Agent will use an automatically generated id.

   - --pinpoint-agentid
   - PINPOINT_GO_AGENTID
   - WithAgentId()
   - string
   - case-sensitive
   - max-length: 24

`AgentName`

:   AgentName configures the Agent name.

   - --pinpoint-agentname
   - PINPOINT_GO_AGENTNAME
   - WithAgentName()
   - string
   - case-sensitive
   - max-length: 255

`Collector.Host`

:   Collector.Host configures the host address of the Pinpoint Collector.

   - --pinpoint-collector-host
   - PINPOINT_GO_COLLECTOR_HOST
   - WithCollectorHost()
   - string
   - default: "localhost"
   - case-sensitive

`Collector.AgentPort`

:   Collector.AgentPort configures the Agent port number of Pinpoint Collector.

   - --pinpoint-collector-agentport
   - PINPOINT_GO_COLLECTOR_AGENTPORT
   - WithCollectorAgentPort()
   - int
   - default: 9991 (The default port number in Datakit Pinpoint Agent is 9991)

`Collector.SpanPort`

:   Collector.SpanPort configures the Span port number of the Pinpoint Collector.

   - --pinpoint-collector-spanport
   - PINPOINT_GO_COLLECTOR_SPANPORT
   - WithCollectorSpanPort()
   - int
   - default: 9993 (The default port number in Datakit Pinpoint Agent is 9991)

`Collector.StatPort`

:   Collector.StatPort configures the Stat port number of Pinpoint Collector.

   - --pinpoint-collector-statport
   - PINPOINT_GO_COLLECTOR_STATPORT
   - WithCollectorStatPort()
   - int
   - default: 9992 (The default port number in Datakit Pinpoint Agent is 9991)

`Sampling.Type`

:   Sampling.Type configures the sampler type, "COUNTER" or "PERCENT".

   - --pinpoint-sampling-type
   - PINPOINT_GO_SAMPLING_TYPE
   - WithSamplingType()
   - string
   - default: "COUNTER"
   - case-insensitive
   - dynamic

`Sampling.CounterRate`

:   Sampling.CounterRate configures the counting sampler sampling rate. The sampling rate is 1/rate. For example, if rate is configured as 1 then the sampling rate is 100%, then if rate is configured as 100 then the sampling rate is 1%.

   - --pinpoint-sampling-counterrate
   - PINPOINT_GO_SAMPLING_COUNTERRATE
   - WithSamplingCounterRate()
   - int
   - default: 1
   - valid range: 0 ~ 100
   - dynamic

`Sampling.PercentRate`

:   Sampling.PercentRate configures the sampling rate of the percentage sampler.

   - --pinpoint-sampling-percentrate
   - PINPOINT_GO_SAMPLING_PERCENTRATE
   - WithSamplingPercentRate()
   - float
   - default: 100
   - valid range: 0.01 ~ 100
   - dynamic

`Span.QueueSize`

:   Span.QueueSize configures the size of the Span Queue.

   - --pinpoint-span-queuesize
   - PINPOINT_GO_SPAN_QUEUESIZE
   - WithSpanQueueSize()
   - type: int
   - default: 1024

`Span.MaxCallStackDepth`

:   Span.MaxCallStackDepth configures the maximum depth of the call stack that Span can detect.

   - --pinpoint-span-maxcallstackdepth
   - PINPOINT_GO_SPAN_MAXCALLSTACKDEPTH
   - WithSpanMaxCallStackDepth()
   - type: int
   - default: 64
   - min: 2
   - ultimate: -1
   - dynamic

`Span.MaxCallStackSequence`

:   Span.MaxCallStackDepth configures the longest sequence of call stations that Span can detect.

   - --pinpoint-span-maxcallstacksequence
   - PINPOINT_GO_SPAN_MAXCALLSTACKSEQUENCE
   - WithSpanMaxCallStackSequence()
   - type: int
   - default: 5000
   - min: 4
   - ultimate: -1
   - dynamic

`Stat.CollectInterval`

:   Stat.CollectInterval configures statistics frequency.

   - --pinpoint-stat-collectinterval
   - PINPOINT_GO_STAT_COLLECTINTERVAL
   - WithStatCollectInterval()
   - type: int
   - default: 5000
   - unit: milliseconds

`Stat.BatchCount`

:   Stat.BatchCount configures the number of batches to send statistical data.

   - --pinpoint-stat-batchcount
   - PINPOINT_GO_STAT_BATCHCOUNT
   - WithStatBatchCount()
   - type: int
   - default: 6

`Log.Level`

:   Log.Level configures the level of Agent logs. trace/debug/info/warn/error must be configured.

   - --pinpoint-log-level
   - PINPOINT_GO_LOG_LEVEL
   - WithLogLevel()
   - type: string
   - default: "info"
   - case-insensitive
   - dynamic

`Log.Output`

:   Log.Output configures the output of the log, stderr/stdout/file path.

   - --pinpoint-log-output
   - PINPOINT_GO_LOG_OUTPUT
   - WithLogOutput()
   - type: string
   - default: "stderr"
   - case-insensitive
   - dynamic

`Log.MaxSize`

:   Log.MaxSize configures the maximum size of the log file.

   - --pinpoint-log-maxsize
   - PINPOINT_GO_LOG_MAXSIZE
   - WithLogMaxSize()
   - type: int
   - default: 10
   - dynamic

<!-- markdownlint-enable -->

## Manual instrumentation of applications {#manual-instrumentation}

For programming languages with virtual machines, such as JAVA, automatic detection can be started by directly injecting the detection agent into the virtual machine. However, for programming languages that need to be compiled and run independently, such as Golang, manual detection is required.

Manual detection can be done in two ways in Pinpoint Golang Agent:

- Use [Pinpoint Golang plug-in library](https://github.com/pinpoint-apm/pinpoint-go-agent/tree/main/plugin){:target="_blank"}
- Manual detection using Pinpoint Agent Golang API

<!-- markdownlint-disable MD006 MD007  MD038 -->
`Span`

:   Span in Pinpoint represents the top-level program operation of service or application, such as creating Span in HTTP handler:

   ```golang
   func doHandle(w http.ResponseWriter, r *http.Request) {
     tracer = pinpoint.GetAgent().NewSpanTracerWithReader("HTTP Server", r.URL.Path, r.Header)
     defer tracer.EndSpan()

     span := tracer.Span()
     span.SetEndPoint(r.Host)
   }
   ```

   You can instrument a single call stack application and generate a span. Tracer.EndSpan() must be called to complete the Span and send it to the remote Collector. The SpanRecorder and Annotation interfaces can be used to record link data in Span.

`SpanEvent`

:   Each SpanEvent in Pinpoint represents a program operation within the scope of a Span detection, such as accessing a database, calling a function, making a request to another service, etc. You can report a span through Tracer.NewSpanEvent(), and Tracer.EndSpanEvent() must be called to complete a span.

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

`Distribution Tracing Context`

:   If a request comes from another node monitored by Pinpoint, a data exchange context will be included in the data exchange. Most data of this type comes from the previous node and is packaged in the request message body. Pinpoint provides two functions to read and write the data exchange context.

   - Tracer.Extract(reader DistributedTracingContextReader) // Extract the distribution context.
   - Tracer.Inject(writer DistributedTracingContextWriter) // Inject context into the request.

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

`Transparent transmission of call context between functions`

:   The transparent transmission of calling context between different APIs in the same service and in different processes is achieved through the operation of context.Context. Pinpoint Golang Agent implements context linking by injecting Tracer into Context.

   - NewContext() // Inject Tracer into Context.
   - FromContext() // Import a Tracer.

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

`Detect Goroutine`

:   Pinpoint Tracer was originally designed to detect applications with a single call stack, so sharing the same Tracer between different threads will cause resource preemption and cause the program to crash. You can create a new Tracer to detect Goroutines by calling Tracer.NewGoroutineTracer().

   Tracers can be passed between threads in the following ways:

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