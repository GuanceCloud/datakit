---
title     : 'DDTrace'
summary   : 'Receive APM data from DDTrace'
__int_icon: 'icon/ddtrace'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# DDTrace
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

DDTrace Agent embedded in Datakit is used to receive, calculate and analyze DataDog Tracing protocol data.

## DDTrace Documentation and Examples {#doc-example}

<!-- markdownlint-disable MD046 MD032 MD030 -->
<div class="grid cards" markdown>
-   :fontawesome-brands-python: __Python__

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-py){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/python?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-python.md)

-   :material-language-java: __Java__

    ---

    [:octicons-code-16: SDK](https://static.guance.com/dd-image/dd-java-agent.jar){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/java?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-java.md)

-   :material-language-ruby: __Ruby__

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-rb){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/ruby){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-java.md)

-   :fontawesome-brands-golang: __Golang__

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-go){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/go?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-golang.md)

-   :material-language-php: __PHP__

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-php){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/php?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-php.md)

-   :fontawesome-brands-node-js: __NodeJS__

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-js){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/nodejs?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-nodejs.md)

-   :material-language-cpp:

    ---

    [:octicons-code-16: SDK](https://github.com/opentracing/opentracing-cpp){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/cpp?tab=containers){:target="_blank"} ·
    [:octicons-arrow-right-24: example](ddtrace-cpp.md)

-   :material-dot-net:

    ---

    [:octicons-code-16: SDK](https://github.com/DataDog/dd-trace-dotnet){:target="_blank"} ·
    [:octicons-book-16: doc](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-framework?tab=windows){:target="_blank"} ·
    [:octicons-book-16: .Net Core doc](https://docs.datadoghq.com/tracing/setup_overview/setup/dotnet-framework?tab=windows){:target="_blank"}
</div>

???+ tip

    The DataKit installation directory, under the `data` directory, has a pre-prepared `dd-java-agent.jar`(recommended). You can also download it directly from [Maven download](https://mvnrepository.com/artifact/com.datadoghq/dd-java-agent){:target="_blank"}

    Guance Cloud also Fork its own branch on the basis of Ddtrace-Java, adding more functions and probes. For more version details, please see [Ddtrace Secondary Development Version Description](../developers/ddtrace-guance.md)

## Configuration {#config}

### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

### Notes on Linking Multiple Line Tools {#trace_propagator}
DDTrace currently supports the following propagation protocols: `datadog/b3multi/tracecontext`. There are two things to note:

- When using `tracecontext`, the `compatible_otel=true` switch needs to be turned on in the configuration because the link ID is 128 bits.
- When using `b3multi`, pay attention to the length of `trace_id`. If it is 64-bit hex encoding, the `trace_id_64_bit_hex=true` needs to be turned on in the configuration file.
- For more propagation protocol and tool usage, please refer to: [Multi-Link Concatenation](tracing-propagator.md){:target="_blank"}

### Add Pod and Node tags {#add-pod-node-info}

When your service deployed on Kubernetes, we can add Pod/Node tags to Span, edit your Pod yaml, here is a Deployment yaml example:

```yaml
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
            - name: POD_NAME    # <------
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: DD_SERVICE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['service']
            - name: DD_TAGS
              value: pod_name:$(POD_NAME),host:$(NODE_NAME)
```

Here we must define `POD_NAME` and `NODE_NAME` before reference them in dedicated environment keys of DDTrace:

After your Pod started, enter the Pod, we can check if environment applied:

```shell
$ env | grep DD_
...
```

Once environment set, the Pod/Node name will attached to related Span tags.

---

???+ attention

    - Don't modify the `endpoints` list here.

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - If you want to turn off sampling (that is, collect all data), the sampling rate field needs to be set as follows:

    ``` toml
    # [inputs.{{.InputName}}.sampler]
    # sampling_rate = 1.0
    ```

    Don't just comment on the line `sampling_rate = 1.0` , it must be commented out along with `[inputs.{{.InputName}}.sampler]` , or the collector will assume that `sampling_rate` is set to 0.0, causing all data to be discarded.

<!-- markdownlint-enable -->

### HTTP Settings {#http}

If Trace data is sent across machines, you need to set [HTTP settings for DataKit](datakit-conf.md#config-http-server).

If you have ddtrace data sent to the DataKit, you can see it on [DataKit's monitor](datakit-monitor.md):

<figure markdown>
  ![input-ddtrace-monitor](https://static.guance.com/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDtrace sends data to the /v0.4/traces interface</figcaption>
</figure>

### Turn on Disk Cache {#disk-cache}

If the amount of Trace data is large, in order to avoid causing a lot of resource overhead to the host, you can temporarily cache the Trace data to disk and delay processing:

``` toml
[inputs.{{.InputName}}.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

### DDtrace SDK Configuration {#sdk}

After configuring the collector, you can also do some configuration on the DDtrace SDK side.

### Environment Variables Setting {#sdk-envs}

- `DD_TRACE_ENABLED`: Enable global tracer (Partial language platform support)
- `DD_AGENT_HOST`: DDtrace agent host address
- `DD_TRACE_AGENT_PORT`: DDtrace agent host port
- `DD_SERVICE`: Service name
- `DD_TRACE_SAMPLE_RATE`: Set sampling rate
- `DD_VERSION`: Application version (optional)
- `DD_TRACE_STARTUP_LOGS`: DDtrace logger
- `DD_TRACE_DEBUG`: DDtrace debug mode
- `DD_ENV`: Application env values
- `DD_TAGS`: Application

In addition to setting the project name, environment name, and version number when initialization is applied, you can also set them in the following two ways:

- Inject environment variables from the command line

```shell
DD_TAGS="project:your_project_name,env=test,version=v1" ddtrace-run python app.py
```

- Configure custom tags directly in ddtrace. conf. This approach affects __all__ data sends to the DataKit tracing service and should be considered carefully:

```toml
# tags is ddtrace configed key value pairs
[inputs.{{.InputName}}.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### Add a Business Tag to your Code {#add-tags}

Starting from DataKit version [1.21.0](../datakit/changelog.md#cl-1.21.0), do not include All in Span.Mate are advanced to the first level label and only select following list labels:

| Mete              | GuanCe tag        | doc                   |
|:------------------|:------------------|:----------------------|
| http.url          | http_url          | HTTP url              |
| http.hostname     | http_hostname     | hostname              |
| http.route        | http_route        | route                 |
| http.status_code  | http_status_code  | status code           |
| http.method       | http_method       | method                |
| http.client_ip    | http_client_ip    | client IP             |
| sampling.priority | sampling_priority | sample                |
| span.kind         | span_kind         | span kind             |
| error             | error             | is error              |
| dd.version        | dd_version        | agent version         |
| error.message     | error_message     | error message         |
| error.stack       | error_stack       | error stack           |
| error.type        | error_type        | error type            |
| system.pid        | pid               | pid                   |
| error.msg         | error_message     | error message         |
| project           | project           | project               |
| version           | version           | version               |
| env               | env               | env                   |
| host              | host              | host from dd.tags     |
| pod_name          | pod_name          | pod_name from dd.tags |
| _dd.base_service  | _dd_base_service  | base service          |

In the link interface of the observation cloud, tags that are not in the list can also be filtered.

Restore whitelist functionality from DataKit version [1.22.0](../datakit/changelog.md#cl-1.22.0). If there are labels that must be extracted from the first level label list, they can be found in the `customer_tags`.

If the configured whitelist label is in the native `message.meta`, Will convert to replace `.` with `_`.

## Tracing {#tracing}

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

## More Readings {#more-reading}

- [DataKit Tracing Field definition](datakit-tracing-struct.md)
- [DataKit general Tracing data collection instructions](datakit-tracing.md)
- [Proper use of regular expressions to configure](datakit-input-conf.md#debug-regex)
