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

DDTrace is an open-source APM (Application Performance Monitoring) product by DataDog. The DDTrace Agent embedded in DataKit is used to receive, process, and analyze data in the DataDog Tracing protocol.

## DDTrace Documentation and Examples {#doc-example}

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

-   :material-language-cpp: **C++**

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

    We have made some [functional extensions](ddtrace-ext-changelog.md) to DDTrace to support more mainstream frameworks and more granular data tracing.

## Configuration {#config}

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

> The `customer_tags` parameter supports regular expressions but requires a fixed prefix format `reg:`. For example, `reg:key_*` matches all keys starting with `key_`.

### Notes on Multi-Tool Tracing Propagation {#trace_propagator}

The TraceID in the DDTrace data structure is of uint64 type. When using the `tracecontext` propagation protocol, a `_dd.p.tid:67c573cf00000000` field is added inside the DDTrace trace details. This is because the `trace_id` in the `tracecontext` protocol is a 128-bit hexadecimal-encoded string, and this high-bit tag is added for compatibility purposes.

Currently, DDTrace supports the following propagation protocols: `datadog/b3multi/tracecontext`. Note the following two scenarios:
- When using `tracecontext`, since the trace ID is 128-bit, you need to enable the `compatible_otel=true` and `trace_128_bit_id` switches in the configuration.
- When using `b3multi`, pay attention to the length of the `trace_id`. If it is a 64-bit hexadecimal encoding, you need to enable `trace_id_64_bit_hex=true` in the configuration file.
- For more propagation protocols and tool usage, refer to: [Multi-Tracing Propagation](tracing-propagator.md){:target="_blank"}

???+ info

    - `compatible_otel`: Converts `span_id` and `parent_id` to hexadecimal strings.
    - `trace_128_bit_id`: Combines `_dd.p.tid` in `meta` with `trace_id` into a 32-character hexadecimal-encoded string.
    - `trace_id_64_bit_hex`: Converts 64-bit `trace_id` to a hexadecimal-encoded string.

### Inject Pod and Node Information {#add-pod-node-info}

When the application is deployed in a container environment such as Kubernetes, you can append Pod/Node information to the final Span data by modifying the application's YAML file. Below is an example YAML for a Kubernetes Deployment:

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
            - name: DD_TAGS
              value: pod_name:$(POD_NAME),host:$(NODE_NAME)
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['service']
```

Note that you need to first define `POD_NAME` and `NODE_NAME`, then embed them into the DDTrace-specific environment variables.

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

    - To disable sampling (i.e., collect all data), set the sampling rate field as follows:

    ``` toml
    # [inputs.{{.InputName}}.sampler]
    # sampling_rate = 1.0
    ```

    Do not only comment out the line `sampling_rate = 1.0`; you must also comment out `[inputs.{{.InputName}}.sampler]`. Otherwise, the collector will treat `sampling_rate` as 0.0, resulting in all data being discarded.

<!-- markdownlint-enable -->

### HTTP Settings {#http}

If Trace data is sent from a remote machine, you need to configure the [HTTP settings of DataKit](../datakit/datakit-conf.md#config-http-server).

If DDTrace data is sent to DataKit, you can view it on the [DataKit monitor](../datakit/datakit-monitor.md):

<figure markdown>
  ![input-ddtrace-monitor](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDTrace sends data to the /v0.4/traces endpoint</figcaption>
</figure>

### Enable Disk Cache {#disk-cache}

If the volume of Trace data is large, to avoid excessive resource consumption on the host, you can temporarily cache Trace data to disk for delayed processing:

``` toml
[inputs.{{.InputName}}.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

### DDTrace SDK Configuration {#sdk}

After configuring the collector, you can also make additional configurations on the DDTrace SDK side.

### Environment Variable Settings {#sdk-envs}

- `DD_TRACE_ENABLED`: Enable global tracer (supported by some language platforms)
- `DD_AGENT_HOST`: DDTrace agent host address
- `DD_TRACE_AGENT_PORT`: DDTrace agent host port
- `DD_SERVICE`: Service name
- `DD_TRACE_SAMPLE_RATE`: Set sampling rate
- `DD_VERSION`: Application version (optional)
- `DD_TRACE_STARTUP_LOGS`: DDTrace logger
- `DD_TRACE_DEBUG`: DDTrace debug mode
- `DD_ENV`: Application environment value
- `DD_TAGS`: Application tags

In addition to setting the project name, environment name, and version number during application initialization, you can also set them in the following two ways:

- Inject environment variables via the command line

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- Configure custom tags directly in *ddtrace.conf*. This method affects all data sent to the DataKit tracing service, so use it with caution:

```toml
# tags are key-value pairs configured for ddtrace
[inputs.{{.InputName}}.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### APMTelemetry {#apm_telemetry}

[:octicons-tag-24: Version-1.35.0](../datakit/changelog.md#cl-1.35.0) · [:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

After the DDTrace agent starts, it continuously reports service-related information through an additional interface, such as startup configuration, heartbeat, and the list of loaded agents. You can view this information in <<<custom_key.brand_name>>> Infrastructure -> Resource Directory. The displayed data is helpful for troubleshooting issues related to startup commands and versions of referenced third-party libraries. It also includes host information, service information, and the number of Spans generated.

Data may vary significantly across different languages and versions; please refer to the actual received data.

### Fixed Tag Extraction {#add-tags}

Starting from DataKit version [1.21.0](../datakit/changelog.md#cl-1.21.0), the blacklist function is deprecated, and not all fields in Span.Meta are extracted into top-level tags anymore—only selected fields are extracted.

The following is a list of tags that may be extracted:

| Original Meta Field     | Extracted Field Name   | Description                                  |
|:--------------------|:--------------------|:--------------------------------------------|
| `http.url`          | `http_url`          | Full HTTP request path                       |
| `http.hostname`     | `http_hostname`     | Hostname                                    |
| `http.route`        | `http_route`        | Route                                       |
| `http.status_code`  | `http_status_code`  | Status code                                 |
| `http.method`       | `http_method`       | Request method                              |
| `http.client_ip`    | `http_client_ip`    | Client IP                                   |
| `sampling.priority` | `sampling_priority` | Sampling status                             |
| `span.kind`         | `span_kind`         | Span type                                   |
| `error`             | `error`             | Whether an error occurred                   |
| `dd.version`        | `dd_version`        | Agent version                               |
| `error.message`     | `error_message`     | Error message                               |
| `error.stack`       | `error_stack`       | Stack trace information                     |
| `error.type`        | `error_type`        | Error type                                  |
| `system.pid`        | `pid`               | Process ID (pid)                            |
| `error.msg`         | `error_message`     | Error message                               |
| `project`           | `project`           | Project name                                |
| `version`           | `version`           | Version                                     |
| `env`               | `env`               | Environment                                 |
| `host`              | `host`              | Hostname in tags                            |
| `pod_name`          | `pod_name`          | Pod name in tags                            |
| `_dd.base_service`  | `_dd_base_service`  | Parent service                              |
| `peer.hostname`     | `db_host`           | May be an IP or domain name (depends on configuration) |
| `db.type`           | `db_system`         | Database type: mysql, oracle, etc.          |
| `db.instance`       | `db_name`           | Database name                               |

In the Studio tracing interface, tags not in the list can also be used for filtering.

Starting from DataKit version [1.22.0](../datakit/changelog.md#cl-1.22.0), the whitelist function is restored. If there are tags that must be extracted into the top-level tag list, you can configure them in `customer_tags`. If the whitelisted tags are in the original `message.meta`, the collector will use `.` as a separator and convert `.` to `_` during extraction.

## Collected Data Field Description {#collected-data}

### Tracing {#tracing}

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

After DDTrace starts, it reports its own configuration information, integration list, dependencies, and service-related information to DataKit. Currently, only Java Agent is supported. The following is a description of each field:

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

## More Readings {#more-reading}

- [DataKit Tracing Field Definition](datakit-tracing-struct.md)
- [DataKit General Tracing Data Collection Description](datakit-tracing.md)
- [Proper Use of Regular Expressions for Configuration](../datakit/datakit-input-conf.md#debug-regex)
- [Multi-Tracing Propagation](tracing-propagator.md)
- [Java Integration and Exception Description](ddtrace-java.md)
- [DDTrace Sampling Strategy and Notes on Multi-Tool Tracing Propagation](tracing-sample.md)
