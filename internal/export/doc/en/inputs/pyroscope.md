---
title: 'Pyroscope'
summary: 'Grafana Pyroscope Application Performance Profiler'
__int_icon: 'icon/profiling'
tags:
  - 'PYROSCOPE'
  - 'PROFILE'
dashboard:
  - desc: 'Not available yet'
    path: '-'
monitor:
  - desc: 'Not available yet'
    path: '-'
---

{{.AvailableArchs}}

---

Starting from the [:octicons-tag-24: Version-1.67.0](../datakit/changelog-2025.md#cl-1.67.0) release, DataKit has added a collector named Pyroscope. It supports the ingestion of data reported by the Grafana Pyroscope Agent, assisting users in identifying performance bottlenecks in aspects such as CPU, memory, and IO within applications.

## Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the *conf.d/{{.Catalog}}* directory under the DataKit installation directory, copy *{{.InputName}}.conf.sample* and name it *{{.InputName}}.conf*. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) to enable the Pyroscope profiler.

=== "Kubernetes"

    Currently, the profiler can be enabled by injecting the profiler configuration through the [ConfigMap method](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

### Custom Tag {#custom-tags}

You can specify other tags in the configuration through `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  #...
```

## Agent Configuration {#app-config}

The Pyroscope profiler currently supports the access of Pyroscope Agents in three languages: [Java](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"}, [Python](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/python/){:target="_blank"}, [Go](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go_push/){:target="_blank"} and
[Rust](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/rust/){:target="_blank"}. Other languages are being added:

<!-- markdownlint-disable MD046 -->
=== "Java"

    Download the latest *pyroscope.jar* package from [Github](https://github.com/grafana/pyroscope-java/releases){:target="_blank"} and start your application as a Java Agent:
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
    For more details, please refer to the [official Grafana documentation](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"}

=== "Python"

    Install the `pyroscope-io` dependency package:
    ```shell
    pip install pyroscope-io
    ```

    Import the `pyroscope-io` package in the code:
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

    Start the application:
    ```shell
    PYROSCOPE_APPLICATION_NAME="python-pyro-demo" python app.py
    ```

=== "Go"

    Add the `pyroscope-go` module:
    ```shell
    go get github.com/grafana/pyroscope-go
    ```

    Import the module and initialize it:
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
    if err!= nil {
        log.Fatal("unable to bootstrap pyroscope profiler: ", err)
    }
    defer profiler.Stop()
    ```

=== "Rust"

    1. Add the dependency crates `pyroscope` and `pyroscope_pprofrs` to your project:
    ```shell
    cargo add pyroscope
    cargo add pyroscope_pprofrs
    ```

    2. Initialize and start the Pyroscope Rust profiling task in your code:
    ```rust
    use pyroscope::{PyroscopeAgent, Result};
    use pyroscope_pprofrs::{pprof_backend, PprofConfig};

    fn main() -> Result<()> {
        let pprof_config = PprofConfig::new().sample_rate(100).report_thread_id().report_thread_name(); // Basic configuration such as sampling rate
        let backend_impl = pprof_backend(pprof_config);

        // Pyroscope agent configuration
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

    ???+ info

        The Pyroscope Rust SDK currently only works properly on the Linux platform.

## Association with OpenTelemetry Tracing Data {#link-to-tracing}

By associating with tracing data (which Grafana calls *Span profiles*), users can more easily identify system performance bottlenecks. Pyroscope provides relevant OpenTelemetry plugins to associate data between the two. Currently, it supports languages such as [Java](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/java-span-profiles/){:target="_blank"}, [Python](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/python-span-profiles/){:target="_blank"}, and [Go](https://grafana.com/docs/pyroscope/latest/configure-client/trace-span-profiles/go-span-profiles/){:target="_blank"}. The following sections introduce them respectively.

???+ note

    To facilitate the differentiation of different instances of the same service, we can randomly generate a UUID when the process is started, and then set this UUID as the `runtime_id` tag on all the tracing and profiling data throughout the entire lifecycle of the process. In this way, the observability cloud can correlate the two sets of data. In addition to the `runtime_id` tag, it is also recommended that all applications add tags such as `host`(hostname), `service`(service name), `version`(service version), `env`(deployment environment), and `process_id`(process ID of the service startup process) to make it easier for the observability cloud to correlate various collected metrics and data.  

=== "Java"

    1. Download the Java agent package [*opentelemetry-javaagent.jar*](https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases){:target="_blank"} provided by the OpenTelemetry official.
    2. Download the Pyroscope OpenTelemetry plugin [*pyroscope-otel.jar*](https://github.com/grafana/otel-profiling-java/releases){:target="_blank"}.
    3. Make the corresponding settings and start your Java application:
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
    PYROSCOPE_LABELS="runtime_id=$uuid,host=$(hostname),service=java-pyro-demo,version=1.2.3,env=dev,other_tag=other_value" \
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

        The above uses the `uuidgen` command to randomly generate a UUID, and sets the `runtime_id` tag for tracing and profiling through the environment variables `OTEL_RESOURCE_ATTRIBUTES` and `PYROSCOPE_LABELS` respectively. Other environment variable settings are for reference only. Please modify them according to the actual situation or refer to the official documentation.

=== "Python"

    1. Install the `pyroscope-otel` dependency library:
    ```shell
    pip install pyroscope-otel
    ```

    2. Import the `pyroscope-otel` and `opentelemetry` libraries and make the corresponding configurations:
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


    UUID = uuid.uuid1() # Randomly generate a UUID when the process starts
    otelExporter = OTLPSpanExporter(endpoint='http://127.0.0.1:4317', insecure=True, timeout=30)

    tracerProvider = TracerProvider(resource=Resource(attributes={
        "service.name": "python-pyro-demo",
        "service.version": "v3.5.7",
        "service.env": "dev",
        "host": socket.gethostname(),
        "process_id": os.getpid(),
        "runtime_id": str(UUID), # Set the runtime_id tag for tracing
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
            "runtime_id": str(UUID), # Set the runtime_id tag for profiling
            "host": socket.gethostname(),
            "service": 'python-pyro-demo',
            "version": 'v0.2.3',
            "env": "testing",
            "process_id": os.getpid(),
        }
    )


    if __name__ == '__main__':
        # your app code

    pyroscope.shutdown()
    ```

=== "Go"

    1. Add the `pyroscope-go` library:
    ```shell
    go get github.com/grafana/pyroscope-go
    ```

    2. Configure and start OpenTelemetry and Pyroscope:
    ```go
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
        UUID       = uuid.NewString() // Generate a global UUID
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
                attribute.String("runtime_id", UUID), // Set the runtime_id tag for tracing
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
                "runtime_id": UUID, // Set the runtime_id tag for profiling
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
