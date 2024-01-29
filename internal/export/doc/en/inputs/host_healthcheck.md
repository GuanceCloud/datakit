---
title     : 'Health Check'
summary   : 'Regularly check the host process and network health status'
__int_icon      : 'icon/healthcheck'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Health check
<!-- markdownlint-enable -->

[:octicons-tag-24: Version-1.24.0](../datakit/changelog.md#cl-1.24.0)

---

{{.AvailableArchs}}

---

The health check collector can regularly monitor the health of processes and networks (such as TCP and HTTP) of the main computer. If it doesn't meet the health requirements, DataKit will collect corresponding information and report the metric data.

## Configuration {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    It supports modifying configuration parameters as environment variables (effective only when the DataKit is running in K8s DaemonSet mode, which is not supported for host-deployed DataKits):

    | Environment Variable Name                              | Corresponding Configuration Parameter Item | Parameter Example                                                     |
    | :---                                 | ---              | ---                                                          |
    | `ENV_INPUT_HEALTHCHECK_INTERVAL`     | `interval`       | `5m`                                               |
    | `ENV_INPUT_HEALTHCHECK_PROCESS`      | `process`        | `[{"names":["nginx","mysql"],"min_run_time":"10m"}]`|
    | `ENV_INPUT_HEALTHCHECK_TCP`          | `tcp`            | `[{"host_ports":["10.100.1.2:3369","192.168.1.2:6379"],"connection_timeout":"3s"}]`|
    | `ENV_INPUT_HEALTHCHECK_HTTP`         | `http`           | `[{"http_urls":["http://local-ip:port/path/to/api?arg1=x&arg2=y"],"method":"GET","expect_status":200,"timeout":"30s","ignore_insecure_tls":false,"headers":{"Header1":"header-value-1","Hedaer2":"header-value-2"}}]`                                               |
    | `ENV_INPUT_HEALTHCHECK_TAGS`         | `tags`           | `{"some_tag":"some_value","more_tag":"some_other_value"}`|

<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}