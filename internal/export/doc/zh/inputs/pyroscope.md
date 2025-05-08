---
title     : 'Pyroscope'
summary   : 'Grafana Pyroscope 应用程序性能采集器'
__int_icon: 'icon/profiling'
tags:
  - 'PYROSCOPE'
  - 'PROFILE'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


{{.AvailableArchs}}

---

Datakit 从[:octicons-tag-24: Version-1.67.0](../datakit/changelog-2025.md#cl-1.67.0) 版本开始增加了 Pyroscope 采集器，支持接入 Grafana Pyroscope Agent 上报的数据，帮助用户定位应用程序中的 CPU、内存、IO 等的性能瓶颈。

## 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 *{{.Catalog}}.conf.sample* 并命名为 *{{.Catalog}}.conf*。配置文件说明如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) ，开启 Pyroscope 采集器。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 设置采集器全局 Tag {#custom-tags}

可以在采集器配置中通过 `[inputs.{{.InputName}}.tags]` 指定额外标签，该标签会统一应用到所有该采集器采到的数据：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

## 客户端 SDK 接入 {#app-config}

Pyroscope 采集器目前支持 [Java](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"}，[Python](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/python/){:target="_blank"}， [Go](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go_push/){:target="_blank"} 和
[Rust](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/rust/){:target="_blank"} 等语言的 Pyroscope Agent 接入，其他语言正在持续支持中：

<!-- markdownlint-disable MD046 -->
=== "Java"

    从 [Github](https://github.com/grafana/pyroscope-java/releases){:target="_blank"} 下载最新的 *pyroscope.jar* 包，作为 Java Agent 启动你的应用：

    ```shell
    PYROSCOPE_APPLICATION_NAME="java-pyro-demo" \
    PYROSCOPE_LOG_LEVEL=debug \
    PYROSCOPE_FORMAT="jfr" \
    PYROSCOPE_PROFILER_EVENT="cpu" \
    PYROSCOPE_LABELS="host=$(hostname),service=java-pyro-demo,version=1.2.3,env=dev,some_other_tag=other_value" \
    PYROSCOPE_UPLOAD_INTERVAL="60s" \
    PYROSCOPE_JAVA_STACK_DEPTH_MAX=512 \
    PYROSCOPE_PROFILING_INTERVAL="10ms" \
    PYROSCOPE_PROFILER_ALLOC=128k \
    PYROSCOPE_PROFILER_LOCK=10ms \
    PYROSCOPE_ALLOC_LIVE=false \
    PYROSCOPE_GC_BEFORE_DUMP=true \
    PYROSCOPE_SERVER_ADDRESS="http://127.0.0.1:9529" \
    java -javaagent:pyroscope.jar -jar your-app.jar
    ```

    更多细节请参考 [Grafana 官方文档](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"} 

=== "Python"

    1. 安装 `pyroscope-io` 依赖包：
    ```shell
    pip install pyroscope-io
    ```
    
    1. 代码引入 `pyroscope-io` 包：
    ```python
    import os
    import pyroscope
    import socket

    pyroscope.configure(
        server_address="http://127.0.0.1:9529",
        detect_subprocesses=True,
        oncpu=True,
        enable_logging=True,
        report_pid=True,
        report_thread_id=True,
        report_thread_name=True,
        tags={
            "host": socket.gethostname(),
            "service": 'python-pyro-demo',
            "version": 'v1.2.3',
            "env": "testing",
            "process_id": os.getpid(),
        }
    )
    ```

    1. 启动应用：
    ```shell
    PYROSCOPE_APPLICATION_NAME="python-pyro-demo" python app.py
    ```

=== "Go"

    1. 添加 `pyroscope-go` 模块：
    ```shell
    go get github.com/grafana/pyroscope-go
    ```

    1. 引入模块并初启动：
    ```go
    import (
        "log"
        "os"
        "runtime"
        "strconv"
        "time"

        "github.com/grafana/pyroscope-go"
    )
    
    func Must[T any](t T, _ error) T {
        return t
    }

    runtime.SetMutexProfileFraction(5)
    runtime.SetBlockProfileRate(5)

    profiler, err := pyroscope.Start(pyroscope.Config{
        ApplicationName: "go-pyroscope-demo",

        // replace this with the address of pyroscope server
        ServerAddress: "http://127.0.0.1:9529",

        // you can disable logging by setting this to nil
        Logger: pyroscope.StandardLogger,

        // uploading interval period
        UploadRate: time.Minute,

        // you can provide static tags via a map:
        Tags: map[string]string{
            "service":    "go-pyroscope-demo",
            "env":        "demo",
            "version":    "1.2.3",
            "host":       Must(os.Hostname()),
            "process_id": strconv.Itoa(os.Getpid()),
            "runtime_id": UUID,
        },

        ProfileTypes: []pyroscope.ProfileType{
            // these profile types are enabled by default:
            pyroscope.ProfileCPU,
            pyroscope.ProfileAllocObjects,
            pyroscope.ProfileAllocSpace,
            pyroscope.ProfileInuseObjects,
            pyroscope.ProfileInuseSpace,

            // these profile types are optional:
            pyroscope.ProfileGoroutines,
            pyroscope.ProfileMutexCount,
            pyroscope.ProfileMutexDuration,
            pyroscope.ProfileBlockCount,
            pyroscope.ProfileBlockDuration,
        },
    })
    if err != nil {
        log.Fatal("unable to bootstrap pyroscope profiler: ", err)
    }

    defer profiler.Stop()
    ```

=== "Rust"

    ???+ attention
    
        Pyroscope Rust agent 目前只能正常工作在 Linux 平台上。

    1. 添加 `pyroscope` 和 `pyroscope_pprofrs` crates 到项目依赖中：
    ```shell
    cargo add pyroscope
    cargo add pyroscope_pprofrs
    ```

    1. 代码中初始化并启动 Pyroscope Rust profiling 任务：
    ```rust
    use pyroscope::{PyroscopeAgent, Result};
    use pyroscope_pprofrs::{pprof_backend, PprofConfig};

    fn main() -> Result<()> {
        let pprof_config = PprofConfig::new().sample_rate(100).report_thread_id().report_thread_name(); // 采样率等基础配置
        let backend_impl = pprof_backend(pprof_config);

        // Pyroscope agent 配置
        let agent = PyroscopeAgent::builder("http://127.0.0.1:9529", "pyroscope-rust-app")
            .backend(backend_impl)
            .tags([("version", "1.23.4"), ("env", "demo"), ("host", "<your-hostname>")].to_vec())
            .build()?;

        // start the Pyroscope agent
        let agent_running = agent.start()?;

        // your application code...

        // gracefully shutdown the Pyroscope agent
        let agent_ready = agent_running.stop()?;
        agent_ready.shutdown();

        Ok(())
    }
    ```

## 与 OpenTelemetry 链路数据进行关联 {#link-to-tracing}

通过与链路数据之间的关联（Grafana 称之为 *Span profiles*），用户可以更容易的洞察到系统的性能瓶颈，Pyroscope 提供了相关的 OpenTelemetry 插件，可以让两者之间的数据关联起来，目前支持 [Java](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/java-span-profiles/){:target="_blank"}，[Python](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/python-span-profiles/){:target="_blank"} 和
[Go](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/go-span-profiles/){:target="_blank"} 等语言，下面分别介绍。


???+ note

    为了便于区分同一服务的不同实例，我们可以在进程启动时随机生成一个 UUID，然后在进程的整个生命周期内把该 UUID 作为 `runtime_id` 标签设置到所有的链路和 profiling 数据上，这样便能关联两者数据。 除了 `runtime_id` 标签，还建议所有应用添加
    `host`（主机名），`service`（服务名），`version`（服务版本），`env`（部署环境）， `process_id`（服务启动进程号）等标签，方便关联各类采集到的指标和数据。

=== "Java"

    1. 下载 OpenTelemetry 官方提供的 Java agent 包 [*opentelemetry-javaagent.jar*](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases){:target="_blank"}。
    1. 下载 Pyroscope OpenTelemetry 插件 [*pyroscope-otel.jar*](https://github.com/grafana/otel-profiling-java/releases){:target="_blank"}。
    1. 进行相应设置并启动你的 Java 应用
    ```shell
    uuid=$(uuidgen);
    OTEL_SERVICE_NAME="java-pyro-demo" \
    OTEL_RESOURCE_ATTRIBUTES="runtime_id=$uuid,host=$(hostname),service.name=java-pyro-demo,service.version=1.3.55,service.env=dev" \
    OTEL_JAVAAGENT_EXTENSIONS=./pyroscope-otel.jar \
    OTEL_PYROSCOPE_ADD_PROFILE_URL=false \
    OTEL_PYROSCOPE_ADD_PROFILE_BASELINE_URL=false \
    OTEL_PYROSCOPE_START_PROFILING=true \
    OTEL_TRACES_EXPORTER=otlp \
    OTEL_EXPORTER_OTLP_PROTOCOL="http/protobuf" \
    OTEL_EXPORTER_OTLP_TRACES_ENDPOINT="http://127.0.0.1:9529/otel/v1/traces" \
    OTEL_EXPORTER_OTLP_METRICS_ENDPOINT="http://127.0.0.1:9529/otel/v1/metrics" \
    OTEL_EXPORTER_OTLP_LOGS_ENDPOINT="http://127.0.0.1:9529/otel/v1/logs" \
    OTEL_EXPORTER_OTLP_COMPRESSION=gzip \
    PYROSCOPE_APPLICATION_NAME="java-pyro-demo" \
    PYROSCOPE_LOG_LEVEL=debug \
    PYROSCOPE_FORMAT="jfr" \
    PYROSCOPE_PROFILER_EVENT="cpu" \
    PYROSCOPE_LABELS="runtime_id=$uuid,service=java-pyro-demo,version=1.2.3,env=dev,host=$(hostname),other_tag=other_value" \
    PYROSCOPE_UPLOAD_INTERVAL="60s" \
    PYROSCOPE_JAVA_STACK_DEPTH_MAX=512 \
    PYROSCOPE_PROFILING_INTERVAL="10ms" \
    PYROSCOPE_PROFILER_ALLOC=128k \
    PYROSCOPE_PROFILER_LOCK=10ms \
    PYROSCOPE_ALLOC_LIVE=false \
    PYROSCOPE_GC_BEFORE_DUMP=true \
    PYROSCOPE_SERVER_ADDRESS="http://127.0.0.1:9529" \
    java -javaagent:opentelemetry-javaagent.jar -jar your-app.jar
    ```

    ???+ tips
    
        上述使用 `uuidgen` 命令随机生成了一个 UUID，并通过环境变量 `OTEL_RESOURCE_ATTRIBUTES` 和 `PYROSCOPE_LABELS` 分别为链路和 profiling 设置 `runtime_id` tag，其它一些环境变量的设置仅供参考，请根据实际情况或参考官方文档酌情增舍修改。

=== "Python"

    1. 安装  `pyroscope-otel` 依赖库
    ```shell
    pip install pyroscope-otel
    ```

    1. 引入 `pyroscope-otel` 和 `opentelemetry` 库并进行相应配置
    ```python
    import uuid
    import socket
    import os
    import pyroscope
    
    from opentelemetry import trace
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor
    from pyroscope.otel import PyroscopeSpanProcessor
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    
    
    UUID = uuid.uuid1() // 进程启动时随机生成一个 UUID
    otelExporter = OTLPSpanExporter(endpoint='http://127.0.0.1:4317', insecure=True, timeout=30)
    
    tracerProvider = TracerProvider(resource=Resource(attributes={
        "service.name": "python-pyro-demo",
        "service.version": "v3.5.7",
        "service.env": "dev",
        "host": socket.gethostname(),
        "process_id": os.getpid(),
        "runtime_id": str(UUID), // 为链路设置 runtime_id tag
    }))
    
    tracerProvider.add_span_processor(PyroscopeSpanProcessor())
    tracerProvider.add_span_processor(BatchSpanProcessor(span_exporter=otelExporter, max_queue_size=100, max_export_batch_size=30))
    trace.set_tracer_provider(tracerProvider)
    tracer = trace.get_tracer("python-pyro-demo")
    
    pyroscope.configure(
        server_address="http://127.0.0.1:9529",
        detect_subprocesses=True,
        oncpu=True,
        enable_logging=True,
        report_pid=True,
        report_thread_id=True,
        report_thread_name=True,
        tags={
            "runtime_id": str(UUID), // 为 profiling 设置 runtime_id tag
            "host": socket.gethostname(),
            "service": 'python-pyro-demo',
            "version": 'v0.2.3',
            "env": "testing",
            "process_id": os.getpid(),
        }
    )
    
    
    if __name__ == '__main__':
        // your app code
    
    pyroscope.shutdown()
    ```

=== "Go"

    1. 添加 `pyroscope-go` 库到项目依赖中
    ```shell
    go get github.com/grafana/pyroscope-go
    ```

    1. 配置并启动 OpenTelemetry 和 Pyroscope
    ```shell
    package main
    
    import (
        "github.com/google/uuid"
        otelpyroscope "github.com/grafana/otel-profiling-go"
        "github.com/grafana/pyroscope-go"
        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/attribute"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
        "go.opentelemetry.io/otel/propagation"
        "go.opentelemetry.io/otel/sdk/resource"
        tracesdk "go.opentelemetry.io/otel/sdk/trace"
        "go.opentelemetry.io/otel/trace"
    )
      
    var (
        UUID       = uuid.NewString() // 生成一个全局的 UUID
        otelTracer trace.Tracer
    )
    
    func hostname() string {
        host, _ := os.Hostname()
        return host
    }
    
    func main() {
        otelExporter, err := otlptracehttp.New(context.Background(),
            otlptracehttp.WithEndpointURL("http://127.0.0.1:9529/otel/v1/traces"),
            otlptracehttp.WithInsecure(),
            otlptracehttp.WithTimeout(time.Second*15),
        )
        if err != nil {
            log.Fatal("unable to init otel tracing exporter: ", err)
        }
        tracerProvider := tracesdk.NewTracerProvider(tracesdk.WithBatcher(otelExporter,
            tracesdk.WithBatchTimeout(time.Second*3)),
            tracesdk.WithResource(resource.NewSchemaless(
                attribute.String("runtime_id", UUID), // 为链路设置 runtime_id tag
                attribute.String("service.name", "go-pyroscope-demo"),
                attribute.String("service.version", "v0.0.1"),
                attribute.String("service.env", "dev"),
                attribute.String("host", hostname()),
                attribute.String("process_id", strconv.Itoa(os.Getpid()))
            )),
        )
        defer tracerProvider.Shutdown(context.Background())
    
        otel.SetTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))
        otelTracer = otel.Tracer("go-pyroscope-demo")
        log.Printf("otel tracing started....\n")
    
        runtime.SetMutexProfileFraction(5)
        runtime.SetBlockProfileRate(5)
    
        profiler, err := pyroscope.Start(pyroscope.Config{
            ApplicationName: "go-pyroscope-demo",
    
            // replace this with the address of pyroscope server
            ServerAddress: "http://127.0.0.1:9529",
    
            // you can disable logging by setting this to nil
            Logger: pyroscope.StandardLogger,
    
            // uploading interval period
            UploadRate: time.Minute,
    
            // you can provide static tags via a map:
            Tags: map[string]string{
                "runtime_id": UUID, // 为 profiling 设置 runtime_id tag
                "env":        "demo",
                "version":    "0.0.1",
                "host":       hostname(),
                "process_id": strconv.Itoa(os.Getpid()),
            },
    
            ProfileTypes: []pyroscope.ProfileType{
                // these profile types are enabled by default:
                pyroscope.ProfileCPU,
                pyroscope.ProfileAllocObjects,
                pyroscope.ProfileAllocSpace,
                pyroscope.ProfileInuseObjects,
                pyroscope.ProfileInuseSpace,
    
                // these profile types are optional:
                pyroscope.ProfileGoroutines,
                pyroscope.ProfileMutexCount,
                pyroscope.ProfileMutexDuration,
                pyroscope.ProfileBlockCount,
                pyroscope.ProfileBlockDuration,
            },
        })
        if err != nil {
            log.Fatal("unable to bootstrap pyroscope profiler: ", err)
        }

        log.Printf("pyroscope profiler started....\n")
        defer profiler.Stop()
        
        // your app code...
    }  
    
    ```
<!-- markdownlint-enable -->
