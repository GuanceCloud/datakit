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

Starting from the [:octicons-tag-24: Version-1.67.0](../datakit/changelog-2025.md#cl-1.67.0) release, Datakit has added a collector named `Pyroscope`. It supports the ingestion of data reported by the Grafana Pyroscope Agent, assisting users in identifying performance bottlenecks in aspects such as CPU, memory, and IO within applications.

## Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Enter the `conf.d/pyroscope` directory under the DataKit installation directory, copy `pyroscope.conf.sample` and name it `pyroscope.conf`. The configuration file description is as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) to enable the Pyroscope profiler.

=== "Kubernetes"

    Currently, the profiler can be enabled by injecting the profiler configuration through the [ConfigMap method](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Agent Configuration {#app-config}

The Pyroscope profiler currently supports the access of Pyroscope Agents in three languages: [Java](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"}, [Python](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/python/){:target="_blank"}, and [Go](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go_push/){:target="_blank"}. Other languages are being added:

<!-- markdownlint-disable MD046 -->
=== "Java"

    Download the latest `pyroscope.jar` package from [Github](https://github.com/grafana/pyroscope-java/releases){:target="_blank"} and start your application as a Java Agent:
    ```shell
    PYROSCOPE_APPLICATION_NAME="java-pyro-demo" \
    PYROSCOPE_LOG_LEVEL=debug \
    PYROSCOPE_FORMAT="jfr" \
    PYROSCOPE_PROFILER_EVENT="cpu" \
    PYROSCOPE_LABELS="service=java-pyro-demo,version=1.2.3,env=dev,some_other_tag=other_value" \
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

<!-- markdownlint-enable -->

## Custom Tag {#custom-tags}

For all the following data collection, a global tag named `host` (the tag value is the host name where DataKit is located) will be added by default. You can also specify other tags in the configuration through `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  #...
```
