
# DDTrace
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

DDTrace Agent embedded in Datakit is used to receive, calculate and analyze DataDog Tracing protocol data.

## DDTrace Documentation and Examples {#doc-example}

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

## Collector Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

    Multiple environment variables supported that can be used in Kubernetes showing below:

    | Envrionment Variable Name              | Type        | Example                                                                          |
    | -------------------------------------- | ----------- | -------------------------------------------------------------------------------- |
    | `ENV_INPUT_DDTRACE_ENDPOINTS`          | JSON string | `["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]`                               |
    | `ENV_INPUT_DDTRACE_IGNORE_TAGS`        | JSON string | `["block1", "block2"]`                                                           |
    | `ENV_INPUT_DDTRACE_KEEP_RARE_RESOURCE` | bool        | true                                                                             |
    | `ENV_INPUT_DDTRACE_OMIT_ERR_STATUS`    | JSON string | `["404", "403", "400"]`                                                          |
    | `ENV_INPUT_DDTRACE_CLOSE_RESOURCE`     | JSON string | `{"service1":["resource1"], "service2":["resource2"], "service3":["resource3"]}` |
    | `ENV_INPUT_DDTRACE_SAMPLER`            | float       | 0.3                                                                              |
    | `ENV_INPUT_DDTRACE_TAGS`               | JSON string | `{"k1":"v1", "k2":"v2", "k3":"v3"}`                                              |
    | `ENV_INPUT_DDTRACE_THREADS`            | JSON string | `{"buffer":1000, "threads":100}`                                                 |
    | `ENV_INPUT_DDTRACE_STORAGE`            | JSON string | `{"storage":"./ddtrace_storage", "capacity": 5120}`                              |

???+ attention

    - Don't modify the `endpoints` list here.

    ```toml
    endpoints = ["/v0.3/traces", "/v0.4/traces", "/v0.5/traces"]
    ```

    - If you want to turn off sampling (that is, collect all data), the sampling rate field needs to be set as follows:

    ``` toml
    # [inputs.ddtrace.sampler]
    # sampling_rate = 1.0
    ```

    Don't just comment on the line `sampling_rate = 1.0` , it must be commented out along with `[inputs.ddtrace.sampler]` , or the collector will assume that `sampling_rate` is set to 0.0, causing all data to be discarded.

### HTTP Settings {#http}

If Trace data is sent across machines, you need to set [HTTP settings for DataKit](datakit-conf.md#config-http-server).

If you have ddtrace data sent to the DataKit, you can see it on [DataKit's monitor](datakit-monitor.md):

<figure markdown>
  ![](https://static.guance.com/images/datakit/input-ddtrace-monitor.png){ width="800" }
  <figcaption> DDtrace sends data to the /v0.4/traces interface</figcaption>
</figure>

### Turn on Disk Cache {#disk-cache}

If the amount of Trace data is large, in order to avoid causing a lot of resource overhead to the host, you can temporarily cache the Trace data to disk and delay processing:

``` toml
[inputs.ddtrace.storage]
  path = "/path/to/ddtrace-disk-storage"
  capacity = 5120
```

## DDtrace SDK Configuration {#sdk}

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

- Configure custom tags directly in ddtrace. conf. This approach affects **all** data sends to the DataKit tracing service and should be considered carefully:

```toml
# tags is ddtrace configed key value pairs
[inputs.ddtrace.tags]
  some_tag = "some_value"
  more_tag = "some_other_value"
```

### Add a Business Tag to your Code {#add-tags}

In the application code, you can set the business custom tag in a way such as `span.SetTag(some-tag-key, some-tag-value)` (different languages have different ways). For these business custom tags, you can identify and extract them by configuring `customer_tags` in ddtrace.conf:

```toml
customer_tags = [
  "order_id",
  "task_id",
  "some.key",  # renamed some_key
]
```

Note that these tag-keys cannot contain the English character '.', and the tag-key with  `.` will be replaced with  `_`.

???+ attention "Considerations for adding business tags to application code"

    - After the corresponding tags are added in the application code, the corresponding tag-key list must also be added simultaneously in `customer_tags` of ddtrace.conf, otherwise DataKit will not extract these business tags
    - Some span with tag added may be discarded when sampling is turned on

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

## More Readings {#more-reading}

- [DataKit Tracing Field definition](datakit-tracing-struct.md)
- [DataKit general Tracing data collection instructions](datakit-tracing.md)
- [Proper use of regular expressions to configure](datakit-input-conf.md#debug-regex)
