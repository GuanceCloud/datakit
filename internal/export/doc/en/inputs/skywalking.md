
# SkyWalking
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

The SkyWalking Agent embedded in Datakit is used to receive, compute and analyze Skywalking Tracing protocol data.

## SkyWalking Doc {#doc}

> APM v8.8. 3 is currently incompatible and cannot be used. V8.5. 0 v8.6. 0 v8.7. 0 is currently supported.

- [Quickstart](https://skywalking.apache.org/docs/skywalking-showcase/latest/readme/){:target="_blank"}
- [Docs](https://skywalking.apache.org/docs/){:target="_blank"}
- [Clients Download](https://skywalking.apache.org/downloads/){:target="_blank"}
- [Souce Code](https://github.com/apache/skywalking){:target="_blank"}

## Configure SkyWalking Client {#client-config}

Open file /path_to_skywalking_agent/config/agent.config to configure.

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## Configure SkyWalking Agent {#agent-config}

=== "Install On Local Host"

    Go to the `conf.d/skywalking` directory under the DataKit installation directory, copy `skywalking.conf.sample` and name it `skywalking.conf`. Examples are as follows:

    ```toml
      {{ CodeBlock .InputSample 4 }}
    ```

    Datakit supports two kinds of Transport Protocol, HTTP & GRPC.

    /v3/profiling API for now used as compatible facility and do not send profiling data to data center.

    HTTP Protocol Config
    ```toml
      ## Skywalking HTTP endpoints for tracing, metric, logging and profiling.
      ## NOTE: DO NOT EDIT.
      endpoints = ["/v3/trace", "/v3/metric", "/v3/logging", "/v3/logs", "/v3/profiling"]
    ```

    GRPC Protocol Config
    ```toml
      ## Skywalking GRPC server listening on address.
      address = "localhost:11800"
    ```

    For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.skywalking.tags]`:

    ```toml
      [inputs.skywalking.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
      # ...
    ```

=== "Install In Kubernetes Cluster"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

    Multiple environment variables supported that can be used in Kubernetes showing below:

    | Envrionment Variable Name                 | Type        | Example                                                                              |
    | ----------------------------------------- | ----------- | ------------------------------------------------------------------------------------ |
    | `ENV_INPUT_SKYWALKING_HTTP_ENDPOINTS`     | JSON string | `["/v3/trace", "/v3/metric", "/v3/logging", "/v3/profiling"]`                        |
    | `ENV_INPUT_SKYWALKING_GRPC_ENDPOINT`      | string      | "127.0.0.1:11800"                                                                    |
    | `ENV_INPUT_SKYWALKING_PLUGINS`            | JSON string | `["db.type", "os.call"]`                                                             |
    | `ENV_INPUT_SKYWALKING_IGNORE_TAGS`        | JSON string | `["block1", "block2"]`                                                               |
    | `ENV_INPUT_SKYWALKING_DEL_MESSAGE`        | bool        | true                                                                                 |
    | `ENV_INPUT_SKYWALKING_KEEP_RARE_RESOURCE` | bool        | true                                                                                 |
    | `ENV_INPUT_SKYWALKING_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":    ["resource3"]}` |
    | `ENV_INPUT_SKYWALKING_SAMPLER`            | float       | 0.3                                                                                  |
    | `ENV_INPUT_SKYWALKING_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                                  |
    | `ENV_INPUT_SKYWALKING_THREADS`            | JSON string | `{"buffer":1000, "threads":100}`                                                     |
    | `ENV_INPUT_SKYWALKING_STORAGE`            | JSON string | `{"storage":"./skywalking_storage", "capacity": 5120}`                               |

## Restart Java Client {#start-java}

```command
java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```

## Send Log to Datakit {#logging}
- log4j2

The toolkit dependency package is added to the maven or gradle.
```xml
  <dependency>
    <groupId>org.apache.skywalking</groupId>
    <artifactId>apm-toolkit-log4j-2.x</artifactId>
    <version>{project.release.version}</version>
  </dependency>
```

Sent through grpc protocol:
```xml
  <GRPCLogClientAppender name="grpc-log">
    <PatternLayout pattern="%d{HH:mm:ss.SSS} %-5level %logger{36} - %msg%n"/>
  </GRPCLogClientAppender>
```

Others:

- [log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}

## SkyWalking JVM Measurement {#jvm-measurements}

jvm metrics collected by skywalking language agent.

- Tag

| Tag Name  | Description  |
| --------- | ------------ |
| `service` | service name |

- Metrics List

| Metrics                            | Description                                                                                                                               | Data Type |  Unit   |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | :-------: | :-----: |
| `class_loaded_count`               | loaded class count.                                                                                                                       |    int    |  count  |
| `class_total_loaded_count`         | total loaded class count.                                                                                                                 |    int    |  count  |
| `class_total_unloaded_class_count` | total unloaded class count.                                                                                                               |    int    |  count  |
| `cpu_usage_percent`                | cpu usage percentile                                                                                                                      |   float   | percent |
| `gc_phrase_old/new_count`          | gc old or new count.                                                                                                                      |    int    |  count  |
| `heap/stack_committed`             | heap or stack committed amount of memory.                                                                                                 |    int    |  count  |
| `heap/stack_init`                  | heap or stack initialized amount of memory.                                                                                               |    int    |  count  |
| `heap/stack_max`                   | heap or stack max amount of memory.                                                                                                       |    int    |  count  |
| `heap/stack_used`                  | heap or stack used amount of memory.                                                                                                      |    int    |  count  |
| `pool_*_committed`                 | committed amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).   |    int    |  count  |
| `pool_*_init`                      | initialized amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage). |    int    |  count  |
| `pool_*_max`                       | max amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).         |    int    |  count  |
| `pool_*_used`                      | used amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).        |    int    |  count  |
| `thread_blocked_state_count`       | blocked state thread count                                                                                                                |    int    |  count  |
| `thread_daemon_count`              | thread daemon count.                                                                                                                      |    int    |  count  |
| `thread_live_count`                | thread live count.                                                                                                                        |    int    |  count  |
| `thread_peak_count`                | thread peak count.                                                                                                                        |    int    |  count  |
| `thread_runnable_state_count`      | runnable state thread count.                                                                                                              |    int    |  count  |
| `thread_time_waiting_state_count`  | time waiting state thread count.                                                                                                          |    int    |  count  |
| `thread_waiting_state_count`       | waiting state thread count.                                                                                                               |    int    |  count  |

## Measurements {#measurements}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}
