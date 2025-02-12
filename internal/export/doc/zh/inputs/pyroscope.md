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

Datakit 从[:octicons-tag-24: Version-1.67.0](../datakit/changelog.md#cl-1.67.0) 版本开始增加了 `Pyroscope` 采集器，支持接入 Grafana Pyroscope Agent 上报的数据，帮助用户定位应用程序中的 CPU、内存、IO 等的性能瓶颈。

## 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/pyroscope` 目录，复制 `pyroscope.conf.sample` 并命名为  `pyroscope.conf`。配置文件说明如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) ，开启 Pyroscope 采集器。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 客户端应用配置 {#app-config}

Pyroscope 采集器目前支持 [Java](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"}，[Python](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/python/){:target="_blank"} 和 [Go](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/go_push/){:target="_blank"} 三种语言的 Pyroscope Agent 接入，其他语言正在陆续接入中：

<!-- markdownlint-disable MD046 -->
=== "Java"

    从 [Github](https://github.com/grafana/pyroscope-java/releases){:target="_blank"} 下载最新的 `pyroscope.jar` 包，作为 Java Agent 启动你的应用：
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
    更多细节请参考 [Grafana 官方文档](https://grafana.com/docs/pyroscope/latest/configure-client/language-sdks/java/){:target="_blank"} 

=== "Python"

    安装 `pyroscope-io` 依赖包：
    ```shell
    pip install pyroscope-io
    ```
    
    代码引入 `pyroscope-io` 包：
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

    启动应用：
    ```shell
    PYROSCOPE_APPLICATION_NAME="python-pyro-demo" python app.py
    ```

=== "Go"

    添加 `pyroscope-go` 模块：
    ```shell
    go get github.com/grafana/pyroscope-go
    ```

    引入模块并初启动 `pyroscope`：
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

<!-- markdownlint-enable -->

## 自定义 Tag {#custom-tags}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```
