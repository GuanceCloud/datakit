---
title     : 'Pinpoint'
summary   : 'Receive Pinpoint Tracing data'
__int_icon      : 'icon/pinpoint'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Pinpoint
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

[:octicons-tag-24: Version-1.6.0](../datakit/changelog.md#cl-1.6.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

---

The built-in Pinpoint Agent in Datakit is used to receive, calculate, and analyze Pinpoint Tracing protocol data.

## Configuration {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host installation"

    Enter the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Datakit Pinpoint Agent listening address configuration items are:

    ```toml
    # Pinpoint GRPC service endpoint for
    # - Span Server
    # - Agent Server(unimplemented, for service intactness and compatibility)
    # - Metadata Server(unimplemented, for service intactness and compatibility)
    # - Profiler Server(unimplemented, for service intactness and compatibility)
    address = "127.0.0.1:9991"
    ```

    After configuration, [Restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

???+ warning "The Pinpoint Agent in Datakit has the following limitations"

    - Currently only supports gRPC protocol
    - Multiple services (Agent/Metadata/Stat/Span) combined into one service use the same port
    - There are differences between Pinpoint links and Datakit links, see [below](pinpoint.md#opentracing-vs-pinpoint) for details

<!-- markdownlint-enable -->

### Pinpoint Agent configuration {#agent-config}

- Download the required Pinpoint APM Agent

Pinpoint supports the multi-language APM Collector. This document uses JAVA Agent for configuration. [Download](https://github.com/pinpoint-apm/pinpoint/releases){:target="_blank"} JAVA APM Collector.

- Configure Pinpoint APM Collector, open */path_to_pinpoint_agent/pinpoint-root.config* and configure the corresponding multi-service ports

    - Configure `profiler.transport.module = GRPC`
    - Configure `profiler.transport.grpc.agent.collector.port = 9991`   (i.e. the port configured in Datakit Pinpoint Agent)
    - Configure `profiler.transport.grpc.metadata.collector.port = 9991`(i.e. the port configured in Datakit Pinpoint Agent)
    - Configure `profiler.transport.grpc.stat.collector.port = 9991`    (i.e. the port configured in Datakit Pinpoint Agent)
    - Configure `profiler.transport.grpc.span.collector.port = 9991`    (i.e. the port configured in Datakit Pinpoint Agent)

- Start Pinpoint APM Agent startup command

```shell
$ java -javaagent:/path_to_pinpoint/pinpoint-bootstrap.jar \
    -Dpinpoint.agentId=agent-id \
    -Dpinpoint.applicationName=app-name \
    -Dpinpoint.config=/path_to_pinpoint/pinpoint-root.config \
    -jar /path_to_your_app.jar
```

Datakit link data follows the OpenTracing protocol. A link in Datakit is concatenated through a simple parent-child (the child span stores the id of the parent span) structure and each span corresponds to a function call.

<figure markdown>
  ![OpenTracing](https://static.guance.com/images/datakit/datakit-opentracing.png){ width="600" }
  <figcaption>OpenTracing</figcaption>
</figure>

Pinpoint APM link data is more complex:

- The parent span is responsible for generating the ID of the child span
- The ID of the parent span must also be stored in the child span.
- Use span event instead of span in OpenTracing
- A span is a response process for a service

<figure markdown>
  ![Pinpoint](https://static.guance.com/images/datakit/datakit-pinpoint.png){ width="600" }
  <figcaption>Pinpoint</figcaption>
</figure>

### PinPointV2 {#pinpointv2}

`DataKit 1.19.0` version has been re-optimized and changed `source` to `PinPointV2`. The new version of link data reorganizes the relationship between `SpanChunk` and `Span`, the relationship between `Event` and `Span`, and the relationship between `Span` and `Span`.
And the time alignment problem between `startElapsed` and `endElapsed` in `Event`.

Main logical points:

- Cache the `serviceType` service table and write it to a file to prevent data loss when DataKit restarts.
- Cache if `parentSpanId` in `Span` is not -1. For example, if `parentSpanId:-1` is used, the `Span` will be fetched from the cache and spliced into a link based on the `nextSpanId` in `spanEvent`.
- Cache all `event` in `SpanChunk`, until the main `Span` is received, all are taken out from the cache and appended to the link.
- Accumulate `startElapsed` in the current `Event` in order as the start time of the next `Event`.
- Determine the parent-child relationship of the current `Event` according to the `Depth` field.
- Database queries will replace the current 'resource' name with `sql` statements.

## Tracing {#tracing}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tags

{{$m.TagsMarkdownTable}}

- fields

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## Metric {#metrics}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}-metric`

{{$m.Desc}}

- tags

{{$m.TagsMarkdownTable}}

- fields

{{$m.FieldsMarkdownTable}}
{{end}}

{{end}}

## Pinpoint References {#references}

- [Pinpoint official documentation](https://pinpoint-apm.gitbook.io/pinpoint/){:target="_blank"}
- [Pinpoint version documentation library](https://pinpoint-apm.github.io/pinpoint/index.html){:target="_blank"}
- [Pinpoint official repository](https://github.com/pinpoint-apm){:target="_blank"}
- [Pinpoint online example](http://125.209.240.10:10123/main){:target="_blank"}
