---
title     : 'DDTrace'
summary   : 'Receive APM Data from DDTrace'
__int_icon: 'icon/ddtrace'
tags      :
  - 'DDTRACE'
  - 'Distributed Tracing'
dashboard :
  - desc  : 'None'
    path  : '-'
monitor   :
  - desc  : 'None'
    path  : '-'
---


{{.AvailableArchs}}

---

The DataKit `ddtrace` collector is a **DataDog Trace protocol receiver**. A DDTrace SDK or Java Agent in your application sends trace payloads to DataKit over HTTP; DataKit then parses, processes, and uploads them. The collector neither installs an SDK in the application nor instruments application code for you.

The data path is:

`application code` → `DDTrace SDK / Java Agent` → `DataKit HTTP receiver` → `<<<custom_key.brand_name>>>`

Traces, profiling data, and runtime/JMX metrics use different receivers:

- This collector receives traces through the DataKit HTTP port, normally `9529`.
- Profiling requires the separate [Profiling collector](profile.md).
- JMX and runtime metrics sent through DogStatsD require the separate [StatsD collector](statsd.md), normally on port `8125`.

## Before You Begin {#overview}

1. Install DataKit, enable this collector, and instrument the application with the SDK or Agent for its language.
1. Ensure that the application can reach DataKit. DataKit listens on `localhost:9529` by default. If the application is on another host or Pod, adjust the [HTTP listener](../datakit/datakit-conf.md#config-http-server) and restrict access with firewalls, a Kubernetes Service, or NetworkPolicy.
1. Explicitly configure the trace destination in the application, for example `DD_AGENT_HOST=<datakit-host>` and `DD_TRACE_AGENT_PORT=9529`. The common upstream Datadog Agent default is `8126`; do not rely on a default and do not confuse the trace port with the StatsD port `8125`.
1. Set stable `DD_SERVICE`, `DD_ENV`, and `DD_VERSION` values. They determine the service, environment, and version shown in the tracing UI.

## Language-Specific Setup {#doc-example}

<!-- markdownlint-disable MD046 MD032 MD030 -->
<div class="grid cards" markdown>
-   :fontawesome-brands-python: **Python**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-py){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/python?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-python.md)

-   :material-language-java: **Java**

    ---

    [SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-java.md)

-   :material-language-ruby: **Ruby**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-rb){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-ruby.md)

-   :fontawesome-brands-golang: **Golang**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-go){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/go?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-golang.md)

-   :material-language-php: **PHP**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-php){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/php?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-php.md)

-   :fontawesome-brands-node-js: **NodeJS**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-js){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/nodejs?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-nodejs.md)

-   :material-language-cpp: **C++ (Legacy Compatibility)**

    ---

    [SDK :material-download:](https://github.com/opentracing/opentracing-cpp){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/setup_overview/setup/cpp?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: Example](ddtrace-cpp.md)

-   :material-dot-net: **.NET**

    ---

    [SDK :material-download:](https://github.com/DataDog/dd-trace-dotnet){:target="_blank"} ·
    [:octicons-book-16: Documentation](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/dotnet-framework?tab=windows){:target="_blank"} ·
    [:octicons-book-16: .NET Core Documentation](https://docs.datadoghq.com/tracing/trace_collection/automatic_instrumentation/dd_libraries/dotnet-core?tab=windows){:target="_blank"}
</div>

???+ info

    <<<custom_key.brand_name>>> provides an extended Java Agent for selected frameworks and more granular collection. These extensions apply only when you use the <<<custom_key.brand_name>>> JAR. Review the [Java extension guide](ddtrace-ext-java.md) and its [changelog](ddtrace-ext-changelog.md), especially the configuration and data-exposure warnings.

## Configuration {#config}

This section configures the **DataKit receiver**. Configure the SDK's destination, service identity, and sampling separately in the application's startup options or environment variables; see the language-specific guide.

=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) to take effect.

=== "Kubernetes"

    You can enable the collector by [injecting collector configuration via ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [configuring ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting).

    You can also modify configuration parameters via environment variables (you need to add the collector to ENV_DEFAULT_ENABLED_INPUTS as a default collector):

{{ CodeBlock .InputENVSample 4 }}

`customer_tags` promotes selected span `meta` and `metrics` entries to the top level. String `meta` values become tags, while numeric `metrics` values become fields; unmatched entries remain in `message`. A `.` in a literal field name becomes `_`; for example, `http.route` becomes `http_route`. A regular expression must begin with `reg:` and uses Go regular-expression syntax; for example, `reg:^key_.*$` matches fields beginning with `key_`. Validate expressions in a test environment first: an invalid expression prevents the collector from initializing correctly.

`sampling_priority_drop_excludes` lists rejected upstream sampling priorities that bypass DDTrace's early drop decision. Supported values are `-3`, `-1`, and `0`. The default empty list preserves the existing behavior of dropping all three values. For example, with `[0]`, a priority=0 trace continues through DataKit's normal filters and sampler; it is not force-kept. This setting cannot restore spans that the upstream did not send and may significantly increase the volume sent downstream.

The DataKit self-monitoring metrics `datakit_input_ddtrace_sampling_priority_trace_total` and `datakit_input_ddtrace_sampling_priority_span_total` record these decisions with `priority`, `action` (`drop` or `bypass`), and `service` labels. They prove how much data DataKit received and then dropped or bypassed, but cannot prove that a priority=0 value originated from W3C `traceparent`.

### Notes on Multi-Tool Trace Propagation {#trace_propagator}

A traditional DDTrace Trace ID is a 64-bit integer, whereas W3C `tracecontext` uses a 128-bit, 32-character hexadecimal Trace ID. DDTrace places the upper 64 bits in `_dd.p.tid`; DataKit uses that field to reconstruct the full 128-bit ID.

Use a consistent propagation protocol across the call path and configure DataKit for the protocol actually emitted:

- When linking `tracecontext` traffic with OpenTelemetry, enable `compatible_otel=true` so span and parent IDs are formatted in hexadecimal. `trace_128_bit_id` is enabled by default and combines `_dd.p.tid` with the lower 64-bit Trace ID.
- When using `b3multi` and the upstream sends a 64-bit hexadecimal Trace ID, enable `trace_id_64_bit_hex=true`.
- After changing propagation, validate one cross-service request. A protocol mismatch normally produces disconnected service topology rather than an ingestion failure.

For more combinations, see [Multi-Tracing Propagation](tracing-propagator.md){:target="_blank"}.

???+ info

    - `compatible_otel`: Formats `span_id` and `parent_id` as hexadecimal strings.
    - `trace_128_bit_id`: Combines `_dd.p.tid` in `meta` and the lower 64-bit `trace_id` into a 32-character hexadecimal string. Default: `true`.
    - `trace_id_64_bit_hex`: Parses an upstream 64-bit `trace_id` as hexadecimal.

### Inject Pod and Node Information {#add-pod-node-info}

When an application runs in Kubernetes, use the Downward API to put Pod, Namespace, and Node values in `DD_TAGS`. They travel with application spans. Add the fields to DataKit `customer_tags` as well if you need them as top-level tags in the trace list.

```yaml hl_lines="21-30"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  selector:
    matchLabels:
      app: my-app
  replicas: 3
  template:
    metadata:
      labels:
        app: my-app
        service: my-service
    spec:
      containers:
        - name: my-app
          image: my-app:v0.0.1
          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: DD_TAGS
              value: pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['service']
```

Kubernetes environment substitution can reference only environment variables defined **earlier in the same container**. Define `POD_NAME`, `POD_NAMESPACE`, and `NODE_NAME` before `DD_TAGS`.

After the application starts, enter the corresponding Pod and verify if the ENV is in effect:

```shell
$ env | grep DD_
...
```

Once injected successfully, you can see the Pod and Node names where the Span is located in the final Span data.

---

???+ warning

    - Do not modify the `endpoints` list here (unless you clearly understand the configuration logic and effects).

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - `[inputs.ddtrace.sampler]` is **receiver-side sampling**, independent of SDK-side `DD_TRACE_SAMPLE_RATE`. To stop sampling at the receiver, remove or comment out the whole table. If the table remains, always set its rate explicitly:

    ``` toml
    # [inputs.{{.InputName}}.sampler]
    # sampling_rate = 1.0
    ```

    `sampling_rate = 1.0` keeps all traces. Do not comment out only `sampling_rate` while keeping the table header: the table is then interpreted as a zero rate and drops every trace. Error traces bypass receiver-side sampling by default; they can be filtered only when matching statuses are configured in `omit_err_status`.

<!-- markdownlint-enable MD046 MD032 MD030 -->

### HTTP Settings {#http}

DataKit listens on `localhost:9529` by default. If traces arrive from a remote host or another Pod, set a reachable listener in `datakit.conf`, for example:

```toml
[http_api]
  listen = "0.0.0.0:9529"
```

Use this only on a controlled network. Do not expose trace endpoints directly to the Internet; apply firewall rules, security groups, Kubernetes Services/NetworkPolicies, or TLS as appropriate. DataKit receives `/v0.3/traces`, `/v0.4/traces`, and `/v0.5/traces` by default. Do not change `endpoints` unless you understand the compatibility impact.

If DDTrace data is sent to DataKit, you can view it on the [DataKit monitor](../datakit/datakit-monitor.md):

<figure markdown>
  ![input-ddtrace-monitor](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDTrace sends data to the /v0.4/traces endpoint</figcaption>
</figure>

### Enable Disk Cache {#disk-cache}

Disk cache defers processing of HTTP request bodies during traffic spikes, reducing memory and processing peaks. It is not long-term archival and does not replace SDK-side sampling. The directory must be writable, have enough capacity, and be on persistent storage when recovery after a container restart matters. `capacity` is in MiB.

``` toml
[inputs.{{.InputName}}.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

### DDTrace SDK Configuration {#sdk}

After configuring the collector, configure the SDK. Supported variables and precedence vary by language and SDK version. The variables below describe common concepts. When an SDK supports `DD_TRACE_AGENT_URL`, that URL normally overrides separate host and port values, so do not set conflicting destinations.

### Environment Variable Settings {#sdk-envs}

| Variable | Purpose | Recommendation |
| --- | --- | --- |
| `DD_AGENT_HOST`, `DD_TRACE_AGENT_PORT` | Trace receiver address and port | Point them at DataKit, such as `datakit-service:9529`; do not use the StatsD port `8125`. |
| `DD_SERVICE`, `DD_ENV`, `DD_VERSION` | Service, environment, and version identity | Use stable, searchable values for every service. |
| `DD_TAGS` | Application-wide tags | Use `key:value` pairs. Never put tokens, request bodies, or personal data here. |
| `DD_TRACE_SAMPLE_RATE` | SDK-side sampling rate | Use a value from `0.0` to `1.0`; control high traffic at the source first. |
| `DD_TRACE_ENABLED` | Enables instrumentation/trace delivery | Exact behavior is language-dependent; during troubleshooting verify it is not `false`. |
| `DD_TRACE_STARTUP_LOGS`, `DD_TRACE_DEBUG` | SDK startup and debug logs | Enable only temporarily for troubleshooting to avoid extra volume or configuration exposure. |

In addition to setting the project name, environment name, and version number during application initialization, you can also set them in the following two ways:

- Inject environment variables via the command line

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- Configure receiver-side tags directly in *ddtrace.conf*. They affect every DDTrace payload entering this DataKit and are suitable for shared deployment tags, not application-specific service identity:

```toml
# tags are key-value pairs configured for ddtrace
[inputs.{{.InputName}}.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### APMTelemetry {#apm_telemetry}

[:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

The Java Agent can report startup configuration, heartbeat, dependencies, and loaded integrations through `/telemetry/proxy/api/v2/apmtelemetry`. DataKit accepts this route by default; disable `apmtelemetry_route_enable` if it is not needed. View the data in the <<<custom_key.brand_name>>> Infrastructure Resource Directory to troubleshoot startup options, dependency versions, and loaded instrumentation.

Data may vary significantly across different languages and versions; please refer to the actual received data.

### Fixed Tag Extraction {#add-tags}

Starting with DataKit [1.21.0](../datakit/changelog.md#cl-1.21.0), DataKit no longer promotes every `Span.Meta` field to a top-level tag. It extracts only the commonly used fields below to control tag cardinality and indexing cost.

The following is a list of tags that may be extracted:

| Original Meta Field | Extracted Field Name | Description                                            |
|:--------------------|:---------------------|:-------------------------------------------------------|
| `http.url`          | `http_url`           | Full HTTP request path                                 |
| `http.hostname`     | `http_hostname`      | Hostname                                               |
| `http.route`        | `http_route`         | Route                                                  |
| `http.status_code`  | `http_status_code`   | Status code                                            |
| `http.method`       | `http_method`        | Request method                                         |
| `http.client_ip`    | `http_client_ip`     | Client IP                                              |
| `sampling.priority` | `sampling_priority`  | Sampling status                                        |
| `span.kind`         | `span_kind`          | Span type                                              |
| `error`             | `error`              | Whether an error occurred                              |
| `runtime.name`      | `runtime_name`       | Runtime name                                           |
| `dd.version`        | `dd_version`         | Agent version                                          |
| `error.message`     | `error_message`      | Error message                                          |
| `error.stack`       | `error_stack`        | Stack trace information                                |
| `error.type`        | `error_type`         | Error type                                             |
| `system.pid`        | `pid`                | Process ID (pid)                                       |
| `error.msg`         | `error_message`      | Error message                                          |
| `project`           | `project`            | Project name                                           |
| `version`           | `version`            | Version                                                |
| `env`               | `env`                | Environment                                            |
| `host`              | `host`               | Hostname in tags                                       |
| `pod_name`          | `pod_name`           | Pod name in tags                                       |
| `pod_namespace`     | `pod_namespace`      | pod namespace                                          |
| `_dd.base_service`  | `_dd_base_service`   | Parent service                                         |
| `peer.hostname`     | `db_host`            | May be an IP or domain name (depends on configuration) |
| `db.type`           | `db_system`          | Database type: mysql, oracle, etc.                     |
| `db.instance`       | `db_name`            | Database name                                          |
| `out.host`          | `out_host`           | Database or Middleware host                            |
| `dd_ext_version`    | `sdk_version`        | SDK extend version                                     |
| `language`          | `sdk_language`       | SDK language                                           |

Fields outside the list remain in the span `meta` and are available in trace details. Whether they can be used as top-level filters depends on the UI and indexing configuration.

Starting with DataKit [1.22.0](../datakit/changelog.md#cl-1.22.0), add fields to the `customer_tags` allowlist when they must become top-level tags. A `.` becomes `_` during extraction. Add only stable, low-cardinality fields; never promote values such as user IDs or request IDs.

### Common Troubleshooting Paths {#troubleshooting}

| Symptom | Check first |
| --- | --- |
| No traces at all | Confirm that `ddtrace` is enabled in DataKit, that the application actually loaded its SDK/Agent, and that `DD_AGENT_HOST`, `DD_TRACE_AGENT_PORT`, Service, NetworkPolicy, and firewall settings are correct. |
| Only some services appear or topology is broken | Ensure propagation protocols match on both sides and that the appropriate `tracecontext`/B3 DataKit ID compatibility switches are enabled. |
| JVM metrics or profiling are missing | They are not received by this collector. Check the `statsd` or `profile` collector and its separate port. |
| Data volume or resource use is too high | Reduce SDK-side sampling first, then evaluate receiver-side sampling, `trace_max_spans`, `max_trace_body_mb`, and disk cache. |

## Collected Data Field Description {#collected-data}

### Tracing {#tracing}

<!-- markdownlint-disable MD024 -->
{{range $i, $m := .Measurements}}

{{if eq $m.Type "tracing"}}

#### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

### Metrics {#metric}

{{range $i, $m := .Measurements}}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

### Custom Objects {#custom-object}

The extended Java Agent can report its configuration, integration list, dependencies, and service metadata after startup. These resource objects currently apply only to the Java Agent. Common events include:

- `app_client_configuration_change`: Contains the agent's configuration information
- `app_dependencies_loaded`: Dependency list (including package names and version information)
- `app_integrations_change`: Integration list (including package names and whether the agent is enabled)
- Other host information, service information, etc.

{{range $i, $m := .Measurements}}

{{if eq $m.Type "custom_object"}}

#### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{end}}

<!-- markdownlint-enable MD024 -->

## More Readings {#more-reading}

- [DataKit Tracing Field Definition](datakit-tracing-struct.md)
- [DataKit General Tracing Data Collection Description](datakit-tracing.md)
- [Proper Use of Regular Expressions for Configuration](../datakit/datakit-input-conf.md#debug-regex)
- [Multi-Tracing Propagation](tracing-propagator.md)
- [Java Integration and Exception Description](ddtrace-java.md)
- [DDTrace Sampling Strategy and Notes on Multi-Tool Tracing Propagation](tracing-sample.md)
