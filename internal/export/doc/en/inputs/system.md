---
title     : 'System'
summary   : 'Collecting metrics data related to the host system'
__int_icon      : 'icon/system'
dashboard :
  - desc  : 'System'
    path  : 'dashboard/en/system'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# System
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

The system collector collects system load, uptime, the number of CPU cores, and the number of users logged in.

## Configuration {#config}

## Preconditions {#requrements}

None

<!-- markdownlint-disable MD046 -->
## Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

---
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration through `[inputs.{{.InputName}}.tags]`:

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### Why no `cpu_total_usage`? {#no-cpu}

Some CPU acquisition features are not supported on some platforms, such as macOS.
