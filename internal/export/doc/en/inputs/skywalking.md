---
title     : 'SkyWalking'
summary   : 'SkyWalking Tracing Data Ingestion'
tags:
  - 'APM'
  - 'TRACING'
  - 'SKYWALKING'
__int_icon      : 'icon/skywalking'
dashboard :
  - desc  : 'Skywalking JVM Monitoring View'
    path  : 'dashboard/en/skywalking'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

The SkyWalking Agent embedded in Datakit is used to receive, compute and analyze SkyWalking Tracing protocol data.

## SkyWalking Doc {#doc}

> APM v8.8. 3 is currently incompatible and cannot be used. V8.5. 0 v8.6. 0 v8.7. 0 is currently supported.

- [Quick Start](https://skywalking.apache.org/docs/skywalking-showcase/latest/readme/){:target="_blank"}
- [Docs](https://skywalking.apache.org/docs/){:target="_blank"}
- [Clients Download](https://skywalking.apache.org/downloads/){:target="_blank"}
- [Source Code](https://github.com/apache/skywalking){:target="_blank"}

## Configure SkyWalking Client {#client-config}

Open file /path_to_skywalking_agent/config/agent.config to configure.

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## Configure SkyWalking Agent {#input-config}

<!-- markdownlint-disable MD046 -->
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

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

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

- [Log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [Logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}

## SkyWalking JVM Measurement {#jvm-measurements}

jvm metrics collected by SkyWalking language agent.

- Tags

| Tag Name  | Description  |
| --------- | ------------ |
| `service` | service name |

- Metrics

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

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}
